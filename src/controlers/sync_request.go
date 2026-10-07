package controlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	apiservices "github.com/shabs76/roro-local-server/api_services"
	"github.com/shabs76/roro-local-server/constants"
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
			slog.Info("Completed task", "task", task.Name, "manifestId", manifestId)

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

// PublishVehiclesAndPackagesInspectionSSE publishes all inspected vehicles and
// packages of a manifest to the remote server and streams the progress.
func PublishVehiclesAndPackagesInspectionSSE(c *gin.Context) {
	runPublishSSE(c, false)
}

// PublishAddedLaterVehiclesAndPackagesSSE publishes only the vehicles and packages
// that were added on site, and streams the progress.
func PublishAddedLaterVehiclesAndPackagesSSE(c *gin.Context) {
	runPublishSSE(c, true)
}
