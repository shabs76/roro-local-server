package manifestdataservices

import (
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/shabs76/roro-local-server/constants"
	"github.com/shabs76/roro-local-server/constants/modules/manifest"
	"github.com/shabs76/roro-local-server/gendb"
)

func SelectVehicleTypes(subQuery string, val []any) (st *constants.AnswerState, data []manifest.VehicleBodyTypes) {
	qr := "SELECT `body_id`, `body_name`, `creation_time` FROM `vehicle_bodies` WHERE " + subQuery

	st, res := gendb.SelectGeneral(qr, val)

	if st.State != constants.SuccessState {
		return st, nil
	}

	defer res.Close()

	var typesList []manifest.VehicleBodyTypes

	for res.Next() {
		var body manifest.VehicleBodyTypes
		ers := res.Scan(&body.BodyId, &body.BodyName, &body.CreationDate)
		if ers != nil {
			return &constants.AnswerState{
				State: constants.ErrorState,
				Data:  "Failed to bind vehicle body type details due to " + ers.Error(),
				Adv:   "none",
			}, nil
		}

		typesList = append(typesList, body)
	}

	return &constants.AnswerState{
		State: constants.SuccessState,
		Data:  "success",
		Adv:   "none",
	}, typesList
}

func SelectVehicleMakers(subQuery string, val []any) (st *constants.AnswerState, data []manifest.VehicleMakers) {
	qr := "SELECT `maker_id`, `maker_name`, `creation_time` FROM `vehicle_makers` WHERE " + subQuery

	st, res := gendb.SelectGeneral(qr, val)

	if st.State != constants.SuccessState {
		return st, nil
	}

	defer res.Close()

	var makersList []manifest.VehicleMakers

	for res.Next() {
		var maker manifest.VehicleMakers
		ers := res.Scan(&maker.MakerId, &maker.MakerName, &maker.CreationDate)
		if ers != nil {
			return &constants.AnswerState{
				State: constants.ErrorState,
				Data:  "Failed to bind vehicle maker details due to " + ers.Error(),
				Adv:   "none",
			}, nil
		}

		makersList = append(makersList, maker)
	}

	return &constants.AnswerState{
		State: constants.SuccessState,
		Data:  "success",
		Adv:   "none",
	}, makersList
}

func SelectInspectionCheckList(subQuery string, val []any) (st *constants.AnswerState, data []manifest.InspectionCheckListDetails) {
	qr := "SELECT `check_id`, `check_name`, `updated_date`, `registered_date` FROM `inspection_checklist` WHERE " + subQuery

	st, res := gendb.SelectGeneral(qr, val)

	if st.State != constants.SuccessState {
		return st, nil
	}

	defer res.Close()

	var checklist []manifest.InspectionCheckListDetails

	for res.Next() {
		var check manifest.InspectionCheckListDetails
		ers := res.Scan(&check.CheckId, &check.CheckName, &check.UpdatedDate, &check.RegisteredDate)
		if ers != nil {
			return &constants.AnswerState{
				State: constants.ErrorState,
				Data:  "Failed to bind inspection checklist details due to " + ers.Error(),
				Adv:   "none",
			}, nil
		}

		checklist = append(checklist, check)
	}

	return &constants.AnswerState{
		State: constants.SuccessState,
		Data:  "success",
		Adv:   "none",
	}, checklist
}

func SelectVehicleModels(subQuery string, val []any) (st *constants.AnswerState, data []manifest.VehicleModelDetails) {
	qr := "SELECT `model_id`, `model_name`, `status`, `created_time` FROM `vehicle_models` WHERE " + subQuery

	st, res := gendb.SelectGeneral(qr, val)

	if st.State != constants.SuccessState {
		return st, nil
	}

	defer res.Close()

	var modelsList []manifest.VehicleModelDetails

	for res.Next() {
		var model manifest.VehicleModelDetails
		ers := res.Scan(&model.ModelId, &model.ModelName, &model.Status, &model.CreationDate)
		if ers != nil {
			return &constants.AnswerState{
				State: constants.ErrorState,
				Data:  "Failed to bind vehicle model details due to " + ers.Error(),
				Adv:   "none",
			}, nil
		}

		modelsList = append(modelsList, model)
	}

	return &constants.AnswerState{
		State: constants.SuccessState,
		Data:  "success",
		Adv:   "none",
	}, modelsList
}

func SelectPackageTypes(subQuery string, val []any) (st *constants.AnswerState, data []manifest.PackageTypeDetails) {
	qr := "SELECT `type_id`, `type_name`, `status`, `created_time` FROM `package_types` WHERE " + subQuery

	st, res := gendb.SelectGeneral(qr, val)

	if st.State != constants.SuccessState {
		return st, nil
	}

	defer res.Close()

	var typesList []manifest.PackageTypeDetails

	for res.Next() {
		var body manifest.PackageTypeDetails
		ers := res.Scan(&body.TypeId, &body.TypeName, &body.Status, &body.CreatedTime)
		if ers != nil {
			return &constants.AnswerState{
				State: constants.ErrorState,
				Data:  "Failed to bind package type details due to " + ers.Error(),
				Adv:   "none",
			}, nil
		}

		typesList = append(typesList, body)
	}

	return &constants.AnswerState{
		State: constants.SuccessState,
		Data:  "success",
		Adv:   "none",
	}, typesList
}

func SelectPackageStatuses(subQuery string, val []any) (st *constants.AnswerState, data []manifest.PackageInspectionStatusDetails) {
	qr := "SELECT `status_id`, `status_name`, `description`, `status_number` FROM `packages_inspection_status` WHERE " + subQuery
	st, res := gendb.SelectGeneral(qr, val)
	if st.State != constants.SuccessState {
		return st, nil
	}

	defer res.Close()

	var statusList []manifest.PackageInspectionStatusDetails

	for res.Next() {
		var status manifest.PackageInspectionStatusDetails
		ers := res.Scan(&status.StatusId, &status.StatusName, &status.StatusDesc, &status.StatusNum)
		if ers != nil {
			return &constants.AnswerState{
				State: constants.ErrorState,
				Data:  "Failed to bind package inspection status details due to " + ers.Error(),
				Adv:   "none",
			}, nil
		}

		statusList = append(statusList, status)
	}

	return &constants.AnswerState{
		State: constants.SuccessState,
		Data:  "success",
		Adv:   "none",
	}, statusList

}

func SelectDeckStowagePlan(subQuery string, val []any) (st *constants.AnswerState, data []manifest.DeckStowagePlanDetails) {
	qr := "SELECT `deck_id`, `manifest_id`, `deck_name`, `units` FROM `decks_numbers` WHERE " + subQuery

	st, res := gendb.SelectGeneral(qr, val)

	if st.State != constants.SuccessState {
		return st, nil
	}

	defer res.Close()

	var stowagePlan []manifest.DeckStowagePlanDetails

	for res.Next() {
		var plan manifest.DeckStowagePlanDetails
		ers := res.Scan(&plan.DeckId, &plan.ManifestId, &plan.DeckName, &plan.Units)
		if ers != nil {
			return &constants.AnswerState{
				State: constants.ErrorState,
				Data:  "Failed to bind deck stowage plan details due to " + ers.Error(),
				Adv:   "none",
			}, nil
		}

		stowagePlan = append(stowagePlan, plan)
	}

	return &constants.AnswerState{
		State: constants.SuccessState,
		Data:  "success",
		Adv:   "none",
	}, stowagePlan
}

