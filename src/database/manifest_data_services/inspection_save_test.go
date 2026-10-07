package manifestdataservices

// Integration tests for the vehicle inspection save. They need a MariaDB database
// with the roro_local schema and run only when RORO_TEST_DB=1, using the usual
// MYSQL_HOST, MYSQL_PORT, MYSQL_USER, MYSQL_PASS and MYSQL_DBNAME variables.
// Set RORO_TEST_MIGRATE=1 to apply gendb migrations before the tests.

import (
	"database/sql"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/shabs76/roro-local-server/constants"
	"github.com/shabs76/roro-local-server/constants/modules/manifest"
	"github.com/shabs76/roro-local-server/gendb"
	"github.com/shabs76/roro-local-server/specials"
)

type testFixture struct {
	db       *sql.DB
	manifest string
	makerId  string
	bodyId   string
	checkIds []string
	userA    string
	userB    string
}

var (
	fixture     *testFixture
	fixtureOnce sync.Once
	fixtureErr  error
)

func setup(t *testing.T) *testFixture {
	t.Helper()
	if os.Getenv("RORO_TEST_DB") != "1" {
		t.Skip("set RORO_TEST_DB=1 to run database integration tests")
	}

	fixtureOnce.Do(func() {
		if os.Getenv("RORO_TEST_MIGRATE") == "1" {
			if err := gendb.RunMigrations(30 * time.Second); err != nil {
				fixtureErr = fmt.Errorf("migrations: %w", err)
				return
			}
		}
		db, err := gendb.InitDb()
		if err != nil {
			fixtureErr = err
			return
		}

		suffix := specials.RandomStringNoSuffix(8)
		f := &testFixture{
			db:       db,
			manifest: "M_" + suffix,
			makerId:  "MK_" + suffix,
			bodyId:   "BD_" + suffix,
			userA:    "UA_" + suffix,
			userB:    "UB_" + suffix,
		}
		stmts := []struct {
			q    string
			args []any
		}{
			{"INSERT INTO clients (client_id, client_name, principal, client_location, creation_date) VALUES (?,?,?,?,NOW())", []any{"C_" + suffix, "Client " + suffix, "p", "loc"}},
			{"INSERT INTO manifest (manifest_id, manifest_name, client_id, vessel_name, voyage_no, arrival_date, received_date, uploaded_date) VALUES (?,?,?,?,?,NOW(),CURDATE(),NOW())", []any{f.manifest, "Manifest " + suffix, "C_" + suffix, "Vessel", "V1"}},
			{"INSERT INTO roles (role_id, role_name, role_number, role_date) VALUES (?,?,?,NOW())", []any{"R_" + suffix, "Role " + suffix, 500}},
			{"INSERT INTO users (user_id, fname, lname, email, phone, password, role, status, creation_date) VALUES (?,?,?,?,?,?,?,?,NOW())", []any{f.userA, "Asha", "One", "a@x", "1", "x", "R_" + suffix, "active"}},
			{"INSERT INTO users (user_id, fname, lname, email, phone, password, role, status, creation_date) VALUES (?,?,?,?,?,?,?,?,NOW())", []any{f.userB, "Baraka", "Two", "b@x", "2", "x", "R_" + suffix, "active"}},
			{"INSERT INTO vehicle_makers (maker_id, maker_name, creation_time) VALUES (?,?,NOW())", []any{f.makerId, "Maker " + suffix}},
			{"INSERT INTO vehicle_bodies (body_id, body_name, creation_time) VALUES (?,?,NOW())", []any{f.bodyId, "Body " + suffix}},
		}
		for i := 0; i < 3; i++ {
			id := fmt.Sprintf("CK%d_%s", i, suffix)
			f.checkIds = append(f.checkIds, id)
			stmts = append(stmts, struct {
				q    string
				args []any
			}{"INSERT INTO inspection_checklist (check_id, check_name, updated_date, registered_date) VALUES (?,?,NOW(),NOW())", []any{id, fmt.Sprintf("Check %d %s", i, suffix)}})
		}
		for _, s := range stmts {
			if _, err := db.Exec(s.q, s.args...); err != nil {
				fixtureErr = fmt.Errorf("seed %q: %w", s.q, err)
				return
			}
		}
		fixture = f
	})

	if fixtureErr != nil {
		t.Fatal(fixtureErr)
	}
	return fixture
}

