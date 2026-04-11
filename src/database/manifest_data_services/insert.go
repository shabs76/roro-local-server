package manifestdataservices

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/shabs76/roro-local-server/constants"
	"github.com/shabs76/roro-local-server/constants/modules/manifest"
	"github.com/shabs76/roro-local-server/gendb"
	"github.com/shabs76/roro-local-server/specials"
)

func InsertVehicleBodyTypes(types []manifest.VehicleBodyTypes) (st *constants.AnswerState) {
	qr := "INSERT INTO `vehicle_bodies`(`body_id`, `body_name`, `creation_time`) VALUES (?,?,?) ON DUPLICATE KEY UPDATE `body_name`=VALUES(`body_name`), `creation_time`=VALUES(`creation_time`)"
	for _, body := range types {
		vals := []any{
			body.BodyId,
			body.BodyName,
			body.CreationDate,
		}
		stx := gendb.SaveGeneral(qr, vals)
		if stx.State != constants.SuccessState {
			return stx
		}
	}

	return &constants.AnswerState{State: constants.SuccessState, Data: "Vehicle body types were successfully synced", Adv: "none"}
}

func InsertVehicleMakers(makers []manifest.VehicleMakers) (st *constants.AnswerState) {
	qr := "INSERT INTO `vehicle_makers`(`maker_id`, `maker_name`, `creation_time`) VALUES (?,?,?) ON DUPLICATE KEY UPDATE `maker_name`=VALUES(`maker_name`), `creation_time`=VALUES(`creation_time`)"
	for _, maker := range makers {
		vals := []any{
			maker.MakerId,
			maker.MakerName,
			maker.CreationDate,
		}
		stx := gendb.SaveGeneral(qr, vals)
		if stx.State != constants.SuccessState {
			return stx
		}
	}

	return &constants.AnswerState{State: constants.SuccessState, Data: "Vehicle makers were successfully synced", Adv: "none"}
}

func InsertVehicleInspectCheckList(data []manifest.InspectionCheckListDetails) (st *constants.AnswerState) {
	qr := "INSERT INTO `inspection_checklist`(`check_id`, `check_name`, `updated_date`, `registered_date`) VALUES (?,?,?,?) ON DUPLICATE KEY UPDATE `check_name`=VALUES(`check_name`), `updated_date`=VALUES(`updated_date`), `registered_date`=VALUES(`registered_date`)"
	for _, check := range data {
		vals := []any{
			check.CheckId,
			check.CheckName,
			check.UpdatedDate,
			check.RegisteredDate,
		}
		stx := gendb.SaveGeneral(qr, vals)
		if stx.State != constants.SuccessState {
			return stx
		}
	}

	return &constants.AnswerState{State: constants.SuccessState, Data: "Vehicle inspect checklist was successfully synced", Adv: "none"}
}

func InsertVehicleModels(data []manifest.VehicleModelDetails) (st *constants.AnswerState) {
	qr := "INSERT INTO `vehicle_models`(`model_id`, `model_name`, `status`, `created_time`) VALUES (?,?,?,?) ON DUPLICATE KEY UPDATE `model_name`=VALUES(`model_name`), `status`=VALUES(`status`), `created_time`=VALUES(`created_time`)"
	for _, model := range data {
		vals := []any{
			model.ModelId,
			model.ModelName,
			model.Status,
			model.CreationDate,
		}
		stx := gendb.SaveGeneral(qr, vals)
		if stx.State != constants.SuccessState {
			return stx
		}
	}

	return &constants.AnswerState{State: constants.SuccessState, Data: "Vehicle models were successfully synced", Adv: "none"}
}

func InsertPackageTypes(data []manifest.PackageTypeDetails) (st *constants.AnswerState) {
	qr := "INSERT INTO `package_types`(`type_id`, `type_name`, `status`, `created_time`) VALUES (?,?,?,?) ON DUPLICATE KEY UPDATE `type_name`=VALUES(`type_name`), `status`=VALUES(`status`), `created_time`=VALUES(`created_time`)"
	for _, pkgType := range data {
		vals := []any{
			pkgType.TypeId,
			pkgType.TypeName,
			pkgType.Status,
			pkgType.CreatedTime,
		}
		stx := gendb.SaveGeneral(qr, vals)
		if stx.State != constants.SuccessState {
			return stx
		}
	}

	return &constants.AnswerState{State: constants.SuccessState, Data: "Package types were successfully synced", Adv: "none"}
}

