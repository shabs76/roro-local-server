package apiservices

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/shabs76/roro-local-server/base_api"
	"github.com/shabs76/roro-local-server/constants"
	baseapi "github.com/shabs76/roro-local-server/constants/modules/base_api"
	"github.com/shabs76/roro-local-server/constants/modules/users"
	usersdataservices "github.com/shabs76/roro-local-server/database/users_data_services"
	"github.com/shabs76/roro-local-server/specials"
)

// ErrRemoteLoginRefused means the remote server answered and refused the login (wrong
// email or password, inactive account). Retrying with the same credentials is useless.
var ErrRemoteLoginRefused = errors.New("remote server refused the login")

// RemoteLogin signs in to the remote server and returns its session.
func RemoteLogin(req users.LoginRequest) (users.UserLoginResponse, error) {
	var session users.UserLoginResponse
	client := base_api.NewAPI(specials.GetEnvVariable("REMOTE_SERVER_URL", "http://localhost:8000"))
	resp, err := client.Post("/auth/login", req, &session, []baseapi.Header{
		{Key: "Content-Type", Value: "application/json"},
	})
	if err != nil {
		return session, fmt.Errorf("could not reach the remote server: %w", err)
	}
	if resp.Status != constants.SuccessState || session.LoginID == "" {
		if resp.StatusCode >= http.StatusInternalServerError {
			return session, fmt.Errorf("remote server error (HTTP %d): %s", resp.StatusCode, resp.Message)
		}
		return session, fmt.Errorf("%w: %s", ErrRemoteLoginRefused, resp.Message)
	}
	return session, nil
}

// SyncFromRemoteAtStart signs in to the remote server as the initial user and pulls
// the roles, the users and the reference lists, so that people can log in to a
// freshly installed local server. Users are filled only from the remote server: by
// this pull, and by each login the remote server accepts (see SaveSignedInUser).
//
// The initial user's remote credentials come from INIT_USER_EMAIL and
// INIT_USER_PASSWORD. The pull runs on every start (so new staff and changed
// passwords arrive) and retries in the background while the remote server is not
// reachable. It stops early when the remote server refuses the login.
func SyncFromRemoteAtStart(ctx context.Context) {
	email := strings.TrimSpace(os.Getenv("INIT_USER_EMAIL"))
	password := os.Getenv("INIT_USER_PASSWORD")
	if email == "" || password == "" {
		if !anyLocalUser() {
			slog.Warn("No users on this server and no initial user configured: set INIT_USER_EMAIL and INIT_USER_PASSWORD " +
				"to the remote credentials of a user allowed to read the company's users, then restart")
		}
		return
	}

	const maxAttempts = 8
	delay := 10 * time.Second
	for attempt := 1; ; attempt++ {
		err := pullUsersAndLists(email, password)
		if err == nil {
			return
		}
		if errors.Is(err, ErrRemoteLoginRefused) {
			slog.Error("Initial user sync stopped: the remote server refused the initial user's login. Check INIT_USER_EMAIL and INIT_USER_PASSWORD.",
				"email", email, "error", err)
			return
		}
		if attempt == maxAttempts {
			slog.Error("Initial user sync gave up; restart the server or run the sync from a tablet", "attempts", attempt, "error", err)
			return
		}
		slog.Warn("Initial user sync failed; retrying", "attempt", attempt, "retryIn", delay.String(), "error", err)
		select {
		case <-ctx.Done():
			return
		case <-time.After(delay):
		}
		delay = min(delay*2, 10*time.Minute)
	}
}

// SaveSignedInUser stores the user the remote server has just signed in, with a hash
// of the password it accepted. The remote users list never includes the user who asks
// for it, so this is how the initial user gets here; it also brings in staff added
// after the last pull, and password changes.
func SaveSignedInUser(session users.UserLoginResponse, password string) error {
	hash, err := specials.GeneratePasswordHash(password)
	if err != nil {
		return fmt.Errorf("could not hash the password: %w", err)
	}
	u := session.User
	st := usersdataservices.InsertUserData([]users.UserData{{
		UserID:       u.UserID,
		FName:        u.FName,
		LName:        u.LName,
		Email:        u.Email,
		Phone:        u.Phone,
		Password:     hash,
		RoleID:       u.RoleID,
		RoleName:     u.RoleName,
		RoleNumber:   u.RoleNumber,
		Status:       u.Status,
		CreationDate: u.CreationDate,
	}})
	if st.State != constants.SuccessState {
		return fmt.Errorf("could not save user %s: %s", u.Email, st.Data)
	}
	return nil
}

// pullUsersAndLists does one sign-in and pull. The users are required; the roles list
// and the reference lists are pulled on a best-effort basis (the users sync adds every
// role its users need).
func pullUsersAndLists(email, password string) error {
	session, err := RemoteLogin(users.LoginRequest{Email: email, Password: password})
	if err != nil {
		return err
	}
	if err := FetchUserRoles(session.LoginID, session.LoginKey); err != nil {
		slog.Warn("Initial user sync: could not pull the roles list; roles come with the users instead", "error", err)
	}
	if err := SaveSignedInUser(session, password); err != nil {
		return fmt.Errorf("initial user: %w", err)
	}
	if err := FetchUsersList(session.LoginID, session.LoginKey); err != nil {
		return fmt.Errorf("users: %w", err)
	}

	st, found := usersdataservices.SelectUserDetailsWithRolesPass(" `email` = ? ", []any{email})
	if st.State != constants.SuccessState || len(found) == 0 {
		return fmt.Errorf("the initial user %s is missing after the pull", email)
	}
	slog.Info("Initial user sync: roles and users pulled from the remote server", "initialUser", email)

	lists := []struct {
		name  string
		fetch func(string, string) error
	}{
		{"vehicle makers", FetchVehicleMakers},
		{"vehicle body types", FetchVehicleBodies},
		{"vehicle models", FetchVehicleModels},
		{"inspection checklist", FetchInspectionCheckList},
		{"package types and statuses", FetchPackageTypesAndStatus},
		{"clients", FetchClientList},
	}
	for _, l := range lists {
		if err := l.fetch(session.LoginID, session.LoginKey); err != nil {
			slog.Warn("Initial user sync: could not pull a reference list; a tablet sync can pull it later", "list", l.name, "error", err)
		}
	}
	return nil
}

func anyLocalUser() bool {
	st, found := usersdataservices.SelectUserDetailsWithRolesPass(" 1 = 1 LIMIT 1", nil)
	return st.State == constants.SuccessState && len(found) > 0
}
