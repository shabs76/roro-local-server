package controlers

import (
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	apiservices "github.com/shabs76/roro-local-server/api_services"
	"github.com/shabs76/roro-local-server/constants"
	"github.com/shabs76/roro-local-server/constants/modules/manifest"
	"github.com/shabs76/roro-local-server/constants/modules/users"
	manifestdataservices "github.com/shabs76/roro-local-server/database/manifest_data_services"
	usersdataservices "github.com/shabs76/roro-local-server/database/users_data_services"
	"github.com/shabs76/roro-local-server/gendb"
	"github.com/shabs76/roro-local-server/pkg/middleware"
)

func GetManifestListFromRemote(c *gin.Context) {
	// 1.0 Obtain AuthSession from context
	val, ok := c.Get("authSession")
	if !ok {
		c.JSON(http.StatusUnauthorized, constants.NormalResponse{
			State:   constants.ErrorState,
			Data:    "Unauthorized",
			Message: "Authentication required to access this resource",
			Adv:     "none",
		})
		return
	}

	authSession, ok := val.(middleware.AuthSession)
	if !ok {
		c.JSON(http.StatusUnauthorized, constants.NormalResponse{
			State:   constants.ErrorState,
			Data:    "Unauthorized",
			Message: "Authentication required to access this resource",
			Adv:     "none",
		})
		return
	}

	// 2.0 Extract query parameters for pagination and filtering
	query := c.Query("query")
	page := c.Query("page")
	pageNum, err := strconv.Atoi(page)
	if err != nil {
		slog.Error(err.Error())
		c.JSON(http.StatusBadRequest, constants.NormalResponse{
			State:   constants.ErrorState,
			Data:    "Invalid page number",
			Message: "Invalid page number provided",
			Adv:     "none",
		})
		return
	}
	limit := c.Query("limit")
	limitNum, err := strconv.Atoi(limit)
	if err != nil {
		slog.Error(err.Error())
		c.JSON(http.StatusBadRequest, constants.NormalResponse{
			State:   constants.ErrorState,
			Data:    "Invalid limit number",
			Message: "Invalid limit number provided",
			Adv:     "none",
		})
		return
	}

	// 3.0 Call the API service to get manifest list
	manifests, err := apiservices.GetManifestLists(authSession.LogID, authSession.LogKey, query, pageNum, limitNum)
	if err != nil {
		slog.Error(err.Error())
		c.JSON(http.StatusInternalServerError, constants.NormalResponse{
			State:   constants.ErrorState,
			Data:    "Failed to fetch manifest list",
			Message: "An error occurred while fetching the manifest list",
			Adv:     "none",
		})
		return
	}

	// 4.0 Return the manifest list in the response
	c.JSON(http.StatusOK, constants.NormalResponse{
		State:   constants.SuccessState,
		Data:    manifests,
		Message: "Manifest list fetched successfully",
		Adv:     "none",
	})

}

func GetManifestesDetails(c *gin.Context) {
	pg := c.DefaultQuery("page", "1")
	ipg, erP := strconv.Atoi(pg)
	if erP != nil {
		slog.Error(erP.Error())
		c.JSON(http.StatusInternalServerError, constants.NormalResponse{
			State:   constants.ErrorState,
			Data:    "Failed to obtain page number",
			Message: "An error occurred while obtaining the page number",
			Adv:     "none",
		})
		return
	}
	// Get query parameter for filtering manifests
	query := c.Query("query")
	limit := c.DefaultQuery("limit", "8")
	ppg, err := strconv.Atoi(limit)
	if err != nil || ppg <= 0 {
		ppg = 8 // default to 8 if conversion fails or invalid value
	}

	// Ensure perPage does not exceed a maximum limit, e.g., 30
	if ppg > 30 {
		ppg = 30
	}

	subQr := " `manifest_id` != ? ORDER BY `uploaded_date` DESC"
	vals := []any{" "}
	if query != "" {
		subQr = " (`manifest_name` LIKE ? OR `vessel_name` LIKE ? OR `voyage_no` LIKE ? OR `manifest_id` = ? OR manifest.client_id = ? OR client_name LIKE ? OR berth_no LIKE ?) ORDER BY `uploaded_date` DESC"
		vals = []any{"%" + query + "%", "%" + query + "%", "%" + query + "%", query, query, "%" + query + "%", "%" + query + "%"}
	}

	qr := "SELECT COUNT(manifest_id) AS numbers FROM `manifest` INNER JOIN clients ON manifest.client_id = clients.client_id WHERE " + subQr

	stx, rez := gendb.PagenationSelect(qr, ipg, ppg, vals)
	if stx.State != constants.SuccessState {
		c.JSON(http.StatusInternalServerError, constants.NormalResponse{
			State:   constants.ErrorState,
			Data:    "Failed to obtain pagination limit",
			Message: "An error occurred while obtaining the pagination limit",
			Adv:     "none",
		})
		return
	} else if rez.State == "end" {
		c.JSON(http.StatusOK, constants.PaginationResponse{
			State:   constants.SuccessState,
			Data:    []any{},
			Message: "No manifests found",
			Pagination: constants.PaginationInfo{
				TotalItems:   rez.ResNum,
				TotalPages:   rez.TotalPages,
				CurrentPage:  ipg,
				ItemsPerPage: ppg,
				HasNextPage:  false,
				HasPrevPage:  false,
			},
		})
		return
	}
	st, manifests := manifestdataservices.SelectManifestInfo(subQr+" "+rez.Limit, vals)
	if st.State != constants.SuccessState {
		slog.Error(st.Data)
		c.JSON(http.StatusInternalServerError, constants.NormalResponse{
			State:   constants.ErrorState,
			Data:    "Failed to obtain manifestes list",
			Message: "An error occurred while obtaining the manifestes list",
			Adv:     "none",
		})
		return
	} else if len(manifests) <= 0 {
		c.JSON(http.StatusOK, constants.PaginationResponse{
			State:   constants.SuccessState,
			Data:    []any{},
			Message: "No manifests found on search criteria",
			Pagination: constants.PaginationInfo{
				TotalItems:   rez.ResNum,
				TotalPages:   rez.TotalPages,
				CurrentPage:  ipg,
				ItemsPerPage: ppg,
				HasNextPage:  false,
				HasPrevPage:  false,
			},
		})
		return
	}
	// Calculate total pages based on the number of manifests
	totalPages := int(math.Ceil(float64(rez.ResNum) / float64(ppg)))

	manifestIDs := make([]string, len(manifests))
	manifestArgs := make([]any, len(manifests))
	manifestAgg := make(map[string]*manifestAggregation, len(manifests))

	for i := range manifests {
		manifestIDs[i] = manifests[i].ManifestId
		manifestArgs[i] = manifests[i].ManifestId
		manifestAgg[manifests[i].ManifestId] = &manifestAggregation{}
	}

	if st, err := populateManifestAggregates(manifestIDs, manifestArgs, manifestAgg); err != nil {
		slog.Error(err.Error())
		if st != nil && st.State != constants.SuccessState {
			c.JSON(http.StatusInternalServerError, constants.NormalResponse{
				State:   constants.ErrorState,
				Data:    st.Data,
				Message: "An error occurred while obtaining manifest statistics",
				Adv:     "none",
			})
		} else {
			c.JSON(http.StatusInternalServerError, constants.NormalResponse{
				State:   constants.ErrorState,
				Data:    "Failed to obtain manifest statistics",
				Message: "An error occurred while obtaining manifest statistics",
				Adv:     "none",
			})
		}
		return
	}

	var senMani = []manifest.ManifestToshow{}
	for i := range manifests {
		agg := manifestAgg[manifests[i].ManifestId]
		senMani = append(senMani, manifest.ManifestToshow{
			ManifestId:        manifests[i].ManifestId,
			ManifestName:      manifests[i].ManifestName,
			ClientName:        manifests[i].ClientName,
			VesselName:        manifests[i].VesselName,
			VoyageNo:          manifests[i].VoyageNo,
			BerthNo:           manifests[i].BerthNo,
			ArrivalDate:       manifests[i].ArrivalDate,
			ReceivedDate:      manifests[i].ReceivedDate,
			UploadedDate:      manifests[i].UploadedDate,
			VehiclesNumber:    agg.totalVehicles,
			Inspected:         agg.inspectedVehicles,
			VehiclesDamaged:   agg.damagedVehicles,
			VehicleDischarged: agg.dischargedVehicles,
			VehicleRemarks:    agg.remarkedVehicles,
			TotalPackages:     agg.totalPackages,
			InspectedPackages: agg.inspectedPackages,
		})
	}

	c.JSON(http.StatusOK, constants.PaginationResponse{
		State:   constants.SuccessState,
		Data:    senMani,
		Message: "Manifest details fetched successfully",
		Pagination: constants.PaginationInfo{
			TotalItems:   rez.ResNum,
			TotalPages:   totalPages,
			CurrentPage:  ipg,
			ItemsPerPage: ppg,
			HasNextPage:  ipg < totalPages,
			HasPrevPage:  ipg > 1,
		},
	})
}