func SelectPackageInfo(subQuery string, val []any) (st *constants.AnswerState, data []manifest.PackageManifestInfo) {
	qr := "SELECT `package_id`, `package_number`, `bl_no`, `manifest_id`, `description`, `is_inspected`, `is_added_later`, `creation_time`, `is_published` FROM `manifest_packages` WHERE " + subQuery

	st, res := gendb.SelectGeneral(qr, val)

	if st.State != constants.SuccessState {
		return st, nil
	}

	defer res.Close()

	var packageList = []manifest.PackageManifestInfo{}

	for res.Next() {
		var pkg manifest.PackageManifestInfo
		ers := res.Scan(&pkg.PackageId, &pkg.PackageNumber, &pkg.BLNumber, &pkg.ManifestId, &pkg.Description, &pkg.IsInspected, &pkg.IsAddedLater, &pkg.CreationTime, &pkg.IsPublished)
		if ers != nil {
			return &constants.AnswerState{
				State: constants.ErrorState,
				Data:  "Failed to bind package info details due to " + ers.Error(),
				Adv:   "none",
			}, nil
		}

		packageList = append(packageList, pkg)
	}

	return &constants.AnswerState{
		State: constants.SuccessState,
		Data:  "success",
		Adv:   "none",
	}, packageList
}

func SelectManifestInfo(subQuery string, val []any) (st *constants.AnswerState, data []manifest.ManifestData) {
	qr := `SELECT 
				manifest_id, 
				manifest_name, 
				manifest.client_id,
				client_name,
				vessel_name, 
				voyage_no,
				berth_no, 
				arrival_date, 
				received_date, 
				uploaded_date
			FROM manifest 
			INNER JOIN clients ON manifest.client_id = clients.client_id
			WHERE ` + subQuery

	st, res := gendb.SelectGeneral(qr, val)

	if st.State != constants.SuccessState {
		return st, nil
	}

	defer res.Close()

	var manifestList []manifest.ManifestData

	for res.Next() {
		var manifest manifest.ManifestData
		ers := res.Scan(&manifest.ManifestId, &manifest.ManifestName, &manifest.ClientId, &manifest.ClientName, &manifest.VesselName,
			&manifest.VoyageNo, &manifest.BerthNo, &manifest.ArrivalDate, &manifest.ReceivedDate, &manifest.UploadedDate,
		)
		if ers != nil {
			return &constants.AnswerState{
				State: constants.ErrorState,
				Data:  "Failed to bind manifest info details due to " + ers.Error(),
				Adv:   "none",
			}, nil
		}

		manifestList = append(manifestList, manifest)
	}

	return &constants.AnswerState{
		State: constants.SuccessState,
		Data:  "success",
		Adv:   "none",
	}, manifestList
}

func SelectVehicleInfo(subQuery string, vals []any) (*constants.AnswerState, []manifest.VehicleData) {
	qr := "SELECT `vehicle_id`, `manifest_id`, `chasis_number`, `model`, `description`, `weight`, `bl_no`, `creation_date`, `inspection_status`, `tallied_status`, `discharged_status`, `is_overland`, `is_added_later`, `inspection_time`, `tallied_time`, `discharge_time`, `is_published` FROM `manifest_vehicles` WHERE " + subQuery

	st, res := gendb.SelectGeneral(qr, vals)

	if st.State != constants.SuccessState {
		return st, nil
	}

	defer res.Close()

	var rows []manifest.VehicleData

	for res.Next() {
		var row manifest.VehicleData

		ers := res.Scan(&row.VehicleId, &row.ManifestId, &row.ChasisNumber, &row.VehicleModel, &row.Description, &row.Weight, &row.BLNumber, &row.CreationDate, &row.InspectionStatus, &row.TalliedStatus, &row.DischargeStatus, &row.OverLandStatus, &row.IsAddedLater, &row.InspectionTime, &row.TalliedTime, &row.DischargeTime, &row.IsPublished)

		if ers != nil {
			slog.Error(ers.Error())
			return &constants.AnswerState{
				State: constants.ErrorState,
				Data:  "Failed to bind checklist results due to " + ers.Error(),
				Adv:   "none",
			}, nil
		}
		rows = append(rows, row)
	}

	return &constants.AnswerState{
		State: constants.SuccessState,
		Data:  "success",
		Adv:   "none",
	}, rows
}

func SelectVehicleHistoryBatchTimes(vehicleId string) (*constants.AnswerState, []time.Time) {
	qr := "SELECT DISTINCT archived_at FROM vehicles_talling_history WHERE vehicle_id = ? ORDER BY archived_at DESC"
	st, rows := gendb.SelectGeneral(qr, []any{vehicleId})
	if st.State != constants.SuccessState {
		return st, nil
	}
	defer rows.Close()

	times := []time.Time{}
	for rows.Next() {
		var at string
		if err := rows.Scan(&at); err != nil {
			slog.Error(err.Error())
			return &constants.AnswerState{State: constants.ErrorState, Data: "Failed to bind history batch times", Adv: "none"}, nil
		}
		parsedTime, err := time.Parse("2006-01-02 15:04:05", at)
		if err != nil {
			slog.Error(err.Error())
			return &constants.AnswerState{State: constants.ErrorState, Data: "Failed to parse history batch times", Adv: "none"}, nil
		}
		times = append(times, parsedTime)
	}

	if err := rows.Err(); err != nil {
		slog.Error(err.Error())
		return &constants.AnswerState{State: constants.ErrorState, Data: "Failed to iterate history batch times", Adv: "none"}, nil
	}

	return &constants.AnswerState{State: constants.SuccessState, Data: "success", Adv: "none"}, times
}

func SelectVehicleTimelineTally(vehicleId string, archivedAt *time.Time) (*constants.AnswerState, []manifest.TallyMoreDetails) {
	table := "vehicles_talling"
	vals := []any{vehicleId}
	where := "t.vehicle_id = ?"
	if archivedAt != nil {
		table = "vehicles_talling_history"
		where = "t.vehicle_id = ? AND t.archived_at = ?"
		vals = append(vals, *archivedAt)
	}

	qr := fmt.Sprintf("SELECT t.tally_id, t.vehicle_id, t.manifest_id, t.maker_id, t.body_id, t.tallied_time, vm.maker_name, vb.body_name, t.image_link, t.deck_number, t.number_of_keys, t.key_type FROM %s t INNER JOIN vehicle_makers vm ON vm.maker_id = t.maker_id INNER JOIN vehicle_bodies vb ON vb.body_id = t.body_id WHERE %s", table, where)

	st, rows := gendb.SelectGeneral(qr, vals)
	if st.State != constants.SuccessState {
		return st, nil
	}
	defer rows.Close()

	result := []manifest.TallyMoreDetails{}
	for rows.Next() {
		var row manifest.TallyMoreDetails
		if err := rows.Scan(&row.TallyId, &row.VehicleId, &row.ManifestId, &row.MakerId, &row.BodyId, &row.TalliedTime, &row.MakerName, &row.BodyName, &row.VehicleImage, &row.DeckNumber, &row.NumberOfKeys, &row.KeyType); err != nil {
			slog.Error(err.Error())
			return &constants.AnswerState{State: constants.ErrorState, Data: "Failed to bind timeline tally data", Adv: "none"}, nil
		}
		result = append(result, row)
	}

	if err := rows.Err(); err != nil {
		slog.Error(err.Error())
		return &constants.AnswerState{State: constants.ErrorState, Data: "Failed to iterate timeline tally data", Adv: "none"}, nil
	}

	return &constants.AnswerState{State: constants.SuccessState, Data: "success", Adv: "none"}, result
}

