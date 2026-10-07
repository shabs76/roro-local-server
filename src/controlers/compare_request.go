package controlers

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shabs76/roro-local-server/constants"
	manifestdataservices "github.com/shabs76/roro-local-server/database/manifest_data_services"
)

// CompareTabletWithServer compares a tablet's summary of a manifest with this
// server's records. The tablet shows the returned items with a suggested action each;
// the report is also stored for GetCompareReports.
func CompareTabletWithServer(c *gin.Context) {
	manifestId := c.Param("manifestId")
	if manifestId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"state": constants.ErrorState, "data": "Manifest ID is required"})
		return
	}

	var req manifestdataservices.CompareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"state": constants.ErrorState, "data": "Invalid comparison data was sent"})
		return
	}

	userId := c.GetString("userID")
	st, result := manifestdataservices.CompareTabletWithServer(manifestId, userId, req)
	if st.State != constants.SuccessState {
		c.JSON(http.StatusInternalServerError, st)
		return
	}

	c.JSON(http.StatusOK, gin.H{"state": constants.SuccessState, "data": result})
}

// GetCompareReports lists the latest comparison report of every tablet for a manifest.
func GetCompareReports(c *gin.Context) {
	manifestId := c.Param("manifestId")
	if manifestId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"state": constants.ErrorState, "data": "Manifest ID is required"})
		return
	}

	st, reports := manifestdataservices.SelectLatestCompareReports(manifestId)
	if st.State != constants.SuccessState {
		c.JSON(http.StatusInternalServerError, st)
		return
	}

	c.JSON(http.StatusOK, gin.H{"state": constants.SuccessState, "data": reports, "message": st.Data})
}

// GetCompareReport returns one stored comparison report with all its items.
func GetCompareReport(c *gin.Context) {
	st, report, overview := manifestdataservices.SelectCompareReport(c.Param("reportId"))
	if st.State != constants.SuccessState {
		status := http.StatusInternalServerError
		if st.Adv == "not_found" {
			status = http.StatusNotFound
		}
		c.JSON(status, st)
		return
	}

	c.JSON(http.StatusOK, gin.H{"state": constants.SuccessState, "data": gin.H{"device": overview, "report": report}})
}
