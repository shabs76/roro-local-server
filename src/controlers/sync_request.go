package controlers

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	apiservices "github.com/shabs76/roro-local-server/api_services"
	"github.com/shabs76/roro-local-server/constants"
	"github.com/shabs76/roro-local-server/constants/modules/manifest"
	manifestdataservices "github.com/shabs76/roro-local-server/database/manifest_data_services"
	"github.com/shabs76/roro-local-server/pkg/middleware"
)

// AutoDataSyncSSE returns an SSE stream updating the client on sync task progress
func AutoDataSyncSSE(c *gin.Context) {
	// 1.0 Obtain AuthSession from context
	val, ok := c.Get("authSession")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"state": "error", "data": "Unauthorized"})
		return
	}

	authSession, ok := val.(middleware.AuthSession)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"state": "error", "data": "Unauthorized"})
		return
	}

	// 1.5 Set Headers for SSE
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")

	// 2. Define the tasks (This is where you would plug in your real logic)
	// You might want to pass channels or callbacks to real functions to get granular progress,
	// or break this down if the tasks are actual complex function calls.
	tasks := []struct {
		Name        string
		Description string
		Execute     func() error
	}{
		{
			Name:        "Init (Health Check)",
			Description: "Checking connection to remote server and if everything is reachable",
			Execute: func() error {
				return apiservices.PerformHealthCheckToRemote(authSession.LogID, authSession.LogKey)
			},
		},
		{
			Name:        "FetchRoles",
			Description: "Fetching user roles from remote server",
			Execute: func() error {
				return apiservices.FetchUserRoles(authSession.LogID, authSession.LogKey)
			},
		},
		{
			Name:        "FetchUsers",
			Description: "Fetching user list from remote server",
			Execute: func() error {
				return apiservices.FetchUsersList(authSession.LogID, authSession.LogKey)
			},
		},
		{
			Name:        "FetchManufacturers",
			Description: "Fetching vehicle manufacturers data from remote server",
			Execute: func() error {
				return apiservices.FetchVehicleMakers(authSession.LogID, authSession.LogKey)
			},
		},
		{
			Name:        "FetchBodies",
			Description: "Fetching vehiclebody types data from remote server",
			Execute: func() error {
				return apiservices.FetchVehicleBodies(authSession.LogID, authSession.LogKey)
			},
		},
		{
			Name:        "FetchModels",
			Description: "Fetching vehicle models data from remote server",
			Execute: func() error {
				return apiservices.FetchVehicleModels(authSession.LogID, authSession.LogKey)
			},
		},
		{
			Name:        "FetchInspectChecklist",
			Description: "Fetching vehicle inspection checklist data from remote server",
			Execute: func() error {
				return apiservices.FetchInspectionCheckList(authSession.LogID, authSession.LogKey)
			},
		},
		{
			Name:        "FetchPackageTypes",
			Description: "Fetching package types and status data from remote server",
			Execute: func() error {
				return apiservices.FetchPackageTypesAndStatus(authSession.LogID, authSession.LogKey)
			},
		},
		{
			Name:        "FetchClientList",
			Description: "Fetching client list data from remote server",
			Execute: func() error {
				return apiservices.FetchClientList(authSession.LogID, authSession.LogKey)
			},
		},
	}

	totalTasks := len(tasks)
	clientGone := c.Request.Context().Done()

	for i, task := range tasks {
		select {
		case <-clientGone:
			// Client disconnected, stop work
			return
		default:
			// Calculate progress
			// We report the start of the task
			progress := int((float64(i) / float64(totalTasks)) * 100)

			// Send 'progress' event
			c.SSEvent("progress", gin.H{
				"state": constants.SuccessState,
				"data": gin.H{
					"task":      task.Name,
					"message":   task.Description,
					"status":    "running",
					"progress":  progress,
					"timestamp": time.Now().Format("2006-01-02 15:04:05"),
				},
			})
			c.Writer.Flush()

			// Execute Task
			err := task.Execute()

			if err != nil {
				// Report Error
				c.SSEvent("error", gin.H{
					"state": constants.ErrorState,
					"data": gin.H{
						"task":      task.Name,
						"message":   fmt.Sprintf("Task failed: %v", err),
						"status":    "error",
						"progress":  progress,
						"timestamp": time.Now().Format("2006-01-02 15:04:05"),
					},
				})
				c.Writer.Flush()
				return // Stop the sequence
			}
		}
	}

	// Send final completion event
	c.SSEvent("complete", gin.H{
		"state": constants.SuccessState,
		"data": gin.H{
			"task":      "All",
			"message":   "Sync completed successfully",
			"status":    "finished",
			"progress":  100,
			"timestamp": time.Now().Format("2006-01-02 15:04:05"),
		},
	})
	c.Writer.Flush()
}