func SelectVehicleTimelineInspections(vehicleId string, archivedAt *time.Time) (*constants.AnswerState, []manifest.InspectionDetailsAndImage) {
	inspectionTable := "vehicles_inspection"
	imageTable := "inspection_image"
	imageMatch := ""
	vals := []any{vehicleId}
	where := "vi.vehicle_id = ?"
	if archivedAt != nil {
		inspectionTable = "vehicles_inspection_history"
		imageTable = "inspection_image_history"
		imageMatch = " AND ii.archived_at = vi.archived_at"
		where = "vi.vehicle_id = ? AND vi.archived_at = ?"
		vals = append(vals, *archivedAt)
	}

	// The latest image per inspection is read by a correlated subquery so the whole
	// timeline is one statement instead of one extra query per check.
	qr := fmt.Sprintf(`SELECT vi.inspection_id, vi.vehicle_id, vi.check_id, ic.check_name, vi.status, vi.check_time,
		COALESCE((SELECT ii.image_link FROM %s ii WHERE ii.inspection_id = vi.inspection_id%s ORDER BY ii.creation_time DESC LIMIT 1), '')
		FROM %s vi INNER JOIN inspection_checklist ic ON ic.check_id = vi.check_id WHERE %s`, imageTable, imageMatch, inspectionTable, where)

	st, rows := gendb.SelectGeneral(qr, vals)
	if st.State != constants.SuccessState {
		return st, nil
	}
	defer rows.Close()

	result := []manifest.InspectionDetailsAndImage{}
	for rows.Next() {
		var row manifest.InspectionDetailsAndImage
		if err := rows.Scan(&row.InspectionId, &row.VehicleId, &row.CheckId, &row.CheckName, &row.Status, &row.Checktime, &row.Image); err != nil {
			slog.Error(err.Error())
			return &constants.AnswerState{State: constants.ErrorState, Data: "Failed to bind timeline inspection data", Adv: "none"}, nil
		}
		result = append(result, row)
	}

	if err := rows.Err(); err != nil {
		slog.Error(err.Error())
		return &constants.AnswerState{State: constants.ErrorState, Data: "Failed to iterate timeline inspection data", Adv: "none"}, nil
	}

	return &constants.AnswerState{State: constants.SuccessState, Data: "success", Adv: "none"}, result
}

func SelectVehicleTimelineRemarks(vehicleId string, archivedAt *time.Time) (*constants.AnswerState, []manifest.InspectionRemarksDetails) {
	table := "inspection_remarks"
	vals := []any{vehicleId}
	where := "vehicle_id = ?"
	if archivedAt != nil {
		table = "inspection_remarks_history"
		where = "vehicle_id = ? AND archived_at = ?"
		vals = append(vals, *archivedAt)
	}

	qr := fmt.Sprintf("SELECT remark_id, vehicle_id, remark, remark_type, image_link, remark_time FROM %s WHERE %s", table, where)
	st, rows := gendb.SelectGeneral(qr, vals)
	if st.State != constants.SuccessState {
		return st, nil
	}
	defer rows.Close()

	result := []manifest.InspectionRemarksDetails{}
	for rows.Next() {
		var row manifest.InspectionRemarksDetails
		if err := rows.Scan(&row.RemarkId, &row.VehicleId, &row.Remark, &row.RemarkType, &row.ImageLink, &row.RemarkTime); err != nil {
			slog.Error(err.Error())
			return &constants.AnswerState{State: constants.ErrorState, Data: "Failed to bind timeline remarks data", Adv: "none"}, nil
		}
		result = append(result, row)
	}

	if err := rows.Err(); err != nil {
		slog.Error(err.Error())
		return &constants.AnswerState{State: constants.ErrorState, Data: "Failed to iterate timeline remarks data", Adv: "none"}, nil
	}

	return &constants.AnswerState{State: constants.SuccessState, Data: "success", Adv: "none"}, result
}

func SelectVehicleTimelinePackages(vehicleId string, archivedAt *time.Time) (*constants.AnswerState, []manifest.OnboardPackageResp) {
	if archivedAt != nil {
		return selectPackagesWithMedia(
			"SELECT package_id, title, remark, vehicle_id, archived_at FROM onboard_packages_history WHERE vehicle_id = ? AND archived_at = ?",
			"onboard_packages_media_history", " AND m.archived_at = p.archived_at",
			[]any{vehicleId, *archivedAt}, true)
	}
	return selectPackagesWithMedia(
		"SELECT package_id, title, remark, vehicle_id FROM onboard_packages WHERE vehicle_id = ?",
		"onboard_packages_media", "", []any{vehicleId}, true)
}

// selectPackagesWithMedia reads onboard packages and their media in one statement.
// packageQuery selects the packages (derived table p); media rows are LEFT JOINed on
// package_id plus mediaMatch. When emptyMedia is true a package without media gets an
// empty slice instead of nil.
func selectPackagesWithMedia(packageQuery, mediaTable, mediaMatch string, vals []any, emptyMedia bool) (*constants.AnswerState, []manifest.OnboardPackageResp) {
	qr := fmt.Sprintf(`SELECT p.package_id, p.title, p.remark, p.vehicle_id, m.media_id, m.media_type, m.media_link
		FROM (%s) p LEFT JOIN %s m ON m.package_id = p.package_id%s`, packageQuery, mediaTable, mediaMatch)

	st, rows := gendb.SelectGeneral(qr, vals)
	if st.State != constants.SuccessState {
		return st, nil
	}
	defer rows.Close()

	result := []manifest.OnboardPackageResp{}
	index := map[string]int{}
	for rows.Next() {
		var pkg manifest.OnboardPackageResp
		var mediaId, mediaType, mediaLink sql.NullString
		if err := rows.Scan(&pkg.PackageId, &pkg.Title, &pkg.Remark, &pkg.VehicleId, &mediaId, &mediaType, &mediaLink); err != nil {
			slog.Error(err.Error())
			return &constants.AnswerState{State: constants.ErrorState, Data: "Failed to bind on board package details", Adv: "none"}, nil
		}

		pos, seen := index[pkg.PackageId]
		if !seen {
			if emptyMedia {
				pkg.Media = []manifest.OnBoardPackageMediaResp{}
			}
			result = append(result, pkg)
			pos = len(result) - 1
			index[pkg.PackageId] = pos
		}
		if mediaId.Valid {
			result[pos].Media = append(result[pos].Media, manifest.OnBoardPackageMediaResp{
				MediaId:   mediaId.String,
				MediaType: mediaType.String,
				MediaLink: mediaLink.String,
				PackageId: pkg.PackageId,
			})
		}
	}

	if err := rows.Err(); err != nil {
		slog.Error(err.Error())
		return &constants.AnswerState{State: constants.ErrorState, Data: "Failed to iterate on board package details", Adv: "none"}, nil
	}

	return &constants.AnswerState{State: constants.SuccessState, Data: "success", Adv: "none"}, result
}

