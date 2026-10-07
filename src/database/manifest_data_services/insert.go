package manifestdataservices

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/go-sql-driver/mysql"
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
	slog.Info(fmt.Sprint(data[0]))
	qr := "INSERT INTO `manifest`(`manifest_id`, `manifest_name`, `client_id`, `vessel_name`, `voyage_no`, `berth_no`, `arrival_date`, `received_date`, `uploaded_date`) VALUES (?,?,?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE `manifest_name`=VALUES(`manifest_name`), `client_id`=VALUES(`client_id`), `vessel_name`=VALUES(`vessel_name`), `voyage_no`=VALUES(`voyage_no`), `berth_no`=VALUES(`berth_no`), `arrival_date`=VALUES(`arrival_date`), `received_date`=VALUES(`received_date`), `uploaded_date`=VALUES(`uploaded_date`)"
	for i, detail := range data {
		slog.Info(fmt.Sprintf("Processing manifest detail %d/%d: Manifest ID %s", i+1, len(data), detail.ManifestId))
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
		if err := ensureClient(detail.ClientId, detail.ClientName); err != nil {
			slog.Error("Could not add the manifest's client", "manifestId", detail.ManifestId, "clientId", detail.ClientId, "error", err)
			return &constants.AnswerState{State: constants.ErrorState, Data: "the manifest's client could not be added: " + err.Error(), Adv: "none"}
		}
		stx := gendb.SaveGeneral(qr, vals)
		if stx.State != constants.SuccessState {
			return stx
		}
		slog.Info(fmt.Sprintf("Successfully inserted manifest detail %d/%d: Manifest ID %s", i+1, len(data), detail.ManifestId))
	}

	return &constants.AnswerState{State: constants.SuccessState, Data: "Manifest details were successfully synced", Adv: "none"}
}