func (f *testFixture) newVehicle(t *testing.T) string {
	t.Helper()
	id := specials.RandomString(20, "_VEH")
	_, err := f.db.Exec("INSERT INTO manifest_vehicles (vehicle_id, manifest_id, chasis_number, description, weight, bl_no) VALUES (?,?,?,?,?,?)",
		id, f.manifest, "CH"+id, "desc", 1000, "BL1")
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func (f *testFixture) request(vehicleId, inspectionTime string) manifest.InspectionChecksRequest {
	checks := []manifest.InspectionCheck{}
	for i, id := range f.checkIds {
		c := manifest.InspectionCheck{CheckId: id, CheckStatus: "seen/okay"}
		if i == 0 {
			c.CheckStatus = "damaged"
			c.FaultImage = "images/fault.jpg"
		}
		checks = append(checks, c)
	}
	return manifest.InspectionChecksRequest{
		VehicleId:      vehicleId,
		VehicleImage:   "images/main.jpg",
		ManifestId:     f.manifest,
		MakerId:        f.makerId,
		BodyId:         f.bodyId,
		ModelName:      "Model",
		InspectionTime: inspectionTime,
		DeckNumber:     "3",
		NumberOfKeys:   2,
		KeyType:        "smart",
		Checks:         checks,
		Remarks:        []manifest.RemarkSaveRequest{{Remark: "side lights", RemarkType: "damage", RemarkImage: "images/r.jpg"}},
		Media:          []manifest.VehicleExtraMediaRequest{{MediaLink: "images/m.jpg", MediaType: "image", Remark: "front"}},
		Packages:       []manifest.OnBoardPackageRequest{{Title: "box", Remark: "small", Media: []manifest.OnBoardPackageMediaRequest{{MediaLink: "images/p.jpg", MediaType: "image"}}}},
	}
}

func (f *testFixture) count(t *testing.T, qr string, args ...any) int {
	t.Helper()
	var n int
	if err := f.db.QueryRow(qr, args...).Scan(&n); err != nil {
		t.Fatalf("%s: %v", qr, err)
	}
	return n
}

func (f *testFixture) historyBatches(t *testing.T, vehicleId string) int {
	return f.count(t, "SELECT COUNT(DISTINCT archived_at) FROM vehicles_talling_history WHERE vehicle_id = ?", vehicleId)
}

func (f *testFixture) publishedFlag(t *testing.T, vehicleId string) string {
	t.Helper()
	var v string
	if err := f.db.QueryRow("SELECT is_published FROM manifest_vehicles WHERE vehicle_id = ?", vehicleId).Scan(&v); err != nil {
		t.Fatal(err)
	}
	return v
}

func mustSave(t *testing.T, req manifest.InspectionChecksRequest, userId, wantAdv string) {
	t.Helper()
	st := InsertInspectionTallyRemarks(req, userId)
	if st.State != constants.SuccessState {
		t.Fatalf("save failed: %+v", st)
	}
	if wantAdv != "" && st.Adv != wantAdv {
		t.Fatalf("save outcome = %q, want %q", st.Adv, wantAdv)
	}
}

func TestResendIsNotArchived(t *testing.T) {
	f := setup(t)
	vid := f.newVehicle(t)
	req := f.request(vid, "2026-10-06 08:00:00")

	mustSave(t, req, f.userA, InspectionSaveNew)
	for i := 0; i < 4; i++ {
		mustSave(t, req, f.userA, InspectionSaveResend)
	}

	if n := f.historyBatches(t, vid); n != 0 {
		t.Fatalf("history batches = %d, want 0", n)
	}
	if n := f.count(t, "SELECT COUNT(*) FROM vehicles_inspection WHERE vehicle_id = ?", vid); n != len(f.checkIds) {
		t.Fatalf("active checks = %d, want %d", n, len(f.checkIds))
	}
	if n := f.count(t, "SELECT COUNT(*) FROM inspection_image ii INNER JOIN vehicles_inspection vi ON vi.inspection_id = ii.inspection_id WHERE vi.vehicle_id = ?", vid); n != 1 {
		t.Fatalf("active fault images = %d, want 1", n)
	}
	for _, tbl := range []string{"vehicles_talling", "inspection_remarks", "vehicle_galllery", "onboard_packages"} {
		if n := f.count(t, "SELECT COUNT(*) FROM "+tbl+" WHERE vehicle_id = ?", vid); n != 1 {
			t.Fatalf("%s rows = %d, want 1", tbl, n)
		}
	}
}

func TestResendKeepsPublishedFlag(t *testing.T) {
	f := setup(t)
	vid := f.newVehicle(t)
	req := f.request(vid, "2026-10-06 08:05:00")

	mustSave(t, req, f.userA, InspectionSaveNew)
	if _, err := f.db.Exec("UPDATE manifest_vehicles SET is_published = 'yes' WHERE vehicle_id = ?", vid); err != nil {
		t.Fatal(err)
	}
	mustSave(t, req, f.userA, InspectionSaveResend)
	if got := f.publishedFlag(t, vid); got != "yes" {
		t.Fatalf("is_published after resend = %q, want yes", got)
	}
}

func TestReinspectionArchivesOnceAndResetsPublished(t *testing.T) {
	f := setup(t)
	vid := f.newVehicle(t)

	mustSave(t, f.request(vid, "2026-10-06 09:00:00"), f.userA, InspectionSaveNew)
	if _, err := f.db.Exec("UPDATE manifest_vehicles SET is_published = 'yes' WHERE vehicle_id = ?", vid); err != nil {
		t.Fatal(err)
	}
	mustSave(t, f.request(vid, "2026-10-06 10:00:00"), f.userB, InspectionSaveReinspect)

	if n := f.historyBatches(t, vid); n != 1 {
		t.Fatalf("history batches = %d, want 1", n)
	}
	if got := f.publishedFlag(t, vid); got != "no" {
		t.Fatalf("is_published after re-inspection = %q, want no", got)
	}

	// Columns are copied by name: maker and body ids land in the right columns, and
	// manifest_id is filled even where vehicles_talling has no manifest_id column.
	var maker, body, manifestId string
	if err := f.db.QueryRow("SELECT maker_id, body_id, manifest_id FROM vehicles_talling_history WHERE vehicle_id = ?", vid).Scan(&maker, &body, &manifestId); err != nil {
		t.Fatal(err)
	}
	if maker != f.makerId || body != f.bodyId || manifestId != f.manifest {
		t.Fatalf("history tally = (%s, %s, %s), want (%s, %s, %s)", maker, body, manifestId, f.makerId, f.bodyId, f.manifest)
	}

	// Every history table of the batch uses the same archived_at value.
	n := f.count(t, `SELECT COUNT(DISTINCT archived_at) FROM (
		SELECT archived_at FROM vehicles_talling_history WHERE vehicle_id = ?
		UNION ALL SELECT archived_at FROM vehicles_inspection_history WHERE vehicle_id = ?
		UNION ALL SELECT archived_at FROM inspection_remarks_history WHERE vehicle_id = ?
		UNION ALL SELECT archived_at FROM vehicle_galllery_history WHERE vehicle_id = ?
		UNION ALL SELECT archived_at FROM onboard_packages_history WHERE vehicle_id = ?) x`, vid, vid, vid, vid, vid)
	if n != 1 {
		t.Fatalf("distinct archived_at across history tables = %d, want 1", n)
	}

	// Later inspections must still archive (no unique key left in history tables),
	// and each one is its own batch even within the same second.
	mustSave(t, f.request(vid, "2026-10-06 11:00:00"), f.userA, InspectionSaveReinspect)
	mustSave(t, f.request(vid, "2026-10-06 12:00:00"), f.userA, InspectionSaveReinspect)
	if n := f.historyBatches(t, vid); n != 3 {
		t.Fatalf("history batches after 4 inspections = %d, want 3", n)
	}
}

func TestSubmissionIdIdentifiesResend(t *testing.T) {
	f := setup(t)
	if !gendb.ColumnExists("vehicles_talling", "submission_id") {
		t.Skip("vehicles_talling.submission_id not present (migration not applied)")
	}
	vid := f.newVehicle(t)
	req := f.request(vid, "2026-10-06 13:00:00")
	req.SubmissionId = "sub-" + vid

	mustSave(t, req, f.userA, InspectionSaveNew)
	req.InspectionTime = "2026-10-06 13:00:09" // tablet rebuilt the payload
	mustSave(t, req, f.userA, InspectionSaveResend)

	other := f.request(vid, "2026-10-06 13:00:00")
	other.SubmissionId = "sub2-" + vid
	mustSave(t, other, f.userA, InspectionSaveReinspect)
}

func TestReinspectFalseReturnsConflict(t *testing.T) {
	f := setup(t)
	vid := f.newVehicle(t)
	mustSave(t, f.request(vid, "2026-10-06 14:00:00"), f.userA, InspectionSaveNew)

	no := false
	req := f.request(vid, "2026-10-06 15:00:00")
	req.Reinspect = &no
	st := InsertInspectionTallyRemarks(req, f.userB)
	if st.Adv != InspectionSaveConflict {
		t.Fatalf("outcome = %+v, want conflict", st)
	}
	if n := f.historyBatches(t, vid); n != 0 {
		t.Fatalf("history batches after conflict = %d, want 0", n)
	}

	yes := true
	req.Reinspect = &yes
	mustSave(t, req, f.userB, InspectionSaveReinspect)
}

func TestDuplicateRemarkTextInOneRequest(t *testing.T) {
	f := setup(t)
	vid := f.newVehicle(t)
	req := f.request(vid, "2026-10-06 16:00:00")
	req.Remarks = append(req.Remarks, manifest.RemarkSaveRequest{Remark: "Side lights", RemarkType: "info", RemarkImage: ""})

	mustSave(t, req, f.userA, InspectionSaveNew)
	if n := f.count(t, "SELECT COUNT(*) FROM inspection_remarks WHERE vehicle_id = ?", vid); n != 1 {
		t.Fatalf("remarks = %d, want 1", n)
	}
}

func TestFailedSaveLeavesNothing(t *testing.T) {
	f := setup(t)
	vid := f.newVehicle(t)
	req := f.request(vid, "2026-10-06 17:00:00")
	req.Checks = append(req.Checks, manifest.InspectionCheck{CheckId: "does-not-exist", CheckStatus: "seen/okay"})

	if st := InsertInspectionTallyRemarks(req, f.userA); st.State == constants.SuccessState {
		t.Fatal("save with unknown check id succeeded")
	}
	if n := f.count(t, "SELECT COUNT(*) FROM vehicles_talling WHERE vehicle_id = ?", vid); n != 0 {
		t.Fatalf("tally rows after failed save = %d, want 0", n)
	}
	if n := f.count(t, "SELECT COUNT(*) FROM vehicles_inspection WHERE vehicle_id = ?", vid); n != 0 {
		t.Fatalf("inspection rows after failed save = %d, want 0", n)
	}
	if n := f.count(t, "SELECT COUNT(*) FROM manifest_vehicles WHERE vehicle_id = ? AND inspection_status = 'yes'", vid); n != 0 {
		t.Fatal("vehicle marked inspected after failed save")
	}
}

func TestConcurrentSavesOfSameVehicle(t *testing.T) {
	f := setup(t)
	vid := f.newVehicle(t)

	var wg sync.WaitGroup
	results := make([]*constants.AnswerState, 6)
	for i := range results {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			// Three tablets, each sending its own inspection twice.
			results[i] = InsertInspectionTallyRemarks(f.request(vid, fmt.Sprintf("2026-10-06 18:00:0%d", i%3)), f.userA)
		}(i)
	}
	wg.Wait()

	for i, st := range results {
		if st.State != constants.SuccessState {
			t.Fatalf("save %d failed: %+v", i, st)
		}
	}
	if n := f.count(t, "SELECT COUNT(*) FROM vehicles_talling WHERE vehicle_id = ?", vid); n != 1 {
		t.Fatalf("active tallies = %d, want 1", n)
	}
	// Every archived batch has its own archived_at, so no batch holds two tallies.
	if n := f.count(t, "SELECT COUNT(*) - COUNT(DISTINCT archived_at) FROM vehicles_talling_history WHERE vehicle_id = ?", vid); n != 0 {
		t.Fatalf("%d history tallies share an archived_at with another batch", n)
	}
	if n := f.count(t, "SELECT COUNT(*) FROM vehicles_inspection WHERE vehicle_id = ?", vid); n != len(f.checkIds) {
		t.Fatalf("active checks = %d, want %d", n, len(f.checkIds))
	}
}