func ManifestDataSyncSSE(c *gin.Context) {
	// 1.0 Obtain AuthSession from context
	val, ok := c.Get("authSession")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"state": "error", "data": "Unauthorized"})
		return
	}

	authSession, ok := val.(middleware.AuthSession)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"state": "error", "data": "Unauthorized"})
		return
	}

	manifestId := c.Param("manifest-id")
	if manifestId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"state": "error", "data": "Missing manifest ID"})
		return
	}

	// 1.5 Set Headers for SSE
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")

	// Create a channel for SSE messages
	type SSEMessage struct {
		Event string
		Data  any
	}
	msgChan := make(chan SSEMessage, 100)
	ctx := c.Request.Context()

	// Launch worker goroutine
	go func() {
		defer close(msgChan)

		defer func() {
			if r := recover(); r != nil {
				slog.Error("Sync goroutine panicked", "panic", r)
				msgChan <- SSEMessage{Event: "error", Data: gin.H{"state": "error", "message": fmt.Sprintf("Internal error: %v", r)}}
			}
		}()

		// tasks for manifest data sync
		tasks := []struct {
			Name        string
			Description string
			Execute     func() error
		}{
			{
				Name:        "FetchManifestDetails",
				Description: "Fetching manifest details from remote server",
				Execute: func() error {
					return apiservices.FetchManifestDetails(authSession.LogID, authSession.LogKey, manifestId)
				},
			},
			{
				Name:        "FetchVehicleList",
				Description: "Fetching vehicle list data from remote server",
				Execute: func() error {
					return apiservices.FetchVehicleList(authSession.LogID, authSession.LogKey, manifestId)
				},
			},
			{
				Name:        "FetchStowagePlan",
				Description: "Fetching stowage plan data from remote server",
				Execute: func() error {
					return apiservices.FetchStowagePlan(authSession.LogID, authSession.LogKey, manifestId)
				},
			},
			{
				Name:        "FetchPackageList",
				Description: "Fetching package list data from remote server",
				Execute: func() error {
					return apiservices.FetchPackageList(authSession.LogID, authSession.LogKey, manifestId)
				},
			},
		}

		totalTasks := len(tasks)

		for i, task := range tasks {
			// Check for cancellation
			select {
			case <-ctx.Done():
				return
			default:
			}

			// Calculate progress
			progress := int((float64(i) / float64(totalTasks)) * 100)

			// Send 'progress' event
			msgChan <- SSEMessage{
				Event: "progress",
				Data: gin.H{
					"state": constants.SuccessState,
					"data": gin.H{
						"task":      task.Name,
						"message":   task.Description,
						"status":    "running",
						"progress":  progress,
						"timestamp": time.Now().Format("2006-01-02 15:04:05"),
					},
				},
			}

			// Execute Task
			err := task.Execute()
			log.Printf("Completed task: %s for manifest ID: %s\n", task.Name, manifestId)

			if err != nil {
				// Report Error
				msgChan <- SSEMessage{
					Event: "error",
					Data: gin.H{
						"state": constants.ErrorState,
						"data": gin.H{
							"task":      task.Name,
							"message":   fmt.Sprintf("Task failed: %v", err),
							"status":    "error",
							"progress":  progress,
							"timestamp": time.Now().Format("2006-01-02 15:04:05"),
						},
					},
				}
				return // Stop the sequence
			}
		}

		// Send final completion event
		msgChan <- SSEMessage{
			Event: "complete",
			Data: gin.H{
				"state": constants.SuccessState,
				"data": gin.H{
					"task":      "All",
					"message":   "Manifest data sync completed successfully",
					"status":    "finished",
					"progress":  100,
					"timestamp": time.Now().Format("2006-01-02 15:04:05"),
				},
			},
		}
	}()

	// Main loop to send events and keep-alives
	// Send updates frequently to prevent proxy dropping for idle read
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	var lastEvent SSEMessage
	firstEvent := true

	for {
		select {
		case msg, ok := <-msgChan:
			if !ok {
				// Channel closed, worker finished
				return
			}
			lastEvent = msg
			firstEvent = false
			c.SSEvent(msg.Event, msg.Data)
			c.Writer.Flush()

			// Break out of loop properly
			if msg.Event == "error" || msg.Event == "complete" {
				return
			}
		case <-ticker.C:
			// Send keep-alive comment
			if _, err := c.Writer.WriteString(": keep-alive\n\n"); err != nil {
				return // socket is gone
			}
			// Re-emit last active progress state
			if !firstEvent && lastEvent.Event == "progress" {
				if d, ok := lastEvent.Data.(gin.H); ok {
					if innerMap, valid := d["data"].(gin.H); valid {
						// Clone maps to prevent concurrent map iterations
						newInner := gin.H{}
						for k, v := range innerMap {
							newInner[k] = v
						}
						newInner["timestamp"] = time.Now().Format("2006-01-02 15:04:05")

						newData := gin.H{}
						for k, v := range d {
							newData[k] = v
						}
						newData["data"] = newInner
						c.SSEvent(lastEvent.Event, newData)
					}
				}
			}
			c.Writer.Flush()
		case <-c.Request.Context().Done():
			// Client disconnected
			return
		}
	}
}