// ensureClient adds the manifest's client when it is not here yet (manifest.client_id
// is a foreign key to clients). The remote client list holds only active clients, so a
// manifest of an inactive client would otherwise fail. The row holds what the manifest
// knows; a later client list sync fills in the rest. client_name is unique, so when
// another client already uses the name, the id is added to it.
func ensureClient(clientId, clientName string) error {
	db, err := gendb.InitDb()
	if err != nil {
		return err
	}
	exists := func() (bool, error) {
		var n int
		err := db.QueryRow("SELECT COUNT(*) FROM clients WHERE client_id = ?", clientId).Scan(&n)
		return n > 0, err
	}
	if ok, err := exists(); err != nil || ok {
		return err
	}
	name := strings.TrimSpace(clientName)
	if name == "" {
		name = clientId
	}
	for _, candidate := range []string{name, name + " (" + clientId + ")"} {
		if len(candidate) > 100 {
			candidate = candidate[:100]
		}
		if _, err := db.Exec("INSERT INTO clients (client_id, client_name, principal, client_location, creation_date) VALUES (?,?,'','',NOW()) ON DUPLICATE KEY UPDATE client_id = client_id", clientId, candidate); err != nil {
			return err
		}
		if ok, err := exists(); err != nil || ok {
			return err
		}
	}
	return fmt.Errorf("client %s could not be added: its name is already used by another client", clientId)
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
	// NOTE: inspection_status, tallied_status, inspection_time, tallied_time are intentionally excluded
	// from ON DUPLICATE KEY UPDATE — these fields are managed locally and must not be overwritten by remote syncs.
	qr := "INSERT INTO `manifest_vehicles`(`vehicle_id`, `manifest_id`, `chasis_number`, `model`, `description`, `weight`, `bl_no`, `creation_date`, `inspection_status`, `tallied_status`, `discharged_status`, `is_overland`, `is_added_later`, `inspection_time`, `tallied_time`, `discharge_time`) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE `manifest_id`=VALUES(`manifest_id`), `chasis_number`=VALUES(`chasis_number`), `model`=VALUES(`model`), `description`=VALUES(`description`), `weight`=VALUES(`weight`), `bl_no`=VALUES(`bl_no`), `creation_date`=VALUES(`creation_date`), `discharged_status`=VALUES(`discharged_status`), `is_overland`=VALUES(`is_overland`), `is_added_later`=VALUES(`is_added_later`), `discharge_time`=VALUES(`discharge_time`)"
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

// Adv values returned by InsertInspectionTallyRemarks so callers can tell the outcomes apart.
const (
	InspectionSaveNew       = "new"
	InspectionSaveResend    = "resend"
	InspectionSaveReinspect = "reinspection"
	InspectionSaveConflict  = "conflict"
)

// Adv values of errors that a retry cannot fix. Handlers answer them with a 4xx code
// so the tablets stop retrying and show the item as needing attention.
const (
	SaveErrNotFound         = "not_found"
	SaveErrInvalidReference = "invalid_reference"
)

// invalidReference reports whether err is a foreign key failure (MySQL 1452): the
// request names a check, maker, body or package that does not exist.
func invalidReference(err error) bool {
	var myErr *mysql.MySQLError
	return errors.As(err, &myErr) && myErr.Number == 1452
}

func InsertInspectionTallyRemarks(req manifest.InspectionChecksRequest, userId string) *constants.AnswerState {
	// Create a context with timeout to prevent hanging operations
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	db, er := gendb.InitDb()
	if er != nil {
		slog.Error(fmt.Sprintf("Database connection error: %v", er))
		return &constants.AnswerState{
			State: constants.ErrorState,
			Data:  er.Error(),
			Adv:   "none",
		}
	}

	// The whole save runs in ONE transaction. A failure leaves nothing behind, so a
	// retry from the tablet never meets a half-saved inspection.
	tx, err := db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		slog.Error(fmt.Sprintf("Transaction start error: %v", err))
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
				slog.Error(fmt.Sprintf("Failed to rollback transaction: %v", err))
			}
		}
	}()

	fail := func(msg string, err error) *constants.AnswerState {
		slog.Error(fmt.Sprintf("%s for vehicle %s: %v", msg, req.VehicleId, err))
		if invalidReference(err) {
			return &constants.AnswerState{
				State: constants.ErrorState,
				Data:  msg + ": the inspection refers to a check, maker or body type that does not exist on this server",
				Adv:   SaveErrInvalidReference,
			}
		}
		return &constants.AnswerState{
			State: constants.ErrorState,
			Data:  msg,
			Adv:   "none",
		}
	}

	// 1. Lock the vehicle row. Saves for the same vehicle from several tablets now run
	// one after another instead of interleaving.
	var currentInspectionTime sql.NullString
	err = tx.QueryRowContext(ctx,
		"SELECT inspection_time FROM manifest_vehicles WHERE vehicle_id = ? FOR UPDATE",
		req.VehicleId).Scan(&currentInspectionTime)
	if err == sql.ErrNoRows {
		return &constants.AnswerState{
			State: constants.ErrorState,
			Data:  "Vehicle was not found in the manifest",
			Adv:   SaveErrNotFound,
		}
	}
	if err != nil {
		return fail("Failed to lock vehicle record", err)
	}

	// 1b. Refuse ids this server does not know, before anything is written. The
	// foreign keys would refuse them too, but this names the exact id for the tablet.
	unknown, err := unknownInspectionReferences(ctx, tx, req)
	if err != nil {
		return fail("Failed to check the inspection's maker, body type and checks", err)
	}
	if len(unknown) > 0 {
		slog.Warn("Inspection refers to unknown records", "vehicle", req.VehicleId, "unknown", strings.Join(unknown, ", "))
		return &constants.AnswerState{
			State: constants.ErrorState,
			Data:  "Not known on this server: " + strings.Join(unknown, ", ") + ". Sync the tablet's lists, then send the inspection again.",
			Adv:   SaveErrInvalidReference,
		}
	}

	// 2. Read the active tally, if there is one. submission_id and manifest_id come
	// from a migration; without them the save falls back to the older behaviour.
	hasSubmissionId := gendb.ColumnExists("vehicles_talling", "submission_id")
	hasTallyManifestId := gendb.ColumnExists("vehicles_talling", "manifest_id")

	var tallyUserId string
	var tallySubmissionId sql.NullString
	tallyExists := true
	if hasSubmissionId {
		err = tx.QueryRowContext(ctx,
			"SELECT user_id, submission_id FROM vehicles_talling WHERE vehicle_id = ?",
			req.VehicleId).Scan(&tallyUserId, &tallySubmissionId)
	} else {
		err = tx.QueryRowContext(ctx,
			"SELECT user_id FROM vehicles_talling WHERE vehicle_id = ?",
			req.VehicleId).Scan(&tallyUserId)
	}
	if err == sql.ErrNoRows {
		tallyExists = false
	} else if err != nil {
		return fail("Failed to check for existing tally data", err)
	}

	// 3. Decide whether this request is a new inspection, a resend of the inspection
	// that is already saved, or a real re-inspection.
	outcome := InspectionSaveNew
	if tallyExists {
		storedSubmissionId := tallySubmissionId.String
		sameSubmission := req.SubmissionId != "" && storedSubmissionId == req.SubmissionId
		// App builds without a submission id resend the same inspection time from the
		// same user, so that pair identifies a resend.
		legacyResend := (req.SubmissionId == "" || storedSubmissionId == "") &&
			currentInspectionTime.String == req.InspectionTime &&
			tallyUserId == userId
		if sameSubmission || legacyResend {
			outcome = InspectionSaveResend
		} else {
			outcome = InspectionSaveReinspect
		}
	}

	if outcome == InspectionSaveReinspect && req.Reinspect != nil && !*req.Reinspect {
		return &constants.AnswerState{
			State: constants.ErrorState,
			Data:  "This vehicle was already inspected. Confirm re-inspection to replace the existing inspection.",
			Adv:   InspectionSaveConflict,
		}
	}

	// 4. Clear the active rows. A real re-inspection is archived first. A resend is
	// overwritten in place, so it never creates a history batch.
	switch outcome {
	case InspectionSaveReinspect:
		if err := copyVehicleDataToHistory(ctx, tx, req.VehicleId); err != nil {
			return fail("Failed to archive existing vehicle data", err)
		}
		if err := deleteActiveVehicleData(ctx, tx, req.VehicleId); err != nil {
			return fail("Failed to archive existing vehicle data", err)
		}
	case InspectionSaveResend:
		if err := deleteActiveVehicleData(ctx, tx, req.VehicleId); err != nil {
			return fail("Failed to replace existing vehicle data", err)
		}
	}

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

	// 5. Insert tally details. Any previous tally row was removed in step 4.
	tallyCols := "tally_id, vehicle_id, maker_id, body_id, image_link, deck_number, number_of_keys, key_type, user_id"
	tallyVals := []any{tallyId, req.VehicleId, req.MakerId, req.BodyId, req.VehicleImage, req.DeckNumber, req.NumberOfKeys, req.KeyType, userId}
	if hasTallyManifestId {
		tallyCols += ", manifest_id"
		tallyVals = append(tallyVals, req.ManifestId)
	}
	if hasSubmissionId {
		var submissionId any
		if req.SubmissionId != "" {
			submissionId = req.SubmissionId
		}
		tallyCols += ", submission_id"
		tallyVals = append(tallyVals, submissionId)
	}
	_, err = tx.ExecContext(ctx,
		"INSERT INTO vehicles_talling ("+tallyCols+", tallied_time) VALUES ("+strings.Repeat("?,", len(tallyVals))+"NOW())",
		tallyVals...)
	if err != nil {
		return fail("Failed to save tally details", err)
	}

	// 6. Update vehicle status. New inspection data must be published again; a resend
	// carries the same inspection, so its published flag stays as it is.
	statusQr := "UPDATE manifest_vehicles SET inspection_status = ?, tallied_status = ?, inspection_time = ?, tallied_time = ?, model = ?"
	statusVals := []any{"yes", "yes", req.InspectionTime, req.InspectionTime, req.ModelName}
	if outcome != InspectionSaveResend {
		statusQr += ", is_published = ?"
		statusVals = append(statusVals, "no")
	}
	statusQr += " WHERE vehicle_id = ?"
	statusVals = append(statusVals, req.VehicleId)

	if _, err = tx.ExecContext(ctx, statusQr, statusVals...); err != nil {
		return fail("Failed to update vehicle status", err)
	}

	// 7. Insert inspections
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
		return fail("Failed to prepare inspection statement", err)
	}
	defer inspStmt.Close()

	for _, insp := range inspections {
		vals := []any{
			insp.inspId, req.VehicleId, insp.inspection.CheckId, insp.inspection.CheckStatus,
			insp.inspId, req.VehicleId, insp.inspection.CheckId, insp.inspection.CheckStatus,
		}
		if _, err := inspStmt.ExecContext(ctx, vals...); err != nil {
			return fail("Failed to save inspection information", err)
		}
	}

	// 8. Inspection images. The parent inspections were inserted above in the same
	// transaction, so the foreign key is satisfied.
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
			return fail("Failed to prepare image statement", err)
		}
		defer inspImgStmt.Close()

		for _, insp := range inspections {
			if insp.inspection.FaultImage != "" {
				vals := []any{
					insp.imgId, insp.inspection.FaultImage, insp.inspId,
					insp.inspection.FaultImage, insp.inspId,
				}
				if _, err := inspImgStmt.ExecContext(ctx, vals...); err != nil {
					return fail("Failed to save image information", err)
				}
			}
		}
	}

	// 9. Handle Onboard packages
	if len(packagez) > 0 {
		// delete existing packages for the vehicle to avoid duplicates, since we are doing a full replace for onboard packages
		if _, err := tx.ExecContext(ctx, `DELETE FROM onboard_packages WHERE vehicle_id = ?`, req.VehicleId); err != nil {
			return fail("Failed to delete existing onboard packages", err)
		}
		packStmt, err := tx.PrepareContext(ctx, `
		INSERT INTO onboard_packages
		(package_id, title, remark, vehicle_id)
		VALUES (?,?,?,?)`)
		if err != nil {
			return fail("Failed to create onboard package statement", err)
		}
		defer packStmt.Close()
		packMeStmt, err := tx.PrepareContext(ctx, `
		INSERT INTO onboard_packages_media
		(media_id, media_type, media_link, package_id)
		VALUES (?,?,?,?)`)
		if err != nil {
			return fail("Failed to create onboard package media statement", err)
		}
		defer packMeStmt.Close()
		for _, pack := range packagez {
			// package first
			vals := []any{
				pack.packageId, pack.packageData.Title, pack.packageData.Remark, req.VehicleId,
			}
			if _, err := packStmt.ExecContext(ctx, vals...); err != nil {
				return fail("Failed to save on board packege information", err)
			}

			// package media
			for i := range pack.packageData.Media {
				valsMed := []any{
					pack.mediaIds[i], pack.packageData.Media[i].MediaType, pack.packageData.Media[i].MediaLink, pack.packageId,
				}
				if _, err := packMeStmt.ExecContext(ctx, valsMed...); err != nil {
					return fail("Failed to save onboard package media information", err)
				}
			}
		}
	}

	// 10. handle remarks if present
	for i, remark := range req.Remarks {
		if err := replaceRemark(ctx, tx, remarkIds[i], req.VehicleId, remark); err != nil {
			return fail("Failed to save remarks information", err)
		}
	}

	// 11. Finally handle media
	if len(req.Media) > 0 {
		// delete existing media for the vehicle to avoid duplicates, since we are doing a full replace for media
		if _, err := tx.ExecContext(ctx, `DELETE FROM vehicle_galllery WHERE vehicle_id = ?`, req.VehicleId); err != nil {
			return fail("Failed to delete existing media", err)
		}
		mediaStmt, err := tx.PrepareContext(ctx, `
			INSERT INTO vehicle_galllery
			(media_id, media_link, media_type, remark, vehicle_id, status)
			VALUES (?,?,?,?,?,?)
		`)
		if err != nil {
			return fail("Failed to prepare media statement", err)
		}
		defer mediaStmt.Close()

		for i, media := range req.Media {
			vals := []any{
				mediaIds[i], media.MediaLink, media.MediaType, media.Remark, req.VehicleId, manifest.GenStatus.Active,
			}
			if _, err := mediaStmt.ExecContext(ctx, vals...); err != nil {
				return fail("Failed to save vehicle media", err)
			}
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fail("Failed to commit changes. Please try again.", err)
	}
	committed = true

	if outcome != InspectionSaveNew {
		slog.Info(fmt.Sprintf("Inspection save for vehicle %s handled as %s", req.VehicleId, outcome))
	}

	return &constants.AnswerState{
		State: constants.SuccessState,
		Data:  "Inspection and tally information were successfully added.",
		Adv:   outcome,
	}
}

