package apiservices

import (
	"fmt"
	"net/http"

	"github.com/shabs76/roro-local-server/base_api"
	"github.com/shabs76/roro-local-server/constants"
	baseapi "github.com/shabs76/roro-local-server/constants/modules/base_api"
	"github.com/shabs76/roro-local-server/constants/modules/manifest"
	"github.com/shabs76/roro-local-server/constants/modules/users"
	manifestdataservices "github.com/shabs76/roro-local-server/database/manifest_data_services"
	usersdataservices "github.com/shabs76/roro-local-server/database/users_data_services"
	"github.com/shabs76/roro-local-server/specials"
)

func PerformHealthCheckToRemote(logId, logKey string) error {
	remoteBaseUrl := specials.GetEnvVariable("REMOTE_SERVER_URL", "http://localhost:8000")
	client := base_api.NewAPI(remoteBaseUrl)

	// 1. Check General API Health
	var healthResponse constants.AnswerState
	headers := []baseapi.Header{
		{Key: "Log-ID", Value: logId},
		{Key: "Log-Key", Value: logKey},
	}
	resp, err := client.Get("/health", &healthResponse, headers)
	if err != nil {
		return fmt.Errorf("failed to contact remote server: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("remote server returned status %d on /health", resp.StatusCode)
	}

	if healthResponse.State != constants.SuccessState {
		return fmt.Errorf("remote server is unhealthy: %s", healthResponse.Data)
	}

	// 2. Check Database Health
	var dbHealthResponse constants.AnswerState
	resp, err = client.Get("/health/db", &dbHealthResponse, []baseapi.Header{})
	if err != nil {
		return fmt.Errorf("failed to check remote database health: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("remote server returned status %d on /health/db", resp.StatusCode)
	}

	if dbHealthResponse.State != constants.SuccessState {
		return fmt.Errorf("remote database is unhealthy: %s", dbHealthResponse.Data)
	}

	// 3. Authentication Check
	var authCheckResponse constants.AnswerState
	resp, err = client.Get("/health/auth", &authCheckResponse, headers)
	if err != nil {
		return fmt.Errorf("failed to authenticate with remote server: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("remote server returned status %d on /health/auth", resp.StatusCode)
	}

	if authCheckResponse.State != constants.SuccessState {
		return fmt.Errorf("authentication with remote server failed: %s", authCheckResponse.Data)
	}

	// All checks passed

	return nil
}

func FetchUserRoles(logId, logKey string) error {
	remoteBaseUrl := specials.GetEnvVariable("REMOTE_SERVER_URL", "http://localhost:8000")
	client := base_api.NewAPI(remoteBaseUrl)

	// 1. Check Roles
	type UserRolesResponse struct {
		State string           `json:"state" binding:"required"`
		Data  []users.UserRole `json:"data" binding:"required"`
	}
	var userRolesResponse UserRolesResponse
	headers := []baseapi.Header{
		{Key: "Log-ID", Value: logId},
		{Key: "Log-Key", Value: logKey},
	}

	resp, err := client.Get("/admin/get/user/roles", &userRolesResponse, headers)
	if err != nil {
		return fmt.Errorf("failed to fetch user roles from remote server: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("remote server returned status %d on /admin/get/user/roles", resp.StatusCode)
	}

	if userRolesResponse.State != constants.SuccessState {
		return fmt.Errorf("failed to fetch user roles: %s", resp.Message)
	}

	// check if data is empty
	if len(userRolesResponse.Data) == 0 {
		return fmt.Errorf("no user roles found on remote server")
	}

	// add data to local database
	st := usersdataservices.InsertUserRole(userRolesResponse.Data)
	if st.State != constants.SuccessState {
		return fmt.Errorf("failed to insert user roles into local database: %s", st.Data)
	}

	return nil
}

func FetchUsersList(logId, logKey string) error {
	remoteBaseUrl := specials.GetEnvVariable("REMOTE_SERVER_URL", "http://localhost:8000")
	client := base_api.NewAPI(remoteBaseUrl)

	// 1. Check Users
	type UsersResponse struct {
		State string           `json:"state" binding:"required"`
		Data  []users.UserData `json:"data" binding:"required"`
	}
	var usersResponse UsersResponse
	headers := []baseapi.Header{
		{Key: "Log-ID", Value: logId},
		{Key: "Log-Key", Value: logKey},
	}

	resp, err := client.Get("/admin/get/company/users/pass", &usersResponse, headers)
	if err != nil {
		return fmt.Errorf("failed to fetch users from remote server: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("remote server returned status %d on /admin/get/company/users/pass", resp.StatusCode)
	}

	if usersResponse.State != constants.SuccessState {
		return fmt.Errorf("failed to fetch users: %s", resp.Message)
	}

	// check if data is empty
	if len(usersResponse.Data) == 0 {
		return fmt.Errorf("no users found on remote server")
	}

	// add data to local database
	st := usersdataservices.InsertUserData(usersResponse.Data)
	if st.State != constants.SuccessState {
		return fmt.Errorf("failed to insert users into local database: %s", st.Data)
	}

	return nil
}

func FetchVehicleBodies(logId, logKey string) error {
	remoteBaseUrl := specials.GetEnvVariable("REMOTE_SERVER_URL", "http://localhost:8000")
	client := base_api.NewAPI(remoteBaseUrl)

	// 1. Check Vehicle Bodies
	type VehicleBodiesResponse struct {
		State string                      `json:"state" binding:"required"`
		Data  []manifest.VehicleBodyTypes `json:"data" binding:"required"`
	}

	page := 1
	limit := 50
	headers := []baseapi.Header{
		{Key: "Log-ID", Value: logId},
		{Key: "Log-Key", Value: logKey},
	}

	for {
		var vehicleBodiesResponse VehicleBodiesResponse
		url := fmt.Sprintf("/admin/get/vehicle/types?page=%d&limit=%d", page, limit)
		resp, err := client.Get(url, &vehicleBodiesResponse, headers)
		if err != nil {
			return fmt.Errorf("failed to fetch vehicle bodies from remote server: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("remote server returned status %d on %s", resp.StatusCode, url)
		}

		if vehicleBodiesResponse.State != constants.SuccessState {
			return fmt.Errorf("failed to fetch vehicle bodies: %v", vehicleBodiesResponse.Data)
		}

		// check if data is empty
		if len(vehicleBodiesResponse.Data) == 0 && page != 1 {
			break
		} else if len(vehicleBodiesResponse.Data) == 0 && page == 1 {
			return fmt.Errorf("no vehicle bodies found on remote server")
		}

		// add data to local database
		st := manifestdataservices.InsertVehicleBodyTypes(vehicleBodiesResponse.Data)
		if st.State != constants.SuccessState {
			return fmt.Errorf("failed to insert vehicle bodies into local database: %s", st.Data)
		}

		page++
	}

	return nil
}

func FetchVehicleMakers(logId, logKey string) error {
	remoteBaseUrl := specials.GetEnvVariable("REMOTE_SERVER_URL", "http://localhost:8000")
	client := base_api.NewAPI(remoteBaseUrl)

	// 1. Check Vehicle Makers
	type VehicleMakersResponse struct {
		State string                   `json:"state" binding:"required"`
		Data  []manifest.VehicleMakers `json:"data" binding:"required"`
	}

	page := 1
	limit := 50
	headers := []baseapi.Header{
		{Key: "Log-ID", Value: logId},
		{Key: "Log-Key", Value: logKey},
	}

	for {
		var vehicleMakersResponse VehicleMakersResponse
		url := fmt.Sprintf("/admin/get/vehicle/makers?page=%d&limit=%d", page, limit)
		resp, err := client.Get(url, &vehicleMakersResponse, headers)
		if err != nil {
			return fmt.Errorf("failed to fetch vehicle makers from remote server: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("remote server returned status %d on %s", resp.StatusCode, url)
		}

		if vehicleMakersResponse.State != constants.SuccessState {
			return fmt.Errorf("failed to fetch vehicle makers: %v", vehicleMakersResponse.Data)
		}

		// check if data is empty
		if len(vehicleMakersResponse.Data) == 0 && page != 1 {
			break
		} else if len(vehicleMakersResponse.Data) == 0 && page == 1 {
			return fmt.Errorf("no vehicle makers found on remote server")
		}

		// add data to local database
		st := manifestdataservices.InsertVehicleMakers(vehicleMakersResponse.Data)
		if st.State != constants.SuccessState {
			return fmt.Errorf("failed to insert vehicle makers into local database: %s", st.Data)
		}

		page++
	}

	return nil
}

func FetchVehicleModels(logId, logKey string) error {
	remoteBaseUrl := specials.GetEnvVariable("REMOTE_SERVER_URL", "http://localhost:8000")
	client := base_api.NewAPI(remoteBaseUrl)

	// 1. Check Vehicle Models
	type VehicleModelsResponse struct {
		State string                         `json:"state" binding:"required"`
		Data  []manifest.VehicleModelDetails `json:"data" binding:"required"`
	}
	page := 1
	limit := 50
	headers := []baseapi.Header{
		{Key: "Log-ID", Value: logId},
		{Key: "Log-Key", Value: logKey},
	}

	for {
		var vehicleModelsResponse VehicleModelsResponse
		url := fmt.Sprintf("/admin/get/vehicle/models?page=%d&limit=%d", page, limit)
		resp, err := client.Get(url, &vehicleModelsResponse, headers)
		if err != nil {
			return fmt.Errorf("failed to fetch vehicle models from remote server: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("remote server returned status %d on %s", resp.StatusCode, url)
		}

		if vehicleModelsResponse.State != constants.SuccessState {
			return fmt.Errorf("failed to fetch vehicle models: %v", vehicleModelsResponse.Data)
		}

		// check if data is empty
		if len(vehicleModelsResponse.Data) == 0 && page != 1 {
			break
		} else if len(vehicleModelsResponse.Data) == 0 && page == 1 {
			return fmt.Errorf("no vehicle models found on remote server")
		}

		// add data to local database
		st := manifestdataservices.InsertVehicleModels(vehicleModelsResponse.Data)
		if st.State != constants.SuccessState {
			return fmt.Errorf("failed to insert vehicle models into local database: %s", st.Data)
		}

		page++
	}

	return nil
}

func FetchInspectionCheckList(logId, logKey string) error {
	remoteBaseUrl := specials.GetEnvVariable("REMOTE_SERVER_URL", "http://localhost:8000")
	client := base_api.NewAPI(remoteBaseUrl)

	// 1. Check Inspection CheckList
	type InspectionCheckListResponse struct {
		State string                                `json:"state" binding:"required"`
		Data  []manifest.InspectionCheckListDetails `json:"data" binding:"required"`
	}

	page := 1
	limit := 50
	headers := []baseapi.Header{
		{Key: "Log-ID", Value: logId},
		{Key: "Log-Key", Value: logKey},
	}

	for {
		var inspectionCheckListResponse InspectionCheckListResponse
		url := fmt.Sprintf("/admin/get/inspection/check/list?page=%d&limit=%d", page, limit)
		resp, err := client.Get(url, &inspectionCheckListResponse, headers)
		if err != nil {
			return fmt.Errorf("failed to fetch inspection checklist from remote server: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("remote server returned status %d on %s", resp.StatusCode, url)
		}

		if inspectionCheckListResponse.State != constants.SuccessState {
			return fmt.Errorf("failed to fetch inspection checklist: %v", inspectionCheckListResponse.Data)
		}

		// check if data is empty
		if len(inspectionCheckListResponse.Data) == 0 && page != 1 {
			break
		} else if len(inspectionCheckListResponse.Data) == 0 && page == 1 {
			return fmt.Errorf("no inspection checklist found on remote server")
		}

		// add data to local database
		st := manifestdataservices.InsertVehicleInspectCheckList(inspectionCheckListResponse.Data)
		if st.State != constants.SuccessState {
			return fmt.Errorf("failed to insert inspection checklist into local database: %s", st.Data)
		}

		page++
	}

	return nil
}

func FetchPackageTypesAndStatus(logId, logKey string) error {
	remoteBaseUrl := specials.GetEnvVariable("REMOTE_SERVER_URL", "http://localhost:8000")
	client := base_api.NewAPI(remoteBaseUrl)

	// 1. Check Package Types
	type PackageTypesResponse struct {
		State string                        `json:"state" binding:"required"`
		Data  []manifest.PackageTypeDetails `json:"data" binding:"required"`
	}

	page := 1
	limit := 50
	headers := []baseapi.Header{
		{Key: "Log-ID", Value: logId},
		{Key: "Log-Key", Value: logKey},
	}

	for {
		var packageTypesResponse PackageTypesResponse
		url := fmt.Sprintf("/admin/get/packages/types?page=%d&limit=%d", page, limit)
		resp, err := client.Get(url, &packageTypesResponse, headers)
		if err != nil {
			return fmt.Errorf("failed to fetch package types from remote server: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("remote server returned status %d on %s", resp.StatusCode, url)
		}

		if packageTypesResponse.State != constants.SuccessState {
			return fmt.Errorf("failed to fetch package types: %v", packageTypesResponse.Data)
		}

		// check if data is empty
		if len(packageTypesResponse.Data) == 0 && page != 1 {
			break
		} else if len(packageTypesResponse.Data) == 0 && page == 1 {
			return fmt.Errorf("no package types found on remote server")
		}

		// add data to local database
		st := manifestdataservices.InsertPackageTypes(packageTypesResponse.Data)
		if st.State != constants.SuccessState {
			return fmt.Errorf("failed to insert package types into local database: %s", st.Data)
		}

		page++
	}

	// 2. Check Package Status
	page = 1
	type PackageStatusResponse struct {
		State string                                    `json:"state" binding:"required"`
		Data  []manifest.PackageInspectionStatusDetails `json:"data" binding:"required"`
	}

	for {
		var packageStatusResponse PackageStatusResponse
		url := fmt.Sprintf("/admin/get/packages/inspection/statuses?page=%d&limit=%d", page, limit)
		resp, err := client.Get(url, &packageStatusResponse, headers)
		if err != nil {
			return fmt.Errorf("failed to fetch package inspection status from remote server: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("remote server returned status %d on %s", resp.StatusCode, url)
		}

		if packageStatusResponse.State != constants.SuccessState {
			return fmt.Errorf("failed to fetch package inspection status: %v", packageStatusResponse.Data)
		}

		// check if data is empty
		if len(packageStatusResponse.Data) == 0 && page != 1 {
			break
		} else if len(packageStatusResponse.Data) == 0 && page == 1 {
			return fmt.Errorf("no package inspection status found on remote server")
		}

		// add data to local database
		st := manifestdataservices.InsertPackageInspectionStatus(packageStatusResponse.Data)
		if st.State != constants.SuccessState {
			return fmt.Errorf("failed to insert package inspection status into local database: %s", st.Data)
		}

		page++
	}

	return nil
}

func FetchClientList(logId, logKey string) error {
	remoteBaseUrl := specials.GetEnvVariable("REMOTE_SERVER_URL", "http://localhost:8000")
	client := base_api.NewAPI(remoteBaseUrl)

	// 1. Check Client List
	type ClientListResponse struct {
		State string                `json:"state" binding:"required"`
		Data  []manifest.ClientInfo `json:"data" binding:"required"`
	}
	var clientListResponse ClientListResponse
	headers := []baseapi.Header{
		{Key: "Log-ID", Value: logId},
		{Key: "Log-Key", Value: logKey},
	}

	resp, err := client.Get("/client/manage/get/clients", &clientListResponse, headers)
	if err != nil {
		return fmt.Errorf("failed to fetch client list from remote server: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("remote server returned status %d on /client/manage/get/clients", resp.StatusCode)
	}

	if clientListResponse.State != constants.SuccessState {
		return fmt.Errorf("failed to fetch client list: %s", clientListResponse.Data)
	}

	// check if data is empty
	if len(clientListResponse.Data) == 0 {
		return fmt.Errorf("no clients found on remote server")
	}

	// add data to local database
	st := manifestdataservices.InsertClientDetail(clientListResponse.Data)
	if st.State != constants.SuccessState {
		return fmt.Errorf("failed to insert client list into local database: %s", st.Data)
	}

	return nil
}