func TestRemarksOnlyRepeat(t *testing.T) {
	f := setup(t)
	vid := f.newVehicle(t)
	mustSave(t, f.request(vid, "2026-10-06 19:00:00"), f.userA, InspectionSaveNew)
	if _, err := f.db.Exec("UPDATE manifest_vehicles SET is_published = 'yes' WHERE vehicle_id = ?", vid); err != nil {
		t.Fatal(err)
	}

	req := manifest.VehicleRemarksOnlyRequest{VehicleId: vid, Remarks: []manifest.RemarkSaveRequest{
		{Remark: "scratch on door", RemarkType: "damage", RemarkImage: ""},
		{Remark: "side lights", RemarkType: "damage", RemarkImage: ""},
	}}
	for i := 0; i < 2; i++ {
		if st := InsertVehicleRemarksOnly(req, f.userA); st.State != constants.SuccessState {
			t.Fatalf("remarks-only save %d failed: %+v", i, st)
		}
	}
	if n := f.count(t, "SELECT COUNT(*) FROM inspection_remarks WHERE vehicle_id = ?", vid); n != 2 {
		t.Fatalf("remarks = %d, want 2", n)
	}
	if got := f.publishedFlag(t, vid); got != "no" {
		t.Fatalf("is_published after remarks-only save = %q, want no", got)
	}
}

func TestVehicleListExtras(t *testing.T) {
	f := setup(t)
	vid := f.newVehicle(t)
	untouched := f.newVehicle(t)
	mustSave(t, f.request(vid, "2026-10-06 20:00:00"), f.userA, InspectionSaveNew)
	mustSave(t, f.request(vid, "2026-10-06 21:00:00"), f.userB, InspectionSaveReinspect)

	st, extras := SelectVehicleListExtras([]string{vid, untouched})
	if st.State != constants.SuccessState {
		t.Fatalf("extras failed: %+v", st)
	}
	e := extras[vid]
	if !e.HasActiveTally || e.InspectedBy != "Baraka Two" || !e.IsDamaged || e.HistoryBatches != 1 || e.NumberOfKeys != 2 {
		t.Fatalf("extras = %+v", *e)
	}
	if u := extras[untouched]; u == nil || u.HasActiveTally || u.IsDamaged || u.HistoryBatches != 0 {
		t.Fatalf("extras for uninspected vehicle = %+v", u)
	}
}
