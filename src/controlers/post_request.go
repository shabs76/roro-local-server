package controlers

import (
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shabs76/roro-local-server/constants"
	"github.com/shabs76/roro-local-server/constants/modules/manifest"
	"github.com/shabs76/roro-local-server/constants/modules/users"
	manifestdataservices "github.com/shabs76/roro-local-server/database/manifest_data_services"
	"github.com/shabs76/roro-local-server/specials"
)

func SaveVehicleInspectionDetails(c *gin.Context) {
	userd, exist := c.Get("userID")
	if !exist {
		c.JSON(http.StatusInternalServerError, gin.H{"state": constants.ErrorState, "data": "Failed to obtain security details"})
		return
	}

	userId, ok := userd.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"state": constants.ErrorState, "data": "Failed to obtain security details #2"})
		return
	}

	var req manifest.InspectionChecksRequest

	if er := c.ShouldBindJSON(&req); er != nil {
		slog.Error(er.Error())
		log.Println(c.Request.Body)
		c.JSON(http.StatusBadRequest, gin.H{"state": constants.ErrorState, "data": "Invalid data was sent"})
		return
	}

	// validate inspection time for datetime format
	if req.InspectionTime == "" {
		c.JSON(http.StatusBadRequest, gin.H{"state": constants.ErrorState, "data": "Inspection time is required"})
		return
	}
	_, err := time.Parse("2006-01-02 15:04:05", req.InspectionTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"state": constants.ErrorState, "data": "Invalid inspection time format, expected YYYY-MM-DD HH:MM:SS"})
		return
	}

	st := manifestdataservices.InsertInspectionTallyRemarks(req, userId)
	if st.State != constants.SuccessState {
		c.JSON(http.StatusInternalServerError, st)
		return
	}

	c.JSON(http.StatusOK, st)

}

func AddSingleVehicleToManifest(c *gin.Context) {
	// user details from context
	userd, exist := c.Get("user")
	if !exist {
		c.JSON(http.StatusInternalServerError, gin.H{"state": constants.ErrorState, "data": "Failed to obtain security details"})
		return
	}

	user, ok := userd.(users.UserData)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"state": constants.ErrorState, "data": "Failed to obtain security details #2"})
		return
	}

	if user.RoleNumber > 300 {
		c.JSON(http.StatusOK, gin.H{"state": constants.ErrorState, "data": "Sorry, you have no permission to perform this action"})
		return
	}

	var req manifest.AddVehicleRequest

	if er := c.ShouldBindJSON(&req); er != nil {
		slog.Error(er.Error())
		log.Println(c.Request.Body)
		c.JSON(http.StatusBadRequest, gin.H{"state": constants.ErrorState, "data": "Invalid data was sent"})
		return
	}

	// check if chasis number already exists for the manifest
	subQuery := " manifest_id = ? AND chasis_number = ? "
	vals := []any{req.ManifestId, req.ChasisNumber}
	stCheck, existingVehicles := manifestdataservices.SelectVehicleInfo(subQuery, vals)
	if stCheck.State != constants.SuccessState {
		c.JSON(http.StatusInternalServerError, stCheck)
		return
	}
	if len(existingVehicles) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"state": constants.ErrorState, "data": "A vehicle with the same chasis number already exists for this manifest"})
		return
	}

	id := specials.RandomString(32, "_VEH")
	curTime := time.Now().Format("2006-01-02 15:04:05")
	data := manifest.VehiclesDetailsToShow{
		VehicleId:        id,
		ManifestId:       req.ManifestId,
		ChasisNumber:     req.ChasisNumber,
		VehicleModel:     req.VehicleModel,
		Description:      req.Description,
		Weight:           float32(req.Weight),
		BLNumber:         req.BLNumber,
		CreationDate:     curTime,
		InspectionStatus: "no",
		TalliedStatus:    "no",
		DischargeStatus:  "no",
		IsOverLand:       req.IsOverLand,
		TalliedTime:      "1000-01-01 00:00:00",
		InspectionTime:   "1000-01-01 00:00:00",
		DischargeTime:    "1000-01-01 00:00:00",
		NumberOfKeys:     0,
	}

	st := manifestdataservices.InsertVehiclesDetails([]manifest.VehiclesDetailsToShow{data}, true)
	if st.State != constants.SuccessState {
		c.JSON(http.StatusInternalServerError, st)
		return
	}
	c.JSON(http.StatusOK, st)
}