func InsertPackageInspectionStatus(data []manifest.PackageInspectionStatusDetails) (st *constants.AnswerState) {
	qr := "INSERT INTO `packages_inspection_status`(`status_id`, `status_name`, `description`, `status_number`) VALUES (?,?,?,?) ON DUPLICATE KEY UPDATE `status_name`=VALUES(`status_name`), `description`=VALUES(`description`), `status_number`=VALUES(`status_number`)"
	for _, status := range data {
		vals := []any{
			status.StatusId,
			status.StatusName,
			status.StatusDesc,
			status.StatusNum,
		}
		stx := gendb.SaveGeneral(qr, vals)
		if stx.State != constants.SuccessState {
			return stx
		}
	}

	return &constants.AnswerState{State: constants.SuccessState, Data: "Package inspection statuses were successfully synced", Adv: "none"}
}

// manifest data insertion functions
func InsertManifestDetails(data []manifest.ManifestData) (st *constants.AnswerState) {
	log.Println(data[0])
	qr := "INSERT INTO `manifest`(`manifest_id`, `manifest_name`, `client_id`, `vessel_name`, `voyage_no`, `berth_no`, `arrival_date`, `received_date`, `uploaded_date`) VALUES (?,?,?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE `manifest_name`=VALUES(`manifest_name`), `client_id`=VALUES(`client_id`), `vessel_name`=VALUES(`vessel_name`), `voyage_no`=VALUES(`voyage_no`), `berth_no`=VALUES(`berth_no`), `arrival_date`=VALUES(`arrival_date`), `received_date`=VALUES(`received_date`), `uploaded_date`=VALUES(`uploaded_date`)"
	for i, detail := range data {
		log.Printf("Processing manifest detail %d/%d: Manifest ID %s\n", i+1, len(data), detail.ManifestId)
		vals := []any{
			detail.ManifestId,
			detail.ManifestName,
			detail.ClientId,
			detail.VesselName,
			detail.VoyageNo,
			detail.BerthNo,
			detail.ArrivalDate,
			detail.ReceivedDate,
			detail.UploadedDate,
		}
		stx := gendb.SaveGeneral(qr, vals)
		if stx.State != constants.SuccessState {
			return stx
		}
		log.Printf("Successfully inserted manifest detail %d/%d: Manifest ID %s\n", i+1, len(data), detail.ManifestId)
	}

	return &constants.AnswerState{State: constants.SuccessState, Data: "Manifest details were successfully synced", Adv: "none"}
}

func InsertDeckStowagePlanDetails(data []manifest.DeckStowagePlanDetails) (st *constants.AnswerState) {
	qr := "INSERT INTO `decks_numbers`(`deck_id`, `manifest_id`, `deck_name`, `units`) VALUES (?,?,?,?) ON DUPLICATE KEY UPDATE `manifest_id`=VALUES(`manifest_id`), `deck_name`=VALUES(`deck_name`), `units`=VALUES(`units`)"
	for _, detail := range data {
		vals := []any{
			detail.DeckId,
			detail.ManifestId,
			detail.DeckName,
			detail.Units,
		}
		stx := gendb.SaveGeneral(qr, vals)
		if stx.State != constants.SuccessState {
			return stx
		}
	}

	return &constants.AnswerState{State: constants.SuccessState, Data: "Deck stowage plan details were successfully synced", Adv: "none"}
}

