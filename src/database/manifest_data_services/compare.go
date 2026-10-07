package manifestdataservices

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"time"

	"github.com/shabs76/roro-local-server/constants"
	"github.com/shabs76/roro-local-server/gendb"
	"github.com/shabs76/roro-local-server/specials"
)

// Tablet-vs-server comparison. A tablet sends a compact summary of what its SQLite
// database holds for one manifest; the server compares it with its own records, tells
// the tablet what to do with each difference, and keeps the report so an operator can
// see which tablet still holds unpublished or mismatched work.

// Statuses of a compared vehicle or package.
const (
	CompareMatch               = "match"
	CompareMissingOnServer     = "missing_on_server"     // tablet inspected it, server has no inspection
	ComparePublishedButMissing = "published_but_missing" // tablet marked it published, server has no inspection
	ComparePublishedNotMarked  = "published_not_marked"  // server has the same inspection, tablet not marked published
	CompareInspectedElsewhere  = "inspected_elsewhere"   // server has an inspection, tablet has none
	CompareContentMismatch     = "content_mismatch"      // same inspection, different check results or remarks
	CompareDifferentInspection = "different_inspection"  // both inspected it, but not the same inspection
	CompareUnknownOnServer     = "unknown_on_server"     // the server does not know the vehicle or package
	CompareMissingOnTablet     = "missing_on_tablet"     // the server has it, the tablet does not
)

// Actions the tablet offers for a status.
var compareActions = map[string]string{
	CompareMissingOnServer:     "publish",
	ComparePublishedButMissing: "republish",
	ComparePublishedNotMarked:  "mark_published",
	CompareInspectedElsewhere:  "pull_status",
	CompareContentMismatch:     "republish",
	CompareDifferentInspection: "review",
	CompareUnknownOnServer:     "review",
	CompareMissingOnTablet:     "refresh_manifest",
}

type CompareDevice struct {
	DeviceId   string `json:"deviceId" binding:"required"`
	DeviceName string `json:"deviceName"`
	AppVersion string `json:"appVersion"`
}

type CompareVehicle struct {
	VehicleId      string `json:"vehicleId" binding:"required"`
	ChasisNumber   string `json:"chasisNumber"`
	Inspected      bool   `json:"inspected"`
	InspectionTime string `json:"inspectionTime"` // the time the tablet sends as inspectionTime when publishing
	SubmissionId   string `json:"submissionId"`
	Published      bool   `json:"published"`
	QueueState     string `json:"queueState"`
	// Checks holds [checkId, value] pairs.
	Checks              [][]string `json:"checks"`
	Remarks             []string   `json:"remarks"`
	MediaCount          int        `json:"mediaCount"`
	OnboardPackageCount int        `json:"onboardPackageCount"`
}

type ComparePackage struct {
	PackageId          string `json:"packageId" binding:"required"`
	PackageNumber      string `json:"packageNumber"`
	Inspected          bool   `json:"inspected"`
	Published          bool   `json:"published"`
	TypeId             string `json:"typeId"`
	InspectionStatusId string `json:"inspectionStatusId"`
	InspectionTime     string `json:"inspectionTime"`
}

type CompareRequest struct {
	Device   CompareDevice    `json:"device" binding:"required"`
	Vehicles []CompareVehicle `json:"vehicles"`
	Packages []ComparePackage `json:"packages"`
}

// CompareServerSnapshot is what the server holds for one item.
type CompareServerSnapshot struct {
	Inspected           bool   `json:"inspected"`
	InspectionTime      string `json:"inspectionTime"`
	InspectedBy         string `json:"inspectedBy"`
	SubmissionId        string `json:"submissionId"`
	IsPublished         string `json:"isPublished"`
	HistoryBatches      int    `json:"historyBatches"`
	CheckCount          int    `json:"checkCount"`
	RemarkCount         int    `json:"remarkCount"`
	MediaCount          int    `json:"mediaCount"`
	OnboardPackageCount int    `json:"onboardPackageCount"`
}

