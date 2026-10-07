package controlers

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shabs76/roro-local-server/constants"
	manifestdataservices "github.com/shabs76/roro-local-server/database/manifest_data_services"
)

// vehicleEvent tells tablets that a vehicle of a manifest changed. Tablets treat it
// only as a signal to call the status endpoint (GetVehicleStatusChanges).
type vehicleEvent struct {
	VehicleId string `json:"vehicleId"`
	Kind      string `json:"kind"` // inspected, reinspected, remarks, published
	At        string `json:"at"`
}

// vehicleEventHub fans vehicle events out to every tablet listening on a manifest.
type vehicleEventHub struct {
	mu   sync.Mutex
	subs map[string]map[chan vehicleEvent]struct{}
}

var vehicleEvents = &vehicleEventHub{subs: map[string]map[chan vehicleEvent]struct{}{}}

func (h *vehicleEventHub) subscribe(manifestId string) chan vehicleEvent {
	ch := make(chan vehicleEvent, 64)
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.subs[manifestId] == nil {
		h.subs[manifestId] = map[chan vehicleEvent]struct{}{}
	}
	h.subs[manifestId][ch] = struct{}{}
	return ch
}

func (h *vehicleEventHub) unsubscribe(manifestId string, ch chan vehicleEvent) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.subs[manifestId], ch)
	if len(h.subs[manifestId]) == 0 {
		delete(h.subs, manifestId)
	}
}

// notify never blocks: a tablet that is not reading misses the event and catches up
// on its next status poll.
func (h *vehicleEventHub) notify(manifestId, vehicleId, kind string) {
	if manifestId == "" {
		return
	}
	ev := vehicleEvent{VehicleId: vehicleId, Kind: kind, At: time.Now().Format("2006-01-02 15:04:05")}
	h.mu.Lock()
	defer h.mu.Unlock()
	for ch := range h.subs[manifestId] {
		select {
		case ch <- ev:
		default:
		}
	}
}

// notifyVehicleChanged looks up the manifest of a vehicle and notifies its listeners.
func notifyVehicleChanged(vehicleId, kind string) {
	st, vehicles := manifestdataservices.SelectVehicleInfo(" `vehicle_id` = ? ", []any{vehicleId})
	if st.State != constants.SuccessState || len(vehicles) == 0 {
		return
	}
	vehicleEvents.notify(vehicles[0].ManifestId, vehicleId, kind)
}

// VehicleEventsSSE streams vehicle change events of one manifest to a tablet.
func VehicleEventsSSE(c *gin.Context) {
	manifestId := c.Param("manifestId")
	if manifestId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"state": constants.ErrorState, "data": "Missing manifest ID"})
		return
	}

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")

	ch := vehicleEvents.subscribe(manifestId)
	defer vehicleEvents.unsubscribe(manifestId, ch)

	// The tablet pulls the status endpoint on "ready", which covers anything it missed
	// while disconnected.
	c.SSEvent("ready", gin.H{"state": constants.SuccessState, "data": gin.H{"manifestId": manifestId}})
	c.Writer.Flush()

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case ev := <-ch:
			c.SSEvent("vehicle", gin.H{"state": constants.SuccessState, "data": ev})
			c.Writer.Flush()
		case <-ticker.C:
			c.Writer.WriteString(": keep-alive\n\n")
			c.Writer.Flush()
		case <-c.Request.Context().Done():
			return
		}
	}
}

// GetVehicleStatusChanges returns the inspection and publish status of the vehicles
// of a manifest that changed since the given cursor (all vehicles without one).
// Tablets call it when the manifest opens, on every event from VehicleEventsSSE, on a
// timer, on app resume and after reconnecting, and merge the rows by vehicleId.
func GetVehicleStatusChanges(c *gin.Context) {
	manifestId := c.Param("manifestId")
	if manifestId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"state": constants.ErrorState, "data": "Manifest ID is required"})
		return
	}

	since := c.Query("since")
	if since != "" {
		if _, err := time.Parse("2006-01-02 15:04:05.999999", since); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"state": constants.ErrorState, "data": "Invalid since cursor, expected YYYY-MM-DD HH:MM:SS[.ffffff]"})
			return
		}
	}

	st, rows, cursor, full := manifestdataservices.SelectVehicleStatusChanges(manifestId, since)
	if st.State != constants.SuccessState {
		c.JSON(http.StatusInternalServerError, gin.H{"state": constants.ErrorState, "data": "Failed to fetch vehicle status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"state": constants.SuccessState, "data": rows, "cursor": cursor, "full": full})
}