func InsertVehiclesDetails(data []manifest.VehiclesDetailsToShow, isAddedLater bool) (st *constants.AnswerState) {
	qr := "INSERT INTO `manifest_vehicles`(`vehicle_id`, `manifest_id`, `chasis_number`, `model`, `description`, `weight`, `bl_no`, `creation_date`, `inspection_status`, `tallied_status`, `discharged_status`, `is_overland`, `is_added_later`, `inspection_time`, `tallied_time`, `discharge_time`) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE `manifest_id`=VALUES(`manifest_id`), `chasis_number`=VALUES(`chasis_number`), `model`=VALUES(`model`), `description`=VALUES(`description`), `weight`=VALUES(`weight`), `bl_no`=VALUES(`bl_no`), `creation_date`=VALUES(`creation_date`), `inspection_status`=VALUES(`inspection_status`), `tallied_status`=VALUES(`tallied_status`), `discharged_status`=VALUES(`discharged_status`), `is_overland`=VALUES(`is_overland`), `is_added_later`=VALUES(`is_added_later`), `inspection_time`=VALUES(`inspection_time`), `tallied_time`=VALUES(`tallied_time`), `discharge_time`=VALUES(`discharge_time`)"
	isAddedLaterVal := "no"
	if isAddedLater {
		isAddedLaterVal = "yes"
	}
	for _, vehicle := range data {
		vals := []any{
			vehicle.VehicleId,
			vehicle.ManifestId,
			vehicle.ChasisNumber,
			vehicle.VehicleModel,
			vehicle.Description,
			vehicle.Weight,
			vehicle.BLNumber,
			vehicle.CreationDate,
			vehicle.InspectionStatus,
			vehicle.TalliedStatus,
			vehicle.DischargeStatus,
			vehicle.IsOverLand,
			isAddedLaterVal,
			vehicle.InspectionTime,
			vehicle.TalliedTime,
			vehicle.DischargeTime,
		}
		stx := gendb.SaveGeneral(qr, vals)
		if stx.State != constants.SuccessState {
			return stx
		}
	}

	return &constants.AnswerState{State: constants.SuccessState, Data: "Vehicle details were successfully synced", Adv: "none"}
}

func InsertPackageDetails(data []manifest.PackageManifestInfo, isAddedLater bool) (st *constants.AnswerState) {
	qr := "INSERT INTO `manifest_packages`(`package_id`, `package_number`, `bl_no`, `manifest_id`, `description`, `is_inspected`, `is_added_later`, `creation_time`) VALUES (?,?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE `package_number`=VALUES(`package_number`), `bl_no`=VALUES(`bl_no`), `manifest_id`=VALUES(`manifest_id`), `description`=VALUES(`description`), `is_inspected`=VALUES(`is_inspected`), `is_added_later`=VALUES(`is_added_later`), `creation_time`=VALUES(`creation_time`)"
	isAddedLaterVal := "no"
	if isAddedLater {
		isAddedLaterVal = "yes"
	}
	for _, pkg := range data {
		vals := []any{
			pkg.PackageId,
			pkg.PackageNumber,
			pkg.BLNumber,
			pkg.ManifestId,
			pkg.Description,
			pkg.IsInspected,
			isAddedLaterVal,
			pkg.CreationTime,
		}
		stx := gendb.SaveGeneral(qr, vals)
		if stx.State != constants.SuccessState {
			return stx
		}
	}

	return &constants.AnswerState{State: constants.SuccessState, Data: "Package details were successfully synced", Adv: "none"}
}

func InsertClientDetail(data []manifest.ClientInfo) (st *constants.AnswerState) {
	qr := "INSERT INTO `clients`(`client_id`, `client_name`, `logo`, `cover`, `principal`, `status`, `client_location`, `creation_date`) VALUES (?,?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE `client_name`=VALUES(`client_name`), `logo`=VALUES(`logo`), `cover`=VALUES(`cover`), `principal`=VALUES(`principal`), `status`=VALUES(`status`), `client_location`=VALUES(`client_location`), `creation_date`=VALUES(`creation_date`)"
	for _, client := range data {
		vals := []any{
			client.ClientID,
			client.ClientName,
			client.Logo,
			client.Cover,
			client.PrincipalName,
			client.Status,
			client.ClientLocation,
			client.CreationDate,
		}
		stx := gendb.SaveGeneral(qr, vals)
		if stx.State != constants.SuccessState {
			return stx
		}
	}

	return &constants.AnswerState{State: constants.SuccessState, Data: "Client details were successfully synced", Adv: "none"}
}

