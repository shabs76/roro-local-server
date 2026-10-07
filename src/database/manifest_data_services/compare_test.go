package manifestdataservices

// Integration tests for the tablet comparison, the package save and the error codes
// of the saves. Same requirements as inspection_save_test.go (RORO_TEST_DB=1).

import (
	"strings"
	"testing"

	"github.com/shabs76/roro-local-server/constants"
	"github.com/shabs76/roro-local-server/constants/modules/manifest"
	"github.com/shabs76/roro-local-server/gendb"
	"github.com/shabs76/roro-local-server/specials"
)

// tabletView builds what a tablet would send for a vehicle saved with f.request.
func (f *testFixture) tabletView(vehicleId, inspectionTime string, published bool) CompareVehicle {
	req := f.request(vehicleId, inspectionTime)
	v := CompareVehicle{VehicleId: vehicleId, Inspected: true, InspectionTime: inspectionTime, Published: published,
		MediaCount: len(req.Media), OnboardPackageCount: len(req.Packages)}
	for _, c := range req.Checks {
		v.Checks = append(v.Checks, []string{c.CheckId, c.CheckStatus})
	}
	for _, r := range req.Remarks {
		v.Remarks = append(v.Remarks, r.Remark)
	}
	return v
}

func statusOf(t *testing.T, res *CompareResult, id string) string {
	t.Helper()
	for _, it := range res.Items {
		if it.Id == id {
			return it.Status
		}
	}
	return CompareMatch
}

func TestCompareClassifiesVehicles(t *testing.T) {
	f := setup(t)

	// Each vehicle sets up one situation; the tablet's view follows below.
	match := f.newVehicle(t)
	mustSave(t, f.request(match, "2026-10-07 08:00:00"), f.userA, InspectionSaveNew)
	notMarked := f.newVehicle(t)
	mustSave(t, f.request(notMarked, "2026-10-07 08:01:00"), f.userA, InspectionSaveNew)
	missing := f.newVehicle(t)
	publishedMissing := f.newVehicle(t)
	elsewhere := f.newVehicle(t)
	mustSave(t, f.request(elsewhere, "2026-10-07 08:02:00"), f.userB, InspectionSaveNew)
	mismatch := f.newVehicle(t)
	mustSave(t, f.request(mismatch, "2026-10-07 08:03:00"), f.userA, InspectionSaveNew)
	different := f.newVehicle(t)
	mustSave(t, f.request(different, "2026-10-07 08:04:00"), f.userB, InspectionSaveNew)
	untouched := f.newVehicle(t)
	onlyOnServer := f.newVehicle(t) // the tablet does not send it

	mismatchView := f.tabletView(mismatch, "2026-10-07 08:03:00", true)
	mismatchView.Checks[1][1] = "missing"
	mismatchView.Remarks = append(mismatchView.Remarks, "Scratch on bonnet")

	req := CompareRequest{
		Device: CompareDevice{DeviceId: "tablet-" + specials.RandomStringNoSuffix(6), DeviceName: "Test tablet", AppVersion: "test"},
		Vehicles: []CompareVehicle{
			f.tabletView(match, "2026-10-07 08:00:00", true),
			f.tabletView(notMarked, "2026-10-07 08:01:00", false),
			f.tabletView(missing, "2026-10-07 08:05:00", false),
			f.tabletView(publishedMissing, "2026-10-07 08:06:00", true),
			{VehicleId: elsewhere},
			mismatchView,
			f.tabletView(different, "2026-10-07 07:00:00", true),
			{VehicleId: untouched},
			{VehicleId: "NOT_ON_SERVER_VEH", Inspected: true},
		},
	}

	st, res := CompareTabletWithServer(f.manifest, f.userA, req)
	if st.State != constants.SuccessState {
		t.Fatalf("compare failed: %+v", st)
	}

	want := map[string]string{
		match:               CompareMatch,
		notMarked:           ComparePublishedNotMarked,
		missing:             CompareMissingOnServer,
		publishedMissing:    ComparePublishedButMissing,
		elsewhere:           CompareInspectedElsewhere,
		mismatch:            CompareContentMismatch,
		different:           CompareDifferentInspection,
		untouched:           CompareMatch,
		"NOT_ON_SERVER_VEH": CompareUnknownOnServer,
		onlyOnServer:        CompareMissingOnTablet,
	}
	for id, status := range want {
		if got := statusOf(t, res, id); got != status {
			t.Errorf("vehicle %s: status %q, want %q", id, got, status)
		}
	}

	for _, it := range res.Items {
		if it.Id == mismatch && len(it.Differences) != 2 {
			t.Errorf("mismatch differences = %v, want one check and one remark", it.Differences)
		}
		if it.Status != CompareMatch && it.Action == "" {
			t.Errorf("item %s (%s) has no action", it.Id, it.Status)
		}
	}

	if gendb.ColumnExists("tablet_compare_reports", "report_id") {
		if !res.Stored {
			t.Fatal("report not stored although the table exists")
		}
		stl, reports := SelectLatestCompareReports(f.manifest)
		if stl.State != constants.SuccessState {
			t.Fatalf("list reports: %+v", stl)
		}
		found := false
		for _, r := range reports {
			if r.DeviceId == req.Device.DeviceId {
				found = r.ReportId == res.ReportId && r.NeedsAction > 0
			}
		}
		if !found {
			t.Fatalf("latest report of the device not listed: %+v", reports)
		}
		sto, stored, _ := SelectCompareReport(res.ReportId)
		if sto.State != constants.SuccessState || len(stored.Items) != len(res.Items) {
			t.Fatalf("stored report = %+v / %d items, want %d", sto, len(stored.Items), len(res.Items))
		}
	} else if res.Stored {
		t.Fatal("report reported as stored without the table")
	}
}