type CompareItem struct {
	Kind        string                 `json:"kind"` // vehicle or package
	Id          string                 `json:"id"`
	Label       string                 `json:"label"`
	Status      string                 `json:"status"`
	Action      string                 `json:"action"`
	Differences []string               `json:"differences"`
	Info        []string               `json:"info"`
	Server      *CompareServerSnapshot `json:"server,omitempty"`
}

type CompareSummary struct {
	Vehicles map[string]int `json:"vehicles"`
	Packages map[string]int `json:"packages"`
}

type CompareResult struct {
	ReportId   string         `json:"reportId"`
	ManifestId string         `json:"manifestId"`
	ComparedAt string         `json:"comparedAt"`
	Stored     bool           `json:"stored"`
	Summary    CompareSummary `json:"summary"`
	Items      []CompareItem  `json:"items"`
}

type serverVehicle struct {
	chassis        string
	inspected      bool
	inspectionTime string
	isPublished    string
	submissionId   string
	inspectedBy    string
	checks         map[string]string
	remarks        map[string]bool
	mediaCount     int
	onboardCount   int
	historyBatches int
}

type serverPackage struct {
	number         string
	inspected      bool
	isPublished    string
	typeId         string
	statusId       string
	inspectionTime string
}

const compareTimeFormat = "%Y-%m-%d %H:%i:%s"

// CompareTabletWithServer compares a tablet's summary of a manifest with the server's
// records and stores the report.
func CompareTabletWithServer(manifestId, userId string, req CompareRequest) (*constants.AnswerState, *CompareResult) {
	vehicles, err := loadServerVehicles(manifestId)
	if err != nil {
		slog.Error("Compare: failed to load vehicles", "manifest", manifestId, "error", err)
		return &constants.AnswerState{State: constants.ErrorState, Data: "Failed to read server vehicles", Adv: "none"}, nil
	}
	packages, err := loadServerPackages(manifestId)
	if err != nil {
		slog.Error("Compare: failed to load packages", "manifest", manifestId, "error", err)
		return &constants.AnswerState{State: constants.ErrorState, Data: "Failed to read server packages", Adv: "none"}, nil
	}
	checkNames, err := loadCheckNames()
	if err != nil {
		slog.Error("Compare: failed to load check names", "error", err)
		checkNames = map[string]string{}
	}

	result := &CompareResult{
		ReportId:   specials.RandomString(24, "_CMP"),
		ManifestId: manifestId,
		ComparedAt: time.Now().Format("2006-01-02 15:04:05"),
		Summary:    CompareSummary{Vehicles: map[string]int{}, Packages: map[string]int{}},
		Items:      []CompareItem{},
	}

	seenVehicles := map[string]bool{}
	for _, v := range req.Vehicles {
		seenVehicles[v.VehicleId] = true
		item := compareVehicle(v, vehicles[v.VehicleId], checkNames)
		result.Summary.Vehicles[item.Status]++
		if item.Status != CompareMatch {
			result.Items = append(result.Items, item)
		}
	}
	// Vehicles the tablet does not have at all. Only reported when the tablet sent its
	// vehicle list, so a tablet that has not downloaded the manifest is not flooded.
	if len(req.Vehicles) > 0 {
		ids := make([]string, 0, len(vehicles))
		for id := range vehicles {
			if !seenVehicles[id] {
				ids = append(ids, id)
			}
		}
		sort.Strings(ids)
		for _, id := range ids {
			sv := vehicles[id]
			item := CompareItem{Kind: "vehicle", Id: id, Label: sv.chassis, Status: CompareMissingOnTablet,
				Action: compareActions[CompareMissingOnTablet], Differences: []string{"The tablet does not have this vehicle. Download the manifest again to get it."},
				Info: []string{}, Server: sv.snapshot()}
			result.Summary.Vehicles[item.Status]++
			result.Items = append(result.Items, item)
		}
	}

	for _, p := range req.Packages {
		item := comparePackage(p, packages[p.PackageId])
		result.Summary.Packages[item.Status]++
		if item.Status != CompareMatch {
			result.Items = append(result.Items, item)
		}
	}

	result.Stored = storeCompareReport(manifestId, userId, req.Device, result)
	return &constants.AnswerState{State: constants.SuccessState, Data: "success", Adv: "none"}, result
}