func SelectVehicleTimelineMedia(vehicleId string, archivedAt *time.Time) (*constants.AnswerState, []manifest.VehicleMediaResp) {
	table := "vehicle_galllery"
	vals := []any{vehicleId}
	where := "vehicle_id = ?"
	if archivedAt != nil {
		table = "vehicle_galllery_history"
		where = "vehicle_id = ? AND archived_at = ?"
		vals = append(vals, *archivedAt)
	}

	qr := fmt.Sprintf("SELECT media_id, media_link, media_type, remark, vehicle_id, status FROM %s WHERE %s", table, where)
	st, rows := gendb.SelectGeneral(qr, vals)
	if st.State != constants.SuccessState {
		return st, nil
	}
	defer rows.Close()

	result := []manifest.VehicleMediaResp{}
	for rows.Next() {
		var row manifest.VehicleMediaResp
		if err := rows.Scan(&row.MediaId, &row.MediaLink, &row.MediaType, &row.Remark, &row.VehicleId, &row.Status); err != nil {
			slog.Error(err.Error())
			return &constants.AnswerState{State: constants.ErrorState, Data: "Failed to bind timeline vehicle media", Adv: "none"}, nil
		}
		result = append(result, row)
	}

	if err := rows.Err(); err != nil {
		slog.Error(err.Error())
		return &constants.AnswerState{State: constants.ErrorState, Data: "Failed to iterate timeline vehicle media", Adv: "none"}, nil
	}

	return &constants.AnswerState{State: constants.SuccessState, Data: "success", Adv: "none"}, result
}

func SelectVehiclesAndInspectionDetails(subQuery string, vals []any) (*constants.AnswerState, []manifest.VehiclesDetailsAndInspection) {
	qr := "SELECT DISTINCT manifest_vehicles.vehicle_id, `manifest_id`, `chasis_number`, `model`, `description`, `weight`, bl_no, `creation_date`, `inspection_status`, `tallied_status`, `discharged_status`, is_overland, `inspection_time`, `tallied_time`, `discharge_time`, vehicles_inspection.status AS insp_status, vehicles_inspection.check_id AS checkId, check_name FROM `manifest_vehicles` INNER JOIN vehicles_inspection ON manifest_vehicles.vehicle_id = vehicles_inspection.vehicle_id INNER JOIN inspection_checklist ON inspection_checklist.check_id = vehicles_inspection.check_id WHERE " + subQuery
	st, res := gendb.SelectGeneral(qr, vals)

	if st.State != constants.SuccessState {
		return st, nil
	}

	defer res.Close()

	var rows []manifest.VehiclesDetailsAndInspection

	for res.Next() {
		var row manifest.VehiclesDetailsAndInspection
		ers := res.Scan(&row.VehicleId, &row.ManifestId, &row.ChasisNumber, &row.VehicleModel, &row.Description, &row.Weight, &row.BLNumber, &row.CreationDate, &row.InspectionStatus, &row.TalliedStatus, &row.DischargeStatus, &row.OverLandStatus, &row.InspectionTime, &row.TalliedTime, &row.DischargeTime, &row.InspectionMark, &row.InspectionCheckId, &row.InspectionCheckName)

		if ers != nil {
			slog.Error(ers.Error())
			return &constants.AnswerState{
				State: constants.ErrorState,
				Data:  "Failed to bind discharge results due to " + ers.Error(),
				Adv:   "none",
			}, nil
		}
		rows = append(rows, row)
	}

	return &constants.AnswerState{
		State: constants.SuccessState,
		Data:  "success",
		Adv:   "none",
	}, rows
}

func SelectInspectionRemarksForGivenManifest(subQuery string, vals []any) (*constants.AnswerState, []manifest.InspectionRemarksDetails) {
	qr := "SELECT `remark_id`, inspection_remarks.vehicle_id, `remark`, `remark_type`, `image_link`, `remark_time` FROM `inspection_remarks` INNER JOIN manifest_vehicles ON manifest_vehicles.vehicle_id = inspection_remarks.vehicle_id WHERE " + subQuery

	st, res := gendb.SelectGeneral(qr, vals)

	if st.State != constants.SuccessState {
		return st, nil
	}

	defer res.Close()

	rows := []manifest.InspectionRemarksDetails{}

	for res.Next() {
		var row manifest.InspectionRemarksDetails

		ers := res.Scan(&row.RemarkId, &row.VehicleId, &row.Remark, &row.RemarkType, &row.ImageLink, &row.RemarkTime)

		if ers != nil {
			slog.Error(ers.Error())
			return &constants.AnswerState{
				State: constants.ErrorState,
				Data:  "Failed to bind inspection remarks results due to " + ers.Error(),
				Adv:   "none",
			}, nil
		}
		rows = append(rows, row)
	}

	return &constants.AnswerState{
		State: constants.SuccessState,
		Data:  "success",
		Adv:   "none",
	}, rows
}

func SelectInspectionDetails(subQuery string, vals []any) (*constants.AnswerState, []manifest.InspectionDetails) {
	qr := "SELECT `inspection_id`, `vehicle_id`, vehicles_inspection.check_id, `check_name`, `status`, `check_time` FROM `vehicles_inspection` INNER JOIN inspection_checklist ON inspection_checklist.check_id = vehicles_inspection.check_id WHERE " + subQuery

	st, res := gendb.SelectGeneral(qr, vals)

	if st.State != constants.SuccessState {
		return st, nil
	}

	defer res.Close()

	var rows []manifest.InspectionDetails

	for res.Next() {
		var row manifest.InspectionDetails

		ers := res.Scan(&row.InspectionId, &row.VehicleId, &row.CheckId, &row.CheckName, &row.Status, &row.Checktime)

		if ers != nil {
			slog.Error(ers.Error())
			return &constants.AnswerState{
				State: constants.ErrorState,
				Data:  "Failed to bind checklist results due to " + ers.Error(),
				Adv:   "none",
			}, nil
		}
		rows = append(rows, row)
	}

	return &constants.AnswerState{
		State: constants.SuccessState,
		Data:  "success",
		Adv:   "none",
	}, rows
}

func SelectTallyDetailsShort(subQuery string, vals []any) (*constants.AnswerState, []manifest.TallyDetails) {
	qr := "SELECT `tally_id`, `vehicle_id`, `manifest_id`, `maker_id`, `body_id`, `image_link`, `deck_number`, `number_of_keys`, `key_type`, `tallied_time` FROM `vehicles_talling` WHERE " + subQuery

	st, res := gendb.SelectGeneral(qr, vals)

	if st.State != constants.SuccessState {
		return st, nil
	}

	defer res.Close()

	var rows []manifest.TallyDetails

	for res.Next() {
		var row manifest.TallyDetails

		ers := res.Scan(&row.TallyId, &row.VehicleId, &row.ManifestId, &row.MakerId, &row.BodyId, &row.VehicleImage, &row.DeckNumber, &row.NumberOfKeys, &row.KeyType, &row.TalliedTime)

		if ers != nil {
			slog.Error(ers.Error())
			return &constants.AnswerState{
				State: constants.ErrorState,
				Data:  "Failed to bind checklist results due to " + ers.Error(),
				Adv:   "none",
			}, nil
		}
		rows = append(rows, row)
	}

	return &constants.AnswerState{
		State: constants.SuccessState,
		Data:  "success",
		Adv:   "none",
	}, rows
}