func AddSinglePackageToManifest(c *gin.Context) {
	// user details from context
	userd, exist := c.Get("user")
	if !exist {
		c.JSON(http.StatusInternalServerError, gin.H{"state": constants.ErrorState, "data": "Failed to obtain security details"})
		return
	}

	user, ok := userd.(users.UserData)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"state": constants.ErrorState, "data": "Failed to obtain security details #2"})
		return
	}

	if user.RoleNumber > 300 {
		c.JSON(http.StatusOK, gin.H{"state": constants.ErrorState, "data": "Sorry, you have no permission to perform this action"})
		return
	}

	var req manifest.AddPackageRequest
	if er := c.ShouldBindJSON(&req); er != nil {
		slog.Error(er.Error())
		c.JSON(http.StatusBadRequest, gin.H{"state": constants.ErrorState, "data": "Invalid data was sent"})
		return
	}

	id := specials.RandomString(32, "_PKG")

	// check if package number already exists for the manifest
	subQuery := " manifest_id = ? AND package_number = ? "
	vals := []any{req.ManifestId, req.PackageNumber}
	stCheck, existingPackages := manifestdataservices.SelectPackageInfo(subQuery, vals)
	if stCheck.State != constants.SuccessState {
		c.JSON(http.StatusInternalServerError, stCheck)
		return
	}
	if len(existingPackages) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"state": constants.ErrorState, "data": "A package with the same number already exists for this manifest"})
		return
	}

	data := manifest.PackageManifestInfo{
		PackageId:     id,
		ManifestId:    req.ManifestId,
		PackageNumber: req.PackageNumber,
		BLNumber:      req.BLNumber,
		Description:   req.Description,
		IsInspected:   "no",
		IsAddedLater:  "yes",
		CreationTime:  time.Now().Format("2006-01-02 15:04:05"),
	}

	st := manifestdataservices.InsertPackageDetails([]manifest.PackageManifestInfo{data}, true)
	if st.State != constants.SuccessState {
		c.JSON(http.StatusInternalServerError, st)
		return
	}
	c.JSON(http.StatusOK, st)
}

func SavePackageInspectionDetails(c *gin.Context) {
	userd, exist := c.Get("user")
	if !exist {
		c.JSON(http.StatusInternalServerError, gin.H{"state": constants.ErrorState, "data": "Failed to obtain security details"})
		return
	}

	user, ok := userd.(users.UserData)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"state": constants.ErrorState, "data": "Failed to obtain security details #2"})
		return
	}

	var req manifest.PackageInspectionSaveRequest

	if er := c.ShouldBindJSON(&req); er != nil {
		c.JSON(http.StatusBadRequest, gin.H{"state": constants.ErrorState, "data": "Invalid data was sent"})
		return
	}

	// validate inspection time for datetime format
	if req.InspectionTime == "" {
		c.JSON(http.StatusBadRequest, gin.H{"state": constants.ErrorState, "data": "Inspection time is required"})
		return
	}
	_, err := time.Parse("2006-01-02 15:04:05", req.InspectionTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"state": constants.ErrorState, "data": "Invalid inspection time format, expected YYYY-MM-DD HH:MM:SS"})
		return
	}

	st := manifestdataservices.InsertPackageInspection(req, user.UserID)
	if st.State != constants.SuccessState {
		c.JSON(http.StatusInternalServerError, st)
		return
	}

	c.JSON(http.StatusOK, st)
}