// unknownInspectionReferences lists the maker, body type and check ids of an
// inspection that do not exist on this server.
func unknownInspectionReferences(ctx context.Context, tx *sql.Tx, req manifest.InspectionChecksRequest) ([]string, error) {
	unknown := []string{}

	exists := func(table, column, id string) (bool, error) {
		var one int
		err := tx.QueryRowContext(ctx, "SELECT 1 FROM "+table+" WHERE "+column+" = ? LIMIT 1", id).Scan(&one)
		if err == sql.ErrNoRows {
			return false, nil
		}
		return err == nil, err
	}
	for _, ref := range []struct{ table, column, id, label string }{
		{"vehicle_makers", "maker_id", req.MakerId, "maker"},
		{"vehicle_bodies", "body_id", req.BodyId, "body type"},
	} {
		ok, err := exists(ref.table, ref.column, ref.id)
		if err != nil {
			return nil, err
		}
		if !ok {
			unknown = append(unknown, fmt.Sprintf("%s %q", ref.label, ref.id))
		}
	}

	ids := []any{}
	seen := map[string]bool{}
	for _, c := range req.Checks {
		if !seen[c.CheckId] {
			seen[c.CheckId] = true
			ids = append(ids, c.CheckId)
		}
	}
	if len(ids) > 0 {
		rows, err := tx.QueryContext(ctx,
			"SELECT check_id FROM inspection_checklist WHERE check_id IN ("+strings.TrimSuffix(strings.Repeat("?,", len(ids)), ",")+")",
			ids...)
		if err != nil {
			return nil, err
		}
		found := map[string]bool{}
		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				rows.Close()
				return nil, err
			}
			found[id] = true
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return nil, err
		}
		for _, id := range ids {
			if !found[id.(string)] {
				unknown = append(unknown, fmt.Sprintf("check %q", id))
			}
		}
	}
	return unknown, nil
}

// replaceRemark writes one remark for a vehicle. inspection_remarks has a
// UNIQUE (vehicle_id, remark) USING HASH key, and ON DUPLICATE KEY UPDATE does not
// catch that key on MariaDB (Error 1062 on 'vehicle_id_2'). The row with the same
// text is therefore deleted first, so a repeated remark text replaces the old one.
func replaceRemark(ctx context.Context, tx *sql.Tx, remarkId, vehicleId string, remark manifest.RemarkSaveRequest) error {
	if _, err := tx.ExecContext(ctx,
		"DELETE FROM inspection_remarks WHERE vehicle_id = ? AND remark = ?",
		vehicleId, remark.Remark); err != nil {
		return err
	}
	_, err := tx.ExecContext(ctx, `
		INSERT INTO inspection_remarks
		(remark_id, vehicle_id, remark, remark_type, image_link, remark_time)
		VALUES (?,?,?,?,?,NOW())`,
		remarkId, vehicleId, remark.Remark, remark.RemarkType, remark.RemarkImage)
	return err
}

// vehicleDataTables lists every active inspection table with its history twin and the
// filter that selects one vehicle's rows. Each child table follows its parent.
var vehicleDataTables = []struct {
	active  string
	history string
	filter  string
}{
	{"vehicles_inspection", "vehicles_inspection_history", "vehicle_id = ?"},
	{"inspection_image", "inspection_image_history", "inspection_id IN (SELECT inspection_id FROM vehicles_inspection WHERE vehicle_id = ?)"},
	{"onboard_packages", "onboard_packages_history", "vehicle_id = ?"},
	{"onboard_packages_media", "onboard_packages_media_history", "package_id IN (SELECT package_id FROM onboard_packages WHERE vehicle_id = ?)"},
	{"inspection_remarks", "inspection_remarks_history", "vehicle_id = ?"},
	{"vehicle_galllery", "vehicle_galllery_history", "vehicle_id = ?"},
	{"vehicles_talling", "vehicles_talling_history", "vehicle_id = ?"},
}

// copyVehicleDataToHistory copies all active inspection records for vehicleId into
// their history tables as one batch. Every row of the batch gets the same archived_at
// value, because the timeline matches the history tables on that value.
//
// Columns are copied by name (see sharedColumns), never with SELECT *. The old
// SELECT * copy failed with "Column count doesn't match" and, where the counts
// happened to match, wrote values into the wrong columns. Every history table needs
// an `archived_at DATETIME` column.
func copyVehicleDataToHistory(ctx context.Context, tx *sql.Tx, vehicleId string) error {
	// archived_at identifies the batch, and it has one-second resolution. Two archives
	// of the same vehicle within one second would merge into one batch, so the value
	// is kept strictly increasing per vehicle. The caller holds the vehicle row lock,
	// so no other save can pick the same value.
	var archivedAt string
	if err := tx.QueryRowContext(ctx, `
		SELECT DATE_FORMAT(GREATEST(NOW(), COALESCE(MAX(archived_at) + INTERVAL 1 SECOND, NOW())), '%Y-%m-%d %H:%i:%s')
		FROM vehicles_talling_history WHERE vehicle_id = ?`, vehicleId).Scan(&archivedAt); err != nil {
		return fmt.Errorf("read archive time: %w", err)
	}

	for _, t := range vehicleDataTables {
		cols, err := sharedColumns(ctx, tx, t.active, t.history)
		if err != nil {
			return err
		}
		qr := fmt.Sprintf("INSERT INTO %s (%s, archived_at) SELECT %s, ? FROM %s WHERE %s", t.history, cols.insert, cols.sel, t.active, t.filter)
		if _, err := tx.ExecContext(ctx, qr, archivedAt, vehicleId); err != nil {
			return fmt.Errorf("archive %s: %w", t.active, err)
		}
	}
	return nil
}

// deleteActiveVehicleData removes all active inspection records for vehicleId. Each
// child table is cleared before its parent to satisfy the foreign keys.
func deleteActiveVehicleData(ctx context.Context, tx *sql.Tx, vehicleId string) error {
	for _, qr := range []string{
		"DELETE FROM inspection_image WHERE inspection_id IN (SELECT inspection_id FROM vehicles_inspection WHERE vehicle_id = ?)",
		"DELETE FROM vehicles_inspection WHERE vehicle_id = ?",
		"DELETE FROM onboard_packages_media WHERE package_id IN (SELECT package_id FROM onboard_packages WHERE vehicle_id = ?)",
		"DELETE FROM onboard_packages WHERE vehicle_id = ?",
		"DELETE FROM inspection_remarks WHERE vehicle_id = ?",
		"DELETE FROM vehicle_galllery WHERE vehicle_id = ?",
		"DELETE FROM vehicles_talling WHERE vehicle_id = ?",
	} {
		if _, err := tx.ExecContext(ctx, qr, vehicleId); err != nil {
			return fmt.Errorf("delete active data: %w", err)
		}
	}
	return nil
}

type archiveColumns struct {
	insert string // column list for INSERT INTO <history> (...)
	sel    string // matching expressions for SELECT ... FROM <active>
}

var archiveColumnsCache sync.Map