func GetManifestData(c *gin.Context) {
	query := c.Query("query")
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		c.JSON(http.StatusBadRequest, constants.NormalResponse{
			State:   constants.ErrorState,
			Data:    "Invalid page number",
			Message: "Invalid page number provided",
			Adv:     "none",
		})
		return
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		c.JSON(http.StatusBadRequest, constants.NormalResponse{
			State:   constants.ErrorState,
			Data:    "Invalid limit number",
			Message: "Invalid limit number provided",
			Adv:     "none",
		})
		return
	}

	subQuery := " `manifest_id` != ?"
	vals := []any{" "}
	if query != "" {
		subQuery = " (`manifest_name` LIKE ? OR `vessel_name` LIKE ? OR `voyage_no` LIKE ? OR `manifest_id` = ? OR manifest.client_id = ? OR client_name LIKE ? OR berth_no LIKE ?) "
		vals = []any{"%" + query + "%", "%" + query + "%", "%" + query + "%", query, query, "%" + query + "%", "%" + query + "%"}
	}
	numQuery := "SELECT COUNT(manifest_id) AS numbers FROM `manifest` INNER JOIN clients ON manifest.client_id = clients.client_id WHERE " + subQuery
	stx, rez := gendb.PagenationSelect(numQuery, page, limit, vals)
	if stx.State != constants.SuccessState {
		c.JSON(http.StatusInternalServerError, constants.NormalResponse{
			State:   constants.ErrorState,
			Data:    "Failed to obtain pagination limit",
			Message: "An error occurred while obtaining pagination limit",
			Adv:     "none",
		})
		return
	} else if rez.State == "end" {
		c.JSON(http.StatusOK, constants.PaginationResponse{
			State:   constants.SuccessState,
			Data:    []any{},
			Message: "No data found",
			Pagination: constants.PaginationInfo{
				TotalItems:   rez.ResNum,
				TotalPages:   rez.TotalPages,
				CurrentPage:  page,
				ItemsPerPage: limit,
				HasNextPage:  page < rez.TotalPages,
				HasPrevPage:  page > 1,
			},
		})
		return
	}

	st, manifests := manifestdataservices.SelectManifestInfo(subQuery+" ORDER BY `uploaded_date` DESC "+rez.Limit, vals)
	if st.State != constants.SuccessState {
		slog.Error(st.Data)
		c.JSON(http.StatusInternalServerError, constants.NormalResponse{
			State:   constants.ErrorState,
			Data:    "Failed to obtain manifests list",
			Message: "An error occurred while obtaining manifests list",
			Adv:     "none",
		})
		return
	} else if len(manifests) <= 0 {
		c.JSON(http.StatusOK, constants.PaginationResponse{
			State:   constants.SuccessState,
			Data:    []any{},
			Message: "No data found",
			Pagination: constants.PaginationInfo{
				TotalItems:   rez.ResNum,
				TotalPages:   rez.TotalPages,
				CurrentPage:  page,
				ItemsPerPage: limit,
				HasNextPage:  page < rez.TotalPages,
				HasPrevPage:  page > 1,
			},
		})
		return
	}
	c.JSON(http.StatusOK, constants.PaginationResponse{
		State:   constants.SuccessState,
		Data:    manifests,
		Message: "Manifest details fetched successfully",
		Pagination: constants.PaginationInfo{
			TotalItems:   rez.ResNum,
			TotalPages:   rez.TotalPages,
			CurrentPage:  page,
			ItemsPerPage: limit,
			HasNextPage:  page < rez.TotalPages,
			HasPrevPage:  page > 1,
		},
	})
}

type manifestAggregation struct {
	totalPackages      int
	inspectedPackages  int
	totalVehicles      int
	inspectedVehicles  int
	dischargedVehicles int
	damagedVehicles    int
	remarkedVehicles   int
}

func buildInPlaceholders(count int) string {
	if count <= 0 {
		return ""
	}

	placeholders := make([]string, count)
	for i := range placeholders {
		placeholders[i] = "?"
	}
	return strings.Join(placeholders, ",")
}

func populateManifestAggregates(manifestIDs []string, manifestArgs []any, aggregates map[string]*manifestAggregation) (*constants.AnswerState, error) {
	if len(manifestIDs) == 0 {
		return &constants.AnswerState{State: constants.SuccessState, Data: "success", Adv: "none"}, nil
	}

	inClause := buildInPlaceholders(len(manifestIDs))

	query := fmt.Sprintf(`
		SELECT 
			m.manifest_id,
			COALESCE(pkg.total_packages, 0) AS totalPackages,
			COALESCE(pkg.inspected_packages, 0) AS inspectedPackages,
			COALESCE(veh.total_vehicles, 0) AS totalVehicles,
			COALESCE(veh.inspected_vehicles, 0) AS inspectedVehicles,
			COALESCE(veh.discharged_vehicles, 0) AS dischargedVehicles,
			COALESCE(dmg.damaged_vehicles, 0) AS damagedVehicles,
			COALESCE(rem.remarked_vehicles, 0) AS remarkedVehicles
		FROM manifest m
		LEFT JOIN (
			SELECT manifest_id,
			       COUNT(*) AS total_packages,
			       SUM(CASE WHEN is_inspected = ? THEN 1 ELSE 0 END) AS inspected_packages
			FROM manifest_packages
			WHERE manifest_id IN (%s)
			GROUP BY manifest_id
		) pkg ON pkg.manifest_id = m.manifest_id
		LEFT JOIN (
			SELECT mv.manifest_id,
			       COUNT(*) AS total_vehicles,
			       COUNT(DISTINCT CASE WHEN mv.inspection_status = ? OR vt.vehicle_id IS NOT NULL THEN mv.vehicle_id END) AS inspected_vehicles,
			       SUM(CASE WHEN mv.discharged_status = ? THEN 1 ELSE 0 END) AS discharged_vehicles
			FROM manifest_vehicles mv
			LEFT JOIN vehicles_talling vt ON vt.vehicle_id = mv.vehicle_id
			WHERE mv.manifest_id IN (%s)
			GROUP BY mv.manifest_id
		) veh ON veh.manifest_id = m.manifest_id
		LEFT JOIN (
			SELECT mv.manifest_id,
			       COUNT(DISTINCT mv.vehicle_id) AS damaged_vehicles
			FROM manifest_vehicles mv
			INNER JOIN (
				SELECT DISTINCT vehicle_id
				FROM vehicles_inspection
				WHERE status = ? OR status = ?
			) dmg_vi ON dmg_vi.vehicle_id = mv.vehicle_id
			WHERE mv.manifest_id IN (%s)
			GROUP BY mv.manifest_id
		) dmg ON dmg.manifest_id = m.manifest_id
		LEFT JOIN (
			SELECT mv.manifest_id,
			       COUNT(DISTINCT ir.vehicle_id) AS remarked_vehicles
			FROM inspection_remarks ir
			INNER JOIN manifest_vehicles mv ON mv.vehicle_id = ir.vehicle_id
			LEFT JOIN (
				SELECT DISTINCT vehicle_id
				FROM vehicles_inspection
				WHERE status = ? OR status = ?
			) rem_dmg ON rem_dmg.vehicle_id = ir.vehicle_id
			WHERE mv.manifest_id IN (%s)
			  AND rem_dmg.vehicle_id IS NULL
			GROUP BY mv.manifest_id
		) rem ON rem.manifest_id = m.manifest_id
		WHERE m.manifest_id IN (%s)
	`, inClause, inClause, inClause, inClause, inClause)

	args := make([]any, 0, len(manifestArgs)*5+7)
	args = append(args, manifest.InspectionStatus.Yes)
	args = append(args, manifestArgs...)
	args = append(args, manifest.InspectionStatus.Yes, manifest.InspectionStatus.Yes)
	args = append(args, manifestArgs...)
	args = append(args, manifest.InspectionMarkStatus.Damaged, manifest.InspectionMarkStatus.Missing)
	args = append(args, manifestArgs...)
	args = append(args, manifest.InspectionMarkStatus.Damaged, manifest.InspectionMarkStatus.Missing)
	args = append(args, manifestArgs...)
	args = append(args, manifestArgs...)

	st, rows := gendb.SelectGeneral(query, args)
	if st.State != constants.SuccessState {
		return st, fmt.Errorf("%s", st.Data)
	}
	defer rows.Close()

	for rows.Next() {
		var manifestID string
		var totalPackages int
		var inspectedPackages int
		var totalVehicles int
		var inspectedVehicles int
		var dischargedVehicles int
		var damagedVehicles int
		var remarkedVehicles int
		if err := rows.Scan(&manifestID, &totalPackages, &inspectedPackages, &totalVehicles, &inspectedVehicles, &dischargedVehicles, &damagedVehicles, &remarkedVehicles); err != nil {
			return &constants.AnswerState{State: constants.ErrorState, Data: "Failed to bind manifest statistics", Adv: "none"}, err
		}
		agg := aggregates[manifestID]
		if agg == nil {
			agg = &manifestAggregation{}
			aggregates[manifestID] = agg
		}
		agg.totalPackages = totalPackages
		agg.inspectedPackages = inspectedPackages
		agg.totalVehicles = totalVehicles
		agg.inspectedVehicles = inspectedVehicles
		agg.dischargedVehicles = dischargedVehicles
		agg.damagedVehicles = damagedVehicles
		agg.remarkedVehicles = remarkedVehicles
	}

	if err := rows.Err(); err != nil {
		return &constants.AnswerState{State: constants.ErrorState, Data: "Failed to iterate manifest statistics", Adv: "none"}, err
	}

	return &constants.AnswerState{State: constants.SuccessState, Data: "success", Adv: "none"}, nil
}