func SelectTallyDetailsLong(subQuery string, vals []any) (*constants.AnswerState, []manifest.TallyMoreDetails) {
	qr := "SELECT `tally_id`, `vehicle_id`, vehicles_talling.maker_id, vehicles_talling.user_id, vehicles_talling.body_id, `tallied_time`, `maker_name`, `body_name`, `image_link`, `deck_number`, number_of_keys, key_type FROM `vehicles_talling` INNER JOIN vehicle_makers ON vehicle_makers.maker_id = vehicles_talling.maker_id INNER JOIN vehicle_bodies ON vehicle_bodies.body_id = vehicles_talling.body_id WHERE " + subQuery

	st, res := gendb.SelectGeneral(qr, vals)

	if st.State != constants.SuccessState {
		return st, nil
	}

	defer res.Close()

	var rows []manifest.TallyMoreDetails

	for res.Next() {
		var row manifest.TallyMoreDetails

		ers := res.Scan(&row.TallyId, &row.VehicleId, &row.MakerId, &row.UserId, &row.BodyId, &row.TalliedTime, &row.MakerName, &row.BodyName, &row.VehicleImage, &row.DeckNumber, &row.NumberOfKeys, &row.KeyType)

		if ers != nil {
			slog.Error(ers.Error())
			return &constants.AnswerState{
				State: constants.ErrorState,
				Data:  "Failed to bind checklist results due to " + ers.Error(),
				Adv:   "none",
			}, nil
		}
		rows = append(rows, row)
	}

	return &constants.AnswerState{
		State: constants.SuccessState,
		Data:  "success",
		Adv:   "none",
	}, rows
}

func SelectInspectionRemarks(subQuery string, vals []any) (*constants.AnswerState, []manifest.InspectionRemarksDetails) {
	qr := "SELECT `remark_id`, `vehicle_id`, `remark`, `remark_type`, `image_link`, `remark_time` FROM `inspection_remarks` WHERE " + subQuery

	st, res := gendb.SelectGeneral(qr, vals)

	if st.State != constants.SuccessState {
		return st, nil
	}

	defer res.Close()

	rows := []manifest.InspectionRemarksDetails{}

	for res.Next() {
		var row manifest.InspectionRemarksDetails

		ers := res.Scan(&row.RemarkId, &row.VehicleId, &row.Remark, &row.RemarkType, &row.ImageLink, &row.RemarkTime)

		if ers != nil {
			slog.Error(ers.Error())
			return &constants.AnswerState{
				State: constants.ErrorState,
				Data:  "Failed to bind checklist results due to " + ers.Error(),
				Adv:   "none",
			}, nil
		}
		rows = append(rows, row)
	}

	return &constants.AnswerState{
		State: constants.SuccessState,
		Data:  "success",
		Adv:   "none",
	}, rows
}

func SelectVehicleAndTallyDetails(subQuery string, vals []any) (*constants.AnswerState, []manifest.VehiclesDetailsAndTally) {
	qr := "SELECT manifest_vehicles.vehicle_id, manifest_vehicles.manifest_id, vehicles_talling.user_id, `chasis_number`, `model`, `description`, `weight`, `bl_no`, `creation_date`, `inspection_status`, `tallied_status`, `discharged_status`, `is_overland`, `is_added_later`, `inspection_time`, manifest_vehicles.tallied_time, `discharge_time`, `tally_id`, vehicles_talling.maker_id, vehicles_talling.body_id, `maker_name`, `body_name`, `image_link`, `deck_number`, `number_of_keys`, `key_type` FROM `vehicles_talling` INNER JOIN vehicle_makers ON vehicle_makers.maker_id = vehicles_talling.maker_id INNER JOIN vehicle_bodies ON vehicle_bodies.body_id = vehicles_talling.body_id INNER JOIN manifest_vehicles ON manifest_vehicles.vehicle_id = vehicles_talling.vehicle_id WHERE " + subQuery
	st, res := gendb.SelectGeneral(qr, vals)

	if st.State != constants.SuccessState {
		return st, nil
	}

	defer res.Close()

	var rows []manifest.VehiclesDetailsAndTally

	for res.Next() {
		var row manifest.VehiclesDetailsAndTally

		ers := res.Scan(&row.VehicleId, &row.ManifestId, &row.UserId, &row.ChasisNumber, &row.VehicleModel, &row.Description, &row.Weight, &row.BLNumber, &row.CreationDate, &row.InspectionStatus, &row.TalliedStatus, &row.DischargeStatus, &row.OverLandStatus, &row.IsAddedLater, &row.InspectionTime, &row.TalliedTime, &row.DischargeTime, &row.TallyId, &row.MakerId, &row.BodyId, &row.MakerName, &row.BodyName, &row.VehicleImage, &row.DeckNumber, &row.NumberOfKeys, &row.KeyType)

		if ers != nil {
			slog.Error(ers.Error())
			return &constants.AnswerState{
				State: constants.ErrorState,
				Data:  "Failed to bind vehicles tally results due to " + ers.Error(),
				Adv:   "none",
			}, nil
		}
		rows = append(rows, row)
	}

	return &constants.AnswerState{
		State: constants.SuccessState,
		Data:  "success",
		Adv:   "none",
	}, rows
}
func SelectInspectionImages(subQuery string, vals []any) (*constants.AnswerState, []manifest.InspectionImageDetails) {
	qr := "SELECT `image_id`, `image_link`, `inspection_id`, `creation_time` FROM `inspection_image` WHERE " + subQuery

	st, res := gendb.SelectGeneral(qr, vals)

	if st.State != constants.SuccessState {
		return st, nil
	}

	defer res.Close()

	var rows []manifest.InspectionImageDetails

	for res.Next() {
		var row manifest.InspectionImageDetails

		ers := res.Scan(&row.ImageId, &row.ImageLink, &row.InspectionId, &row.CreationTime)

		if ers != nil {
			slog.Error(ers.Error())
			return &constants.AnswerState{
				State: constants.ErrorState,
				Data:  "Failed to bind checklist results due to " + ers.Error(),
				Adv:   "none",
			}, nil
		}
		rows = append(rows, row)
	}

	return &constants.AnswerState{
		State: constants.SuccessState,
		Data:  "success",
		Adv:   "none",
	}, rows
}
func SelectOnBoardPackage(subQuery string, vals []any) (*constants.AnswerState, []manifest.OnboardPackageResp) {
	return selectPackagesWithMedia(
		"SELECT `package_id`, `title`, `remark`, `vehicle_id` FROM `onboard_packages` WHERE "+subQuery,
		"onboard_packages_media", "", vals, false)
}

func SelectVehicleMedia(subQuery string, vals []any) (*constants.AnswerState, []manifest.VehicleMediaResp) {
	qr := "SELECT `media_id`, `media_link`, `media_type`, `remark`, `vehicle_id`, `status` FROM `vehicle_galllery` WHERE " + subQuery
	st, res := gendb.SelectGeneral(qr, vals)

	if st.State != constants.SuccessState {
		return st, nil
	}

	defer res.Close()

	var rows []manifest.VehicleMediaResp

	for res.Next() {
		var row manifest.VehicleMediaResp
		ers := res.Scan(&row.MediaId, &row.MediaLink, &row.MediaType, &row.Remark, &row.VehicleId, &row.Status)

		if ers != nil {
			slog.Error(ers.Error())
			return &constants.AnswerState{
				State: constants.ErrorState,
				Data:  "Failed to bind vehicle media results due to " + ers.Error(),
				Adv:   "none",
			}, nil
		}
		rows = append(rows, row)
	}

	return &constants.AnswerState{
		State: constants.SuccessState,
		Data:  "success",
		Adv:   "none",
	}, rows
}