func (sv *serverVehicle) snapshot() *CompareServerSnapshot {
	if sv == nil {
		return nil
	}
	return &CompareServerSnapshot{
		Inspected:           sv.inspected,
		InspectionTime:      sv.inspectionTime,
		InspectedBy:         sv.inspectedBy,
		SubmissionId:        sv.submissionId,
		IsPublished:         sv.isPublished,
		HistoryBatches:      sv.historyBatches,
		CheckCount:          len(sv.checks),
		RemarkCount:         len(sv.remarks),
		MediaCount:          sv.mediaCount,
		OnboardPackageCount: sv.onboardCount,
	}
}

func compareVehicle(v CompareVehicle, sv *serverVehicle, checkNames map[string]string) CompareItem {
	item := CompareItem{Kind: "vehicle", Id: v.VehicleId, Label: v.ChasisNumber, Differences: []string{}, Info: []string{}}
	if sv == nil {
		item.Status = CompareUnknownOnServer
		item.Action = compareActions[item.Status]
		item.Differences = append(item.Differences, "The server has no vehicle with this id in this manifest.")
		return item
	}
	if item.Label == "" {
		item.Label = sv.chassis
	}
	item.Server = sv.snapshot()

	switch {
	case !v.Inspected && !sv.inspected:
		item.Status = CompareMatch
	case v.Inspected && !sv.inspected:
		if v.Published {
			item.Status = ComparePublishedButMissing
			item.Differences = append(item.Differences, "The tablet marked this inspection as published, but the server has no inspection.")
		} else {
			item.Status = CompareMissingOnServer
			item.Differences = append(item.Differences, "The server has not received this inspection yet.")
			if v.QueueState != "" {
				item.Info = append(item.Info, "Tablet publish state: "+v.QueueState)
			}
		}
	case !v.Inspected && sv.inspected:
		item.Status = CompareInspectedElsewhere
		item.Differences = append(item.Differences, fmt.Sprintf("Inspected on the server by %s at %s; this tablet has no inspection.", orUnknown(sv.inspectedBy), sv.inspectionTime))
	default:
		if !sameSubmission(v, sv) {
			item.Status = CompareDifferentInspection
			item.Differences = append(item.Differences, fmt.Sprintf("Tablet inspection time %s; server inspection by %s at %s.",
				orUnknown(v.InspectionTime), orUnknown(sv.inspectedBy), sv.inspectionTime))
			break
		}
		diffs := compareChecks(v.Checks, sv.checks, checkNames)
		diffs = append(diffs, compareRemarks(v.Remarks, sv.remarks)...)
		switch {
		case len(diffs) > 0:
			item.Status = CompareContentMismatch
			item.Differences = append(item.Differences, diffs...)
		case !v.Published:
			item.Status = ComparePublishedNotMarked
			item.Differences = append(item.Differences, "The server already has this inspection, but the tablet has not marked it as published.")
		default:
			item.Status = CompareMatch
		}
	}

	if v.Inspected && sv.inspected {
		if v.MediaCount != sv.mediaCount {
			item.Info = append(item.Info, fmt.Sprintf("Extra media: tablet %d, server %d", v.MediaCount, sv.mediaCount))
		}
		if v.OnboardPackageCount != sv.onboardCount {
			item.Info = append(item.Info, fmt.Sprintf("Onboard packages: tablet %d, server %d", v.OnboardPackageCount, sv.onboardCount))
		}
	}
	if sv.historyBatches > 0 {
		item.Info = append(item.Info, fmt.Sprintf("The server holds %d earlier inspection(s) of this vehicle", sv.historyBatches))
	}
	item.Action = compareActions[item.Status]
	return item
}

// sameSubmission reports whether the tablet's inspection is the one the server holds:
// the same submission id, or, when either side has none, the same inspection time.
func sameSubmission(v CompareVehicle, sv *serverVehicle) bool {
	if v.SubmissionId != "" && sv.submissionId != "" {
		return v.SubmissionId == sv.submissionId
	}
	return v.InspectionTime != "" && v.InspectionTime == sv.inspectionTime
}

