package apiservices_test

// A fresh installation: the database holds only the schema (roro_local.sql). On
// start the server signs in to the remote server as the initial user, pulls roles
// and users, and after that users can log in. Runs only when RORO_TEST_DB=1, with the
// usual MYSQL_* variables pointing at a database loaded from roro_local.sql.

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	apiservices "github.com/shabs76/roro-local-server/api_services"
	"github.com/shabs76/roro-local-server/controlers"
	"github.com/shabs76/roro-local-server/gendb"
	"golang.org/x/crypto/bcrypt"
)

const (
	initialEmail    = "initial.user@walls.test"
	initialPassword = "initial-pass-123"
	staffEmail      = "inspector@walls.test"
	staffPassword   = "inspector-pass-123"
	bossEmail       = "boss@walls.test"
	bossPassword    = "boss-pass-123"
	lateEmail       = "late.joiner@walls.test"
	latePassword    = "late-pass-123"
)

// fakeRemote answers like the remote server's login, roles, users, makers and
// manifest routes, including the quirks that matter for foreign keys:
//   - the roles list holds only roles numbered above 100;
//   - the users list leaves out the user who asks for it;
//   - the client list holds only active clients.
//
// passwords is what the remote login accepts; a test may change it.
func fakeRemote(t *testing.T, passwords map[string]string) *httptest.Server {
	t.Helper()
	hash := func(p string) string {
		h, err := bcrypt.GenerateFromPassword([]byte(p), bcrypt.MinCost)
		if err != nil {
			t.Fatal(err)
		}
		return string(h)
	}
	user := func(id, fname, email, roleId, roleName string, roleNumber int) map[string]any {
		return map[string]any{"userId": id, "fname": fname, "lname": "Test", "email": email, "phone": "1", "password": hash(passwords[email]),
			"roleId": roleId, "roleName": roleName, "roleNumber": roleNumber, "status": "active", "creationDate": "2026-01-01 00:00:00"}
	}
	users := []map[string]any{
		user("RU_INIT", "Initial", initialEmail, "RR_ADMIN", "Admin", 100),
		user("RU_STAFF", "Asha", staffEmail, "RR_INSPECTOR", "Inspector", 500),
		user("RU_BOSS", "Baraka", bossEmail, "RR_SUPER", "Super Admin", 50),
		// Added on the remote server after the local server pulled the users.
		user("RU_LATE", "Neema", lateEmail, "RR_INSPECTOR", "Inspector", 500),
	}
	inUsersList := map[string]bool{"RU_INIT": true, "RU_STAFF": true, "RU_BOSS": true}
	reply := func(w http.ResponseWriter, status int, body any) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		json.NewEncoder(w).Encode(body)
	}
	caller := func(r *http.Request) string { return strings.TrimPrefix(r.Header.Get("Log-Id"), "RL_") }

	mux := http.NewServeMux()
	mux.HandleFunc("/auth/login", func(w http.ResponseWriter, r *http.Request) {
		var req struct{ Email, Password string }
		json.NewDecoder(r.Body).Decode(&req)
		if passwords[req.Email] == "" || passwords[req.Email] != req.Password {
			reply(w, http.StatusUnauthorized, map[string]any{"state": "error", "data": "Invalid email or password"})
			return
		}
		for _, u := range users {
			if u["email"] == req.Email {
				session := map[string]any{}
				for k, v := range u {
					if k != "password" {
						session[k] = v
					}
				}
				reply(w, http.StatusOK, map[string]any{"state": "success", "login_id": "RL_" + u["userId"].(string), "login_key": "RK", "user": session})
				return
			}
		}
	})
	mux.HandleFunc("/admin/get/user/roles", func(w http.ResponseWriter, r *http.Request) {
		if caller(r) != "RU_INIT" {
			reply(w, http.StatusForbidden, map[string]any{"state": "error", "data": "not allowed"})
			return
		}
		reply(w, http.StatusOK, map[string]any{"state": "success", "data": []map[string]any{
			{"roleId": "RR_INSPECTOR", "roleName": "Inspector", "roleNumber": 500, "roleDate": "2026-01-01 00:00:00"},
		}})
	})
	mux.HandleFunc("/admin/get/company/users/pass", func(w http.ResponseWriter, r *http.Request) {
		if caller(r) != "RU_INIT" {
			reply(w, http.StatusForbidden, map[string]any{"state": "error", "data": "not allowed"})
			return
		}
		list := []map[string]any{}
		for _, u := range users {
			if inUsersList[u["userId"].(string)] && u["userId"] != caller(r) {
				list = append(list, u)
			}
		}
		reply(w, http.StatusOK, map[string]any{"state": "success", "data": list})
	})
	mux.HandleFunc("/admin/get/vehicle/makers", func(w http.ResponseWriter, r *http.Request) {
		data := []map[string]any{}
		if r.URL.Query().Get("page") == "1" {
			data = append(data, map[string]any{"makerId": "RM_TOYOTA", "makerName": "Toyota", "creationDate": "2026-01-01 00:00:00"})
		}
		reply(w, http.StatusOK, map[string]any{"state": "success", "data": data})
	})
	mux.HandleFunc("/client/manage/get/clients", func(w http.ResponseWriter, r *http.Request) {
		reply(w, http.StatusOK, map[string]any{"state": "success", "data": []map[string]any{
			{"clientId": "CL_ACTIVE", "clientName": "Active Client", "logo": "notset", "cover": "notset", "principalName": "p",
				"status": "active", "clientLocation": "Dar", "creationDate": "2026-01-01 00:00:00"},
		}})
	})
	mux.HandleFunc("/admin/get/manifest/data/M_INACTIVE_CLIENT", func(w http.ResponseWriter, r *http.Request) {
		reply(w, http.StatusOK, map[string]any{"state": "success", "data": map[string]any{
			"manifestId": "M_INACTIVE_CLIENT", "manifestName": "Old client", "clientId": "CL_INACTIVE", "clientName": "Active Client",
			"vesselName": "Vessel", "voyageNo": "V1", "berthNo": "1", "arrivalDate": "2026-10-01 00:00:00",
			"receivedDate": "2026-10-01", "uploadedDate": "2026-10-01 00:00:00"}})
	})
	// Other reference lists are not served; the sync must carry on without them.
	return httptest.NewServer(mux)
}