// sharedColumns returns the columns to copy from the active table into its history
// table: every column present in both, in the active table's order. A required
// history column that the active table lacks is filled from manifest_vehicles when it
// is manifest_id (older vehicles_talling tables have no manifest_id); any other one is
// an error. The result is cached for the life of the process.
func sharedColumns(ctx context.Context, tx *sql.Tx, active, history string) (archiveColumns, error) {
	if cols, ok := archiveColumnsCache.Load(active); ok {
		return cols.(archiveColumns), nil
	}

	rows, err := tx.QueryContext(ctx, `
		SELECT h.COLUMN_NAME,
		       a.COLUMN_NAME IS NOT NULL AS in_active,
		       h.IS_NULLABLE = 'NO' AND h.COLUMN_DEFAULT IS NULL AS required
		FROM information_schema.COLUMNS h
		LEFT JOIN information_schema.COLUMNS a
		  ON a.TABLE_SCHEMA = h.TABLE_SCHEMA AND a.TABLE_NAME = ? AND a.COLUMN_NAME = h.COLUMN_NAME
		WHERE h.TABLE_SCHEMA = DATABASE() AND h.TABLE_NAME = ? AND h.COLUMN_NAME <> 'archived_at'
		ORDER BY h.ORDINAL_POSITION`, active, history)
	if err != nil {
		return archiveColumns{}, fmt.Errorf("read columns of %s: %w", history, err)
	}
	defer rows.Close()

	insertCols := []string{}
	selectExprs := []string{}
	for rows.Next() {
		var name string
		var inActive, required bool
		if err := rows.Scan(&name, &inActive, &required); err != nil {
			return archiveColumns{}, fmt.Errorf("read columns of %s: %w", history, err)
		}
		switch {
		case inActive:
			insertCols = append(insertCols, "`"+name+"`")
			selectExprs = append(selectExprs, "`"+name+"`")
		case required && name == "manifest_id":
			insertCols = append(insertCols, "`manifest_id`")
			selectExprs = append(selectExprs, fmt.Sprintf("(SELECT mv.manifest_id FROM manifest_vehicles mv WHERE mv.vehicle_id = %s.vehicle_id)", active))
		case required:
			return archiveColumns{}, fmt.Errorf("history table %s requires column %s, which %s does not have", history, name, active)
		}
	}
	if err := rows.Err(); err != nil {
		return archiveColumns{}, fmt.Errorf("read columns of %s: %w", history, err)
	}
	if len(insertCols) == 0 {
		return archiveColumns{}, fmt.Errorf("history table %s is missing or shares no columns with %s", history, active)
	}

	cols := archiveColumns{insert: strings.Join(insertCols, ", "), sel: strings.Join(selectExprs, ", ")}
	archiveColumnsCache.Store(active, cols)
	return cols, nil
}

// nullArchiveTime scans a DATETIME such as MAX(archived_at). The DSN does not set
// parseTime, so the driver returns the value as text, which sql.NullTime cannot scan
// ("unsupported Scan, storing driver.Value type []uint8 into type *time.Time").
type nullArchiveTime struct {
	Time  string
	Valid bool
}

func (n *nullArchiveTime) Scan(value any) error {
	switch v := value.(type) {
	case nil:
		n.Time, n.Valid = "", false
	case []byte:
		n.Time, n.Valid = string(v), true
	case string:
		n.Time, n.Valid = v, true
	case time.Time:
		n.Time, n.Valid = v.Format("2006-01-02 15:04:05"), true
	default:
		return fmt.Errorf("unsupported archived_at value of type %T", value)
	}
	return nil
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
		slog.Error(fmt.Sprintf("Database connection error: %v", er))
		return &constants.AnswerState{
			State: constants.ErrorState,
			Data:  er.Error(),
			Adv:   "none",
		}
	}

	// Use ReadCommitted isolation level for better performance while maintaining data integrity
	tx, err := db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		slog.Error(fmt.Sprintf("Transaction start error: %v", err))
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
				slog.Error(fmt.Sprintf("Failed to rollback transaction: %v", err))
			}
		}
	}()

	// Lock the package row so saves from several tablets do not interleave, and find
	// out whether this request changes the stored inspection.
	var lockedId string
	err = tx.QueryRowContext(ctx, "SELECT package_id FROM manifest_packages WHERE package_id = ? FOR UPDATE", req.PackageId).Scan(&lockedId)
	if err == sql.ErrNoRows {
		return &constants.AnswerState{State: constants.ErrorState, Data: "Package was not found in the manifest", Adv: SaveErrNotFound}
	}
	if err != nil {
		slog.Error(fmt.Sprintf("Package lock error: %v", err))
		return &constants.AnswerState{State: constants.ErrorState, Data: "Failed to lock package record", Adv: "none"}
	}

	var oldType, oldPicture, oldStatus, oldTime string
	changed := true
	err = tx.QueryRowContext(ctx,
		"SELECT type_id, picture, inspection_status, DATE_FORMAT(inspection_time, '%Y-%m-%d %H:%i:%s') FROM packages_inspection WHERE package_id = ?",
		req.PackageId).Scan(&oldType, &oldPicture, &oldStatus, &oldTime)
	if err == nil {
		changed = oldType != req.TypeId || oldPicture != req.PackageImage || oldStatus != req.InspectionStatusId || oldTime != req.InspectionTime
	} else if err != sql.ErrNoRows {
		slog.Error(fmt.Sprintf("Package inspection read error: %v", err))
		return &constants.AnswerState{State: constants.ErrorState, Data: "Failed to read package inspection", Adv: "none"}
	}

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

	inspeStmt, err := tx.PrepareContext(ctx, qr)
	if err != nil {
		slog.Error(fmt.Sprintf("Statement preparation error: %v", err))
		return &constants.AnswerState{
			State: constants.ErrorState,
			Data:  "Failed to prepare statement",
			Adv:   "none",
		}
	}
	defer inspeStmt.Close()

	_, err = inspeStmt.ExecContext(ctx, vals...)
	if err != nil {
		slog.Error(fmt.Sprintf("Statement execution error: %v", err))
		if invalidReference(err) {
			return &constants.AnswerState{
				State: constants.ErrorState,
				Data:  "The package inspection refers to a package type or status that does not exist on this server",
				Adv:   SaveErrInvalidReference,
			}
		}
		return &constants.AnswerState{
			State: constants.ErrorState,
			Data:  "Failed to execute statement",
			Adv:   "none",
		}
	}

	// udpate package manifest inspection status
	// A changed inspection has to reach the remote server again.
	manifestUpdateQr := `UPDATE manifest_packages SET is_inspected = ? WHERE package_id = ?`
	manifestUpdateVals := []any{manifest.InspectionStatus.Yes, req.PackageId}
	if changed && gendb.ColumnExists("manifest_packages", "is_published") {
		manifestUpdateQr = `UPDATE manifest_packages SET is_inspected = ?, is_published = 'no' WHERE package_id = ?`
	}
	_, err = tx.ExecContext(ctx, manifestUpdateQr, manifestUpdateVals...)
	if err != nil {
		slog.Error(fmt.Sprintf("Manifest update error: %v", err))
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
			slog.Error(fmt.Sprintf("Failed to delete existing media: %v", err))
			return &constants.AnswerState{
				State: constants.ErrorState,
				Data:  "Failed to delete existing media",
				Adv:   "none",
			}
		}

		mediaQr := `INSERT INTO package_gallery(media_id, media_link, media_type, remark, package_id, status) VALUES (?,?,?,?,?,?)`
		mediaStmt, err := tx.PrepareContext(ctx, mediaQr)
		if err != nil {
			slog.Error(fmt.Sprintf("Media statement preparation error: %v", err))
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
				slog.Error(fmt.Sprintf("Media statement execution error: %v", err))
				return &constants.AnswerState{
					State: constants.ErrorState,
					Data:  "Failed to save media information",
					Adv:   "none",
				}
			}
		}
	}

	if err := tx.Commit(); err != nil {
		slog.Error(fmt.Sprintf("Transaction commit error: %v", err))
		return &constants.AnswerState{
			State: constants.ErrorState,
			Data:  "Failed to commit changes. Please try again.",
			Adv:   "none",
		}
	}
	committed = true

	return &constants.AnswerState{
		State: constants.SuccessState,
		Data:  "Package inspection added successfully",
		Adv:   "none",
	}
}