func GetVehicleListOfManifest(c *gin.Context) {
	// 1. Pagination & Limits
	page := c.DefaultQuery("page", "1")
	ipg, err := strconv.Atoi(page)
	if err != nil {
		c.JSON(http.StatusInternalServerError, constants.NormalResponse{
			State:   constants.ErrorState,
			Data:    "Failed to obtain page number",
			Message: "An error occurred while obtaining the page number",
			Adv:     "none",
		})
		return
	}

	limit := 250
	if l := c.Query("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 {
			limit = val
		}
	}

	// 2. Params
	mId := c.Param("manifestId")
	searchQuery := c.Query("query")
	reqType := c.Query("type")

	// 3. Logic Branching
	if reqType == "damaged" {
		getManifestDamagedVehiclesLogic(c, mId, ipg, limit, searchQuery)
		return
	}

	// Standard logic for others
	subQr := " `manifest_id` = ? "
	vals := []any{mId}

	switch strings.TrimSpace(reqType) {
	case "discharged":
		subQr += " AND `discharged_status` = ? "
		vals = append(vals, "yes")
	case "undischarged":
		subQr += " AND `discharged_status` != ? "
		vals = append(vals, "yes")
	case "inspected":
		subQr += " AND (`inspection_status` = ? OR EXISTS (SELECT 1 FROM `vehicles_talling` WHERE `vehicles_talling`.`vehicle_id` = `manifest_vehicles`.`vehicle_id`)) "
		vals = append(vals, "yes")
	case "uninspected":
		subQr += " AND `inspection_status` != ? AND NOT EXISTS (SELECT 1 FROM `vehicles_talling` WHERE `vehicles_talling`.`vehicle_id` = `manifest_vehicles`.`vehicle_id`) "
		vals = append(vals, "yes")
	}

	if searchQuery != "" {
		subQr += " AND ( `chasis_number` LIKE ? OR `model` LIKE ? OR `description` LIKE ? OR `weight` LIKE ? OR `bl_no` LIKE ? )"
		s := "%" + searchQuery + "%"
		vals = append(vals, s, s, s, s, s)
	}

	// 4. Count & Paging
	qr := "SELECT COUNT(vehicle_id) AS numbers FROM `manifest_vehicles` WHERE " + subQr

	stx, rez := gendb.PagenationSelect(qr, ipg, limit, vals)
	if stx.State != constants.SuccessState {
		c.JSON(http.StatusInternalServerError, constants.NormalResponse{
			State:   constants.ErrorState,
			Data:    "Failed to obtain pagination limit",
			Message: "An error occurred while obtaining the pagination limit",
			Adv:     "none",
		})
		return
	}
	totalPages := int(math.Ceil(float64(rez.ResNum) / float64(limit)))

	if rez.State == "end" {
		c.JSON(http.StatusOK, constants.PaginationResponse{
			State:   constants.SuccessState,
			Data:    []any{},
			Message: "No vehicles found",
			Pagination: constants.PaginationInfo{
				TotalItems:   rez.ResNum,
				TotalPages:   totalPages,
				CurrentPage:  ipg,
				ItemsPerPage: limit,
				HasNextPage:  false,
				HasPrevPage:  ipg > 1,
			},
		})
		return
	}

	// 5. Fetch Data
	st, vehicles := manifestdataservices.SelectVehicleInfo(subQr+" ORDER BY `vehicle_id` DESC "+rez.Limit, vals)
	if st.State != constants.SuccessState {
		slog.Error(st.Data)
		c.JSON(http.StatusInternalServerError, constants.NormalResponse{
			State:   constants.ErrorState,
			Data:    "Failed to obtain vehicles list",
			Message: "An error occurred while fetching the vehicles list",
			Adv:     "none",
		})
		return
	}

	if len(vehicles) == 0 {
		c.JSON(http.StatusOK, constants.PaginationResponse{
			State:   constants.SuccessState,
			Data:    []any{},
			Message: "No vehicles found",
			Pagination: constants.PaginationInfo{
				TotalItems:   rez.ResNum,
				TotalPages:   totalPages,
				CurrentPage:  ipg,
				ItemsPerPage: limit,
				HasNextPage:  false,
				HasPrevPage:  ipg > 1,
			},
		})
		return
	}

	// 6. Enrich
	vhShow := enrichVehicleDetails(vehicles)
	c.JSON(http.StatusOK, constants.PaginationResponse{
		State:   constants.SuccessState,
		Data:    vhShow,
		Message: "Vehicles fetched successfully",
		Pagination: constants.PaginationInfo{
			TotalItems:   rez.ResNum,
			TotalPages:   totalPages,
			CurrentPage:  ipg,
			ItemsPerPage: limit,
			HasNextPage:  ipg < totalPages,
			HasPrevPage:  ipg > 1,
		},
	})
}

func getManifestDamagedVehiclesLogic(c *gin.Context, mId string, page, limit int, query string) {
	subQr := " `manifest_id` = ? AND (vehicles_inspection.status = ? OR vehicles_inspection.status = ?) "
	vals := []any{mId, manifest.InspectionMarkStatus.Damaged, manifest.InspectionMarkStatus.Missing}

	if query != "" {
		subQr += " AND (chasis_number LIKE ? OR model LIKE ? OR description LIKE ? OR weight LIKE ? OR bl_no LIKE ?)"
		s := "%" + query + "%"
		vals = append(vals, s, s, s, s, s)
	}

	qr := "SELECT COUNT(DISTINCT manifest_vehicles.vehicle_id) FROM `manifest_vehicles` INNER JOIN vehicles_inspection ON manifest_vehicles.vehicle_id = vehicles_inspection.vehicle_id WHERE " + subQr

	stx, rez := gendb.PagenationSelect(qr, page, limit, vals)
	if stx.State != constants.SuccessState {
		c.JSON(http.StatusInternalServerError, constants.NormalResponse{
			State:   constants.ErrorState,
			Data:    "Failed to obtain pagination limit",
			Message: "An error occurred while obtaining the pagination limit",
			Adv:     "none",
		})
		return
	}
	totalPages := int(math.Ceil(float64(rez.ResNum) / float64(limit)))

	if rez.State == "end" {
		c.JSON(http.StatusOK, constants.PaginationResponse{
			State:   constants.SuccessState,
			Data:    []any{},
			Message: "No damaged vehicles found",
			Pagination: constants.PaginationInfo{
				TotalItems:   rez.ResNum,
				TotalPages:   totalPages,
				CurrentPage:  page,
				ItemsPerPage: limit,
				HasNextPage:  false,
				HasPrevPage:  page > 1,
			},
		})
		return
	}

	st, vehicles := manifestdataservices.SelectVehiclesAndInspectionDetails(subQr+" GROUP BY manifest_vehicles.vehicle_id  "+rez.Limit, vals)
	if st.State != constants.SuccessState {
		slog.Error(st.Data)
		c.JSON(http.StatusInternalServerError, constants.NormalResponse{
			State:   constants.ErrorState,
			Data:    "Failed to obtain vehicles list",
			Message: "An error occurred while fetching the vehicles list",
			Adv:     "none",
		})
		return
	}

	// Remarks Logic
	strem, remarks := manifestdataservices.SelectInspectionRemarksForGivenManifest(" `manifest_id` = ? AND remark_type = ? GROUP BY manifest_vehicles.vehicle_id", []any{mId, manifest.RemarkStatus.Damaged})
	if strem.State != constants.SuccessState {
		slog.Error(strem.Data)
		c.JSON(http.StatusInternalServerError, constants.NormalResponse{
			State:   constants.ErrorState,
			Data:    "Failed to fetch vehicle inspection remarks details",
			Message: "An error occurred while fetching vehicle inspection remarks details: " + strem.Data,
			Adv:     "none",
		})
		return
	}

	existingVehicleIDs := make(map[string]struct{}, len(vehicles))
	for _, insp := range vehicles {
		existingVehicleIDs[insp.VehicleId] = struct{}{}
	}

	remarkOnlyVehicles := make([]manifest.InspectionRemarksDetails, 0, len(remarks))
	processedRemarkIDs := make(map[string]struct{}, len(remarks))
	for _, remark := range remarks {
		if _, exists := existingVehicleIDs[remark.VehicleId]; exists {
			continue
		}
		if _, alreadyProcessed := processedRemarkIDs[remark.VehicleId]; alreadyProcessed {
			continue
		}
		remarkOnlyVehicles = append(remarkOnlyVehicles, remark)
		processedRemarkIDs[remark.VehicleId] = struct{}{}
	}

	vehicleIds := []any{}
	subQuery := ""
	for _, remark := range remarkOnlyVehicles {
		if subQuery != "" {
			subQuery += " OR "
		}
		subQuery += " manifest_vehicles.vehicle_id = ? "
		vehicleIds = append(vehicleIds, remark.VehicleId)
	}
	if subQuery != "" {
		st, vehiclesRem := manifestdataservices.SelectVehiclesAndInspectionDetails(subQuery+" GROUP BY manifest_vehicles.vehicle_id  "+rez.Limit, vehicleIds)
		if st.State != constants.SuccessState {
			slog.Error(st.Data)
			c.JSON(http.StatusInternalServerError, constants.NormalResponse{
				State:   constants.ErrorState,
				Data:    "Failed to fetch vehicles details for remarks",
				Message: "An error occurred while fetching vehicles details for remarks: " + st.Data,
				Adv:     "none",
			})
			return
		}
		for _, veh := range vehiclesRem {
			if _, exists := existingVehicleIDs[veh.VehicleId]; exists {
				continue
			}
			vehicles = append(vehicles, veh)
			existingVehicleIDs[veh.VehicleId] = struct{}{}
		}
	}

	// Convert to standard VehiclesDetails for enrichment
	stdVehicles := make([]manifest.VehicleData, len(vehicles))
	for i, v := range vehicles {
		stdVehicles[i] = manifest.VehicleData{
			VehicleId:        v.VehicleId,
			ManifestId:       v.ManifestId,
			ChasisNumber:     v.ChasisNumber,
			VehicleModel:     v.VehicleModel,
			Description:      v.Description,
			Weight:           v.Weight,
			BLNumber:         v.BLNumber,
			CreationDate:     v.CreationDate,
			InspectionStatus: v.InspectionStatus,
			TalliedStatus:    v.TalliedStatus,
			DischargeStatus:  v.DischargeStatus,
			InspectionTime:   v.InspectionTime,
			TalliedTime:      v.TalliedTime,
			DischargeTime:    v.DischargeTime,
			OverLandStatus:   v.OverLandStatus,
		}
	}

	// Enrich
	vhShow := enrichVehicleDetails(stdVehicles)
	c.JSON(http.StatusOK, constants.PaginationResponse{
		State:   constants.SuccessState,
		Data:    vhShow,
		Message: "Damaged vehicles fetched successfully",
		Pagination: constants.PaginationInfo{
			TotalItems:   rez.ResNum,
			TotalPages:   totalPages,
			CurrentPage:  page,
			ItemsPerPage: limit,
			HasNextPage:  page < totalPages,
			HasPrevPage:  page > 1,
		},
	})
}