func InsertInspectionTallyRemarks(req manifest.InspectionChecksRequest, userId string) *constants.AnswerState {
	// Create a context with timeout to prevent hanging operations
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	db, er := gendb.InitDb()
	if er != nil {
		log.Printf("Database connection error: %v", er)
		return &constants.AnswerState{
			State: constants.ErrorState,
			Data:  er.Error(),
			Adv:   "none",
		}
	}
	defer db.Close()

	// Use ReadCommitted isolation level for better performance while maintaining data integrity
	tx, err := db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		log.Printf("Transaction start error: %v", err)
		return &constants.AnswerState{
			State: constants.ErrorState,
			Data:  "Failed to start transaction",
			Adv:   "none",
		}
	}

	var committed bool
	defer func() {
		if !committed {
			if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
				log.Printf("Failed to rollback transaction: %v", err)
			}
		}
	}()

	// Pre-generate all IDs to avoid generating them during database operations
	tallyId := specials.RandomString(36, "_TALLY")

	// Generate inspection IDs upfront
	type inspectionData struct {
		inspection manifest.InspectionCheck
		inspId     string
		imgId      string // Only used if there's an image
	}

	type OnboardPackageData struct {
		packageData manifest.OnBoardPackageRequest
		packageId   string
		mediaIds    []string
	}

	inspections := make([]inspectionData, len(req.Checks))
	for i, insp := range req.Checks {
		inspections[i] = inspectionData{
			inspection: insp,
			inspId:     specials.RandomString(36, "_INSP"),
		}
		if insp.FaultImage != "" {
			inspections[i].imgId = specials.RandomString(36, "_I_IMG")
		}
	}

	// Generate remark IDs upfront
	remarkIds := make([]string, len(req.Remarks))
	for i := range req.Remarks {
		remarkIds[i] = specials.RandomString(36, "_RMK")
	}

	// Generate media IDS upfront
	mediaIds := make([]string, len(req.Media))
	for i := range req.Media {
		mediaIds[i] = specials.RandomString(36, "_MEDIA")
	}

	// Generate onboard package ids upfront
	packagez := make([]OnboardPackageData, len(req.Packages))
	for i, packagex := range req.Packages {
		packagez[i].packageId = specials.RandomString(36, "_O_PACKAGE")
		packagez[i].packageData = packagex
		medzIds := make([]string, len(packagex.Media))
		for j := range packagex.Media {
			medzIds[j] = specials.RandomString(36, "_O_PKG_MEDIA")
		}
		packagez[i].mediaIds = medzIds
	}

	// 1. Insert tally details
	tallyQr := `INSERT INTO vehicles_talling
                (tally_id, vehicle_id, maker_id, body_id, image_link, deck_number, number_of_keys, key_type, user_id, tallied_time) 
                VALUES (?,?,?,?,?,?,?,?,?,NOW())
                ON DUPLICATE KEY UPDATE
                maker_id = ?,
                body_id = ?,
                image_link = ?,
                deck_number = ?,
                number_of_keys = ?,
                key_type = ?,
                user_id = ?,
                tallied_time = NOW()`

	tallyVals := []any{
		tallyId, req.VehicleId, req.MakerId, req.BodyId, req.VehicleImage, req.DeckNumber, req.NumberOfKeys, req.KeyType, userId,
		req.MakerId, req.BodyId, req.VehicleImage, req.DeckNumber, req.NumberOfKeys, req.KeyType, userId,
	}

	_, err = tx.ExecContext(ctx, tallyQr, tallyVals...)
	if err != nil {
		log.Printf("Tally insert error: %v", err)
		return &constants.AnswerState{
			State: constants.ErrorState,
			Data:  "Failed to save tally details",
			Adv:   "none",
		}
	}

	// 2. Update vehicle status
	tallyStQr := `UPDATE manifest_vehicles SET 
                inspection_status = ?, tallied_status = ?, inspection_time = ?, tallied_time = ?, model = ?
                WHERE vehicle_id = ?`

	_, err = tx.ExecContext(ctx, tallyStQr, "yes", "yes", req.InspectionTime, req.InspectionTime, req.ModelName, req.VehicleId)
	if err != nil {
		log.Printf("Status update error: %v", err)
		return &constants.AnswerState{
			State: constants.ErrorState,
			Data:  "Failed to update vehicle status",
			Adv:   "none",
		}
	}

	// 3. Prepare statements for batch operations
	inspStmt, err := tx.PrepareContext(ctx, `
        INSERT INTO vehicles_inspection
        (inspection_id, vehicle_id, check_id, status, check_time) 
        VALUES (?,?,?,?,NOW())
        ON DUPLICATE KEY UPDATE
		inspection_id = ?,
        vehicle_id = ?,
        check_id = ?,
        status = ?,
        check_time = NOW()
    `)
	if err != nil {
		log.Printf("Inspection statement prep error: %v", err)
		return &constants.AnswerState{
			State: constants.ErrorState,
			Data:  "Failed to prepare inspection statement",
			Adv:   "none",
		}
	}
	defer inspStmt.Close()

	// 4. First insert ALL inspections to avoid foreign key issues
	for _, insp := range inspections {
		vals := []any{
			insp.inspId, req.VehicleId, insp.inspection.CheckId, insp.inspection.CheckStatus,
			insp.inspId, req.VehicleId, insp.inspection.CheckId, insp.inspection.CheckStatus,
		}
		_, err := inspStmt.ExecContext(ctx, vals...)
		if err != nil {
			log.Printf("Inspection insert error: %v", err)
			return &constants.AnswerState{
				State: constants.ErrorState,
				Data:  "Failed to save inspection information",
				Adv:   "none",
			}
		}
	}

	// commit the transaction here to ensure all inspections are saved before images
	if err := tx.Commit(); err != nil {
		log.Printf("Transaction commit error: %v", err)
		return &constants.AnswerState{
			State: constants.ErrorState,
			Data:  "Failed to commit changes. Please try again.",
			Adv:   "none",
		}
	}
	// Reopen the transaction to handle images
	tx, err = db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		log.Printf("Transaction start error: %v", err)
		return &constants.AnswerState{
			State: constants.ErrorState,
			Data:  "Failed to start transaction",
			Adv:   "none",
		}
	}

	// 5. Now prepare and execute image inserts after all inspections are committed
	if hasImages(req.Checks) {
		inspImgStmt, err := tx.PrepareContext(ctx, `
            INSERT INTO inspection_image
            (image_id, image_link, inspection_id, creation_time) 
            VALUES (?,?,?,NOW())
            ON DUPLICATE KEY UPDATE
            image_link = ?,
            inspection_id = ?,
            creation_time = NOW()
        `)
		if err != nil {
			log.Printf("Image statement prep error: %v", err)
			return &constants.AnswerState{
				State: constants.ErrorState,
				Data:  "Failed to prepare image statement",
				Adv:   "none",
			}
		}
		defer inspImgStmt.Close()

		for _, insp := range inspections {
			if insp.inspection.FaultImage != "" {
				vals := []any{
					insp.imgId, insp.inspection.FaultImage, insp.inspId,
					insp.inspection.FaultImage, insp.inspId,
				}
				_, err := inspImgStmt.ExecContext(ctx, vals...)
				if err != nil {
					log.Printf("Image insert error: %v", err)
					return &constants.AnswerState{
						State: constants.ErrorState,
						Data:  "Failed to save image information",
						Adv:   "none",
					}
				}
			}
		}
	}

	// 6. Handle Onboard packages
	if len(packagez) > 0 {
		// delete existing packages for the vehicle to avoid duplicates, since we are doing a full replace for onboard packages
		delPkgQr := `DELETE FROM onboard_packages WHERE vehicle_id = ?`
		_, err := tx.ExecContext(ctx, delPkgQr, req.VehicleId)
		if err != nil {
			log.Printf("Failed to delete existing onboard packages: %v", err)
			return &constants.AnswerState{
				State: constants.ErrorState,
				Data:  "Failed to delete existing onboard packages",
				Adv:   "none",
			}
		}
		packStmt, err := tx.PrepareContext(ctx, `
		INSERT INTO onboard_packages
		(package_id, title, remark, vehicle_id) 
		VALUES (?,?,?,?)`)
		if err != nil {
			log.Printf("Failed to create on board package statement: %v", err)
			return &constants.AnswerState{
				State: constants.ErrorState,
				Data:  "Failed to create onboard package statement",
				Adv:   "none",
			}
		}
		defer packStmt.Close()
		packMeStmt, err := tx.PrepareContext(ctx, `
		INSERT INTO onboard_packages_media
		(media_id, media_type, media_link, package_id) 
		VALUES (?,?,?,?)`)
		if err != nil {
			log.Printf("Failed to create on board package media statement %v", err)
			return &constants.AnswerState{
				State: constants.ErrorState,
				Data:  "Failed to create onboard package media statement",
				Adv:   "none",
			}
		}
		defer packMeStmt.Close()
		for _, pack := range packagez {
			// package first
			vals := []any{
				pack.packageId, pack.packageData.Title, pack.packageData.Remark, req.VehicleId,
			}

			_, err := packStmt.ExecContext(ctx, vals...)
			if err != nil {
				log.Printf("Failed to save on board package information: %v", err)
				return &constants.AnswerState{
					State: constants.ErrorState,
					Data:  "Failed to save on board packege information",
					Adv:   "none",
				}
			}

			// package media
			for i := range pack.packageData.Media {
				valsMed := []any{
					pack.mediaIds[i], pack.packageData.Media[i].MediaType, pack.packageData.Media[i].MediaLink, pack.packageId,
				}

				_, err := packMeStmt.ExecContext(ctx, valsMed...)
				if err != nil {
					log.Printf("Failed to save onboard package media information: %v", err)
					return &constants.AnswerState{
						State: constants.ErrorState,
						Data:  "Failed to save onboard package media information",
						Adv:   "none",
					}
				}
			}
		}
	}

	// 7. handle remarks if present
	if len(req.Remarks) > 0 {
		rmkStmt, err := tx.PrepareContext(ctx, `
            INSERT INTO inspection_remarks
            (remark_id, vehicle_id, remark, remark_type, image_link, remark_time) 
            VALUES (?,?,?,?,?,NOW())
            ON DUPLICATE KEY UPDATE
            vehicle_id = ?,
            remark = ?,
            image_link = ?,
            remark_time = NOW()
        `)
		if err != nil {
			log.Printf("Remarks statement prep error: %v", err)
			return &constants.AnswerState{
				State: constants.ErrorState,
				Data:  "Failed to prepare remarks statement",
				Adv:   "none",
			}
		}
		defer rmkStmt.Close()

		for i, remark := range req.Remarks {
			vals := []any{
				remarkIds[i], req.VehicleId, remark.Remark, remark.RemarkType, remark.RemarkImage,
				req.VehicleId, remark.Remark, remark.RemarkImage,
			}
			_, err := rmkStmt.ExecContext(ctx, vals...)
			if err != nil {
				log.Printf("Remark insert error: %v", err)
				return &constants.AnswerState{
					State: constants.ErrorState,
					Data:  "Failed to save remarks information",
					Adv:   "none",
				}
			}
		}
	}

	// 8. Finally handle media
	if len(req.Media) > 0 {
		// delete existing media for the vehicle to avoid duplicates, since we are doing a full replace for media
		delMediaQr := `DELETE FROM vehicle_galllery WHERE vehicle_id = ?`
		_, err := tx.ExecContext(ctx, delMediaQr, req.VehicleId)
		if err != nil {
			log.Printf("Failed to delete existing media: %v", err)
			return &constants.AnswerState{
				State: constants.ErrorState,
				Data:  "Failed to delete existing media",
				Adv:   "none",
			}
		}
		mediaStmt, err := tx.PrepareContext(ctx, `
			INSERT INTO vehicle_galllery
			(media_id, media_link, media_type, remark, vehicle_id, status) 
			VALUES (?,?,?,?,?,?)
		`)

		if err != nil {
			log.Printf("Media statement prep error: %v", err)
			return &constants.AnswerState{
				State: constants.ErrorState,
				Data:  "Failed to prepare media statement",
				Adv:   "none",
			}
		}

		defer mediaStmt.Close()

		for i, media := range req.Media {
			vals := []any{
				mediaIds[i], media.MediaLink, media.MediaType, media.Remark, req.VehicleId, manifest.GenStatus.Active,
			}

			_, err := mediaStmt.ExecContext(ctx, vals...)
			if err != nil {
				log.Printf("Media insert has failed due to: %v", err)
				return &constants.AnswerState{
					State: constants.ErrorState,
					Data:  "Failed to save vehicle media",
					Adv:   "none",
				}
			}
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		log.Printf("Transaction commit error: %v", err)
		return &constants.AnswerState{
			State: constants.ErrorState,
			Data:  "Failed to commit changes. Please try again.",
			Adv:   "none",
		}
	}

	committed = true

	return &constants.AnswerState{
		State: constants.SuccessState,
		Data:  "Inspection and tally information were successfully added.",
		Adv:   "none",
	}
}