// InsertVehicleRemarksOnly inserts a list of remarks for a vehicle without touching inspection data
func InsertVehicleRemarksOnly(req manifest.VehicleRemarksOnlyRequest, userId string) *constants.AnswerState {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	db, er := gendb.InitDb()
	if er != nil {
		slog.Error(fmt.Sprintf("Database connection error: %v", er))
		return &constants.AnswerState{State: constants.ErrorState, Data: er.Error(), Adv: "none"}
	}

	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		slog.Error(fmt.Sprintf("Transaction start error: %v", err))
		return &constants.AnswerState{State: constants.ErrorState, Data: "Failed to start transaction", Adv: "none"}
	}

	var committed bool
	defer func() {
		if !committed {
			if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
				slog.Error(fmt.Sprintf("Failed to rollback transaction: %v", err))
			}
		}
	}()

	// Lock the vehicle row so remarks-only saves and full saves do not interleave.
	var lockedId string
	err = tx.QueryRowContext(ctx, "SELECT vehicle_id FROM manifest_vehicles WHERE vehicle_id = ? FOR UPDATE", req.VehicleId).Scan(&lockedId)
	if err == sql.ErrNoRows {
		return &constants.AnswerState{State: constants.ErrorState, Data: "Vehicle was not found in the manifest", Adv: SaveErrNotFound}
	}
	if err != nil {
		slog.Error(fmt.Sprintf("Vehicle lock error: %v", err))
		return &constants.AnswerState{State: constants.ErrorState, Data: "Failed to lock vehicle record", Adv: "none"}
	}

	// A remark with the same text replaces the old one, so a resend from the tablet
	// does not fail on the (vehicle_id, remark) unique key.
	for _, remark := range req.Remarks {
		remarkId := specials.RandomString(36, "_RMK")
		if err := replaceRemark(ctx, tx, remarkId, req.VehicleId, remark); err != nil {
			slog.Error(fmt.Sprintf("Remark insert error: %v", err))
			return &constants.AnswerState{State: constants.ErrorState, Data: "Failed to save remark information", Adv: "none"}
		}
	}

	// The new remarks have to reach the remote server on the next publish.
	if _, err := tx.ExecContext(ctx, "UPDATE manifest_vehicles SET is_published = ? WHERE vehicle_id = ?", "no", req.VehicleId); err != nil {
		slog.Error(fmt.Sprintf("Publish status reset error: %v", err))
		return &constants.AnswerState{State: constants.ErrorState, Data: "Failed to update vehicle status", Adv: "none"}
	}

	if err := tx.Commit(); err != nil {
		slog.Error(fmt.Sprintf("Transaction commit error: %v", err))
		return &constants.AnswerState{State: constants.ErrorState, Data: "Failed to commit changes. Please try again.", Adv: "none"}
	}
	committed = true

	return &constants.AnswerState{State: constants.SuccessState, Data: "Remarks were successfully saved.", Adv: "none"}
}