func enrichVehicleDetails(vehicles []manifest.VehicleData) []manifest.VehiclesDetailsToShow {
	var vhShow = []manifest.VehiclesDetailsToShow{}
	if len(vehicles) == 0 {
		return vhShow
	}

	ids := make([]string, len(vehicles))
	for i := range vehicles {
		ids[i] = vehicles[i].VehicleId
	}
	ste, extras := manifestdataservices.SelectVehicleListExtras(ids)
	if ste.State != constants.SuccessState {
		slog.Error(ste.Data)
	}

	for i := range vehicles {
		maker := "notset"
		body := "notset"
		deck := "notset"
		image := ""
		damaged := "no"
		numberOfKeys := 0
		inspectedBy := ""
		inspectionCount := 0

		if e := extras[vehicles[i].VehicleId]; e != nil {
			if e.HasActiveTally {
				maker = e.Maker
				body = e.BodyType
				deck = e.DeckNumber
				image = e.Image
				numberOfKeys = e.NumberOfKeys
				inspectedBy = e.InspectedBy
				inspectionCount = 1
			}
			if e.IsDamaged {
				damaged = "yes"
			}
			inspectionCount += e.HistoryBatches
		}

		vhShow = append(vhShow, manifest.VehiclesDetailsToShow{
			VehicleId:        vehicles[i].VehicleId,
			ManifestId:       vehicles[i].ManifestId,
			ChasisNumber:     vehicles[i].ChasisNumber,
			VehicleModel:     vehicles[i].VehicleModel,
			Description:      vehicles[i].Description,
			Weight:           vehicles[i].Weight,
			BLNumber:         vehicles[i].BLNumber,
			CreationDate:     vehicles[i].CreationDate,
			InspectionStatus: vehicles[i].InspectionStatus,
			TalliedStatus:    vehicles[i].TalliedStatus,
			DischargeStatus:  vehicles[i].DischargeStatus,
			InspectionTime:   vehicles[i].InspectionTime,
			TalliedTime:      vehicles[i].TalliedTime,
			DischargeTime:    vehicles[i].DischargeTime,
			IsOverLand:       vehicles[i].OverLandStatus,
			Maker:            maker,
			BodyType:         body,
			Image:            image,
			DeckNumber:       deck,
			IsDamaged:        damaged,
			NumberOfKeys:     numberOfKeys,
			IsPublished:      vehicles[i].IsPublished,
			InspectedBy:      inspectedBy,
			InspectionCount:  inspectionCount,
		})
	}
	return vhShow
}

func GetVehicleInspectionsDetails(c *gin.Context) {
	vId := c.Param("vehicleId")

	st, vehicleDets := manifestdataservices.SelectVehicleAndTallyDetails(" manifest_vehicles.vehicle_id = ? AND `inspection_status` = ?", []any{vId, "yes"})
	if st.State != constants.SuccessState {
		slog.Error(st.Data)
		c.JSON(http.StatusInternalServerError, constants.NormalResponse{
			State:   constants.ErrorState,
			Data:    "Failed to fetch vehicle details",
			Message: "An error occurred while fetching vehicle details: " + st.Data,
			Adv:     "none",
		})
		return
	} else if len(vehicleDets) <= 0 {
		c.JSON(http.StatusOK, constants.NormalResponse{
			State:   constants.ErrorState,
			Data:    []any{},
			Message: "Vehicle details not found",
			Adv:     "none",
		})
		return
	}

	// get inspection detail list
	sti, insps := manifestdataservices.SelectInspectionDetails(" `vehicle_id` = ? ", []any{vId})
	if sti.State != constants.SuccessState {
		slog.Error(sti.Data)
		c.JSON(http.StatusInternalServerError, constants.NormalResponse{
			State:   constants.ErrorState,
			Data:    "Failed to fetch vehicle inspection details",
			Message: "An error occurred while fetching vehicle inspection details: " + sti.Data,
			Adv:     "none",
		})
		return
	}
	var inspImg = []manifest.InspectionDetailsAndImage{}
	for i := range insps {
		img := ""
		stm, imgs := manifestdataservices.SelectInspectionImages(" `inspection_id` = ?", []any{insps[i].InspectionId})
		if stm.State != constants.SuccessState {
			slog.Error(stm.Data)
			c.JSON(http.StatusInternalServerError, constants.NormalResponse{
				State:   constants.ErrorState,
				Data:    "Failed to fetch inspection images",
				Message: "An error occurred while fetching inspection images: " + stm.Data,
				Adv:     "none",
			})
			return
		} else if len(imgs) > 0 {
			img = imgs[0].ImageLink
		}

		inspImg = append(inspImg, manifest.InspectionDetailsAndImage{
			InspectionId: insps[i].InspectionId,
			VehicleId:    insps[i].VehicleId,
			CheckId:      insps[i].CheckId,
			CheckName:    insps[i].CheckName,
			UserId:       vehicleDets[0].UserId,
			Status:       insps[i].Status,
			Checktime:    insps[i].Checktime,
			Image:        img,
		})
	}

	stu, users := usersdataservices.SelectUserDetailsWithRolesPass(" `user_id` = ? ", []any{inspImg[0].UserId})
	if stu.State != constants.SuccessState {
		slog.Error(stu.Data)
		c.JSON(http.StatusInternalServerError, constants.NormalResponse{
			State:   constants.ErrorState,
			Data:    "Failed to fetch user details",
			Message: "An error occurred while fetching user details: " + stu.Data,
			Adv:     "none",
		})
		return
	} else if len(users) <= 0 {
		c.JSON(http.StatusInternalServerError, constants.NormalResponse{
			State:   constants.ErrorState,
			Data:    "Failed to fetch user details",
			Message: "User details not found",
			Adv:     "none",
		})
		return
	}

	// get inspection remarks
	str, rem := manifestdataservices.SelectInspectionRemarks(" `vehicle_id` = ? ", []any{vId})
	if str.State != constants.SuccessState {
		slog.Error(str.Data)
		c.JSON(http.StatusInternalServerError, constants.NormalResponse{
			State:   constants.ErrorState,
			Data:    "Failed to fetch vehicle inspection remarks details",
			Message: "An error occurred while fetching vehicle inspection remarks details: " + str.Data,
			Adv:     "none",
		})
		return
	}

	// get packages on board
	st, packs := manifestdataservices.SelectOnBoardPackage(" `vehicle_id` = ?", []any{vId})
	if st.State != constants.SuccessState {
		slog.Error(st.Data)
		c.JSON(http.StatusInternalServerError, constants.NormalResponse{
			State:   constants.ErrorState,
			Data:    "Failed to fetch on board package details",
			Message: "An error occurred while fetching on board package details: " + st.Data,
			Adv:     "none",
		})
		return
	}

	slog.Info("Packages fetched for vehicle ", vId, " are ")
	for _, p := range packs {
		slog.Info(fmt.Sprintf("PackID: %s, Pack Title %s", p.PackageId, p.Title))
	}

	// get other media
	st, media := manifestdataservices.SelectVehicleMedia(" vehicle_id = ? ", []any{vId})
	if st.State != constants.SuccessState {
		slog.Error(st.Data)
		c.JSON(http.StatusInternalServerError, constants.NormalResponse{
			State:   constants.ErrorState,
			Data:    "Failed to fetch vehicle media details",
			Message: "An error occurred while fetching vehicle media details: " + st.Data,
			Adv:     "none",
		})
		return
	}

	// discharge image
	stim, dis := manifestdataservices.SelectDischargeDetails(" `vehicle_id` = ? ", []any{vId})
	if stim.State != constants.SuccessState {
		slog.Error(stim.Data)
		c.JSON(http.StatusInternalServerError, constants.NormalResponse{
			State:   constants.ErrorState,
			Data:    "Failed to fetch vehicle discharge details",
			Message: "An error occurred while fetching vehicle discharge details: " + stim.Data,
			Adv:     "none",
		})
		return
	} else if len(dis) < 1 {
		vdataz := VehicleInspectionDetailsRes{
			VehicleDets:   vehicleDets[0],
			Inspections:   inspImg,
			Remarks:       rem,
			DischargeInfo: []manifest.VehicleDischargeDetails{},
			UserDetails:   users[0],
			Packages:      packs,
			Media:         media,
		}
		c.JSON(http.StatusOK, constants.NormalResponse{
			State:   constants.SuccessState,
			Data:    vdataz,
			Message: "Vehicle inspection details fetched successfully",
			Adv:     "none",
		})
		return
	}

	userDets := users[0]
	userDets.Password = "" // remove password from response

	vdata := VehicleInspectionDetailsRes{
		VehicleDets:   vehicleDets[0],
		Inspections:   inspImg,
		Remarks:       rem,
		DischargeInfo: dis,
		UserDetails:   userDets,
		Media:         media,
		Packages:      packs,
	}

	c.JSON(http.StatusOK, constants.NormalResponse{
		State:   constants.SuccessState,
		Data:    vdata,
		Message: "Vehicle inspection details fetched successfully",
		Adv:     "none",
	})
}

