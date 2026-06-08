package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/shabs76/roro-local-server/controlers"
	"github.com/shabs76/roro-local-server/pkg/middleware"
)

func SetupManifestRoutes(router *gin.Engine) {
	manifestGroup := router.Group("/manifest", middleware.AuthMiddleware())
	{
		// Media routes
		media := manifestGroup.Group("/media")
		{
			media.POST("/upload/file", controlers.UploadMedia)
			media.GET("/get/file", controlers.GetFile)
			media.GET("/get/full/file", controlers.PlayMedia)
		}
		// manifest routes
		mani := manifestGroup.Group("/manifest")
		{
			mani.GET("/list/remote", controlers.GetManifestListFromRemote)
			mani.GET("/list/local", controlers.GetManifestData)
			mani.GET("/list/local/extended", controlers.GetManifestesDetails)
			mani.GET("/stowage/plan/list/:manifestId", controlers.GetDeckStowagePlan)
		}
		// vehicle routes
		veh := manifestGroup.Group("/vehicles")
		{
			veh.POST("/add/single/vehicle", controlers.AddSingleVehicleToManifest)
			veh.GET("/list/local/:manifestId", controlers.GetVehicleListOfManifest)
			veh.GET("/list/with/multiple/inspections/:manifestId", controlers.GetVehicleWithMultipleInspections)
			veh.GET("/multiple/inspection/details/:vehicleId", controlers.GetVehicleInspectionTallyTimeline)
			veh.GET("/list/inspection/check/:manifestId", controlers.GetVehicleShortInfo) // this is used to check if the vehicle is inspected or not in the manifest list page
		}
		// package routes
		pack := manifestGroup.Group("/packages")
		{
			pack.POST("/add/single/package", controlers.AddSinglePackageToManifest)
			pack.GET("/list/local/:manifestId", controlers.GetManifestPackagesList)
			pack.GET("/details/:packageId", controlers.GetInspectedPackageDetails)
		}

		// inspection routes
		ins := manifestGroup.Group("/inspection")
		{
			ins.POST("/save/vehicle/inspection", controlers.SaveVehicleInspectionDetails)
			ins.POST("/save/package/inspection", controlers.SavePackageInspectionDetails)
			ins.POST("/save/remarks/only", controlers.SaveVehicleRemarksOnly)
			// update
			ins.PUT("/swap/vehicle/inspection", controlers.SaveTransferVehicleData) // not tested yet
			// data
			ins.GET("/get/vehicle/inspection/:vehicleId", controlers.GetVehicleInspectionsDetails)
			ins.GET("/get/package/inspection/:packageId", controlers.GetInspectedPackageDetails)
			// reports
			ins.GET("/get/vehicle/tally/reports/:manifestId", controlers.GenerateTallyReportPdf)
			ins.GET("/get/vehicle/damage/report/:manifestId", controlers.GenerateDamagedVahiclePdf)
			ins.GET("/get/package/report/:manifestId", controlers.GeneratePackageReportPdf)
		}

		// general information routes
		info := manifestGroup.Group("/info")
		{
			info.GET("/get/vehicle/body/types/list", controlers.GetVehicleTypes)
			info.GET("/get/vehicle/manufacturers/list", controlers.GetVehicleMakers)
			info.GET("/get/vehicle/inspection/check-list", controlers.GetInspectionCheckList)
			info.GET("/get/vehicle/models/list", controlers.GetVehicleModels)
			// package info
			info.GET("/get/package/types/list", controlers.GetPackageTypes)
			info.GET("/get/package/statuses", controlers.GetPackageStatusList)
		}
	}
}