func count(t *testing.T, qr string, args ...any) int {
	t.Helper()
	db, err := gendb.InitDb()
	if err != nil {
		t.Fatal(err)
	}
	var n int
	if err := db.QueryRow(qr, args...).Scan(&n); err != nil {
		t.Fatalf("%s: %v", qr, err)
	}
	return n
}

func login(t *testing.T, router *gin.Engine, email, password string) int {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"email": email, "password": password})
	req := httptest.NewRequest(http.MethodPost, "/users/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec.Code
}

func TestFreshInstallPullsUsersThenUsersCanLogIn(t *testing.T) {
	if os.Getenv("RORO_TEST_DB") != "1" {
		t.Skip("set RORO_TEST_DB=1 to run database integration tests")
	}
	passwords := map[string]string{initialEmail: initialPassword, staffEmail: staffPassword, bossEmail: bossPassword, lateEmail: latePassword}
	remote := fakeRemote(t, passwords)
	defer remote.Close()
	t.Setenv("REMOTE_SERVER_URL", remote.URL)

	if n := count(t, "SELECT COUNT(*) FROM users"); n != 0 {
		t.Fatalf("expected an empty users table on a fresh install, found %d users", n)
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/users/login", controlers.LoginUser)

	// Wrong initial credentials: the pull stops at once instead of retrying.
	t.Setenv("INIT_USER_EMAIL", initialEmail)
	t.Setenv("INIT_USER_PASSWORD", "wrong-password")
	started := time.Now()
	apiservices.SyncFromRemoteAtStart(context.Background())
	if time.Since(started) > 5*time.Second {
		t.Fatal("a refused initial login was retried")
	}
	if n := count(t, "SELECT COUNT(*) FROM users"); n != 0 {
		t.Fatalf("users pulled with a refused login: %d", n)
	}

	// Correct initial credentials. The users list holds a user whose role (number 50)
	// is not in the roles list, and leaves out the initial user itself.
	t.Setenv("INIT_USER_PASSWORD", initialPassword)
	apiservices.SyncFromRemoteAtStart(context.Background())
	if n := count(t, "SELECT COUNT(*) FROM users WHERE user_id IN ('RU_INIT','RU_STAFF','RU_BOSS')"); n != 3 {
		t.Fatalf("users = %d, want 3 (initial user, inspector, super admin)", n)
	}
	if n := count(t, "SELECT COUNT(*) FROM roles WHERE role_id IN ('RR_ADMIN','RR_INSPECTOR','RR_SUPER')"); n != 3 {
		t.Fatalf("roles = %d, want 3", n)
	}
	if n := count(t, "SELECT COUNT(*) FROM vehicle_makers WHERE maker_id = 'RM_TOYOTA'"); n != 1 {
		t.Fatal("makers list not pulled")
	}

	// Pulled users log in; their remote sessions are stored for the tablets.
	for email, password := range map[string]string{initialEmail: initialPassword, staffEmail: staffPassword, bossEmail: bossPassword} {
		if code := login(t, router, email, password); code != http.StatusOK {
			t.Fatalf("%s login answered %d, want 200", email, code)
		}
	}
	if code := login(t, router, staffEmail, "not-the-password"); code != http.StatusUnauthorized {
		t.Fatalf("wrong password answered %d, want 401", code)
	}

	// A user added on the remote server after the pull logs in without a restart.
	if code := login(t, router, lateEmail, latePassword); code != http.StatusOK {
		t.Fatalf("late joiner login answered %d, want 200", code)
	}
	if n := count(t, "SELECT COUNT(*) FROM users WHERE user_id = 'RU_LATE'"); n != 1 {
		t.Fatal("late joiner was not stored")
	}

	// A password changed on the remote server works here at once; the old one stops working.
	passwords[staffEmail] = "changed-pass-456"
	if code := login(t, router, staffEmail, "changed-pass-456"); code != http.StatusOK {
		t.Fatalf("changed password answered %d, want 200", code)
	}
	if code := login(t, router, staffEmail, staffPassword); code != http.StatusUnauthorized {
		t.Fatalf("old password answered %d, want 401", code)
	}

	// Pulling again on the next start keeps everyone and their sessions.
	sessions := count(t, "SELECT COUNT(*) FROM logins WHERE status = 'active'")
	if sessions < 4 {
		t.Fatalf("stored sessions = %d, want at least 4", sessions)
	}
	apiservices.SyncFromRemoteAtStart(context.Background())
	if n := count(t, "SELECT COUNT(*) FROM users"); n != 4 {
		t.Fatalf("users after a second pull = %d, want 4", n)
	}
	if n := count(t, "SELECT COUNT(*) FROM logins WHERE status = 'active'"); n != sessions {
		t.Fatalf("sessions after a second pull = %d, want %d", n, sessions)
	}

	// A manifest of a client missing from the (active-only) client list is still saved.
	if err := apiservices.FetchClientList("RL_RU_INIT", "RK"); err != nil {
		t.Fatal(err)
	}
	if err := apiservices.FetchManifestDetails("RL_RU_INIT", "RK", "M_INACTIVE_CLIENT"); err != nil {
		t.Fatalf("manifest of an inactive client: %v", err)
	}
	if n := count(t, "SELECT COUNT(*) FROM manifest m JOIN clients c ON c.client_id = m.client_id WHERE m.manifest_id = 'M_INACTIVE_CLIENT' AND c.client_id = 'CL_INACTIVE'"); n != 1 {
		t.Fatal("manifest or its client missing")
	}
}