func GetVehicleShortInfo(c *gin.Context) {
	manifestId := c.Param("manifestId")
	query := c.Query("query")

	if query == "" {
		c.JSON(http.StatusBadRequest, constants.NormalResponse{
			State:   constants.ErrorState,
			Data:    "Query parameter is required",
			Message: "Please provide a query to search for the vehicle",
			Adv:     "none",
		})
		return
	}

	page := c.DefaultQuery("page", "1")
	ipg, err := strconv.Atoi(page)
	if err != nil {
		c.JSON(http.StatusInternalServerError, constants.NormalResponse{
			State:   constants.ErrorState,
			Data:    "Failed to obtain page number",
			Message: "An error occurred while obtaining the page number",
			Adv:     "none",
		})
		return
	}

	limit := 10
	if l := c.Query("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 {
			limit = val
		}
	}

	subQr := " `manifest_id` = ? AND ( `chasis_number` LIKE ? OR `bl_no` LIKE ? OR vehicle_id = ?)"
	q := "%" + query + "%"
	vals := []any{manifestId, q, q, query}

	qr := "SELECT COUNT(vehicle_id) AS numbers FROM `manifest_vehicles` WHERE " + subQr

	stx, rez := gendb.PagenationSelect(qr, ipg, limit, vals)
	if stx.State != constants.SuccessState {
		c.JSON(http.StatusInternalServerError, constants.NormalResponse{
			State:   constants.ErrorState,
			Data:    "Failed to obtain pagination limit",
			Message: "An error occurred while obtaining the pagination limit",
			Adv:     "none",
		})
		return
	}
	totalPages := int(math.Ceil(float64(rez.ResNum) / float64(limit)))

	if rez.State == "end" {
		c.JSON(http.StatusOK, constants.PaginationResponse{
			State:   constants.SuccessState,
			Data:    []any{},
			Message: "Vehicle not found",
			Pagination: constants.PaginationInfo{
				TotalItems:   rez.ResNum,
				TotalPages:   totalPages,
				CurrentPage:  ipg,
				ItemsPerPage: limit,
				HasNextPage:  false,
				HasPrevPage:  ipg > 1,
			},
		})
		return
	}

	st, vehicles := manifestdataservices.SelectVehicleInfo(subQr+" "+rez.Limit, vals)

	if st.State != constants.SuccessState {
		slog.Error(st.Data)
		c.JSON(http.StatusInternalServerError, constants.NormalResponse{
			State:   constants.ErrorState,
			Data:    "Failed to search for vehicle",
			Message: "An error occurred while searching for the vehicle: " + st.Data,
			Adv:     "none",
		})
		return
	}

	if len(vehicles) == 0 {
		c.JSON(http.StatusOK, constants.PaginationResponse{
			State:   constants.SuccessState,
			Data:    []any{},
			Message: "Vehicle not found",
			Pagination: constants.PaginationInfo{
				TotalItems:   rez.ResNum,
				TotalPages:   totalPages,
				CurrentPage:  ipg,
				ItemsPerPage: limit,
				HasNextPage:  false,
				HasPrevPage:  ipg > 1,
			},
		})
		return
	}

	var shortInfos []VehicleShortInfo

	for _, vehicle := range vehicles {
		inspectedBy := "Not Inspected"
		mainImage := ""

		// Get tally details for the main image
		stt, tallyDets := manifestdataservices.SelectTallyDetailsLong(" `vehicle_id` = ? ", []any{vehicle.VehicleId})
		if stt.State != constants.SuccessState {
			slog.Error(stt.Data)
		} else if len(tallyDets) > 0 {
			mainImage = tallyDets[0].VehicleImage
		}

		if vehicle.InspectionStatus == "yes" && len(tallyDets) > 0 {
			stu, users := usersdataservices.SelectUserDetailsWithRolesPass(" `user_id` = ? ", []any{tallyDets[0].UserId})
			if stu.State == constants.SuccessState && len(users) > 0 {
				inspectedBy = users[0].FName + " " + users[0].LName
			}
		}

		shortInfos = append(shortInfos, VehicleShortInfo{
			VehicleId:     vehicle.VehicleId,
			ChasisNumber:  vehicle.ChasisNumber,
			BlNumber:      vehicle.BLNumber,
			Description:   vehicle.Description,
			Model:         vehicle.VehicleModel,
			IsInspected:   vehicle.InspectionStatus == "yes",
			InspectedBy:   inspectedBy,
			InspectedTime: vehicle.InspectionTime,
			Image:         mainImage,
		})
	}

	c.JSON(http.StatusOK, constants.PaginationResponse{
		State:   constants.SuccessState,
		Data:    shortInfos,
		Message: "Vehicle information fetched successfully",
		Pagination: constants.PaginationInfo{
			TotalItems:   rez.ResNum,
			TotalPages:   totalPages,
			CurrentPage:  ipg,
			ItemsPerPage: limit,
			HasNextPage:  ipg < totalPages,
			HasPrevPage:  ipg > 1,
		},
	})
}

type VehicleShortInfo struct {
	VehicleId     string `json:"vehicleId"`
	ChasisNumber  string `json:"chasisNumber"`
	BlNumber      string `json:"blNumber"`
	Description   string `json:"description"`
	Model         string `json:"model"`
	IsInspected   bool   `json:"isInspected"`
	InspectedBy   string `json:"inspectedBy"`
	InspectedTime string `json:"inspectedTime"`
	Image         string `json:"image"`
}

type VehicleInspectionDetailsRes struct {
	VehicleDets   manifest.VehiclesDetailsAndTally     `json:"vehicleDetails" binding:"required"`
	Inspections   []manifest.InspectionDetailsAndImage `json:"inspections" binding:"required"`
	Remarks       []manifest.InspectionRemarksDetails  `json:"remarks" binding:"required"`
	DischargeInfo []manifest.VehicleDischargeDetails   `json:"dischargeInfo" binding:"required"`
	UserDetails   users.UserData                       `json:"userDetails" binding:"required"`
	Media         []manifest.VehicleMediaResp          `json:"media" binding:"required"`
	Packages      []manifest.OnboardPackageResp        `json:"packageOnboard" binding:"required"`
}

func GetVehicleTypes(c *gin.Context) {
	page := c.DefaultQuery("page", "1")
	ipg, err := strconv.Atoi(page)
	if err != nil {
		c.JSON(http.StatusInternalServerError, constants.NormalResponse{
			State:   constants.ErrorState,
			Data:    "Failed to obtain page number",
			Message: "An error occurred while obtaining the page number",
			Adv:     "none",
		})
		return
	}

	limit := 10
	if l := c.Query("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 {
			limit = val
		}
	}

	query := c.Query("query")

	subQr := " `body_name` LIKE ? "
	q := "%" + query + "%"
	vals := []any{q}
	if query == "" {
		subQr = " `body_id` != ? "
		vals = []any{""}
	}

	qr := "SELECT COUNT(body_id) AS numbers FROM `vehicle_bodies` WHERE " + subQr

	stx, rez := gendb.PagenationSelect(qr, ipg, limit, vals)

	if stx.State != constants.SuccessState {
		c.JSON(http.StatusInternalServerError, constants.NormalResponse{
			State:   constants.ErrorState,
			Data:    "Failed to obtain pagination limit",
			Message: "An error occurred while obtaining the pagination limit",
			Adv:     "none",
		})
		return
	} else if rez.State == "end" {
		c.JSON(http.StatusOK, constants.PaginationResponse{
			State:   constants.SuccessState,
			Data:    []VehicleShortInfo{},
			Message: "No more vehicle types available",
			Pagination: constants.PaginationInfo{
				TotalItems:   rez.ResNum,
				TotalPages:   0,
				CurrentPage:  ipg,
				ItemsPerPage: limit,
				HasNextPage:  false,
				HasPrevPage:  ipg > 1,
			},
		})
		return
	}
	totalPages := int(math.Ceil(float64(rez.ResNum) / float64(limit)))
	st, types := manifestdataservices.SelectVehicleTypes(subQr+" ORDER BY `creation_time` DESC "+rez.Limit, vals)

	if st.State != constants.SuccessState {
		slog.Error(st.Data)
		c.JSON(http.StatusInternalServerError, constants.NormalResponse{
			State:   constants.ErrorState,
			Data:    "Failed to fetch vehicle types",
			Message: "An error occurred while fetching vehicle types: " + st.Data,
			Adv:     "none",
		})
		return
	}

	c.JSON(http.StatusOK, constants.PaginationResponse{
		State:   constants.SuccessState,
		Data:    types,
		Message: "Vehicle types fetched successfully",
		Pagination: constants.PaginationInfo{
			TotalItems:   rez.ResNum,
			TotalPages:   totalPages,
			CurrentPage:  ipg,
			ItemsPerPage: limit,
			HasNextPage:  ipg < totalPages,
			HasPrevPage:  ipg > 1,
		},
	})
}