func compareChecks(tablet [][]string, server map[string]string, names map[string]string) []string {
	tabletChecks := map[string]string{}
	for _, c := range tablet {
		if len(c) == 2 {
			tabletChecks[c[0]] = c[1]
		}
	}
	ids := map[string]bool{}
	for id := range tabletChecks {
		ids[id] = true
	}
	for id := range server {
		ids[id] = true
	}
	sorted := make([]string, 0, len(ids))
	for id := range ids {
		sorted = append(sorted, id)
	}
	sort.Strings(sorted)

	diffs := []string{}
	for _, id := range sorted {
		name := names[id]
		if name == "" {
			name = id
		}
		tv, onTablet := tabletChecks[id]
		svv, onServer := server[id]
		switch {
		case onTablet && !onServer:
			diffs = append(diffs, fmt.Sprintf("Check %q: tablet %q, missing on server", name, tv))
		case !onTablet && onServer:
			diffs = append(diffs, fmt.Sprintf("Check %q: server %q, missing on tablet", name, svv))
		case !strings.EqualFold(strings.TrimSpace(tv), strings.TrimSpace(svv)):
			diffs = append(diffs, fmt.Sprintf("Check %q: tablet %q, server %q", name, tv, svv))
		}
	}
	return diffs
}

// compareRemarks compares remark texts the way the database does: case and
// surrounding spaces do not matter.
func compareRemarks(tablet []string, server map[string]bool) []string {
	tabletSet := map[string]string{}
	for _, r := range tablet {
		tabletSet[normalizeRemark(r)] = r
	}
	diffs := []string{}
	for key, text := range tabletSet {
		if key != "" && !server[key] {
			diffs = append(diffs, fmt.Sprintf("Remark %q is on the tablet but not on the server", text))
		}
	}
	for key := range server {
		if _, ok := tabletSet[key]; !ok {
			diffs = append(diffs, fmt.Sprintf("Remark %q is on the server but not on the tablet", key))
		}
	}
	sort.Strings(diffs)
	return diffs
}

func normalizeRemark(r string) string {
	return strings.ToLower(strings.TrimSpace(r))
}

func comparePackage(p ComparePackage, sp *serverPackage) CompareItem {
	item := CompareItem{Kind: "package", Id: p.PackageId, Label: p.PackageNumber, Differences: []string{}, Info: []string{}}
	if sp == nil {
		item.Status = CompareUnknownOnServer
		item.Action = compareActions[item.Status]
		item.Differences = append(item.Differences, "The server has no package with this id in this manifest.")
		return item
	}
	if item.Label == "" {
		item.Label = sp.number
	}

	switch {
	case !p.Inspected && !sp.inspected:
		item.Status = CompareMatch
	case p.Inspected && !sp.inspected:
		if p.Published {
			item.Status = ComparePublishedButMissing
			item.Differences = append(item.Differences, "The tablet marked this package inspection as published, but the server has none.")
		} else {
			item.Status = CompareMissingOnServer
			item.Differences = append(item.Differences, "The server has not received this package inspection yet.")
		}
	case !p.Inspected && sp.inspected:
		item.Status = CompareInspectedElsewhere
		item.Differences = append(item.Differences, fmt.Sprintf("Inspected on the server at %s; this tablet has no inspection.", sp.inspectionTime))
	default:
		if p.InspectionTime != "" && sp.inspectionTime != "" && p.InspectionTime != sp.inspectionTime {
			item.Status = CompareDifferentInspection
			item.Differences = append(item.Differences, fmt.Sprintf("Tablet inspection time %s; server inspection time %s.", p.InspectionTime, sp.inspectionTime))
			break
		}
		if p.InspectionStatusId != "" && p.InspectionStatusId != sp.statusId {
			item.Differences = append(item.Differences, "Package condition differs between tablet and server")
		}
		if p.TypeId != "" && p.TypeId != sp.typeId {
			item.Differences = append(item.Differences, "Package type differs between tablet and server")
		}
		switch {
		case len(item.Differences) > 0:
			item.Status = CompareContentMismatch
		case !p.Published:
			item.Status = ComparePublishedNotMarked
			item.Differences = append(item.Differences, "The server already has this package inspection, but the tablet has not marked it as published.")
		default:
			item.Status = CompareMatch
		}
	}
	item.Action = compareActions[item.Status]
	return item
}