func TestSaveErrorCodes(t *testing.T) {
	f := setup(t)

	st := InsertInspectionTallyRemarks(f.request("NO_SUCH_VEH", "2026-10-07 09:00:00"), f.userA)
	if st.Adv != SaveErrNotFound {
		t.Fatalf("unknown vehicle: %+v, want adv %s", st, SaveErrNotFound)
	}

	for _, c := range []struct {
		edit func(*manifest.InspectionChecksRequest)
		want string
	}{
		{func(r *manifest.InspectionChecksRequest) { r.MakerId = "NO_SUCH_MAKER" }, `maker "NO_SUCH_MAKER"`},
		{func(r *manifest.InspectionChecksRequest) { r.BodyId = "NO_SUCH_BODY" }, `body type "NO_SUCH_BODY"`},
		{func(r *manifest.InspectionChecksRequest) {
			r.Checks = append(r.Checks, manifest.InspectionCheck{CheckId: "NO_SUCH_CHECK", CheckStatus: "seen/okay"})
		}, `check "NO_SUCH_CHECK"`},
	} {
		vid := f.newVehicle(t)
		req := f.request(vid, "2026-10-07 09:01:00")
		c.edit(&req)
		st := InsertInspectionTallyRemarks(req, f.userA)
		if st.Adv != SaveErrInvalidReference || !strings.Contains(st.Data, c.want) {
			t.Fatalf("unknown reference: %+v, want adv %s naming %s", st, SaveErrInvalidReference, c.want)
		}
		if n := f.count(t, "SELECT COUNT(*) FROM vehicles_talling WHERE vehicle_id = ?", vid); n != 0 {
			t.Fatalf("tally saved for an inspection with an unknown %s", c.want)
		}
	}

	rem := manifest.VehicleRemarksOnlyRequest{VehicleId: "NO_SUCH_VEH", Remarks: []manifest.RemarkSaveRequest{{Remark: "x", RemarkType: "info"}}}
	if st := InsertVehicleRemarksOnly(rem, f.userA); st.Adv != SaveErrNotFound {
		t.Fatalf("remarks-only unknown vehicle: %+v", st)
	}
}