func PublishVehiclesAndPackagesInspectionSSE(c *gin.Context) {
	// 1.0 Obtain AuthSession from context
	val, ok := c.Get("authSession")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"state": "error", "data": "Unauthorized"})
		return
	}

	authSession, ok := val.(middleware.AuthSession)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"state": "error", "data": "Unauthorized"})
		return
	}

	// 1.5 Set Headers for SSE
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")

	// 2 Get parameters
	manifestId := c.Param("manifestId")
	if manifestId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"state": "error", "data": "Missing manifest ID"})
		return
	}
	// 2.5 get type of publish (continue or new)
	publishStatus := c.DefaultQuery("status", "continue")
	isContinue := publishStatus == "continue"

	// Helper to upload media
	uploadMedia := func(localPath string) (string, error) {
		if localPath == "" {
			return "", nil
		}

		cleanPath := strings.TrimPrefix(localPath, "/")

		// Attempt to resolve the file path relative to CWD or strict location
		candidates := []string{cleanPath}
		// Also try with media base dir prefix
		candidates = append(candidates, filepath.Join(constants.MediaBaseDir, cleanPath))

		// Also try looking in parent directory (useful if running from src/)
		// For each candidate, add "../candidate"
		extra := []string{}
		for _, p := range candidates {
			extra = append(extra, filepath.Join("..", p))
		}
		candidates = append(candidates, extra...)

		// Find first existing file
		var finalPath string
		found := false
		for _, p := range candidates {
			if info, err := os.Stat(p); err == nil && !info.IsDir() {
				finalPath = p
				found = true
				break
			}
		}

		if !found {
			return "", fmt.Errorf("file not found: %s (checked: %v)", localPath, candidates)
		}

		// Determine type and call appropriate API
		ext := strings.ToLower(filepath.Ext(finalPath))
		var remoteURL string
		var err error

		switch ext {
		case ".jpg", ".jpeg", ".png", ".webp", ".gif", ".heic", ".avif":
			remoteURL, err = apiservices.UploadImage(authSession.LogID, authSession.LogKey, finalPath)
		case ".mp4", ".avi", ".mov", ".mkv":
			remoteURL, err = apiservices.UploadVideo(authSession.LogID, authSession.LogKey, finalPath)
		case ".pdf", ".doc", ".docx":
			remoteURL, err = apiservices.UploadPdf(authSession.LogID, authSession.LogKey, finalPath)
		default:
			// Fallback treat as image if not known
			remoteURL, err = apiservices.UploadImage(authSession.LogID, authSession.LogKey, finalPath)
		}
		return remoteURL, err
	}

	// Prepare Queries
	pkgSubQuery := " manifest_packages.manifest_id = ? "
	pkgArgs := []any{manifestId}

	vehSubQuery := " manifest_vehicles.manifest_id = ? AND manifest_vehicles.inspection_status = 'yes' "
	vehArgs := []any{manifestId}

	if isContinue {
		pkgSubQuery += " AND manifest_packages.is_published != 'yes' "
		vehSubQuery += " AND manifest_vehicles.is_published != 'yes' "
	}

	// Fetch Data
	stP, packages := manifestdataservices.SelectPackageInspectionData(pkgSubQuery, pkgArgs)
	if stP.State != constants.SuccessState {
		c.SSEvent("error", gin.H{"state": "error", "message": "Failed to fetch packages"})
		c.Writer.Flush()
		return
	}

	stV, vehicles := manifestdataservices.SelectVehicleAndTallyDetails(vehSubQuery, vehArgs)
	if stV.State != constants.SuccessState {
		c.SSEvent("error", gin.H{"state": "error", "message": "Failed to fetch vehicles"})
		c.Writer.Flush()
		return
	}

	type SSEMessage struct {
		Event string
		Data  any
	}
	msgChan := make(chan SSEMessage, 100)
	doneChan := make(chan struct{})
	ctx := c.Request.Context()

	go func() {
		defer close(doneChan)
		defer close(msgChan)

		defer func() {
			if r := recover(); r != nil {
				slog.Error("Sync goroutine panicked", "panic", r)
				msgChan <- SSEMessage{Event: "error", Data: gin.H{"state": "error", "message": fmt.Sprintf("Internal error: %v", r)}}
			}
		}()

		totalItems := len(packages) + len(vehicles)
		currentItem := 0

		// Process Packages
		for _, pkg := range packages {
			// Check for cancellation
			select {
			case <-ctx.Done():
				return
			default:
			}

			currentItem++
			progress := int((float64(currentItem) / float64(totalItems)) * 100)
			if progress >= 100 {
				progress = 99
			}
			senMsg := fmt.Sprintf("Processing Package %s [%s]", pkg.PackageNumber, pkg.Description)

			if pkg.IsAddedLater == "yes" {
				msgChan <- SSEMessage{Event: "progress", Data: gin.H{"state": "success", "data": gin.H{"task": "Package Upload", "message": senMsg + ": Registering New Package", "status": "running", "progress": progress}}}

				newPkgId, err := apiservices.UploadSinglePackageToRemote(authSession.LogID, authSession.LogKey, manifest.AddPackageRequest{
					ManifestId:    manifestId,
					BLNumber:      pkg.BLNumber,
					PackageNumber: pkg.PackageNumber,
					Description:   pkg.Description,
				})

				if err != nil {
					slog.Error("Failed to upload new package", "err", err)
					msgChan <- SSEMessage{Event: "error", Data: gin.H{"state": "error", "message": fmt.Sprintf("Failed to upload new package %s: %v", pkg.PackageNumber, err)}}
					continue
				}

				// Update IDs
				manifestdataservices.UpdatePackageIdToRemote(pkg.PackageId, newPkgId)
				pkg.PackageId = newPkgId
			}

			msgChan <- SSEMessage{Event: "progress", Data: gin.H{"state": "success", "data": gin.H{"task": "Package Upload", "message": senMsg + ": Uploading Image", "status": "running", "progress": progress}}}

			// Upload Image
			imgUrl, err := uploadMedia(pkg.Picture)
			if err != nil {
				slog.Error("Failed to upload package image", "err", err)
				msgChan <- SSEMessage{Event: "error", Data: gin.H{"state": "error", "message": fmt.Sprintf("Failed to upload image for package %s: %v", pkg.PackageNumber, err)}}
				continue
			}

			// Media
			msgChan <- SSEMessage{Event: "progress", Data: gin.H{"state": "success", "data": gin.H{"task": "Package Upload", "message": senMsg + ": Processing Extra Media", "status": "running", "progress": progress}}}
			packageMedias := []manifest.PackageExtraMediaRequest{}
			for _, m := range pkg.Media {
				if m.MediaLink == "" {
					continue
				}
				mLink, err := uploadMedia(m.MediaLink)
				if err != nil {
					slog.Warn("Failed to upload extra media", "error", err, "file", m.MediaLink)
					msgChan <- SSEMessage{Event: "error", Data: gin.H{"state": "error", "message": fmt.Sprintf("Failed to upload extra media %s: %v", m.MediaLink, err)}}
					continue
				}
				packageMedias = append(packageMedias, manifest.PackageExtraMediaRequest{
					MediaLink: mLink,
					MediaType: m.MediaType,
					Remark:    m.Remark,
				})
			}

			req := manifest.PackageInspectionSaveRequest{
				PackageId:          pkg.PackageId,
				PackageImage:       imgUrl,
				TypeId:             pkg.TypeId,
				InspectionTime:     pkg.InspectionTime,
				InspectionStatusId: pkg.InspectionStatusId,
				Media:              packageMedias,
			}

			msgChan <- SSEMessage{Event: "progress", Data: gin.H{"state": "success", "data": gin.H{"task": "Package Upload", "message": senMsg + ": Saving Data", "status": "running", "progress": progress}}}

			_, err = apiservices.UploadPackageInspectionData(authSession.LogID, authSession.LogKey, req)
			if err != nil {
				slog.Error("Failed to upload package data", "err", err)
				msgChan <- SSEMessage{Event: "error", Data: gin.H{"state": "error", "message": fmt.Sprintf("Failed to save data for package %s: %v", pkg.PackageNumber, err)}}
				continue
			}

			// Update Local
			manifestdataservices.UpdatePackagePublishedStatus(pkg.PackageId, true)
		}

		// Process Vehicles
		for _, veh := range vehicles {
			// Check for cancellation
			select {
			case <-ctx.Done():
				return
			default:
			}

			currentItem++
			progress := int((float64(currentItem) / float64(totalItems)) * 100)
			if progress >= 100 {
				progress = 99
			}
			senMsg := fmt.Sprintf("Processing Vehicle %s [%s]", veh.ChasisNumber, veh.Description)

			if veh.IsAddedLater == "yes" {
				msgChan <- SSEMessage{Event: "progress", Data: gin.H{"state": "success", "data": gin.H{"task": "Vehicle Upload", "message": senMsg + ": Registering New Vehicle", "status": "running", "progress": progress}}}

				isOverLand := "no"
				if veh.OverLandStatus == "yes" {
					isOverLand = "yes"
				}
				newVehId, err := apiservices.UploadSingleVehicleToRemote(authSession.LogID, authSession.LogKey, manifest.AddVehicleRequest{
					ManifestId:   manifestId,
					ChasisNumber: veh.ChasisNumber,
					VehicleModel: veh.VehicleModel,
					Description:  veh.Description,
					Weight:       float64(veh.Weight),
					IsOverLand:   isOverLand,
					BLNumber:     veh.BLNumber,
				})

				if err != nil {
					slog.Error("Failed to upload new vehicle", "err", err)
					msgChan <- SSEMessage{Event: "error", Data: gin.H{"state": "error", "message": fmt.Sprintf("Failed to upload new vehicle %s: %v", veh.ChasisNumber, err)}}
					continue
				}

				// Update IDs
				// Need to update local DB with new ID
				manifestdataservices.UpdateVehicleIdToRemote(veh.VehicleId, newVehId)
				veh.VehicleId = newVehId
			}

			msgChan <- SSEMessage{Event: "progress", Data: gin.H{"state": "success", "data": gin.H{"task": "Vehicle Upload", "message": senMsg + ": Uploading Vehicle Image", "status": "running", "progress": progress}}}

			vImgUrl, err := uploadMedia(veh.VehicleImage)
			if err != nil {
				slog.Error("Failed to upload vehicle image", "err", err)
				msgChan <- SSEMessage{Event: "error", Data: gin.H{"state": "error", "message": fmt.Sprintf("Failed to upload image for vehicle %s: %v", veh.ChasisNumber, err)}}
				continue
			}

			// Checks
			msgChan <- SSEMessage{Event: "progress", Data: gin.H{"state": "success", "data": gin.H{"task": "Vehicle Upload", "message": senMsg + ": Processing Checks", "status": "running", "progress": progress}}}

			_, insps := manifestdataservices.SelectInspectionDetails(" vehicle_id = ? ", []any{veh.VehicleId})
			checks := []manifest.InspectionCheck{}
			for _, insp := range insps {
				faultUrl := ""
				// Check for images
				_, imgs := manifestdataservices.SelectInspectionImages(" inspection_id = ? ", []any{insp.InspectionId})
				if len(imgs) > 0 {
					faultUrl, _ = uploadMedia(imgs[0].ImageLink)
				}
				checks = append(checks, manifest.InspectionCheck{
					CheckId:     insp.CheckId,
					CheckStatus: insp.Status,
					FaultImage:  faultUrl,
				})
			}

			// Remarks
			msgChan <- SSEMessage{Event: "progress", Data: gin.H{"state": "success", "data": gin.H{"task": "Vehicle Upload", "message": senMsg + ": Processing Remarks", "status": "running", "progress": progress}}}

			_, rems := manifestdataservices.SelectInspectionRemarks(" vehicle_id = ? ", []any{veh.VehicleId})
			remarks := []manifest.RemarkSaveRequest{}
			for _, r := range rems {
				rImg, _ := uploadMedia(r.ImageLink)
				remarks = append(remarks, manifest.RemarkSaveRequest{
					Remark:      r.Remark,
					RemarkType:  r.RemarkType,
					RemarkImage: rImg,
				})
			}

			// Media
			msgChan <- SSEMessage{Event: "progress", Data: gin.H{"state": "success", "data": gin.H{"task": "Vehicle Upload", "message": senMsg + ": Processing Extra Media", "status": "running", "progress": progress}}}
			_, meds := manifestdataservices.SelectVehicleMedia(" vehicle_id = ? ", []any{veh.VehicleId})
			medias := []manifest.VehicleExtraMediaRequest{}
			for _, m := range meds {
				if m.MediaLink == "" {
					continue
				}
				mLink, err := uploadMedia(m.MediaLink)
				if err != nil {
					slog.Warn("Failed to upload extra media", "error", err, "file", m.MediaLink)
					// Optionally report error to client or just skip
					msgChan <- SSEMessage{Event: "error", Data: gin.H{"state": "error", "message": fmt.Sprintf("Failed to upload extra media %s: %v", m.MediaLink, err)}}
					continue
				}
				medias = append(medias, manifest.VehicleExtraMediaRequest{
					MediaLink: mLink,
					MediaType: m.MediaType,
					Remark:    m.Remark,
				})
			}

			// OnBoard
			msgChan <- SSEMessage{Event: "progress", Data: gin.H{"state": "success", "data": gin.H{"task": "Vehicle Upload", "message": senMsg + ": Processing OnBoard Packages", "status": "running", "progress": progress}}}
			_, obs := manifestdataservices.SelectOnBoardPackage(" vehicle_id = ? ", []any{veh.VehicleId})
			onBoards := []manifest.OnBoardPackageRequest{}
			for _, o := range obs {
				obMedias := []manifest.OnBoardPackageMediaRequest{}
				for _, om := range o.Media {
					if om.MediaLink == "" {
						continue
					}
					omLink, err := uploadMedia(om.MediaLink)
					if err != nil {
						slog.Warn("Failed to upload on-board media", "error", err, "file", om.MediaLink)
						msgChan <- SSEMessage{Event: "error", Data: gin.H{"state": "error", "message": fmt.Sprintf("Failed to upload on-board media %s: %v", om.MediaLink, err)}}
						continue
					}
					obMedias = append(obMedias, manifest.OnBoardPackageMediaRequest{
						MediaLink: omLink,
						MediaType: om.MediaType,
					})
				}
				onBoards = append(onBoards, manifest.OnBoardPackageRequest{
					Title:  o.Title,
					Remark: o.Remark,
					Media:  obMedias,
				})
			}

			req := manifest.InspectionChecksRequest{
				VehicleId:      veh.VehicleId,
				VehicleImage:   vImgUrl,
				ManifestId:     manifestId,
				MakerId:        veh.MakerId,
				BodyId:         veh.BodyId,
				ModelName:      veh.VehicleModel,
				InspectionTime: veh.InspectionTime,
				DeckNumber:     veh.DeckNumber,
				NumberOfKeys:   veh.NumberOfKeys,
				KeyType:        veh.KeyType,
				Checks:         checks,
				Remarks:        remarks,
				Media:          medias,
				Packages:       onBoards,
			}

			msgChan <- SSEMessage{Event: "progress", Data: gin.H{"state": "success", "data": gin.H{"task": "Vehicle Upload", "message": senMsg + ": Saving Data", "status": "running", "progress": progress}}}

			_, err = apiservices.UploadVehicleInspectionData(authSession.LogID, authSession.LogKey, req)
			if err != nil {
				slog.Error("Failed to upload vehicle data", "err", err)
				msgChan <- SSEMessage{Event: "error", Data: gin.H{"state": "error", "message": fmt.Sprintf("Failed to save data for vehicle %s: %v", veh.ChasisNumber, err)}}
				continue
			}

			manifestdataservices.UpdateVehiclePublishedStatus(veh.VehicleId, true)
		}

		// Final
		msgChan <- SSEMessage{Event: "complete", Data: gin.H{
			"state": "success",
			"data": gin.H{
				"task":      "All",
				"message":   "Upload completed successfully",
				"status":    "finished",
				"progress":  100,
				"timestamp": time.Now().Format("2006-01-02 15:04:05"),
			},
		}}
	}()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case msg, ok := <-msgChan:
			if !ok {
				return
			}
			c.SSEvent(msg.Event, msg.Data)
			c.Writer.Flush()
		case <-ticker.C:
			c.Writer.WriteString(": keep-alive\n\n")
			c.Writer.Flush()
		case <-c.Request.Context().Done():
			return
		}
	}

}