func GetVehicleMakers(c *gin.Context) {
	page := c.DefaultQuery("page", "1")
	ipg, err := strconv.Atoi(page)
	if err != nil {
		c.JSON(http.StatusInternalServerError, constants.NormalResponse{
			State:   constants.ErrorState,
			Data:    "Failed to obtain page number",
			Message: "An error occurred while obtaining the page number",
			Adv:     "none",
		})
		return
	}

	limit := 10
	if l := c.Query("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 {
			limit = val
		}
	}

	query := c.Query("query")

	subQr := " `maker_name` LIKE ? "
	q := "%" + query + "%"
	vals := []any{q}
	if query == "" {
		subQr = " `maker_id` != ? "
		vals = []any{""}
	}

	qr := "SELECT COUNT(maker_id) AS numbers FROM `vehicle_makers` WHERE " + subQr

	stx, rez := gendb.PagenationSelect(qr, ipg, limit, vals)
	if stx.State != constants.SuccessState {
		c.JSON(http.StatusInternalServerError, constants.NormalResponse{
			State:   constants.ErrorState,
			Data:    "Failed to obtain pagination limit",
			Message: "An error occurred while obtaining the pagination limit",
			Adv:     "none",
		})
		return
	} else if rez.State == "end" {
		c.JSON(http.StatusOK, constants.PaginationResponse{
			State:   constants.SuccessState,
			Data:    []VehicleShortInfo{},
			Message: "No more vehicle makers available",
			Pagination: constants.PaginationInfo{
				TotalItems:   rez.ResNum,
				TotalPages:   0,
				CurrentPage:  ipg,
				ItemsPerPage: limit,
				HasNextPage:  false,
				HasPrevPage:  ipg > 1,
			},
		})
		return
	}
	totalPages := int(math.Ceil(float64(rez.ResNum) / float64(limit)))

	st, makers := manifestdataservices.SelectVehicleMakers(subQr+" ORDER BY `maker_id` DESC "+rez.Limit, vals)

	if st.State != constants.SuccessState {
		slog.Error(st.Data)
		c.JSON(http.StatusInternalServerError, constants.NormalResponse{
			State:   constants.ErrorState,
			Data:    "Failed to fetch vehicle makers",
			Message: "An error occurred while fetching vehicle makers: " + st.Data,
			Adv:     "none",
		})
		return
	}

	c.JSON(http.StatusOK, constants.PaginationResponse{
		State:   constants.SuccessState,
		Data:    makers,
		Message: "Vehicle makers fetched successfully",
		Pagination: constants.PaginationInfo{
			TotalItems:   rez.ResNum,
			TotalPages:   totalPages,
			CurrentPage:  ipg,
			ItemsPerPage: limit,
			HasNextPage:  ipg < totalPages,
			HasPrevPage:  ipg > 1,
		},
	})
}

func GetInspectionCheckList(c *gin.Context) {
	page := c.DefaultQuery("page", "1")
	ipg, err := strconv.Atoi(page)
	if err != nil {
		c.JSON(http.StatusInternalServerError, constants.NormalResponse{
			State:   constants.ErrorState,
			Data:    "Failed to obtain page number",
			Message: "An error occurred while obtaining the page number",
			Adv:     "none",
		})
		return
	}

	limit := 10
	if l := c.Query("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 {
			limit = val
		}
	}

	query := c.Query("query")

	subQr := " `check_name` LIKE ? "
	q := "%" + query + "%"
	vals := []any{q}
	if query == "" {
		subQr = " `check_id` != ? "
		vals = []any{""}
	}

	qr := "SELECT COUNT(check_id) AS numbers FROM `inspection_checklist` WHERE " + subQr

	stx, rez := gendb.PagenationSelect(qr, ipg, limit, vals)
	if stx.State != constants.SuccessState {
		c.JSON(http.StatusInternalServerError, constants.NormalResponse{
			State:   constants.ErrorState,
			Data:    "Failed to obtain pagination limit",
			Message: "An error occurred while obtaining the pagination limit",
			Adv:     "none",
		})
		return
	} else if rez.State == "end" {
		c.JSON(http.StatusOK, constants.PaginationResponse{
			State:   constants.SuccessState,
			Data:    []VehicleShortInfo{},
			Message: "No more inspection checklist available",
			Pagination: constants.PaginationInfo{
				TotalItems:   rez.ResNum,
				TotalPages:   0,
				CurrentPage:  ipg,
				ItemsPerPage: limit,
				HasNextPage:  false,
				HasPrevPage:  ipg > 1,
			},
		})
		return
	}
	totalPages := int(math.Ceil(float64(rez.ResNum) / float64(limit)))

	st, checks := manifestdataservices.SelectInspectionCheckList(subQr+" ORDER BY `check_id` DESC "+rez.Limit, vals)

	if st.State != constants.SuccessState {
		slog.Error(st.Data)
		c.JSON(http.StatusInternalServerError, constants.NormalResponse{
			State:   constants.ErrorState,
			Data:    "Failed to fetch inspection checklist",
			Message: "An error occurred while fetching inspection checklist: " + st.Data,
			Adv:     "none",
		})
		return
	}

	c.JSON(http.StatusOK, constants.PaginationResponse{
		State:   constants.SuccessState,
		Data:    checks,
		Message: "Inspection checklist fetched successfully",
		Pagination: constants.PaginationInfo{
			TotalItems:   rez.ResNum,
			TotalPages:   totalPages,
			CurrentPage:  ipg,
			ItemsPerPage: limit,
			HasNextPage:  ipg < totalPages,
			HasPrevPage:  ipg > 1,
		},
	})
}

func GetVehicleModels(c *gin.Context) {
	page := c.DefaultQuery("page", "1")
	ipg, err := strconv.Atoi(page)
	if err != nil {
		c.JSON(http.StatusInternalServerError, constants.NormalResponse{
			State:   constants.ErrorState,
			Data:    "Failed to obtain page number",
			Message: "An error occurred while obtaining the page number",
			Adv:     "none",
		})
		return
	}

	limit := 10
	if l := c.Query("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 {
			limit = val
		}
	}

	query := c.Query("query")

	subQr := " `model_name` LIKE ? "
	q := "%" + query + "%"
	vals := []any{q}
	if query == "" {
		subQr = " `model_id` != ? "
		vals = []any{""}
	}

	qr := "SELECT COUNT(model_id) AS numbers FROM `vehicle_models` WHERE " + subQr

	stx, rez := gendb.PagenationSelect(qr, ipg, limit, vals)
	if stx.State != constants.SuccessState {
		c.JSON(http.StatusInternalServerError, constants.NormalResponse{
			State:   constants.ErrorState,
			Data:    "Failed to obtain pagination limit",
			Message: "An error occurred while obtaining the pagination limit",
			Adv:     "none",
		})
		return
	} else if rez.State == "end" {
		c.JSON(http.StatusOK, constants.PaginationResponse{
			State:   constants.SuccessState,
			Data:    []VehicleShortInfo{},
			Message: "No more vehicle models available",
			Pagination: constants.PaginationInfo{
				TotalItems:   rez.ResNum,
				TotalPages:   0,
				CurrentPage:  ipg,
				ItemsPerPage: limit,
				HasNextPage:  false,
				HasPrevPage:  ipg > 1,
			},
		})
		return
	}
	totalPages := int(math.Ceil(float64(rez.ResNum) / float64(limit)))

	st, models := manifestdataservices.SelectVehicleModels(subQr+" ORDER BY `model_id` DESC "+rez.Limit, vals)

	if st.State != constants.SuccessState {
		slog.Error(st.Data)
		c.JSON(http.StatusInternalServerError, constants.NormalResponse{
			State:   constants.ErrorState,
			Data:    "Failed to fetch vehicle models",
			Message: "An error occurred while fetching vehicle models: " + st.Data,
			Adv:     "none",
		})
		return
	}

	c.JSON(http.StatusOK, constants.PaginationResponse{
		State:   constants.SuccessState,
		Data:    models,
		Message: "Vehicle models fetched successfully",
		Pagination: constants.PaginationInfo{
			TotalItems:   rez.ResNum,
			TotalPages:   totalPages,
			CurrentPage:  ipg,
			ItemsPerPage: limit,
			HasNextPage:  ipg < totalPages,
			HasPrevPage:  ipg > 1,
		},
	})
}

func GetDeckStowagePlan(c *gin.Context) {
	manifestId := c.Param("manifestId")
	page := c.DefaultQuery("page", "1")
	ipg, err := strconv.Atoi(page)
	if err != nil {
		c.JSON(http.StatusInternalServerError, constants.NormalResponse{
			State:   constants.ErrorState,
			Data:    "Failed to obtain page number",
			Message: "An error occurred while obtaining the page number",
			Adv:     "none",
		})
		return
	}

	limit := 40
	if l := c.Query("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 {
			limit = val
		}
	}

	query := c.Query("query")

	subQr := " `manifest_id` = ? AND `deck_name` LIKE ? "
	q := "%" + query + "%"
	vals := []any{manifestId, q}
	if query == "" {
		subQr = " `manifest_id` = ? "
		vals = []any{manifestId}

		if manifestId == "" {
			subQr = " `deck_name` LIKE ? "
			vals = []any{q}
			if query == "" {
				subQr = " `deck_id` != ? "
				vals = []any{""}
			}
		}
	}

	qr := "SELECT COUNT(deck_id) AS numbers FROM `decks_numbers` WHERE " + subQr

	stx, rez := gendb.PagenationSelect(qr, ipg, limit, vals)
	if stx.State != constants.SuccessState {
		c.JSON(http.StatusInternalServerError, constants.NormalResponse{
			State:   constants.ErrorState,
			Data:    "Failed to obtain pagination limit",
			Message: "An error occurred while obtaining the pagination limit",
			Adv:     "none",
		})
		return
	} else if rez.State == "end" {
		c.JSON(http.StatusOK, constants.PaginationResponse{
			State:   constants.SuccessState,
			Data:    []VehicleShortInfo{},
			Message: "No more deck stowage plans available",
			Pagination: constants.PaginationInfo{
				TotalItems:   rez.ResNum,
				TotalPages:   0,
				CurrentPage:  ipg,
				ItemsPerPage: limit,
				HasNextPage:  false,
				HasPrevPage:  ipg > 1,
			},
		})
		return
	}
	totalPages := int(math.Ceil(float64(rez.ResNum) / float64(limit)))

	st, plans := manifestdataservices.SelectDeckStowagePlan(subQr+" ORDER BY `deck_name` ASC "+rez.Limit, vals)

	if st.State != constants.SuccessState {
		slog.Error(st.Data)
		c.JSON(http.StatusInternalServerError, constants.NormalResponse{
			State:   constants.ErrorState,
			Data:    "Failed to fetch deck stowage plans",
			Message: "An error occurred while fetching deck stowage plans: " + st.Data,
			Adv:     "none",
		})
		return
	}

	c.JSON(http.StatusOK, constants.PaginationResponse{
		State:   constants.SuccessState,
		Data:    plans,
		Message: "Deck stowage plans fetched successfully",
		Pagination: constants.PaginationInfo{
			TotalItems:   rez.ResNum,
			TotalPages:   totalPages,
			CurrentPage:  ipg,
			ItemsPerPage: limit,
			HasNextPage:  ipg < totalPages,
			HasPrevPage:  ipg > 1,
		},
	})
}