func SelectDischargeDetails(subQuery string, vals []any) (*constants.AnswerState, []manifest.VehicleDischargeDetails) {
	qr := `SELECT 
				discharge_id, 
				vehicle_id, 
				discharge_image, 
				users.user_id,
				users.fname,
				users.lname,
				users.phone,
				drivers.driver_id,
				drivers.first_name,
				drivers.last_name,
				drivers.phone,
				discharge_time
			FROM 
				vehicle_discharge_tally
			INNER JOIN users ON users.user_id = vehicle_discharge_tally.user_id
			INNER JOIN drivers ON drivers.driver_id = vehicle_discharge_tally.driver_id
			WHERE ` + subQuery

	st, res := gendb.SelectGeneral(qr, vals)

	if st.State != constants.SuccessState {
		return st, nil
	}

	defer res.Close()

	rows := []manifest.VehicleDischargeDetails{}

	for res.Next() {
		var row manifest.VehicleDischargeDetails
		ers := res.Scan(&row.DischargeId, &row.VehicleId, &row.DischargeImage, &row.UserId, &row.UserFname, &row.UserLname, &row.UserPhone, &row.DriverId, &row.DriverFname, &row.DriverLname, &row.DriverPhone, &row.DischargeTime)

		if ers != nil {
			slog.Error(ers.Error())
			return &constants.AnswerState{
				State: constants.ErrorState,
				Data:  "Failed to bind discharge results due to " + ers.Error(),
				Adv:   "none",
			}, nil
		}
		rows = append(rows, row)
	}

	return &constants.AnswerState{
		State: constants.SuccessState,
		Data:  "success",
		Adv:   "none",
	}, rows
}

// SelectVehicleCountByMaker returns a count of vehicles grouped by maker
func SelectVehicleCountByMaker(manifestId string) (*constants.AnswerState, []manifest.MakerSummary) {
	var values []any
	var whereClause string

	if manifestId != "" {
		whereClause = "WHERE mv.manifest_id = ?"
		values = append(values, manifestId)
	} else {
		whereClause = "WHERE 1" // all records
	}

	qr := `SELECT 
        vm.maker_name, 
        COUNT(*) as count
    FROM 
        vehicles_talling vt
    INNER JOIN 
        vehicle_makers vm ON vm.maker_id = vt.maker_id
	INNER JOIN manifest_vehicles mv ON mv.vehicle_id = vt.vehicle_id
    ` + whereClause + `
    GROUP BY 
        vm.maker_name
    ORDER BY 
        count DESC`

	st, res := gendb.SelectGeneral(qr, values)

	if st.State != constants.SuccessState {
		return st, nil
	}

	defer res.Close()

	rows := []manifest.MakerSummary{}

	for res.Next() {
		var row manifest.MakerSummary
		ers := res.Scan(&row.MakerName, &row.Count)

		if ers != nil {
			slog.Error(ers.Error())
			return &constants.AnswerState{
				State: constants.ErrorState,
				Data:  "Failed to bind vehicle maker summary results: " + ers.Error(),
				Adv:   "none",
			}, nil
		}
		rows = append(rows, row)
	}

	return &constants.AnswerState{
		State: constants.SuccessState,
		Data:  "success",
		Adv:   "none",
	}, rows
}

// SelectVehicleCountByBodyType returns a count of vehicles grouped by body type
func SelectVehicleCountByBodyType(manifestId string) (*constants.AnswerState, []manifest.BodySummary) {
	var values []any
	var whereClause string

	if manifestId != "" {
		whereClause = "WHERE mv.manifest_id = ?"
		values = append(values, manifestId)
	} else {
		whereClause = "WHERE 1" // all records
	}

	qr := `SELECT 
        vb.body_name, 
        COUNT(*) as count
    FROM 
        vehicles_talling vt
    INNER JOIN 
        vehicle_bodies vb ON vb.body_id = vt.body_id 
	INNER JOIN manifest_vehicles mv ON mv.vehicle_id = vt.vehicle_id
    ` + whereClause + ` 
    GROUP BY 
        vb.body_name
    ORDER BY 
        count DESC`

	st, res := gendb.SelectGeneral(qr, values)

	if st.State != constants.SuccessState {
		return st, nil
	}

	defer res.Close()

	rows := []manifest.BodySummary{}

	for res.Next() {
		var row manifest.BodySummary
		ers := res.Scan(&row.BodyName, &row.Count)

		if ers != nil {
			slog.Error(ers.Error())
			return &constants.AnswerState{
				State: constants.ErrorState,
				Data:  "Failed to bind vehicle body summary results: " + ers.Error(),
				Adv:   "none",
			}, nil
		}
		rows = append(rows, row)
	}

	return &constants.AnswerState{
		State: constants.SuccessState,
		Data:  "success",
		Adv:   "none",
	}, rows
}

func SelectOverLandVehicleCount(manifestId string) (*constants.AnswerState, *int) {
	qr := "SELECT COUNT(vehicle_id) AS overLand FROM manifest_vehicles WHERE is_overland = 'yes' AND manifest_id = ?"
	vals := []any{manifestId}
	st, res := gendb.SelectGeneral(qr, vals)
	if st.State != constants.SuccessState {
		return st, nil
	}

	defer res.Close()

	var count int
	if res.Next() {
		err := res.Scan(&count)
		if err != nil {
			slog.Error(err.Error())
			return &constants.AnswerState{
				State: constants.ErrorState,
				Data:  "Failed to bind vehicle overland count results: " + err.Error(),
				Adv:   "none",
			}, nil
		}
	}

	return &constants.AnswerState{
		State: constants.SuccessState,
		Data:  "success",
		Adv:   "none",
	}, &count
}

func SelectTalliedAndInspectedVehicleCount(manifestId string) (*constants.AnswerState, *int) {
	qr := "SELECT COUNT(vehicle_id) AS talliedAndInspected FROM manifest_vehicles WHERE tallied_status = 'yes' AND inspection_status = 'yes' AND manifest_id = ?"
	vals := []any{manifestId}
	st, res := gendb.SelectGeneral(qr, vals)
	if st.State != constants.SuccessState {
		return st, nil
	}

	defer res.Close()

	var count int
	if res.Next() {
		err := res.Scan(&count)
		if err != nil {
			slog.Error(err.Error())
			return &constants.AnswerState{
				State: constants.ErrorState,
				Data:  "Failed to bind vehicle overland count results: " + err.Error(),
				Adv:   "none",
			}, nil
		}
	}

	return &constants.AnswerState{
		State: constants.SuccessState,
		Data:  "success",
		Adv:   "none",
	}, &count

}

func SelectDishargedVehicleCount(manifestId string) (*constants.AnswerState, *int) {
	qr := "SELECT COUNT(vehicle_id) AS discharged FROM manifest_vehicles WHERE discharged_status = 'yes' AND manifest_id = ?"
	vals := []any{manifestId}
	st, res := gendb.SelectGeneral(qr, vals)
	if st.State != constants.SuccessState {
		return st, nil
	}

	defer res.Close()

	var count int
	if res.Next() {
		err := res.Scan(&count)
		if err != nil {
			slog.Error(err.Error())
			return &constants.AnswerState{
				State: constants.ErrorState,
				Data:  "Failed to bind vehicle overland count results: " + err.Error(),
				Adv:   "none",
			}, nil
		}
	}

	return &constants.AnswerState{
		State: constants.SuccessState,
		Data:  "success",
		Adv:   "none",
	}, &count

}