// Helper function to check if any checks have images
func hasImages(checks []manifest.InspectionCheck) bool {
	for _, check := range checks {
		if check.FaultImage != "" {
			return true
		}
	}
	return false
}

func InsertPackageInspection(req manifest.PackageInspectionSaveRequest, userId string) *constants.AnswerState {
	// Create a context with timeout to prevent hanging operations
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	db, er := gendb.InitDb()
	if er != nil {
		log.Printf("Database connection error: %v", er)
		return &constants.AnswerState{
			State: constants.ErrorState,
			Data:  er.Error(),
			Adv:   "none",
		}
	}
	defer db.Close()

	// Use ReadCommitted isolation level for better performance while maintaining data integrity
	tx, err := db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		log.Printf("Transaction start error: %v", err)
		return &constants.AnswerState{
			State: constants.ErrorState,
			Data:  "Failed to start transaction",
			Adv:   "none",
		}
	}

	var committed bool
	defer func() {
		if !committed {
			if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
				log.Printf("Failed to rollback transaction: %v", err)
			}
		}
	}()

	qr := ` INSERT INTO 
					packages_inspection(inspection_id, package_id, type_id, picture, inspection_status, user_id, inspection_time, creation_time) 
			VALUES (?,?,?,?,?,?,?,NOW())
			ON DUPLICATE KEY UPDATE
					type_id = ?,
					picture = ?,
					inspection_status = ?,
					user_id = ?,
					inspection_time = ?
		`
	id := specials.RandomString(24, "_PACKAGE_INSPECTION")
	vals := []any{
		id, req.PackageId, req.TypeId, req.PackageImage, req.InspectionStatusId, userId, req.InspectionTime,
		req.TypeId, req.PackageImage, req.InspectionStatusId, userId, req.InspectionTime,
	}

	inspeStmt, err := tx.PrepareContext(context.Background(), qr)
	if err != nil {
		log.Printf("Statement preparation error: %v", err)
		return &constants.AnswerState{
			State: constants.ErrorState,
			Data:  "Failed to prepare statement",
			Adv:   "none",
		}
	}
	defer inspeStmt.Close()

	_, err = inspeStmt.ExecContext(context.Background(), vals...)
	if err != nil {
		log.Printf("Statement execution error: %v", err)
		return &constants.AnswerState{
			State: constants.ErrorState,
			Data:  "Failed to execute statement",
			Adv:   "none",
		}
	}

	// udpate package manifest inspection status
	manifestUpdateQr := `UPDATE manifest_packages SET is_inspected = ? WHERE package_id = ?`
	_, err = tx.ExecContext(ctx, manifestUpdateQr, manifest.InspectionStatus.Yes, req.PackageId)
	if err != nil {
		log.Printf("Manifest update error: %v", err)
		return &constants.AnswerState{
			State: constants.ErrorState,
			Data:  "Failed to update package manifest inspection status",
			Adv:   "none",
		}
	}

	// add media if present
	if len(req.Media) > 0 {
		// delete existing media for the package to avoid duplicates, since we are doing a full replace for media
		delMediaQr := `DELETE FROM package_gallery WHERE package_id = ?`
		_, err := tx.ExecContext(ctx, delMediaQr, req.PackageId)
		if err != nil {
			log.Printf("Failed to delete existing media: %v", err)
			return &constants.AnswerState{
				State: constants.ErrorState,
				Data:  "Failed to delete existing media",
				Adv:   "none",
			}
		}

		mediaQr := `INSERT INTO package_gallery(media_id, media_link, media_type, remark, package_id, status) VALUES (?,?,?,?,?,?)`
		mediaStmt, err := tx.PrepareContext(ctx, mediaQr)
		if err != nil {
			log.Printf("Media statement preparation error: %v", err)
			return &constants.AnswerState{
				State: constants.ErrorState,
				Data:  "Failed to prepare media statement",
				Adv:   "none",
			}
		}
		defer mediaStmt.Close()

		for _, media := range req.Media {
			mediaId := specials.RandomString(24, "_PKG_INSP_MEDIA")
			vals := []any{
				mediaId, media.MediaLink, media.MediaType, media.Remark, req.PackageId, manifest.GenStatus.Active,
			}
			_, err := mediaStmt.ExecContext(ctx, vals...)
			if err != nil {
				log.Printf("Media statement execution error: %v", err)
				return &constants.AnswerState{
					State: constants.ErrorState,
					Data:  "Failed to save media information",
					Adv:   "none",
				}
			}
		}
	}

	committed = true
	if err := tx.Commit(); err != nil {
		log.Printf("Transaction commit error: %v", err)
		return &constants.AnswerState{
			State: constants.ErrorState,
			Data:  "Failed to commit changes. Please try again.",
			Adv:   "none",
		}
	}

	return &constants.AnswerState{
		State: constants.SuccessState,
		Data:  "Package inspection added successfully",
		Adv:   "none",
	}
}