// FixMistakenVehicleIdentification corrects the scenario where a user submitted an
// inspection using the wrong vehicle ID. It executes entirely within a single transaction
// and performs the following steps:
//
//  1. Migrate all active inspection data from wrongVehicleId → correctVehicleId.
//  2. Mark correctVehicleId as inspected/tallied on manifest_vehicles.
//  3. Check whether wrongVehicleId has a prior archive batch (i.e. the archive block in
//     InsertInspectionTallyRemarks ran and displaced its legitimate historical record when
//     the mistaken submission arrived).
//  4. If history exists, restore the most-recent archived batch back into the active tables
//     for wrongVehicleId and remove those rows from the history tables to prevent duplication.
//  5. Update manifest_vehicles for wrongVehicleId accordingly (restored → 'yes'/'yes';
//     no history → reset to 'no'/'no').
func FixMistakenVehicleIdentification(wrongVehicleId string, correctVehicleId string) *constants.AnswerState {
	if wrongVehicleId == "" || correctVehicleId == "" || wrongVehicleId == correctVehicleId {
		return &constants.AnswerState{
			State: constants.ErrorState,
			Data:  "Invalid vehicle IDs: both must be non-empty and distinct",
			Adv:   "none",
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	db, er := gendb.InitDb()
	if er != nil {
		slog.Error(fmt.Sprintf("Database connection error: %v", er))
		return &constants.AnswerState{State: constants.ErrorState, Data: er.Error(), Adv: "none"}
	}

	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		slog.Error(fmt.Sprintf("Transaction start error: %v", err))
		return &constants.AnswerState{State: constants.ErrorState, Data: "Failed to start transaction", Adv: "none"}
	}

	var committed bool
	defer func() {
		if !committed {
			if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
				slog.Error(fmt.Sprintf("Failed to rollback transaction: %v", err))
			}
		}
	}()

	// ── Step 1: Re-attribute active data wrongVehicleId → correctVehicleId ────
	// inspection_image and onboard_packages_media carry no vehicle_id; they follow
	// their parent rows automatically since inspection_id / package_id are unchanged.
	for _, tbl := range [5]string{
		"vehicles_talling",
		"vehicles_inspection",
		"onboard_packages",
		"inspection_remarks",
		"vehicle_galllery",
	} {
		if _, err = tx.ExecContext(ctx,
			"UPDATE "+tbl+" SET vehicle_id = ? WHERE vehicle_id = ?",
			correctVehicleId, wrongVehicleId); err != nil {
			slog.Error(fmt.Sprintf("Migrate %s error: %v", tbl, err))
			return &constants.AnswerState{State: constants.ErrorState, Data: "Failed to migrate data to correct vehicle", Adv: "none"}
		}
	}

	// ── Step 2: Mark the correct vehicle as fully inspected / tallied ──────────
	if _, err = tx.ExecContext(ctx,
		"UPDATE manifest_vehicles SET inspection_status = 'yes', tallied_status = 'yes' WHERE vehicle_id = ?",
		correctVehicleId); err != nil {
		slog.Error(fmt.Sprintf("Status update (correct vehicle) error: %v", err))
		return &constants.AnswerState{State: constants.ErrorState, Data: "Failed to update correct vehicle status", Adv: "none"}
	}

	// ── Step 3: Check for a prior archive batch on the wrong vehicle ───────────
	// The archive block in InsertInspectionTallyRemarks stamps every archived row
	// with NOW() at the time of archival. We use the MAX per table as the batch key.
	var tallyArchivedAt nullArchiveTime
	if err = tx.QueryRowContext(ctx,
		"SELECT MAX(archived_at) FROM vehicles_talling_history WHERE vehicle_id = ?",
		wrongVehicleId).Scan(&tallyArchivedAt); err != nil {
		slog.Error(fmt.Sprintf("History check (tally) error: %v", err))
		return &constants.AnswerState{State: constants.ErrorState, Data: "Failed to check vehicle tally history", Adv: "none"}
	}

	if tallyArchivedAt.Valid {
		// ── Step 4: Restore the most-recent archived batch ──────────────────────

		// 4a. Restore vehicles_talling
		if _, err = tx.ExecContext(ctx, `
			INSERT INTO vehicles_talling
			    (tally_id, vehicle_id, manifest_id, maker_id, body_id, image_link, deck_number, number_of_keys, key_type, tallied_time)
			SELECT tally_id, vehicle_id, manifest_id, maker_id, body_id, image_link, deck_number, number_of_keys, key_type, tallied_time
			FROM vehicles_talling_history
			WHERE vehicle_id = ? AND archived_at = ?`,
			wrongVehicleId, tallyArchivedAt.Time); err != nil {
			slog.Error(fmt.Sprintf("Restore vehicles_talling error: %v", err))
			return &constants.AnswerState{State: constants.ErrorState, Data: "Failed to restore tally record", Adv: "none"}
		}

		// 4b. Restore vehicles_inspection (use its own MAX to be precise)
		var inspArchivedAt nullArchiveTime
		if err = tx.QueryRowContext(ctx,
			"SELECT MAX(archived_at) FROM vehicles_inspection_history WHERE vehicle_id = ?",
			wrongVehicleId).Scan(&inspArchivedAt); err != nil {
			slog.Error(fmt.Sprintf("History check (inspection) error: %v", err))
			return &constants.AnswerState{State: constants.ErrorState, Data: "Failed to check inspection history", Adv: "none"}
		}
		if inspArchivedAt.Valid {
			if _, err = tx.ExecContext(ctx, `
				INSERT INTO vehicles_inspection
				    (inspection_id, vehicle_id, check_id, user_id, status, check_time)
				SELECT inspection_id, vehicle_id, check_id, user_id, status, check_time
				FROM vehicles_inspection_history
				WHERE vehicle_id = ? AND archived_at = ?`,
				wrongVehicleId, inspArchivedAt.Time); err != nil {
				slog.Error(fmt.Sprintf("Restore vehicles_inspection error: %v", err))
				return &constants.AnswerState{State: constants.ErrorState, Data: "Failed to restore inspection records", Adv: "none"}
			}

			// 4c. Restore inspection_image via its parent inspection_id values
			if _, err = tx.ExecContext(ctx, `
				INSERT INTO inspection_image
				    (image_id, image_link, inspection_id, creation_time)
				SELECT image_id, image_link, inspection_id, creation_time
				FROM inspection_image_history
				WHERE inspection_id IN (
				    SELECT inspection_id FROM vehicles_inspection_history
				    WHERE vehicle_id = ? AND archived_at = ?
				)`, wrongVehicleId, inspArchivedAt.Time); err != nil {
				slog.Error(fmt.Sprintf("Restore inspection_image error: %v", err))
				return &constants.AnswerState{State: constants.ErrorState, Data: "Failed to restore inspection images", Adv: "none"}
			}

			// Purge the restored rows from their history tables
			if _, err = tx.ExecContext(ctx, `
				DELETE FROM inspection_image_history
				WHERE inspection_id IN (
				    SELECT inspection_id FROM vehicles_inspection_history
				    WHERE vehicle_id = ? AND archived_at = ?
				)`, wrongVehicleId, inspArchivedAt.Time); err != nil {
				slog.Error(fmt.Sprintf("Purge inspection_image_history error: %v", err))
				return &constants.AnswerState{State: constants.ErrorState, Data: "Failed to purge inspection image history", Adv: "none"}
			}
			if _, err = tx.ExecContext(ctx,
				"DELETE FROM vehicles_inspection_history WHERE vehicle_id = ? AND archived_at = ?",
				wrongVehicleId, inspArchivedAt.Time); err != nil {
				slog.Error(fmt.Sprintf("Purge vehicles_inspection_history error: %v", err))
				return &constants.AnswerState{State: constants.ErrorState, Data: "Failed to purge inspection history", Adv: "none"}
			}
		}

		// 4d. Restore onboard_packages (use its own MAX)
		var packArchivedAt nullArchiveTime
		if err = tx.QueryRowContext(ctx,
			"SELECT MAX(archived_at) FROM onboard_packages_history WHERE vehicle_id = ?",
			wrongVehicleId).Scan(&packArchivedAt); err != nil {
			slog.Error(fmt.Sprintf("History check (packages) error: %v", err))
			return &constants.AnswerState{State: constants.ErrorState, Data: "Failed to check onboard packages history", Adv: "none"}
		}
		if packArchivedAt.Valid {
			if _, err = tx.ExecContext(ctx, `
				INSERT INTO onboard_packages
				    (package_id, title, remark, vehicle_id)
				SELECT package_id, title, remark, vehicle_id
				FROM onboard_packages_history
				WHERE vehicle_id = ? AND archived_at = ?`,
				wrongVehicleId, packArchivedAt.Time); err != nil {
				slog.Error(fmt.Sprintf("Restore onboard_packages error: %v", err))
				return &constants.AnswerState{State: constants.ErrorState, Data: "Failed to restore onboard packages", Adv: "none"}
			}

			// 4e. Restore onboard_packages_media via parent package_id values
			if _, err = tx.ExecContext(ctx, `
				INSERT INTO onboard_packages_media
				    (media_id, media_type, media_link, package_id)
				SELECT media_id, media_type, media_link, package_id
				FROM onboard_packages_media_history
				WHERE package_id IN (
				    SELECT package_id FROM onboard_packages_history
				    WHERE vehicle_id = ? AND archived_at = ?
				)`, wrongVehicleId, packArchivedAt.Time); err != nil {
				slog.Error(fmt.Sprintf("Restore onboard_packages_media error: %v", err))
				return &constants.AnswerState{State: constants.ErrorState, Data: "Failed to restore onboard package media", Adv: "none"}
			}

			// Purge the restored rows from their history tables
			if _, err = tx.ExecContext(ctx, `
				DELETE FROM onboard_packages_media_history
				WHERE package_id IN (
				    SELECT package_id FROM onboard_packages_history
				    WHERE vehicle_id = ? AND archived_at = ?
				)`, wrongVehicleId, packArchivedAt.Time); err != nil {
				slog.Error(fmt.Sprintf("Purge onboard_packages_media_history error: %v", err))
				return &constants.AnswerState{State: constants.ErrorState, Data: "Failed to purge onboard package media history", Adv: "none"}
			}
			if _, err = tx.ExecContext(ctx,
				"DELETE FROM onboard_packages_history WHERE vehicle_id = ? AND archived_at = ?",
				wrongVehicleId, packArchivedAt.Time); err != nil {
				slog.Error(fmt.Sprintf("Purge onboard_packages_history error: %v", err))
				return &constants.AnswerState{State: constants.ErrorState, Data: "Failed to purge onboard packages history", Adv: "none"}
			}
		}

		// 4f. Restore inspection_remarks
		var rmkArchivedAt nullArchiveTime
		if err = tx.QueryRowContext(ctx,
			"SELECT MAX(archived_at) FROM inspection_remarks_history WHERE vehicle_id = ?",
			wrongVehicleId).Scan(&rmkArchivedAt); err != nil {
			slog.Error(fmt.Sprintf("History check (remarks) error: %v", err))
			return &constants.AnswerState{State: constants.ErrorState, Data: "Failed to check remarks history", Adv: "none"}
		}
		if rmkArchivedAt.Valid {
			if _, err = tx.ExecContext(ctx, `
				INSERT INTO inspection_remarks
				    (remark_id, vehicle_id, user_id, remark, remark_type, image_link, remark_time)
				SELECT remark_id, vehicle_id, user_id, remark, remark_type, image_link, remark_time
				FROM inspection_remarks_history
				WHERE vehicle_id = ? AND archived_at = ?`,
				wrongVehicleId, rmkArchivedAt.Time); err != nil {
				slog.Error(fmt.Sprintf("Restore inspection_remarks error: %v", err))
				return &constants.AnswerState{State: constants.ErrorState, Data: "Failed to restore inspection remarks", Adv: "none"}
			}
			if _, err = tx.ExecContext(ctx,
				"DELETE FROM inspection_remarks_history WHERE vehicle_id = ? AND archived_at = ?",
				wrongVehicleId, rmkArchivedAt.Time); err != nil {
				slog.Error(fmt.Sprintf("Purge inspection_remarks_history error: %v", err))
				return &constants.AnswerState{State: constants.ErrorState, Data: "Failed to purge remarks history", Adv: "none"}
			}
		}

		// 4g. Restore vehicle_galllery (preserving the intentional triple-l spelling)
		var galleryArchivedAt nullArchiveTime
		if err = tx.QueryRowContext(ctx,
			"SELECT MAX(archived_at) FROM vehicle_galllery_history WHERE vehicle_id = ?",
			wrongVehicleId).Scan(&galleryArchivedAt); err != nil {
			slog.Error(fmt.Sprintf("History check (gallery) error: %v", err))
			return &constants.AnswerState{State: constants.ErrorState, Data: "Failed to check gallery history", Adv: "none"}
		}
		if galleryArchivedAt.Valid {
			if _, err = tx.ExecContext(ctx, `
				INSERT INTO vehicle_galllery
				    (media_id, media_link, media_type, remark, vehicle_id, status)
				SELECT media_id, media_link, media_type, remark, vehicle_id, status
				FROM vehicle_galllery_history
				WHERE vehicle_id = ? AND archived_at = ?`,
				wrongVehicleId, galleryArchivedAt.Time); err != nil {
				slog.Error(fmt.Sprintf("Restore vehicle_galllery error: %v", err))
				return &constants.AnswerState{State: constants.ErrorState, Data: "Failed to restore vehicle gallery", Adv: "none"}
			}
			if _, err = tx.ExecContext(ctx,
				"DELETE FROM vehicle_galllery_history WHERE vehicle_id = ? AND archived_at = ?",
				wrongVehicleId, galleryArchivedAt.Time); err != nil {
				slog.Error(fmt.Sprintf("Purge vehicle_galllery_history error: %v", err))
				return &constants.AnswerState{State: constants.ErrorState, Data: "Failed to purge gallery history", Adv: "none"}
			}
		}

		// Purge the tally history row last (it is the anchor for the archive batch).
		if _, err = tx.ExecContext(ctx,
			"DELETE FROM vehicles_talling_history WHERE vehicle_id = ? AND archived_at = ?",
			wrongVehicleId, tallyArchivedAt.Time); err != nil {
			slog.Error(fmt.Sprintf("Purge vehicles_talling_history error: %v", err))
			return &constants.AnswerState{State: constants.ErrorState, Data: "Failed to purge tally history", Adv: "none"}
		}

		// ── Step 5a: wrongVehicleId now has its original data restored ─────────
		if _, err = tx.ExecContext(ctx,
			"UPDATE manifest_vehicles SET inspection_status = 'yes', tallied_status = 'yes' WHERE vehicle_id = ?",
			wrongVehicleId); err != nil {
			slog.Error(fmt.Sprintf("Status update (wrong vehicle restore) error: %v", err))
			return &constants.AnswerState{State: constants.ErrorState, Data: "Failed to restore wrong vehicle status", Adv: "none"}
		}
	} else {
		// ── Step 5b: wrongVehicleId was never legitimately inspected ──────────
		// Reset its manifest status so it re-appears as pending in all dashboards.
		if _, err = tx.ExecContext(ctx, `
			UPDATE manifest_vehicles SET
			    inspection_status = 'no',
			    tallied_status    = 'no',
			    inspection_time   = '1000-01-01 00:00:00',
			    tallied_time      = '1000-01-01 00:00:00'
			WHERE vehicle_id = ?`, wrongVehicleId); err != nil {
			slog.Error(fmt.Sprintf("Status reset (wrong vehicle) error: %v", err))
			return &constants.AnswerState{State: constants.ErrorState, Data: "Failed to reset wrong vehicle status", Adv: "none"}
		}
	}

	if err := tx.Commit(); err != nil {
		slog.Error(fmt.Sprintf("Transaction commit error: %v", err))
		return &constants.AnswerState{State: constants.ErrorState, Data: "Failed to commit changes. Please try again.", Adv: "none"}
	}
	committed = true

	return &constants.AnswerState{
		State: constants.SuccessState,
		Data:  "Vehicle identification corrected successfully.",
		Adv:   "none",
	}
}