func TestPackageSaveIsIdempotentAndResetsPublished(t *testing.T) {
	f := setup(t)
	suffix := specials.RandomStringNoSuffix(8)
	pkgId, typeA, typeB, status := "PK_"+suffix, "PT1_"+suffix, "PT2_"+suffix, "PS_"+suffix
	for _, s := range []struct {
		q    string
		args []any
	}{
		{"INSERT INTO manifest_packages (package_id, package_number, bl_no, manifest_id, description, creation_time) VALUES (?,?,?,?,?,NOW())", []any{pkgId, "N" + suffix, "BL", f.manifest, "crate"}},
		{"INSERT INTO package_types (type_id, type_name, status, created_time) VALUES (?,?,?,NOW())", []any{typeA, "Type A " + suffix, "active"}},
		{"INSERT INTO package_types (type_id, type_name, status, created_time) VALUES (?,?,?,NOW())", []any{typeB, "Type B " + suffix, "active"}},
		{"INSERT INTO packages_inspection_status (status_id, status_name, description, status_number) VALUES (?,?,?,FLOOR(RAND()*900000)+1000)", []any{status, "Good " + suffix, "ok"}},
	} {
		if _, err := f.db.Exec(s.q, s.args...); err != nil {
			t.Fatal(err)
		}
	}

	req := manifest.PackageInspectionSaveRequest{PackageId: pkgId, PackageImage: "images/p.jpg", TypeId: typeA, InspectionTime: "2026-10-07 10:00:00", InspectionStatusId: status,
		Media: []manifest.PackageExtraMediaRequest{{MediaLink: "images/pm.jpg", MediaType: "image", Remark: "side"}}}
	published := func() string {
		var v string
		if err := f.db.QueryRow("SELECT is_published FROM manifest_packages WHERE package_id = ?", pkgId).Scan(&v); err != nil {
			t.Fatal(err)
		}
		return v
	}
	setPublished := func() {
		if _, err := f.db.Exec("UPDATE manifest_packages SET is_published = 'yes' WHERE package_id = ?", pkgId); err != nil {
			t.Fatal(err)
		}
	}

	for i := 0; i < 3; i++ {
		if st := InsertPackageInspection(req, f.userA); st.State != constants.SuccessState {
			t.Fatalf("package save %d: %+v", i, st)
		}
		if i == 0 {
			setPublished()
		}
	}
	if n := f.count(t, "SELECT COUNT(*) FROM packages_inspection WHERE package_id = ?", pkgId); n != 1 {
		t.Fatalf("package inspections = %d, want 1", n)
	}
	if n := f.count(t, "SELECT COUNT(*) FROM package_gallery WHERE package_id = ?", pkgId); n != 1 {
		t.Fatalf("package media = %d, want 1", n)
	}
	if got := published(); got != "yes" {
		t.Fatalf("is_published after identical resends = %q, want yes", got)
	}

	req.TypeId = typeB
	if st := InsertPackageInspection(req, f.userA); st.State != constants.SuccessState {
		t.Fatalf("changed package save: %+v", st)
	}
	if got := published(); got != "no" {
		t.Fatalf("is_published after a changed inspection = %q, want no", got)
	}

	req.TypeId = "NO_SUCH_TYPE"
	if st := InsertPackageInspection(req, f.userA); st.Adv != SaveErrInvalidReference {
		t.Fatalf("unknown type: %+v", st)
	}
	req.PackageId = "NO_SUCH_PACKAGE"
	if st := InsertPackageInspection(req, f.userA); st.Adv != SaveErrNotFound {
		t.Fatalf("unknown package: %+v", st)
	}
}

func TestMediaFileLookup(t *testing.T) {
	setup(t)
	if !gendb.ColumnExists("media_files", "sha256") {
		t.Skip("media_files not present (migration not applied)")
	}
	sum := specials.RandomStringNoSuffix(64)
	if _, ok := FindMediaFile(sum); ok {
		t.Fatal("unknown hash found")
	}
	SaveMediaFile(sum, "images/x_1.jpg", 10)
	if url, ok := FindMediaFile(sum); !ok || url != "images/x_1.jpg" {
		t.Fatalf("FindMediaFile = %q, %v", url, ok)
	}
}