func orUnknown(s string) string {
	if strings.TrimSpace(s) == "" {
		return "unknown"
	}
	return s
}

// queryEach runs a query and calls scan for every row.
func queryEach(qr string, vals []any, scan func(*sql.Rows) error) error {
	st, rows := gendb.SelectGeneral(qr, vals)
	if st.State != constants.SuccessState {
		return fmt.Errorf("%s", st.Data)
	}
	defer rows.Close()
	for rows.Next() {
		if err := scan(rows); err != nil {
			return err
		}
	}
	return rows.Err()
}

func loadServerVehicles(manifestId string) (map[string]*serverVehicle, error) {
	vehicles := map[string]*serverVehicle{}

	submission := "''"
	if gendb.ColumnExists("vehicles_talling", "submission_id") {
		submission = "COALESCE(vt.submission_id, '')"
	}
	qr := `SELECT mv.vehicle_id, mv.chasis_number, mv.inspection_status, DATE_FORMAT(mv.inspection_time, '` + compareTimeFormat + `'),
		mv.is_published, vt.vehicle_id IS NOT NULL, ` + submission + `, COALESCE(CONCAT(u.fname, ' ', u.lname), '')
		FROM manifest_vehicles mv
		LEFT JOIN vehicles_talling vt ON vt.vehicle_id = mv.vehicle_id
		LEFT JOIN users u ON u.user_id = vt.user_id
		WHERE mv.manifest_id = ?`
	err := queryEach(qr, []any{manifestId}, func(rows *sql.Rows) error {
		var id, status string
		var hasTally bool
		sv := &serverVehicle{checks: map[string]string{}, remarks: map[string]bool{}}
		if err := rows.Scan(&id, &sv.chassis, &status, &sv.inspectionTime, &sv.isPublished, &hasTally, &sv.submissionId, &sv.inspectedBy); err != nil {
			return err
		}
		sv.inspected = status == "yes" || hasTally
		vehicles[id] = sv
		return nil
	})
	if err != nil {
		return nil, err
	}

	byVehicle := ` INNER JOIN manifest_vehicles mv ON mv.vehicle_id = x.vehicle_id WHERE mv.manifest_id = ?`
	if err := queryEach("SELECT x.vehicle_id, x.check_id, x.status FROM vehicles_inspection x"+byVehicle, []any{manifestId}, func(rows *sql.Rows) error {
		var id, check, status string
		if err := rows.Scan(&id, &check, &status); err != nil {
			return err
		}
		if sv := vehicles[id]; sv != nil {
			sv.checks[check] = status
		}
		return nil
	}); err != nil {
		return nil, err
	}

	if err := queryEach("SELECT x.vehicle_id, x.remark FROM inspection_remarks x"+byVehicle, []any{manifestId}, func(rows *sql.Rows) error {
		var id, remark string
		if err := rows.Scan(&id, &remark); err != nil {
			return err
		}
		if sv := vehicles[id]; sv != nil {
			sv.remarks[normalizeRemark(remark)] = true
		}
		return nil
	}); err != nil {
		return nil, err
	}

	counts := []struct {
		qr  string
		set func(*serverVehicle, int)
	}{
		{"SELECT x.vehicle_id, COUNT(*) FROM vehicle_galllery x" + byVehicle + " GROUP BY x.vehicle_id", func(sv *serverVehicle, n int) { sv.mediaCount = n }},
		{"SELECT x.vehicle_id, COUNT(*) FROM onboard_packages x" + byVehicle + " GROUP BY x.vehicle_id", func(sv *serverVehicle, n int) { sv.onboardCount = n }},
		{"SELECT x.vehicle_id, COUNT(DISTINCT x.archived_at) FROM vehicles_talling_history x" + byVehicle + " GROUP BY x.vehicle_id", func(sv *serverVehicle, n int) { sv.historyBatches = n }},
	}
	for _, c := range counts {
		if err := queryEach(c.qr, []any{manifestId}, func(rows *sql.Rows) error {
			var id string
			var n int
			if err := rows.Scan(&id, &n); err != nil {
				return err
			}
			if sv := vehicles[id]; sv != nil {
				c.set(sv, n)
			}
			return nil
		}); err != nil {
			return nil, err
		}
	}

	return vehicles, nil
}