// TransferVehicleData moves inspection data from sourceVehicleId to targetVehicleId.
//
// dataSource controls which set of records targetVehicleId receives:
//
//   - "active":  the live active-table rows of sourceVehicleId are re-attributed to
//     targetVehicleId. The source's active records are gone after this call.
//     All source history is also purged and source is reset to uninspected.
//
//   - "history": the most-recent archived batch for sourceVehicleId is restored into
//     the active tables under targetVehicleId. The source keeps its own active records
//     intact. All source history is purged after the restore.
//
// targetVehicleId must have no existing inspection records in the active tables.
func TransferVehicleData(sourceVehicleId, targetVehicleId, dataSource string) *constants.AnswerState {
	if sourceVehicleId == "" || targetVehicleId == "" || sourceVehicleId == targetVehicleId {
		return &constants.AnswerState{State: constants.ErrorState, Data: "Invalid vehicle IDs: both must be non-empty and distinct", Adv: "none"}
	}
	if dataSource != "active" && dataSource != "history" {
		return &constants.AnswerState{State: constants.ErrorState, Data: `dataSource must be "active" or "history"`, Adv: "none"}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	db, er := gendb.InitDb()
	if er != nil {
		slog.Error(fmt.Sprintf("Database connection error: %s", er.Error()))
		return &constants.AnswerState{State: constants.ErrorState, Data: er.Error(), Adv: "none"}
	}

	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		slog.Error(fmt.Sprintf("Transaction start error: %s", err.Error()))
		return &constants.AnswerState{State: constants.ErrorState, Data: "Failed to start transaction", Adv: "none"}
	}

	var committed bool
	defer func() {
		if !committed {
			if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
				slog.Error(fmt.Sprintf("Failed to rollback transaction: %s", err.Error()))
			}
		}
	}()

	switch dataSource {

	// ── "active": re-attribute all live rows to the target vehicle ─────────────
	// inspection_image and onboard_packages_media have no vehicle_id column; they
	// follow automatically because their parent inspection_id / package_id keys are
	// unchanged by the UPDATE.
	case "active":
		for _, tbl := range [5]string{
			"vehicles_talling",
			"vehicles_inspection",
			"onboard_packages",
			"inspection_remarks",
			"vehicle_galllery",
		} {
			if _, err = tx.ExecContext(ctx,
				"UPDATE "+tbl+" SET vehicle_id = ? WHERE vehicle_id = ?",
				targetVehicleId, sourceVehicleId); err != nil {
				slog.Error(fmt.Sprintf("Transfer active %s error: %v", tbl, err))
				return &constants.AnswerState{State: constants.ErrorState, Data: "Failed to transfer active records", Adv: "none"}
			}
		}

	// ── "history": restore the most-recent archive batch to the target vehicle ─
	// vehicle_id is substituted with targetVehicleId in every INSERT … SELECT so
	// the restored rows are owned by the correct vehicle from the start.
	// inspection_image and onboard_packages_media carry no vehicle_id, so they are
	// copied verbatim using the same inspection_id / package_id values.
	case "history":
		var tallyAt nullArchiveTime
		if err = tx.QueryRowContext(ctx,
			"SELECT MAX(archived_at) FROM vehicles_talling_history WHERE vehicle_id = ?",
			sourceVehicleId).Scan(&tallyAt); err != nil {
			slog.Error(fmt.Sprintf("History read (tally) error: %s", err.Error()))
			return &constants.AnswerState{State: constants.ErrorState, Data: "Failed to read tally history", Adv: "none"}
		}
		if !tallyAt.Valid {
			return &constants.AnswerState{State: constants.ErrorState, Data: "No history found for source vehicle", Adv: "none"}
		}

		if _, err = tx.ExecContext(ctx, `
			INSERT INTO vehicles_talling
			    (tally_id, vehicle_id, manifest_id, maker_id, body_id, image_link, deck_number, number_of_keys, key_type, tallied_time)
			SELECT tally_id, ?, manifest_id, maker_id, body_id, image_link, deck_number, number_of_keys, key_type, tallied_time
			FROM vehicles_talling_history
			WHERE vehicle_id = ? AND archived_at = ?`,
			targetVehicleId, sourceVehicleId, tallyAt.Time); err != nil {
			slog.Error(fmt.Sprintf("Restore tally to target error: %s", err.Error()))
			return &constants.AnswerState{State: constants.ErrorState, Data: "Failed to restore tally record", Adv: "none"}
		}

		var inspAt nullArchiveTime
		if err = tx.QueryRowContext(ctx,
			"SELECT MAX(archived_at) FROM vehicles_inspection_history WHERE vehicle_id = ?",
			sourceVehicleId).Scan(&inspAt); err != nil {
			slog.Error(fmt.Sprintf("History read (inspection) error: %s", err.Error()))
			return &constants.AnswerState{State: constants.ErrorState, Data: "Failed to read inspection history", Adv: "none"}
		}
		if inspAt.Valid {
			if _, err = tx.ExecContext(ctx, `
				INSERT INTO vehicles_inspection
				    (inspection_id, vehicle_id, check_id, user_id, status, check_time)
				SELECT inspection_id, ?, check_id, user_id, status, check_time
				FROM vehicles_inspection_history
				WHERE vehicle_id = ? AND archived_at = ?`,
				targetVehicleId, sourceVehicleId, inspAt.Time); err != nil {
				slog.Error(fmt.Sprintf("Restore inspection to target error: %s", err.Error()))
				return &constants.AnswerState{State: constants.ErrorState, Data: "Failed to restore inspection records", Adv: "none"}
			}
			if _, err = tx.ExecContext(ctx, `
				INSERT INTO inspection_image
				    (image_id, image_link, inspection_id, creation_time)
				SELECT image_id, image_link, inspection_id, creation_time
				FROM inspection_image_history
				WHERE inspection_id IN (
				    SELECT inspection_id FROM vehicles_inspection_history
				    WHERE vehicle_id = ? AND archived_at = ?
				)`, sourceVehicleId, inspAt.Time); err != nil {
				slog.Error(fmt.Sprintf("Restore inspection_image to target error: %s", err.Error()))
				return &constants.AnswerState{State: constants.ErrorState, Data: "Failed to restore inspection images", Adv: "none"}
			}
		}

		var packAt nullArchiveTime
		if err = tx.QueryRowContext(ctx,
			"SELECT MAX(archived_at) FROM onboard_packages_history WHERE vehicle_id = ?",
			sourceVehicleId).Scan(&packAt); err != nil {
			slog.Error(fmt.Sprintf("History read (packages) error: %s", err.Error()))
			return &constants.AnswerState{State: constants.ErrorState, Data: "Failed to read package history", Adv: "none"}
		}
		if packAt.Valid {
			if _, err = tx.ExecContext(ctx, `
				INSERT INTO onboard_packages
				    (package_id, title, remark, vehicle_id)
				SELECT package_id, title, remark, ?
				FROM onboard_packages_history
				WHERE vehicle_id = ? AND archived_at = ?`,
				targetVehicleId, sourceVehicleId, packAt.Time); err != nil {
				slog.Error(fmt.Sprintf("Restore packages to target error: %s", err.Error()))
				return &constants.AnswerState{State: constants.ErrorState, Data: "Failed to restore onboard packages", Adv: "none"}
			}
			if _, err = tx.ExecContext(ctx, `
				INSERT INTO onboard_packages_media
				    (media_id, media_type, media_link, package_id)
				SELECT media_id, media_type, media_link, package_id
				FROM onboard_packages_media_history
				WHERE package_id IN (
				    SELECT package_id FROM onboard_packages_history
				    WHERE vehicle_id = ? AND archived_at = ?
				)`, sourceVehicleId, packAt.Time); err != nil {
				slog.Error(fmt.Sprintf("Restore package media to target error: %s", err.Error()))
				return &constants.AnswerState{State: constants.ErrorState, Data: "Failed to restore onboard package media", Adv: "none"}
			}
		}

		var rmkAt nullArchiveTime
		if err = tx.QueryRowContext(ctx,
			"SELECT MAX(archived_at) FROM inspection_remarks_history WHERE vehicle_id = ?",
			sourceVehicleId).Scan(&rmkAt); err != nil {
			slog.Error(fmt.Sprintf("History read (remarks) error: %s", err.Error()))
			return &constants.AnswerState{State: constants.ErrorState, Data: "Failed to read remarks history", Adv: "none"}
		}
		if rmkAt.Valid {
			if _, err = tx.ExecContext(ctx, `
				INSERT INTO inspection_remarks
				    (remark_id, vehicle_id, user_id, remark, remark_type, image_link, remark_time)
				SELECT remark_id, ?, user_id, remark, remark_type, image_link, remark_time
				FROM inspection_remarks_history
				WHERE vehicle_id = ? AND archived_at = ?`,
				targetVehicleId, sourceVehicleId, rmkAt.Time); err != nil {
				slog.Error(fmt.Sprintf("Restore remarks to target error: %s", err.Error()))
				return &constants.AnswerState{State: constants.ErrorState, Data: "Failed to restore inspection remarks", Adv: "none"}
			}
		}

		var gallAt nullArchiveTime
		if err = tx.QueryRowContext(ctx,
			"SELECT MAX(archived_at) FROM vehicle_galllery_history WHERE vehicle_id = ?",
			sourceVehicleId).Scan(&gallAt); err != nil {
			slog.Error(fmt.Sprintf("History read (gallery) error: %s", err.Error()))
			return &constants.AnswerState{State: constants.ErrorState, Data: "Failed to read gallery history", Adv: "none"}
		}
		if gallAt.Valid {
			if _, err = tx.ExecContext(ctx, `
				INSERT INTO vehicle_galllery
				    (media_id, media_link, media_type, remark, vehicle_id, status)
				SELECT media_id, media_link, media_type, remark, ?, status
				FROM vehicle_galllery_history
				WHERE vehicle_id = ? AND archived_at = ?`,
				targetVehicleId, sourceVehicleId, gallAt.Time); err != nil {
				slog.Error(fmt.Sprintf("Restore gallery to target error: %s", err.Error()))
				return &constants.AnswerState{State: constants.ErrorState, Data: "Failed to restore vehicle gallery", Adv: "none"}
			}
		}
	}

	// Mark targetVehicleId as fully inspected / tallied regardless of which source was used.
	if _, err = tx.ExecContext(ctx,
		"UPDATE manifest_vehicles SET inspection_status = 'yes', tallied_status = 'yes' WHERE vehicle_id = ?",
		targetVehicleId); err != nil {
		slog.Error(fmt.Sprintf("Status update (target) error: %s", err.Error()))
		return &constants.AnswerState{State: constants.ErrorState, Data: "Failed to update target vehicle status", Adv: "none"}
	}

	// ── Purge ALL history for sourceVehicleId ─────────────────────────────────
	// Child tables first to respect FK order.
	if _, err = tx.ExecContext(ctx, `
		DELETE FROM inspection_image_history
		WHERE inspection_id IN (
		    SELECT inspection_id FROM vehicles_inspection_history WHERE vehicle_id = ?
		)`, sourceVehicleId); err != nil {
		slog.Error(fmt.Sprintf("Purge inspection_image_history error: %s", err.Error()))
		return &constants.AnswerState{State: constants.ErrorState, Data: "Failed to purge inspection image history", Adv: "none"}
	}
	if _, err = tx.ExecContext(ctx, `
		DELETE FROM onboard_packages_media_history
		WHERE package_id IN (
		    SELECT package_id FROM onboard_packages_history WHERE vehicle_id = ?
		)`, sourceVehicleId); err != nil {
		slog.Error(fmt.Sprintf("Purge onboard_packages_media_history error: %s", err.Error()))
		return &constants.AnswerState{State: constants.ErrorState, Data: "Failed to purge onboard package media history", Adv: "none"}
	}
	for _, tbl := range [5]string{
		"vehicles_inspection_history",
		"onboard_packages_history",
		"inspection_remarks_history",
		"vehicle_galllery_history",
		"vehicles_talling_history",
	} {
		if _, err = tx.ExecContext(ctx, "DELETE FROM "+tbl+" WHERE vehicle_id = ?", sourceVehicleId); err != nil {
			slog.Error(fmt.Sprintf("Purge %s error: %s", tbl, err.Error()))
			return &constants.AnswerState{State: constants.ErrorState, Data: "Failed to purge history records", Adv: "none"}
		}
	}

	// When active data was transferred the source has no records left at all — reset
	// its manifest status. When history was transferred the source still has its own
	// active inspection intact, so leave its status unchanged.
	if dataSource == "active" {
		if _, err = tx.ExecContext(ctx, `
			UPDATE manifest_vehicles SET
			    inspection_status = 'no',
			    tallied_status    = 'no',
			    inspection_time   = '1000-01-01 00:00:00',
			    tallied_time      = '1000-01-01 00:00:00'
			WHERE vehicle_id = ?`, sourceVehicleId); err != nil {
			slog.Error(fmt.Sprintf("Status reset (source) error: %s", err.Error()))
			return &constants.AnswerState{State: constants.ErrorState, Data: "Failed to reset source vehicle status", Adv: "none"}
		}
	}

	if err := tx.Commit(); err != nil {
		slog.Error(fmt.Sprintf("Transaction commit error: %s", err.Error()))
		return &constants.AnswerState{State: constants.ErrorState, Data: "Failed to commit changes. Please try again.", Adv: "none"}
	}
	committed = true

	return &constants.AnswerState{
		State: constants.SuccessState,
		Data:  "Vehicle data transferred successfully.",
		Adv:   "none",
	}
}
