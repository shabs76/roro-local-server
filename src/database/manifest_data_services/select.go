package manifestdataservices

import (
	"log/slog"

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
	qr := "SELECT `package_id`, `title`, `remark`, `vehicle_id` FROM `onboard_packages` WHERE " + subQuery

	st, res := gendb.SelectGeneral(qr, vals)
	if st.State != constants.SuccessState {
		return st, nil
	}

	rows := []manifest.OnboardPackageResp{}

	defer res.Close()

	for res.Next() {
		var row manifest.OnboardPackageResp
		ers := res.Scan(&row.PackageId, &row.Title, &row.Remark, &row.VehicleId)
		if ers != nil {
			slog.Error(ers.Error())
			return &constants.AnswerState{
				State: constants.ErrorState,
				Data:  "Failed to binding main on board package details",
				Adv:   "none",
			}, nil
		}
		// media
		qri := "SELECT `media_id`, `media_type`, `media_link`, `package_id` FROM `onboard_packages_media` WHERE package_id = ?"
		st, resm := gendb.SelectGeneral(qri, []any{row.PackageId})
		if st.State != constants.SuccessState {
			return st, nil
		}

		for resm.Next() {
			var rowM manifest.OnBoardPackageMediaResp
			ers := resm.Scan(&rowM.MediaId, &rowM.MediaType, &rowM.MediaLink, &rowM.PackageId)
			if ers != nil {
				slog.Error(ers.Error())
				return &constants.AnswerState{
					State: constants.ErrorState,
					Data:  "Failed to binding media on board package details",
					Adv:   "none",
				}, nil
			}
			row.Media = append(row.Media, rowM)
		}
		rows = append(rows, row)
	}

	return &constants.AnswerState{
		State: constants.SuccessState,
		Data:  "success",
		Adv:   "none",
	}, rows

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

	var rows []manifest.VehicleDischargeDetails

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
			inspection_id, 
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
			manifest_packages.is_added_later
		FROM 
			packages_inspection
		INNER JOIN manifest_packages ON manifest_packages.package_id = packages_inspection.package_id
		INNER JOIN package_types ON package_types.type_id = packages_inspection.type_id
		INNER JOIN users ON users.user_id = packages_inspection.user_id
		INNER JOIN packages_inspection_status ON packages_inspection_status.status_id = packages_inspection.inspection_status
		WHERE ` + subQuery

	st, res := gendb.SelectGeneral(qr, vals)

	if st.State != constants.SuccessState {
		return st, nil
	}

	defer res.Close()

	rows := []manifest.PackageInspectionDetails{}

	for res.Next() {
		var row manifest.PackageInspectionDetails

		ers := res.Scan(&row.InspectionId, &row.PackageId, &row.TypeId, &row.TypeName, &row.Picture, &row.InspectionStatusId, &row.InspectionStatus, &row.InspectionDescription,
			&row.InspectionNumber, &row.InspectionTime, &row.CreationTime, &row.ManifestId, &row.BLNumber, &row.PackageNumber, &row.Description, &row.IsInspected,
			&row.UserId, &row.UserFname, &row.UserLname, &row.UserPhone, &row.IsAddedLater,
		)

		if ers != nil {
			slog.Error(ers.Error())
			return &constants.AnswerState{
				State: constants.ErrorState,
				Data:  "Failed to bind package inspection data results due to " + ers.Error(),
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