func PublishAddedLaterVehiclesAndPackagesSSE(c *gin.Context) {
	// 1.0 Obtain AuthSession from context
	val, ok := c.Get("authSession")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"state": "error", "data": "Unauthorized"})
		return
	}

	authSession, ok := val.(middleware.AuthSession)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"state": "error", "data": "Unauthorized"})
		return
	}

	// 1.5 Set Headers for SSE
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")

	// 2 Get parameters
	manifestId := c.Param("manifestId")
	if manifestId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"state": "error", "data": "Missing manifest ID"})
		return
	}
	// 2.5 get type of publish (continue or new)
	publishStatus := c.DefaultQuery("status", "continue")
	isContinue := publishStatus == "continue"

	// Helper to upload media
	uploadMedia := func(localPath string) (string, error) {
		if localPath == "" {
			return "", nil
		}

		cleanPath := strings.TrimPrefix(localPath, "/")

		// Attempt to resolve the file path relative to CWD or strict location
		candidates := []string{cleanPath}
		// Also try with media base dir prefix
		candidates = append(candidates, filepath.Join(constants.MediaBaseDir, cleanPath))

		// Also try looking in parent directory (useful if running from src/)
		// For each candidate, add "../candidate"
		extra := []string{}
		for _, p := range candidates {
			extra = append(extra, filepath.Join("..", p))
		}
		candidates = append(candidates, extra...)

		// Find first existing file
		var finalPath string
		found := false
		for _, p := range candidates {
			if info, err := os.Stat(p); err == nil && !info.IsDir() {
				finalPath = p
				found = true
				break
			}
		}

		if !found {
			return "", fmt.Errorf("file not found: %s (checked: %v)", localPath, candidates)
		}

		// Determine type and call appropriate API
		ext := strings.ToLower(filepath.Ext(finalPath))
		var remoteURL string
		var err error

		switch ext {
		case ".jpg", ".jpeg", ".png", ".webp", ".gif", ".heic", ".avif":
			remoteURL, err = apiservices.UploadImage(authSession.LogID, authSession.LogKey, finalPath)
		case ".mp4", ".avi", ".mov", ".mkv":
			remoteURL, err = apiservices.UploadVideo(authSession.LogID, authSession.LogKey, finalPath)
		case ".pdf", ".doc", ".docx":
			remoteURL, err = apiservices.UploadPdf(authSession.LogID, authSession.LogKey, finalPath)
		default:
			// Fallback treat as image if not known
			remoteURL, err = apiservices.UploadImage(authSession.LogID, authSession.LogKey, finalPath)
		}
		return remoteURL, err
	}

	// Prepare Queries
	pkgSubQuery := " manifest_packages.manifest_id = ? AND manifest_packages.is_added_later = 'yes' "
	pkgArgs := []any{manifestId}

	vehSubQuery := " manifest_vehicles.manifest_id = ? AND manifest_vehicles.inspection_status = 'yes' AND manifest_vehicles.is_added_later = 'yes' "
	vehArgs := []any{manifestId}

	if isContinue {
		pkgSubQuery += " AND manifest_packages.is_published != 'yes' "
		vehSubQuery += " AND manifest_vehicles.is_published != 'yes' "
	}

	// Fetch Data
	stP, packages := manifestdataservices.SelectPackageInspectionData(pkgSubQuery, pkgArgs)
	if stP.State != constants.SuccessState {
		c.SSEvent("error", gin.H{"state": "error", "message": "Failed to fetch packages"})
		c.Writer.Flush()
		return
	}

	stV, vehicles := manifestdataservices.SelectVehicleAndTallyDetails(vehSubQuery, vehArgs)
	if stV.State != constants.SuccessState {
		c.SSEvent("error", gin.H{"state": "error", "message": "Failed to fetch vehicles"})
		c.Writer.Flush()
		return
	}

	type SSEMessage struct {
		Event string
		Data  any
	}
	msgChan := make(chan SSEMessage, 100)
	doneChan := make(chan struct{})
	ctx := c.Request.Context()

	go func() {
		defer close(doneChan)
		defer close(msgChan)

		defer func() {
			if r := recover(); r != nil {
				slog.Error("Sync goroutine panicked", "panic", r)
				msgChan <- SSEMessage{Event: "error", Data: gin.H{"state": "error", "message": fmt.Sprintf("Internal error: %v", r)}}
			}
		}()

		totalItems := len(packages) + len(vehicles)
		currentItem := 0

		// Process Packages
		for _, pkg := range packages {
			// Check for cancellation
			select {
			case <-ctx.Done():
				return
			default:
			}

			currentItem++
			progress := int((float64(currentItem) / float64(totalItems)) * 100)
			if progress >= 100 {
				progress = 99
			}
			senMsg := fmt.Sprintf("Processing Package %s [%s]", pkg.PackageNumber, pkg.Description)

			if pkg.IsAddedLater == "yes" {
				msgChan <- SSEMessage{Event: "progress", Data: gin.H{"state": "success", "data": gin.H{"task": "Package Upload", "message": senMsg + ": Registering New Package", "status": "running", "progress": progress}}}

				newPkgId, err := apiservices.UploadSinglePackageToRemote(authSession.LogID, authSession.LogKey, manifest.AddPackageRequest{
					ManifestId:    manifestId,
					BLNumber:      pkg.BLNumber,
					PackageNumber: pkg.PackageNumber,
					Description:   pkg.Description,
				})

				if err != nil {
					slog.Error("Failed to upload new package", "err", err)
					msgChan <- SSEMessage{Event: "error", Data: gin.H{"state": "error", "message": fmt.Sprintf("Failed to upload new package %s: %v", pkg.PackageNumber, err)}}
					continue
				}

				// Update IDs
				manifestdataservices.UpdatePackageIdToRemote(pkg.PackageId, newPkgId)
				pkg.PackageId = newPkgId
			}

			msgChan <- SSEMessage{Event: "progress", Data: gin.H{"state": "success", "data": gin.H{"task": "Package Upload", "message": senMsg + ": Uploading Image", "status": "running", "progress": progress}}}

			// Upload Image
			imgUrl, err := uploadMedia(pkg.Picture)
			if err != nil {
				slog.Error("Failed to upload package image", "err", err.Error())
				msgChan <- SSEMessage{Event: "error", Data: gin.H{"state": "error", "message": fmt.Sprintf("Failed to upload image for package %s: %v", pkg.PackageNumber, err)}}
				continue
			}

			// Media
			msgChan <- SSEMessage{Event: "progress", Data: gin.H{"state": "success", "data": gin.H{"task": "Package Upload", "message": senMsg + ": Processing Extra Media", "status": "running", "progress": progress}}}
			packageMedias := []manifest.PackageExtraMediaRequest{}
			for _, m := range pkg.Media {
				if m.MediaLink == "" {
					continue
				}
				mLink, err := uploadMedia(m.MediaLink)
				if err != nil {
					slog.Warn("Failed to upload extra media", "error", err, "file", m.MediaLink)
					msgChan <- SSEMessage{Event: "error", Data: gin.H{"state": "error", "message": fmt.Sprintf("Failed to upload extra media %s: %v", m.MediaLink, err)}}
					continue
				}
				packageMedias = append(packageMedias, manifest.PackageExtraMediaRequest{
					MediaLink: mLink,
					MediaType: m.MediaType,
					Remark:    m.Remark,
				})
			}

			req := manifest.PackageInspectionSaveRequest{
				PackageId:          pkg.PackageId,
				PackageImage:       imgUrl,
				TypeId:             pkg.TypeId,
				InspectionTime:     pkg.InspectionTime,
				InspectionStatusId: pkg.InspectionStatusId,
				Media:              packageMedias,
			}

			msgChan <- SSEMessage{Event: "progress", Data: gin.H{"state": "success", "data": gin.H{"task": "Package Upload", "message": senMsg + ": Saving Data", "status": "running", "progress": progress}}}

			_, err = apiservices.UploadPackageInspectionData(authSession.LogID, authSession.LogKey, req)
			if err != nil {
				slog.Error("Failed to upload package data", "err", err)
				msgChan <- SSEMessage{Event: "error", Data: gin.H{"state": "error", "message": fmt.Sprintf("Failed to save data for package %s: %v", pkg.PackageNumber, err)}}
				continue
			}

			// Update Local
			manifestdataservices.UpdatePackagePublishedStatus(pkg.PackageId, true)
		}

		// Process Vehicles
		for _, veh := range vehicles {
			// Check for cancellation
			select {
			case <-ctx.Done():
				return
			default:
			}

			currentItem++
			progress := int((float64(currentItem) / float64(totalItems)) * 100)
			if progress >= 100 {
				progress = 99
			}
			senMsg := fmt.Sprintf("Processing Vehicle %s [%s]", veh.ChasisNumber, veh.Description)

			if veh.IsAddedLater == "yes" {
				msgChan <- SSEMessage{Event: "progress", Data: gin.H{"state": "success", "data": gin.H{"task": "Vehicle Upload", "message": senMsg + ": Registering New Vehicle", "status": "running", "progress": progress}}}

				isOverLand := "no"
				if veh.OverLandStatus == "yes" {
					isOverLand = "yes"
				}
				newVehId, err := apiservices.UploadSingleVehicleToRemote(authSession.LogID, authSession.LogKey, manifest.AddVehicleRequest{
					ManifestId:   manifestId,
					ChasisNumber: veh.ChasisNumber,
					VehicleModel: veh.VehicleModel,
					Description:  veh.Description,
					Weight:       float64(veh.Weight),
					IsOverLand:   isOverLand,
					BLNumber:     veh.BLNumber,
				})

				if err != nil {
					slog.Error("Failed to upload new vehicle", "err", err)
					msgChan <- SSEMessage{Event: "error", Data: gin.H{"state": "error", "message": fmt.Sprintf("Failed to upload new vehicle %s: %v", veh.ChasisNumber, err)}}
					continue
				}

				// Update IDs
				// Need to update local DB with new ID
				manifestdataservices.UpdateVehicleIdToRemote(veh.VehicleId, newVehId)
				veh.VehicleId = newVehId
			}

			msgChan <- SSEMessage{Event: "progress", Data: gin.H{"state": "success", "data": gin.H{"task": "Vehicle Upload", "message": senMsg + ": Uploading Vehicle Image", "status": "running", "progress": progress}}}

			vImgUrl, err := uploadMedia(veh.VehicleImage)
			if err != nil {
				slog.Error("Failed to upload vehicle image", "err", err)
				msgChan <- SSEMessage{Event: "error", Data: gin.H{"state": "error", "message": fmt.Sprintf("Failed to upload image for vehicle %s: %v", veh.ChasisNumber, err)}}
				continue
			}

			// Checks
			msgChan <- SSEMessage{Event: "progress", Data: gin.H{"state": "success", "data": gin.H{"task": "Vehicle Upload", "message": senMsg + ": Processing Checks", "status": "running", "progress": progress}}}

			_, insps := manifestdataservices.SelectInspectionDetails(" vehicle_id = ? ", []any{veh.VehicleId})
			checks := []manifest.InspectionCheck{}
			for _, insp := range insps {
				faultUrl := ""
				// Check for images
				_, imgs := manifestdataservices.SelectInspectionImages(" inspection_id = ? ", []any{insp.InspectionId})
				if len(imgs) > 0 {
					faultUrl, _ = uploadMedia(imgs[0].ImageLink)
				}
				checks = append(checks, manifest.InspectionCheck{
					CheckId:     insp.CheckId,
					CheckStatus: insp.Status,
					FaultImage:  faultUrl,
				})
			}

			// Remarks
			msgChan <- SSEMessage{Event: "progress", Data: gin.H{"state": "success", "data": gin.H{"task": "Vehicle Upload", "message": senMsg + ": Processing Remarks", "status": "running", "progress": progress}}}

			_, rems := manifestdataservices.SelectInspectionRemarks(" vehicle_id = ? ", []any{veh.VehicleId})
			remarks := []manifest.RemarkSaveRequest{}
			for _, r := range rems {
				rImg, _ := uploadMedia(r.ImageLink)
				remarks = append(remarks, manifest.RemarkSaveRequest{
					Remark:      r.Remark,
					RemarkType:  r.RemarkType,
					RemarkImage: rImg,
				})
			}

			// Media
			msgChan <- SSEMessage{Event: "progress", Data: gin.H{"state": "success", "data": gin.H{"task": "Vehicle Upload", "message": senMsg + ": Processing Extra Media", "status": "running", "progress": progress}}}
			_, meds := manifestdataservices.SelectVehicleMedia(" vehicle_id = ? ", []any{veh.VehicleId})
			medias := []manifest.VehicleExtraMediaRequest{}
			for _, m := range meds {
				if m.MediaLink == "" {
					continue
				}
				mLink, err := uploadMedia(m.MediaLink)
				if err != nil {
					slog.Warn("Failed to upload extra media", "error", err, "file", m.MediaLink)
					// Optionally report error to client or just skip
					msgChan <- SSEMessage{Event: "error", Data: gin.H{"state": "error", "message": fmt.Sprintf("Failed to upload extra media %s: %v", m.MediaLink, err)}}
					continue
				}
				medias = append(medias, manifest.VehicleExtraMediaRequest{
					MediaLink: mLink,
					MediaType: m.MediaType,
					Remark:    m.Remark,
				})
			}

			// OnBoard
			msgChan <- SSEMessage{Event: "progress", Data: gin.H{"state": "success", "data": gin.H{"task": "Vehicle Upload", "message": senMsg + ": Processing OnBoard Packages", "status": "running", "progress": progress}}}
			_, obs := manifestdataservices.SelectOnBoardPackage(" vehicle_id = ? ", []any{veh.VehicleId})
			onBoards := []manifest.OnBoardPackageRequest{}
			for _, o := range obs {
				obMedias := []manifest.OnBoardPackageMediaRequest{}
				for _, om := range o.Media {
					if om.MediaLink == "" {
						continue
					}
					omLink, err := uploadMedia(om.MediaLink)
					if err != nil {
						slog.Warn("Failed to upload on-board media", "error", err, "file", om.MediaLink)
						msgChan <- SSEMessage{Event: "error", Data: gin.H{"state": "error", "message": fmt.Sprintf("Failed to upload on-board media %s: %v", om.MediaLink, err)}}
						continue
					}
					obMedias = append(obMedias, manifest.OnBoardPackageMediaRequest{
						MediaLink: omLink,
						MediaType: om.MediaType,
					})
				}
				onBoards = append(onBoards, manifest.OnBoardPackageRequest{
					Title:  o.Title,
					Remark: o.Remark,
					Media:  obMedias,
				})
			}

			req := manifest.InspectionChecksRequest{
				VehicleId:      veh.VehicleId,
				VehicleImage:   vImgUrl,
				ManifestId:     manifestId,
				MakerId:        veh.MakerId,
				BodyId:         veh.BodyId,
				ModelName:      veh.VehicleModel,
				InspectionTime: veh.InspectionTime,
				DeckNumber:     veh.DeckNumber,
				NumberOfKeys:   veh.NumberOfKeys,
				KeyType:        veh.KeyType,
				Checks:         checks,
				Remarks:        remarks,
				Media:          medias,
				Packages:       onBoards,
			}

			msgChan <- SSEMessage{Event: "progress", Data: gin.H{"state": "success", "data": gin.H{"task": "Vehicle Upload", "message": senMsg + ": Saving Data", "status": "running", "progress": progress}}}

			_, err = apiservices.UploadVehicleInspectionData(authSession.LogID, authSession.LogKey, req)
			if err != nil {
				slog.Error("Failed to upload vehicle data", "err", err)
				msgChan <- SSEMessage{Event: "error", Data: gin.H{"state": "error", "message": fmt.Sprintf("Failed to save data for vehicle %s: %v", veh.ChasisNumber, err)}}
				continue
			}

			manifestdataservices.UpdateVehiclePublishedStatus(veh.VehicleId, true)
		}

		// Final
		msgChan <- SSEMessage{Event: "complete", Data: gin.H{
			"state": "success",
			"data": gin.H{
				"task":      "All",
				"message":   "Upload completed successfully",
				"status":    "finished",
				"progress":  100,
				"timestamp": time.Now().Format("2006-01-02 15:04:05"),
			},
		}}
	}()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case msg, ok := <-msgChan:
			if !ok {
				return
			}
			c.SSEvent(msg.Event, msg.Data)
			c.Writer.Flush()
		case <-ticker.C:
			c.Writer.WriteString(": keep-alive\n\n")
			c.Writer.Flush()
		case <-c.Request.Context().Done():
			return
		}
	}

}