// PACKAGE STARTS HERE
func GetPackageTypes(c *gin.Context) {
	page := c.DefaultQuery("page", "1")
	ipg, err := strconv.Atoi(page)
	if err != nil {
		c.JSON(http.StatusInternalServerError, constants.NormalResponse{
			State:   constants.ErrorState,
			Data:    "Failed to obtain page number",
			Message: "An error occurred while obtaining the page number",
			Adv:     "none",
		})
		return
	}

	limit := 10
	if l := c.Query("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 {
			limit = val
		}
	}

	query := c.Query("query")
	subQr := " `type_name` LIKE ? "
	q := "%" + query + "%"
	vals := []any{q}
	if query == "" {
		subQr = " `type_id` != ? "
		vals = []any{""}
	}

	qr := "SELECT COUNT(type_id) AS numbers FROM `package_types` WHERE " + subQr
	stx, rez := gendb.PagenationSelect(qr, ipg, limit, vals)
	if stx.State != constants.SuccessState {
		c.JSON(http.StatusInternalServerError, constants.NormalResponse{
			State:   constants.ErrorState,
			Data:    "Failed to obtain pagination limit",
			Message: "An error occurred while obtaining the pagination limit",
			Adv:     "none",
		})
		return
	} else if rez.State == "end" {
		c.JSON(http.StatusOK, constants.PaginationResponse{
			State:   constants.SuccessState,
			Data:    []VehicleShortInfo{},
			Message: "No more package types available",
			Pagination: constants.PaginationInfo{
				TotalItems:   rez.ResNum,
				TotalPages:   0,
				CurrentPage:  ipg,
				ItemsPerPage: limit,
				HasNextPage:  false,
				HasPrevPage:  ipg > 1,
			},
		})
		return
	}
	totalPages := int(math.Ceil(float64(rez.ResNum) / float64(limit)))
	st, types := manifestdataservices.SelectPackageTypes(subQr+" ORDER BY `type_id` DESC "+rez.Limit, vals)

	if st.State != constants.SuccessState {
		slog.Error(st.Data)
		c.JSON(http.StatusInternalServerError, constants.NormalResponse{
			State:   constants.ErrorState,
			Data:    "Failed to fetch package types",
			Message: "An error occurred while fetching package types: " + st.Data,
			Adv:     "none",
		})
		return
	}

	c.JSON(http.StatusOK, constants.PaginationResponse{
		State:   constants.SuccessState,
		Data:    types,
		Message: "Package types fetched successfully",
		Pagination: constants.PaginationInfo{
			TotalItems:   rez.ResNum,
			TotalPages:   totalPages,
			CurrentPage:  ipg,
			ItemsPerPage: limit,
			HasNextPage:  ipg < totalPages,
			HasPrevPage:  ipg > 1,
		},
	})
}

func GetPackageStatusList(c *gin.Context) {
	page := c.DefaultQuery("page", "1")
	ipg, err := strconv.Atoi(page)
	if err != nil {
		c.JSON(http.StatusInternalServerError, constants.NormalResponse{
			State:   constants.ErrorState,
			Data:    "Failed to obtain page number",
			Message: "An error occurred while obtaining the page number",
			Adv:     "none",
		})
		return
	}

	limit := 10
	if l := c.Query("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 {
			limit = val
		}
	}

	query := c.Query("query")
	subQr := " `status_name` LIKE ? "
	q := "%" + query + "%"
	vals := []any{q}
	if query == "" {
		subQr = " `status_id` != ? "
		vals = []any{""}
	}

	qr := "SELECT COUNT(status_id) AS numbers FROM `packages_inspection_status` WHERE " + subQr
	stx, rez := gendb.PagenationSelect(qr, ipg, limit, vals)
	if stx.State != constants.SuccessState {
		c.JSON(http.StatusInternalServerError, constants.NormalResponse{
			State:   constants.ErrorState,
			Data:    "Failed to obtain pagination limit",
			Message: "An error occurred while obtaining the pagination limit",
			Adv:     "none",
		})
		return
	} else if rez.State == "end" {
		c.JSON(http.StatusOK, constants.PaginationResponse{
			State:   constants.SuccessState,
			Data:    []VehicleShortInfo{},
			Message: "No more package inspection statuses available",
			Pagination: constants.PaginationInfo{
				TotalItems:   rez.ResNum,
				TotalPages:   0,
				CurrentPage:  ipg,
				ItemsPerPage: limit,
				HasNextPage:  false,
				HasPrevPage:  ipg > 1,
			},
		})
		return
	}
	totalPages := int(math.Ceil(float64(rez.ResNum) / float64(limit)))
	st, types := manifestdataservices.SelectPackageStatuses(subQr+" ORDER BY `status_id` DESC "+rez.Limit, vals)

	if st.State != constants.SuccessState {
		slog.Error(st.Data)
		c.JSON(http.StatusInternalServerError, constants.NormalResponse{
			State:   constants.ErrorState,
			Data:    "Failed to fetch package inspection status list",
			Message: "An error occurred while fetching package inspection status list: " + st.Data,
			Adv:     "none",
		})
		return
	}

	c.JSON(http.StatusOK, constants.PaginationResponse{
		State:   constants.SuccessState,
		Data:    types,
		Message: "Package inspection status list fetched successfully",
		Pagination: constants.PaginationInfo{
			TotalItems:   rez.ResNum,
			TotalPages:   totalPages,
			CurrentPage:  ipg,
			ItemsPerPage: limit,
			HasNextPage:  ipg < totalPages,
			HasPrevPage:  ipg > 1,
		},
	})
}

func GetManifestPackagesList(c *gin.Context) {
	manifestId := c.Param("manifestId")

	if manifestId == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"state": constants.ErrorState,
			"data":  "Manifest ID is required",
		})
		return
	}

	// Pagination parameters
	pg := c.DefaultQuery("page", "1")
	ipg, err := strconv.Atoi(pg)
	if err != nil {
		ipg = 1
	}

	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 50
	}
	// Cap the limit to avoid excessive load
	if limit > 100 {
		limit = 100
	}

	status := c.Query("status")

	// Handle Inspected Packages (Detailed View)
	if status == "inspected" {
		subQuery := " manifest_packages.manifest_id = ? AND manifest_packages.is_inspected = ?"
		vals := []any{manifestId, manifest.InspectionStatus.Yes}

		// Count query for pagination
		// Note: The main query INNER JOINs manifest_packages, package_types, users, packages_inspection_status
		// We need a count query that matches the WHERE clause of SelectPackageInspectionData
		// SelectPackageInspectionData FROM packages_inspection ... INNER JOIN manifest_packages ...
		qrN := fmt.Sprintf("SELECT COUNT(packages_inspection.inspection_id) FROM packages_inspection INNER JOIN manifest_packages ON manifest_packages.package_id = packages_inspection.package_id WHERE %s", subQuery)

		st, rez := gendb.PagenationSelect(qrN, ipg, limit, vals)
		if st.State != constants.SuccessState {
			slog.Error(st.Data)
			c.JSON(http.StatusInternalServerError, gin.H{
				"state": constants.ErrorState,
				"data":  "Failed to retrieve pagination data",
			})
			return
		}

		if rez.State == "end" {
			c.JSON(http.StatusOK, gin.H{
				"state": constants.SuccessState,
				"data":  []any{},
				"adv":   rez.ResNum,
				"per":   limit,
			})
			return
		}

		// Fetch data with limit
		stFetch, inspections := manifestdataservices.SelectPackageInspectionData(subQuery+rez.Limit, vals)
		if stFetch.State != constants.SuccessState {
			slog.Error(stFetch.Data)
			c.JSON(http.StatusInternalServerError, gin.H{
				"state": constants.ErrorState,
				"data":  "Failed to retrieve package inspection",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"state": constants.SuccessState,
			"data":  inspections,
			"adv":   rez.ResNum,
			"per":   limit,
		})
		return
	}

	// Handle All or Uninspected Packages (Manifest Info View)
	subQuery := " `manifest_id` = ? "
	vals := []any{manifestId}

	if status == "uninspected" {
		subQuery += " AND `is_inspected` = ? "
		vals = append(vals, manifest.InspectionStatus.No)
	}

	// Count query for pagination
	qrN := fmt.Sprintf("SELECT COUNT(package_id) FROM manifest_packages WHERE %s", subQuery)

	st, rez := gendb.PagenationSelect(qrN, ipg, limit, vals)
	if st.State != constants.SuccessState {
		slog.Error(st.Data)
		c.JSON(http.StatusInternalServerError, gin.H{
			"state": constants.ErrorState,
			"data":  "Failed to retrieve pagination data",
		})
		return
	}

	if rez.State == "end" {
		c.JSON(http.StatusOK, gin.H{
			"state": constants.SuccessState,
			"data":  []any{},
			"adv":   rez.ResNum,
			"per":   limit,
		})
		return
	}

	// Fetch data with limit
	stFetch, packages := manifestdataservices.SelectPackageInfo(subQuery+rez.Limit, vals)
	if stFetch.State != constants.SuccessState {
		slog.Error(stFetch.Data)
		c.JSON(http.StatusInternalServerError, gin.H{
			"state": constants.ErrorState,
			"data":  "Failed to retrieve manifest packages",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"state": constants.SuccessState,
		"data":  packages,
		"adv":   rez.ResNum,
		"per":   limit,
	})
}

func GetInspectedPackageDetails(c *gin.Context) {
	packageId := c.Param("packageId")

	if packageId == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"state": constants.ErrorState,
			"data":  "Package ID is required",
		})
		return
	}

	subQuery := " packages_inspection.package_id = ? "
	vals := []any{packageId}
	st, packageDetails := manifestdataservices.SelectPackageInspectionData(subQuery, vals)
	if st.State != constants.SuccessState {
		slog.Error(st.Data)
		c.JSON(http.StatusInternalServerError, gin.H{
			"state": constants.ErrorState,
			"data":  "Failed to retrieve package details",
		})
		return
	} else if len(packageDetails) < 1 {
		c.JSON(http.StatusNotFound, gin.H{
			"state": constants.ErrorState,
			"data":  "Package not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"state": constants.SuccessState,
		"data":  packageDetails[0],
	})
}