func loadServerPackages(manifestId string) (map[string]*serverPackage, error) {
	packages := map[string]*serverPackage{}
	qr := `SELECT mp.package_id, mp.package_number, mp.is_inspected, mp.is_published, pi.package_id IS NOT NULL,
		COALESCE(pi.type_id, ''), COALESCE(pi.inspection_status, ''), COALESCE(DATE_FORMAT(pi.inspection_time, '` + compareTimeFormat + `'), '')
		FROM manifest_packages mp
		LEFT JOIN packages_inspection pi ON pi.package_id = mp.package_id
		WHERE mp.manifest_id = ?`
	err := queryEach(qr, []any{manifestId}, func(rows *sql.Rows) error {
		var id, isInspected string
		var hasInspection bool
		sp := &serverPackage{}
		if err := rows.Scan(&id, &sp.number, &isInspected, &sp.isPublished, &hasInspection, &sp.typeId, &sp.statusId, &sp.inspectionTime); err != nil {
			return err
		}
		sp.inspected = isInspected == "yes" || hasInspection
		packages[id] = sp
		return nil
	})
	return packages, err
}

func loadCheckNames() (map[string]string, error) {
	names := map[string]string{}
	err := queryEach("SELECT check_id, check_name FROM inspection_checklist", nil, func(rows *sql.Rows) error {
		var id, name string
		if err := rows.Scan(&id, &name); err != nil {
			return err
		}
		names[id] = name
		return nil
	})
	return names, err
}

func compareReportsReady() bool {
	return gendb.ColumnExists("tablet_compare_reports", "report_id")
}

// storeCompareReport keeps the report for the server-side overview. It reports false
// when the table does not exist (migration not applied) or the insert failed.
func storeCompareReport(manifestId, userId string, device CompareDevice, result *CompareResult) bool {
	if !compareReportsReady() {
		return false
	}
	summary, err := json.Marshal(result.Summary)
	if err != nil {
		return false
	}
	items, err := json.Marshal(result.Items)
	if err != nil {
		return false
	}
	db, err := gendb.InitDb()
	if err != nil {
		return false
	}
	if _, err := db.Exec(`INSERT INTO tablet_compare_reports
		(report_id, manifest_id, device_id, device_name, app_version, user_id, summary_json, items_json)
		VALUES (?,?,?,?,?,?,?,?)`,
		result.ReportId, manifestId, device.DeviceId, device.DeviceName, device.AppVersion, userId, string(summary), string(items)); err != nil {
		slog.Error("Failed to store compare report", "manifest", manifestId, "device", device.DeviceId, "error", err)
		return false
	}
	return true
}

// CompareReportOverview is the latest report of one device.
type CompareReportOverview struct {
	ReportId    string         `json:"reportId"`
	DeviceId    string         `json:"deviceId"`
	DeviceName  string         `json:"deviceName"`
	AppVersion  string         `json:"appVersion"`
	UserId      string         `json:"userId"`
	UserName    string         `json:"userName"`
	ReportedAt  string         `json:"reportedAt"`
	Summary     CompareSummary `json:"summary"`
	NeedsAction int            `json:"needsAction"`
}