// GetCompleteSummary returns both maker and body type summaries in a single call
func GetCompleteSummary(manifestId string) (*constants.AnswerState, manifest.VehicleSummary) {
	// Get maker summary
	stMaker, makers := SelectVehicleCountByMaker(manifestId)
	if stMaker.State != constants.SuccessState {
		return stMaker, manifest.VehicleSummary{}
	}

	// Get body type summary
	stBody, bodies := SelectVehicleCountByBodyType(manifestId)
	if stBody.State != constants.SuccessState {
		return stBody, manifest.VehicleSummary{}
	}
	// get overland count
	stOverland, overlandCount := SelectOverLandVehicleCount(manifestId)
	if stOverland.State != constants.SuccessState {
		return stOverland, manifest.VehicleSummary{}
	}

	// get tallied and inspected count
	stTallied, talliedCount := SelectTalliedAndInspectedVehicleCount(manifestId)
	if stTallied.State != constants.SuccessState {
		return stTallied, manifest.VehicleSummary{}
	}
	// get discharged count
	stDischarged, dischargedCount := SelectDishargedVehicleCount(manifestId)
	if stDischarged.State != constants.SuccessState {
		return stDischarged, manifest.VehicleSummary{}
	}

	// Get total count
	var totalCount int
	qr := "SELECT COUNT(*) FROM vehicles_talling"
	var values []any

	if manifestId != "" {
		qr += " WHERE manifest_id = ?"
		values = append(values, manifestId)
	}

	st, res := gendb.SelectGeneral(qr, values)
	if st.State == constants.SuccessState {
		if res.Next() {
			res.Scan(&totalCount)
		}
		res.Close()
	}

	return &constants.AnswerState{
			State: constants.SuccessState,
			Data:  "success",
			Adv:   "none",
		}, manifest.VehicleSummary{
			Makers:     makers,
			BodyTypes:  bodies,
			TotalCount: totalCount,
			OverLand:   *overlandCount,
			Inspected:  *talliedCount,
			Discharged: *dischargedCount,
		}
}

func SelectPackageInspectionData(subQuery string, vals []any) (*constants.AnswerState, []manifest.PackageInspectionDetails) {
	qr := `SELECT 
			packages_inspection.inspection_id, 
			packages_inspection.package_id, 
			package_types.type_id, 
			package_types.type_name,
			picture,
			packages_inspection_status.status_id,
			packages_inspection_status.status_name,
			packages_inspection_status.description,
			packages_inspection_status.status_number,
			packages_inspection.inspection_time, 
			packages_inspection.creation_time,
			manifest_packages.manifest_id,
			manifest_packages.bl_no,
			manifest_packages.package_number,
			manifest_packages.description,
			manifest_packages.is_inspected,
			users.user_id,
            users.fname,
            users.lname,
            users.phone,
			manifest_packages.is_added_later,
			package_gallery.media_id,
			package_gallery.media_link,
			package_gallery.media_type,
			package_gallery.remark,
			package_gallery.status
		FROM 
			packages_inspection
		INNER JOIN manifest_packages ON manifest_packages.package_id = packages_inspection.package_id
		INNER JOIN package_types ON package_types.type_id = packages_inspection.type_id
		INNER JOIN users ON users.user_id = packages_inspection.user_id
		INNER JOIN packages_inspection_status ON packages_inspection_status.status_id = packages_inspection.inspection_status
		LEFT JOIN package_gallery ON package_gallery.package_id = packages_inspection.package_id
		WHERE ` + subQuery

	st, res := gendb.SelectGeneral(qr, vals)

	if st.State != constants.SuccessState {
		return st, nil
	}

	defer res.Close()

	packageMap := make(map[string]*manifest.PackageInspectionDetails)
	var orderedKeys []string

	for res.Next() {
		var row manifest.PackageInspectionDetails
		var mediaId, mediaLink, mediaType, mediaRemark, mediaStatus sql.NullString

		ers := res.Scan(&row.InspectionId, &row.PackageId, &row.TypeId, &row.TypeName, &row.Picture, &row.InspectionStatusId, &row.InspectionStatus, &row.InspectionDescription,
			&row.InspectionNumber, &row.InspectionTime, &row.CreationTime, &row.ManifestId, &row.BLNumber, &row.PackageNumber, &row.Description, &row.IsInspected,
			&row.UserId, &row.UserFname, &row.UserLname, &row.UserPhone, &row.IsAddedLater,
			&mediaId, &mediaLink, &mediaType, &mediaRemark, &mediaStatus,
		)

		if ers != nil {
			slog.Error(ers.Error())
			return &constants.AnswerState{
				State: constants.ErrorState,
				Data:  "Failed to bind package inspection data results due to " + ers.Error(),
				Adv:   "none",
			}, nil
		}

		if pkg, exists := packageMap[row.PackageId]; exists {
			if mediaId.Valid {
				pkg.Media = append(pkg.Media, manifest.PackageMediaDetails{
					MediaId:   mediaId.String,
					PackageId: pkg.PackageId,
					MediaLink: mediaLink.String,
					MediaType: mediaType.String,
					Remark:    mediaRemark.String,
					Status:    mediaStatus.String,
				})
			}
		} else {
			if mediaId.Valid {
				row.Media = []manifest.PackageMediaDetails{
					{
						MediaId:   mediaId.String,
						PackageId: row.PackageId,
						MediaLink: mediaLink.String,
						MediaType: mediaType.String,
						Remark:    mediaRemark.String,
						Status:    mediaStatus.String,
					},
				}
			} else {
				row.Media = []manifest.PackageMediaDetails{}
			}
			packageMap[row.PackageId] = &row
			orderedKeys = append(orderedKeys, row.PackageId)
		}
	}

	rows := make([]manifest.PackageInspectionDetails, 0, len(orderedKeys))
	for _, key := range orderedKeys {
		rows = append(rows, *packageMap[key])
	}

	return &constants.AnswerState{
		State: constants.SuccessState,
		Data:  "success",
		Adv:   "none",
	}, rows

}

// VehicleListExtras holds the per-vehicle details that list endpoints show next to
// the manifest_vehicles row.
type VehicleListExtras struct {
	Maker          string
	BodyType       string
	DeckNumber     string
	Image          string
	NumberOfKeys   int
	InspectedBy    string
	IsDamaged      bool
	HistoryBatches int
	HasActiveTally bool
}