type VehicleInspectionTallyTimelineRes struct {
	VehicleId      string                        `json:"vehicleId" binding:"required"`
	ManifestId     string                        `json:"manifestId" binding:"required"`
	ActiveData     VehicleInspectionDetailsRes   `json:"activeData"`
	HistoricalData []VehicleInspectionDetailsRes `json:"historicalData"`
}

func GetVehicleWithMultipleInspections(c *gin.Context) {
	manifestId := c.Param("manifestId")
	if manifestId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"state": constants.ErrorState, "data": "Manifest ID is required"})
		return
	}

	page := c.DefaultQuery("page", "1")
	ipg, err := strconv.Atoi(page)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"state": constants.ErrorState, "data": "Failed to obtain page number"})
		return
	}

	limit := 250
	if l := c.Query("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 {
			limit = val
		}
	}

	subQr := " `manifest_id` = ? AND EXISTS (SELECT 1 FROM `vehicles_talling_history` h WHERE h.`vehicle_id` = `manifest_vehicles`.`vehicle_id`) "
	vals := []any{manifestId}

	if searchQuery := c.Query("query"); searchQuery != "" {
		subQr += " AND ( `chasis_number` LIKE ? OR `model` LIKE ? OR `description` LIKE ? OR `weight` LIKE ? OR `bl_no` LIKE ? )"
		s := "%" + searchQuery + "%"
		vals = append(vals, s, s, s, s, s)
	}

	qr := "SELECT COUNT(vehicle_id) AS numbers FROM `manifest_vehicles` WHERE " + subQr
	stx, rez := gendb.PagenationSelect(qr, ipg, limit, vals)
	if stx.State != constants.SuccessState {
		c.JSON(http.StatusInternalServerError, gin.H{"state": constants.ErrorState, "data": "Failed to obtain pagination limit"})
		return
	}
	if rez.State == "end" {
		c.JSON(http.StatusOK, gin.H{"state": constants.SuccessState, "data": []any{}, "adv": rez.ResNum, "per": limit})
		return
	}

	st, vehicles := manifestdataservices.SelectVehicleInfo(subQr+" ORDER BY `vehicle_id` DESC "+rez.Limit, vals)
	if st.State != constants.SuccessState {
		slog.Error(st.Data)
		c.JSON(http.StatusInternalServerError, gin.H{"state": constants.ErrorState, "data": "Failed to fetch vehicles for manifest"})
		return
	}

	if len(vehicles) == 0 {
		c.JSON(http.StatusOK, gin.H{"state": constants.SuccessState, "data": []any{}, "adv": rez.ResNum, "per": limit})
		return
	}

	vhShow := enrichVehicleDetails(vehicles)
	c.JSON(http.StatusOK, gin.H{"state": constants.SuccessState, "data": vhShow, "adv": rez.ResNum, "per": limit})
}

func GetVehicleInspectionTallyTimeline(c *gin.Context) {
	vehicleId := c.Param("vehicleId")
	if vehicleId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"state": constants.ErrorState, "data": "Vehicle ID is required"})
		return
	}

	st, vehicles := manifestdataservices.SelectVehicleInfo(" `vehicle_id` = ? ", []any{vehicleId})
	if st.State != constants.SuccessState {
		slog.Error(st.Data)
		c.JSON(http.StatusInternalServerError, gin.H{"state": constants.ErrorState, "data": "Failed to fetch vehicle details"})
		return
	}
	if len(vehicles) == 0 {
		c.JSON(http.StatusOK, gin.H{"state": constants.ErrorState, "data": []any{}})
		return
	}

	stx, timeline := buildVehicleInspectionTallyTimeline(vehicles[0].VehicleId, vehicles[0].ManifestId)
	if stx.State != constants.SuccessState {
		slog.Error(stx.Data)
		c.JSON(http.StatusInternalServerError, gin.H{"state": constants.ErrorState, "data": stx.Data})
		return
	}

	c.JSON(http.StatusOK, gin.H{"state": constants.SuccessState, "data": timeline})
}

func buildVehicleInspectionTallyTimeline(vehicleId, manifestId string) (*constants.AnswerState, VehicleInspectionTallyTimelineRes) {
	st, vehicleRows := manifestdataservices.SelectVehicleAndTallyDetails(" manifest_vehicles.vehicle_id = ? ", []any{vehicleId})
	if st.State != constants.SuccessState {
		return st, VehicleInspectionTallyTimelineRes{}
	}
	if len(vehicleRows) == 0 {
		return &constants.AnswerState{State: constants.ErrorState, Data: "Vehicle not found", Adv: "none"}, VehicleInspectionTallyTimelineRes{}
	}

	std, discharge := manifestdataservices.SelectDischargeDetails(" `vehicle_id` = ? ", []any{vehicleId})
	if std.State != constants.SuccessState {
		return &constants.AnswerState{State: constants.ErrorState, Data: "Failed to fetch vehicle discharge details", Adv: "none"}, VehicleInspectionTallyTimelineRes{}
	}

	activeDetails, stActive := buildInspectionDetailsFromTimeline(vehicleId, nil, vehicleRows[0], discharge)
	if stActive.State != constants.SuccessState {
		return stActive, VehicleInspectionTallyTimelineRes{}
	}

	sth, historyTimes := manifestdataservices.SelectVehicleHistoryBatchTimes(vehicleId)
	if sth.State != constants.SuccessState {
		return sth, VehicleInspectionTallyTimelineRes{}
	}

	historicalDetails := make([]VehicleInspectionDetailsRes, 0, len(historyTimes))
	for i := range historyTimes {
		det, stBatch := buildInspectionDetailsFromTimeline(vehicleId, &historyTimes[i], vehicleRows[0], discharge)
		if stBatch.State != constants.SuccessState {
			return stBatch, VehicleInspectionTallyTimelineRes{}
		}
		historicalDetails = append(historicalDetails, det)
	}

	return &constants.AnswerState{State: constants.SuccessState, Data: "success", Adv: "none"}, VehicleInspectionTallyTimelineRes{
		VehicleId:      vehicleId,
		ManifestId:     manifestId,
		ActiveData:     activeDetails,
		HistoricalData: historicalDetails,
	}
}

func buildInspectionDetailsFromTimeline(vehicleId string, archivedAt *time.Time, vehicleDets manifest.VehiclesDetailsAndTally, discharge []manifest.VehicleDischargeDetails) (VehicleInspectionDetailsRes, *constants.AnswerState) {
	if archivedAt != nil {
		// vehicleDets holds the active tally. A history batch shows its own archived
		// tally (and its inspector), not the current one.
		stt, tallies := manifestdataservices.SelectVehicleTimelineTally(vehicleId, archivedAt)
		if stt.State != constants.SuccessState {
			return VehicleInspectionDetailsRes{}, stt
		}
		if len(tallies) > 0 {
			t := tallies[0]
			vehicleDets.TallyId = t.TallyId
			vehicleDets.UserId = t.UserId
			vehicleDets.MakerId = t.MakerId
			vehicleDets.BodyId = t.BodyId
			vehicleDets.MakerName = t.MakerName
			vehicleDets.BodyName = t.BodyName
			vehicleDets.VehicleImage = t.VehicleImage
			vehicleDets.DeckNumber = t.DeckNumber
			vehicleDets.NumberOfKeys = t.NumberOfKeys
			vehicleDets.KeyType = t.KeyType
			vehicleDets.TalliedTime = t.TalliedTime
		}
	}

	sti, inspections := manifestdataservices.SelectVehicleTimelineInspections(vehicleId, archivedAt)
	if sti.State != constants.SuccessState {
		return VehicleInspectionDetailsRes{}, sti
	}

	str, remarks := manifestdataservices.SelectVehicleTimelineRemarks(vehicleId, archivedAt)
	if str.State != constants.SuccessState {
		return VehicleInspectionDetailsRes{}, str
	}

	stp, packages := manifestdataservices.SelectVehicleTimelinePackages(vehicleId, archivedAt)
	if stp.State != constants.SuccessState {
		return VehicleInspectionDetailsRes{}, stp
	}

	stm, media := manifestdataservices.SelectVehicleTimelineMedia(vehicleId, archivedAt)
	if stm.State != constants.SuccessState {
		return VehicleInspectionDetailsRes{}, stm
	}

	userId := vehicleDets.UserId
	if len(inspections) > 0 && inspections[0].UserId != "" {
		userId = inspections[0].UserId
	}

	var userDets users.UserData
	stu, userList := usersdataservices.SelectUserDetailsWithRolesPass(" `user_id` = ? ", []any{userId})
	if stu.State == constants.SuccessState && len(userList) > 0 {
		userDets = userList[0]
		userDets.Password = ""
	}

	return VehicleInspectionDetailsRes{
		VehicleDets:   vehicleDets,
		Inspections:   inspections,
		Remarks:       remarks,
		DischargeInfo: discharge,
		UserDetails:   userDets,
		Media:         media,
		Packages:      packages,
	}, &constants.AnswerState{State: constants.SuccessState, Data: "success", Adv: "none"}
}