// SelectLatestCompareReports returns the most recent report of every device for a
// manifest, newest first.
func SelectLatestCompareReports(manifestId string) (*constants.AnswerState, []CompareReportOverview) {
	result := []CompareReportOverview{}
	if !compareReportsReady() {
		return &constants.AnswerState{State: constants.SuccessState, Data: "Compare reports are not stored on this server yet (migration not applied)", Adv: "none"}, result
	}
	qr := `SELECT r.report_id, r.device_id, r.device_name, r.app_version, r.user_id, COALESCE(CONCAT(u.fname, ' ', u.lname), ''),
		DATE_FORMAT(r.created_at, '` + compareTimeFormat + `'), r.summary_json
		FROM tablet_compare_reports r
		INNER JOIN (
			SELECT device_id, MAX(created_at) AS latest FROM tablet_compare_reports WHERE manifest_id = ? GROUP BY device_id
		) l ON l.device_id = r.device_id AND l.latest = r.created_at
		LEFT JOIN users u ON u.user_id = r.user_id
		WHERE r.manifest_id = ?
		ORDER BY r.created_at DESC`
	seen := map[string]bool{}
	err := queryEach(qr, []any{manifestId, manifestId}, func(rows *sql.Rows) error {
		var o CompareReportOverview
		var summary string
		if err := rows.Scan(&o.ReportId, &o.DeviceId, &o.DeviceName, &o.AppVersion, &o.UserId, &o.UserName, &o.ReportedAt, &summary); err != nil {
			return err
		}
		if seen[o.DeviceId] { // two reports in the same second
			return nil
		}
		seen[o.DeviceId] = true
		if err := json.Unmarshal([]byte(summary), &o.Summary); err != nil {
			return err
		}
		for status, n := range o.Summary.Vehicles {
			if status != CompareMatch {
				o.NeedsAction += n
			}
		}
		for status, n := range o.Summary.Packages {
			if status != CompareMatch {
				o.NeedsAction += n
			}
		}
		result = append(result, o)
		return nil
	})
	if err != nil {
		slog.Error("Failed to read compare reports", "manifest", manifestId, "error", err)
		return &constants.AnswerState{State: constants.ErrorState, Data: "Failed to read compare reports", Adv: "none"}, nil
	}
	return &constants.AnswerState{State: constants.SuccessState, Data: "success", Adv: "none"}, result
}

// SelectCompareReport returns one stored report with all its items.
func SelectCompareReport(reportId string) (*constants.AnswerState, *CompareResult, *CompareReportOverview) {
	if !compareReportsReady() {
		return &constants.AnswerState{State: constants.ErrorState, Data: "Compare reports are not stored on this server", Adv: "none"}, nil, nil
	}
	var o CompareReportOverview
	var r CompareResult
	var summary, items string
	found := false
	err := queryEach(`SELECT r.report_id, r.manifest_id, r.device_id, r.device_name, r.app_version, r.user_id,
		COALESCE(CONCAT(u.fname, ' ', u.lname), ''), DATE_FORMAT(r.created_at, '`+compareTimeFormat+`'), r.summary_json, r.items_json
		FROM tablet_compare_reports r LEFT JOIN users u ON u.user_id = r.user_id WHERE r.report_id = ?`,
		[]any{reportId}, func(rows *sql.Rows) error {
			found = true
			return rows.Scan(&o.ReportId, &r.ManifestId, &o.DeviceId, &o.DeviceName, &o.AppVersion, &o.UserId, &o.UserName, &o.ReportedAt, &summary, &items)
		})
	if err != nil {
		slog.Error("Failed to read compare report", "report", reportId, "error", err)
		return &constants.AnswerState{State: constants.ErrorState, Data: "Failed to read compare report", Adv: "none"}, nil, nil
	}
	if !found {
		return &constants.AnswerState{State: constants.ErrorState, Data: "Compare report not found", Adv: "not_found"}, nil, nil
	}
	if err := json.Unmarshal([]byte(summary), &r.Summary); err != nil {
		return &constants.AnswerState{State: constants.ErrorState, Data: "Stored report is unreadable", Adv: "none"}, nil, nil
	}
	if err := json.Unmarshal([]byte(items), &r.Items); err != nil {
		return &constants.AnswerState{State: constants.ErrorState, Data: "Stored report is unreadable", Adv: "none"}, nil, nil
	}
	r.ReportId, r.ComparedAt, r.Stored = o.ReportId, o.ReportedAt, true
	o.Summary = r.Summary
	return &constants.AnswerState{State: constants.SuccessState, Data: "success", Adv: "none"}, &r, &o
}