// SelectVehicleListExtras loads tally, inspector, damage and history details for many
// vehicles with three queries in total, instead of several queries per vehicle.
func SelectVehicleListExtras(vehicleIds []string) (*constants.AnswerState, map[string]*VehicleListExtras) {
	extras := make(map[string]*VehicleListExtras, len(vehicleIds))
	if len(vehicleIds) == 0 {
		return &constants.AnswerState{State: constants.SuccessState, Data: "success", Adv: "none"}, extras
	}

	ids := make([]any, len(vehicleIds))
	for i, id := range vehicleIds {
		ids[i] = id
		extras[id] = &VehicleListExtras{}
	}
	in := strings.TrimSuffix(strings.Repeat("?,", len(ids)), ",")

	// 1. Tally details with maker, body and inspector names.
	qr := `SELECT vt.vehicle_id, vm.maker_name, vb.body_name, vt.deck_number, vt.image_link, vt.number_of_keys,
		COALESCE(CONCAT(u.fname, ' ', u.lname), '')
		FROM vehicles_talling vt
		INNER JOIN vehicle_makers vm ON vm.maker_id = vt.maker_id
		INNER JOIN vehicle_bodies vb ON vb.body_id = vt.body_id
		LEFT JOIN users u ON u.user_id = vt.user_id
		WHERE vt.vehicle_id IN (` + in + `)`
	st, rows := gendb.SelectGeneral(qr, ids)
	if st.State != constants.SuccessState {
		return st, extras
	}
	for rows.Next() {
		var id string
		var e VehicleListExtras
		if err := rows.Scan(&id, &e.Maker, &e.BodyType, &e.DeckNumber, &e.Image, &e.NumberOfKeys, &e.InspectedBy); err != nil {
			rows.Close()
			slog.Error(err.Error())
			return &constants.AnswerState{State: constants.ErrorState, Data: "Failed to bind vehicle tally details", Adv: "none"}, extras
		}
		e.HasActiveTally = true
		extras[id] = &e
	}
	rows.Close()

	// 2. Damaged vehicles: a damaged or missing check, or a damage remark.
	damageVals := append(append([]any{}, ids...), manifest.InspectionMarkStatus.Damaged, manifest.InspectionMarkStatus.Missing)
	damageVals = append(append(damageVals, ids...), manifest.RemarkStatus.Damaged)
	qr = `SELECT DISTINCT vehicle_id FROM vehicles_inspection WHERE vehicle_id IN (` + in + `) AND (status = ? OR status = ?)
		UNION
		SELECT DISTINCT vehicle_id FROM inspection_remarks WHERE vehicle_id IN (` + in + `) AND remark_type = ?`
	st, rows = gendb.SelectGeneral(qr, damageVals)
	if st.State != constants.SuccessState {
		return st, extras
	}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			slog.Error(err.Error())
			return &constants.AnswerState{State: constants.ErrorState, Data: "Failed to bind vehicle damage details", Adv: "none"}, extras
		}
		if e, ok := extras[id]; ok {
			e.IsDamaged = true
		}
	}
	rows.Close()

	// 3. Number of archived inspection batches per vehicle.
	qr = `SELECT vehicle_id, COUNT(DISTINCT archived_at) FROM vehicles_talling_history WHERE vehicle_id IN (` + in + `) GROUP BY vehicle_id`
	st, rows = gendb.SelectGeneral(qr, ids)
	if st.State != constants.SuccessState {
		return st, extras
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		var batches int
		if err := rows.Scan(&id, &batches); err != nil {
			slog.Error(err.Error())
			return &constants.AnswerState{State: constants.ErrorState, Data: "Failed to bind vehicle history details", Adv: "none"}, extras
		}
		if e, ok := extras[id]; ok {
			e.HistoryBatches = batches
		}
	}

	return &constants.AnswerState{State: constants.SuccessState, Data: "success", Adv: "none"}, extras
}

// InspectionCheckForPublish is one check result with its fault image, as sent to the
// remote server.
type InspectionCheckForPublish struct {
	CheckId    string
	Status     string
	FaultImage string
}

// SelectInspectionChecksForPublish reads all check results of a vehicle with their
// fault image in one query.
func SelectInspectionChecksForPublish(vehicleId string) (*constants.AnswerState, []InspectionCheckForPublish) {
	qr := `SELECT vi.check_id, vi.status,
		COALESCE((SELECT ii.image_link FROM inspection_image ii WHERE ii.inspection_id = vi.inspection_id ORDER BY ii.creation_time DESC LIMIT 1), '')
		FROM vehicles_inspection vi WHERE vi.vehicle_id = ?`
	st, rows := gendb.SelectGeneral(qr, []any{vehicleId})
	if st.State != constants.SuccessState {
		return st, nil
	}
	defer rows.Close()

	result := []InspectionCheckForPublish{}
	for rows.Next() {
		var row InspectionCheckForPublish
		if err := rows.Scan(&row.CheckId, &row.Status, &row.FaultImage); err != nil {
			slog.Error(err.Error())
			return &constants.AnswerState{State: constants.ErrorState, Data: "Failed to bind inspection checks", Adv: "none"}, nil
		}
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		slog.Error(err.Error())
		return &constants.AnswerState{State: constants.ErrorState, Data: "Failed to iterate inspection checks", Adv: "none"}, nil
	}

	return &constants.AnswerState{State: constants.SuccessState, Data: "success", Adv: "none"}, result
}

// VehicleStatusRow is the compact inspection and publish status of one vehicle, as
// the tablets keep it in their local list.
type VehicleStatusRow struct {
	VehicleId        string `json:"vehicleId"`
	ChasisNumber     string `json:"chasisNumber"`
	BLNumber         string `json:"blNumber"`
	InspectionStatus string `json:"inspectionStatus"`
	IsInspected      bool   `json:"isInspected"`
	InspectionTime   string `json:"inspectionTime"`
	InspectedBy      string `json:"inspectedBy"`
	IsPublished      string `json:"isPublished"`
	HasHistory       bool   `json:"hasHistory"`
	UpdatedAt        string `json:"updatedAt"`
}

// SelectVehicleStatusChanges returns the vehicles of a manifest that changed at or
// after since, or all of them when since is empty or manifest_vehicles has no
// updated_at column yet (full is then true). cursor is the value the tablet sends as
// since on its next call; it lies a few seconds in the past so that a save that was
// still committing is not missed. Rows can therefore repeat; tablets merge them by
// vehicleId.
func SelectVehicleStatusChanges(manifestId, since string) (st *constants.AnswerState, rows []VehicleStatusRow, cursor string, full bool) {
	db, err := gendb.InitDb()
	if err != nil {
		return &constants.AnswerState{State: constants.ErrorState, Data: err.Error(), Adv: "none"}, nil, "", false
	}

	hasUpdatedAt := gendb.ColumnExists("manifest_vehicles", "updated_at")
	if err := db.QueryRow("SELECT DATE_FORMAT(NOW(3) - INTERVAL 10 SECOND, '%Y-%m-%d %H:%i:%s.%f')").Scan(&cursor); err != nil {
		slog.Error(err.Error())
		return &constants.AnswerState{State: constants.ErrorState, Data: "Failed to read server time", Adv: "none"}, nil, "", false
	}

	updatedAt := "''"
	if hasUpdatedAt {
		updatedAt = "DATE_FORMAT(mv.updated_at, '%Y-%m-%d %H:%i:%s.%f')"
	}
	qr := `SELECT mv.vehicle_id, mv.chasis_number, mv.bl_no, mv.inspection_status, mv.inspection_time, mv.is_published,
		COALESCE(CONCAT(u.fname, ' ', u.lname), ''),
		EXISTS (SELECT 1 FROM vehicles_talling_history h WHERE h.vehicle_id = mv.vehicle_id),
		` + updatedAt + `
		FROM manifest_vehicles mv
		LEFT JOIN vehicles_talling vt ON vt.vehicle_id = mv.vehicle_id
		LEFT JOIN users u ON u.user_id = vt.user_id
		WHERE mv.manifest_id = ?`
	vals := []any{manifestId}
	full = since == "" || !hasUpdatedAt
	if !full {
		qr += " AND mv.updated_at >= ?"
		vals = append(vals, since)
	}

	stq, res := gendb.SelectGeneral(qr, vals)
	if stq.State != constants.SuccessState {
		return stq, nil, "", false
	}
	defer res.Close()

	rows = []VehicleStatusRow{}
	for res.Next() {
		var row VehicleStatusRow
		if err := res.Scan(&row.VehicleId, &row.ChasisNumber, &row.BLNumber, &row.InspectionStatus, &row.InspectionTime,
			&row.IsPublished, &row.InspectedBy, &row.HasHistory, &row.UpdatedAt); err != nil {
			slog.Error(err.Error())
			return &constants.AnswerState{State: constants.ErrorState, Data: "Failed to bind vehicle status", Adv: "none"}, nil, "", false
		}
		row.IsInspected = row.InspectionStatus == manifest.InspectionStatus.Yes
		rows = append(rows, row)
	}
	if err := res.Err(); err != nil {
		slog.Error(err.Error())
		return &constants.AnswerState{State: constants.ErrorState, Data: "Failed to iterate vehicle status", Adv: "none"}, nil, "", false
	}

	return &constants.AnswerState{State: constants.SuccessState, Data: "success", Adv: "none"}, rows, cursor, full
}
