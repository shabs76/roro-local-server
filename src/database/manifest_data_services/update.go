package manifestdataservices

import (
	"github.com/shabs76/roro-local-server/constants"
	"github.com/shabs76/roro-local-server/constants/modules/manifest"
	"github.com/shabs76/roro-local-server/gendb"
)

func UpdateVehicleIdToRemote(oldID, newId string) *constants.AnswerState {
	qr := "UPDATE `manifest_vehicles` SET `vehicle_id`= ?,`is_added_later`= ? WHERE `vehicle_id` = ?"
	vals := []any{newId, manifest.InspectionStatus.No, oldID}
	stx := gendb.UpdateGeneral(qr, vals)
	if stx.State != constants.SuccessState {
		// No changed row means the old id does not exist: the remote id was not stored.
		return &constants.AnswerState{State: constants.ErrorState, Data: "Failed to store remote id: " + stx.Data, Adv: "none"}
	}
	return &constants.AnswerState{State: constants.SuccessState, Data: "Vehicle ID updated successfully", Adv: "okay"}
}

func UpdatePackageIdToRemote(oldID, newId string) *constants.AnswerState {
	qr := "UPDATE `manifest_packages` SET `package_id`= ?,`is_added_later`= ? WHERE `package_id` = ?"
	vals := []any{newId, manifest.InspectionStatus.No, oldID}
	stx := gendb.UpdateGeneral(qr, vals)
	if stx.State != constants.SuccessState {
		// No changed row means the old id does not exist: the remote id was not stored.
		return &constants.AnswerState{State: constants.ErrorState, Data: "Failed to store remote id: " + stx.Data, Adv: "none"}
	}
	return &constants.AnswerState{State: constants.SuccessState, Data: "Package ID updated successfully", Adv: "okay"}
}

func UpdateVehiclePublishedStatus(vehicleId string, isPublished bool) *constants.AnswerState {
	qr := "UPDATE `manifest_vehicles` SET `is_published`= ? WHERE `vehicle_id` = ?"
	publishVal := "no"
	if isPublished {
		publishVal = "yes"
	}
	vals := []any{publishVal, vehicleId}
	stx := gendb.UpdateGeneral(qr, vals)
	if stx.State != constants.SuccessState && stx.Adv != "okay" {
		return stx
	}
	return &constants.AnswerState{State: constants.SuccessState, Data: "Vehicle published status updated successfully", Adv: "okay"}
}

func UpdatePackagePublishedStatus(packageId string, isPublished bool) *constants.AnswerState {
	qr := "UPDATE `manifest_packages` SET `is_published`= ? WHERE `package_id` = ?"
	publishVal := "no"
	if isPublished {
		publishVal = "yes"
	}
	vals := []any{publishVal, packageId}
	stx := gendb.UpdateGeneral(qr, vals)
	if stx.State != constants.SuccessState && stx.Adv != "okay" {
		return stx
	}
	return &constants.AnswerState{State: constants.SuccessState, Data: "Package published status updated successfully", Adv: "okay"}
}
