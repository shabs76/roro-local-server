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
)

// fakeRemote answers like the remote server's login, roles, users and makers routes.
func fakeRemote(t *testing.T) *httptest.Server {
	t.Helper()
	hash := func(p string) string {
		h, err := bcrypt.GenerateFromPassword([]byte(p), bcrypt.MinCost)
		if err != nil {
			t.Fatal(err)
		}
		return string(h)
	}
	users := []map[string]any{
		{"userId": "RU_INIT", "fname": "Initial", "lname": "User", "email": initialEmail, "phone": "1", "password": hash(initialPassword),
			"roleId": "RR_ADMIN", "roleName": "Admin", "roleNumber": 100, "status": "active", "creationDate": "2026-01-01 00:00:00"},
		{"userId": "RU_STAFF", "fname": "Asha", "lname": "Inspector", "email": staffEmail, "phone": "2", "password": hash(staffPassword),
			"roleId": "RR_INSPECTOR", "roleName": "Inspector", "roleNumber": 500, "status": "active", "creationDate": "2026-01-01 00:00:00"},
	}
	passwords := map[string]string{initialEmail: initialPassword, staffEmail: staffPassword}
	reply := func(w http.ResponseWriter, status int, body any) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		json.NewEncoder(w).Encode(body)
	}
	authorized := func(r *http.Request) bool { return r.Header.Get("Log-Id") == "RL_RU_INIT" }

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
				reply(w, http.StatusOK, map[string]any{"state": "success", "login_id": "RL_" + u["userId"].(string), "login_key": "RK", "user": u})
				return
			}
		}
	})
	mux.HandleFunc("/admin/get/user/roles", func(w http.ResponseWriter, r *http.Request) {
		if !authorized(r) {
			reply(w, http.StatusForbidden, map[string]any{"state": "error", "data": "not allowed"})
			return
		}
		reply(w, http.StatusOK, map[string]any{"state": "success", "data": []map[string]any{
			{"roleId": "RR_ADMIN", "roleName": "Admin", "roleNumber": 100, "roleDate": "2026-01-01 00:00:00"},
			{"roleId": "RR_INSPECTOR", "roleName": "Inspector", "roleNumber": 500, "roleDate": "2026-01-01 00:00:00"},
		}})
	})
	mux.HandleFunc("/admin/get/company/users/pass", func(w http.ResponseWriter, r *http.Request) {
		if !authorized(r) {
			reply(w, http.StatusForbidden, map[string]any{"state": "error", "data": "not allowed"})
			return
		}
		reply(w, http.StatusOK, map[string]any{"state": "success", "data": users})
	})
	mux.HandleFunc("/admin/get/vehicle/makers", func(w http.ResponseWriter, r *http.Request) {
		data := []map[string]any{}
		if r.URL.Query().Get("page") == "1" {
			data = append(data, map[string]any{"makerId": "RM_TOYOTA", "makerName": "Toyota", "creationDate": "2026-01-01 00:00:00"})
		}
		reply(w, http.StatusOK, map[string]any{"state": "success", "data": data})
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
	remote := fakeRemote(t)
	defer remote.Close()
	t.Setenv("REMOTE_SERVER_URL", remote.URL)

	if n := count(t, "SELECT COUNT(*) FROM users"); n != 0 {
		t.Fatalf("expected an empty users table on a fresh install, found %d users", n)
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/users/login", controlers.LoginUser)

	// Before the pull nobody can log in.
	if code := login(t, router, initialEmail, initialPassword); code != http.StatusUnauthorized {
		t.Fatalf("login before the pull answered %d, want 401", code)
	}

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

	// Correct initial credentials: roles, users and the makers list arrive.
	t.Setenv("INIT_USER_PASSWORD", initialPassword)
	apiservices.SyncFromRemoteAtStart(context.Background())
	if n := count(t, "SELECT COUNT(*) FROM roles"); n != 2 {
		t.Fatalf("roles = %d, want 2", n)
	}
	if n := count(t, "SELECT COUNT(*) FROM users WHERE user_id IN ('RU_INIT','RU_STAFF')"); n != 2 {
		t.Fatalf("users = %d, want 2", n)
	}
	if n := count(t, "SELECT COUNT(*) FROM vehicle_makers WHERE maker_id = 'RM_TOYOTA'"); n != 1 {
		t.Fatal("makers list not pulled")
	}

	// Now the initial user and the other staff can log in, and their remote
	// sessions are stored for the tablets.
	if code := login(t, router, initialEmail, initialPassword); code != http.StatusOK {
		t.Fatalf("initial user login answered %d, want 200", code)
	}
	if code := login(t, router, staffEmail, staffPassword); code != http.StatusOK {
		t.Fatalf("staff login answered %d, want 200", code)
	}
	if code := login(t, router, staffEmail, "not-the-password"); code != http.StatusUnauthorized {
		t.Fatalf("wrong password answered %d, want 401", code)
	}
	if n := count(t, "SELECT COUNT(*) FROM logins WHERE user_id IN ('RU_INIT','RU_STAFF') AND status = 'active'"); n != 2 {
		t.Fatalf("stored sessions = %d, want 2", n)
	}

	// Pulling again on the next start is harmless.
	apiservices.SyncFromRemoteAtStart(context.Background())
	if n := count(t, "SELECT COUNT(*) FROM users"); n != 2 {
		t.Fatalf("users after a second pull = %d, want 2", n)
	}
}
