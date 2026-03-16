package apiservices

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"

	"github.com/shabs76/roro-local-server/base_api"
	"github.com/shabs76/roro-local-server/constants"
	baseapi "github.com/shabs76/roro-local-server/constants/modules/base_api"
	"github.com/shabs76/roro-local-server/constants/modules/manifest"
	manifestdataservices "github.com/shabs76/roro-local-server/database/manifest_data_services"
	"github.com/shabs76/roro-local-server/specials"
)

func FetchManifestDetails(logId, logKey, manifestId string) error {
	remoteBaseUrl := specials.GetEnvVariable("REMOTE_SERVER_URL", "http://localhost:8000")
	client := base_api.NewAPI(remoteBaseUrl)

	type ManifestDetailsResponse struct {
		State string                `json:"state"`
		Data  manifest.ManifestData `json:"data"`
	}
	// 1. Check General API Health
	var manifestDetailsResponse ManifestDetailsResponse
	headers := []baseapi.Header{
		{Key: "Log-ID", Value: logId},
		{Key: "Log-Key", Value: logKey},
	}
	resp, err := client.Get("/admin/get/manifest/data/"+manifestId, &manifestDetailsResponse, headers)
	if err != nil {
		return fmt.Errorf("failed to fetch manifest details from remote server: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("remote server returned status %d on /admin/get/manifest/data/%s", resp.StatusCode, manifestId)
	}

	if manifestDetailsResponse.State != constants.SuccessState {
		return fmt.Errorf("failed to fetch manifest details: %s", resp.Message)
	}

	if manifestDetailsResponse.Data.ManifestId == "" {
		return fmt.Errorf("no manifest details found for manifest ID: %s", manifestId)
	}

	log.Println(manifestDetailsResponse.Data)

	// 2. Process and Store Manifest Details in Local Database
	st := manifestdataservices.InsertManifestDetails([]manifest.ManifestData{manifestDetailsResponse.Data})
	if st.State != constants.SuccessState {
		return fmt.Errorf("failed to insert manifest details: %s", st.Data)
	}
	return nil
}

func FetchVehicleList(logId, logKey, manifestId string) error {
	remoteBaseUrl := specials.GetEnvVariable("REMOTE_SERVER_URL", "http://localhost:8000")
	client := base_api.NewAPI(remoteBaseUrl)

	type VehicleOnManifestResponse struct {
		State string                           `json:"state"`
		Data  []manifest.VehiclesDetailsToShow `json:"data"`
	}
	headers := []baseapi.Header{
		{Key: "Log-ID", Value: logId},
		{Key: "Log-Key", Value: logKey},
	}
	page := 1
	limit := 250
	for {
		var vehicleListResponse VehicleOnManifestResponse
		paginatedPath := fmt.Sprintf("/admin/get/manifest/vehicles/%s?page=%d&limit=%d", manifestId, page, limit)
		resp, err := client.Get(paginatedPath, &vehicleListResponse, headers)
		if err != nil {
			return fmt.Errorf("failed to fetch vehicle list for manifest from remote server: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("remote server returned status %d on %s", resp.StatusCode, paginatedPath)
		}

		if vehicleListResponse.State != constants.SuccessState {
			return fmt.Errorf("failed to fetch vehicle list for manifest: %s", resp.Message)
		}

		if len(vehicleListResponse.Data) == 0 && page == 1 {
			return fmt.Errorf("no vehicle list found for manifest ID: %s", manifestId)
		} else if len(vehicleListResponse.Data) == 0 {
			// No more data to fetch
			break
		}
		// Process and Store Vehicle List in Local Database
		st := manifestdataservices.InsertVehiclesDetails(vehicleListResponse.Data, false)
		if st.State != constants.SuccessState {
			return fmt.Errorf("failed to insert vehicle list for manifest: %s", st.Data)
		}

		page++ // Move to next page
	}
	slog.Info(fmt.Sprintf("Successfully fetched and stored vehicle list for manifest ID: %s", manifestId))
	return nil
}

func FetchStowagePlan(logId, logKey, manifestId string) error {
	remoteBaseUrl := specials.GetEnvVariable("REMOTE_SERVER_URL", "http://localhost:8000")
	client := base_api.NewAPI(remoteBaseUrl)

	type StowagePlanResponse struct {
		State string                            `json:"state"`
		Data  []manifest.DeckStowagePlanDetails `json:"data"`
	}
	var stowagePlanResponse StowagePlanResponse
	headers := []baseapi.Header{
		{Key: "Log-ID", Value: logId},
		{Key: "Log-Key", Value: logKey},
	}
	resp, err := client.Get("/admin/get/deck/stowage/plan/"+manifestId, &stowagePlanResponse, headers)
	if err != nil {
		return fmt.Errorf("failed to fetch stowage plan for manifest from remote server: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("remote server returned status %d on /admin/get/deck/stowage/plan/%s", resp.StatusCode, manifestId)
	}

	if stowagePlanResponse.State != constants.SuccessState {
		return fmt.Errorf("failed to fetch stowage plan for manifest: %s", resp.Message)
	}

	if len(stowagePlanResponse.Data) == 0 {
		return fmt.Errorf("no stowage plan details found for manifest ID: %s", manifestId)
	}

	// Process and Store Stowage Plan in Local Database
	st := manifestdataservices.InsertDeckStowagePlanDetails(stowagePlanResponse.Data)
	if st.State != constants.SuccessState {
		return fmt.Errorf("failed to insert stowage plan for manifest: %s", st.Data)
	}
	slog.Info(fmt.Sprintf("Successfully fetched and stored stowage plan for manifest ID: %s", manifestId))
	return nil
}

func FetchPackageList(logId, logKey, manifestId string) error {
	remoteBaseUrl := specials.GetEnvVariable("REMOTE_SERVER_URL", "http://localhost:8000")
	client := base_api.NewAPI(remoteBaseUrl)

	type PackageListResponse struct {
		State string                         `json:"state"`
		Data  []manifest.PackageManifestInfo `json:"data"`
	}
	headers := []baseapi.Header{
		{Key: "Log-ID", Value: logId},
		{Key: "Log-Key", Value: logKey},
	}
	page := 1
	limit := 50
	for {
		var packageListResponse PackageListResponse
		paginatedPath := fmt.Sprintf("/admin/get/manifest/packages/%s?page=%d&limit=%d", manifestId, page, limit)
		resp, err := client.Get(paginatedPath, &packageListResponse, headers)
		if err != nil {
			return fmt.Errorf("failed to fetch package list for manifest from remote server: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("remote server returned status %d on %s", resp.StatusCode, paginatedPath)
		}

		if packageListResponse.State != constants.SuccessState {
			return fmt.Errorf("failed to fetch package list for manifest: %s", resp.Message)
		}

		if len(packageListResponse.Data) == 0 {
			break
		}
		// Process and Store Package List in Local Database
		st := manifestdataservices.InsertPackageDetails(packageListResponse.Data, false)
		if st.State != constants.SuccessState {
			return fmt.Errorf("failed to insert package list for manifest: %s", st.Data)
		}
		page++ // Move to next page
	}
	slog.Info(fmt.Sprintf("Successfully fetched and stored package list for manifest ID: %s", manifestId))
	return nil
}

func UploadVehicleInspectionData(logId, logKey string, req manifest.InspectionChecksRequest) (resp *constants.AnswerState, err error) {
	remoteBaseUrl := specials.GetEnvVariable("REMOTE_SERVER_URL", "http://localhost:8000")
	client := base_api.NewAPI(remoteBaseUrl)
	headers := []baseapi.Header{
		{Key: "Log-ID", Value: logId},
		{Key: "Log-Key", Value: logKey},
	}
	var response constants.AnswerState
	url := "/admin/save/vehicle/inspection/details"
	respx, err := client.Post(url, req, &response, headers)
	if err != nil {
		return nil, err
	}

	if respx.StatusCode != http.StatusOK {
		slog.Error(fmt.Sprintf("Failed to upload inspection data, status code: %d and error %v", respx.StatusCode, err))
		return nil, fmt.Errorf("failed to upload inspection data, status code: %d", respx.StatusCode)
	}

	if response.State != constants.SuccessState {
		slog.Error(fmt.Sprintf("Failed to upload inspection data, state: %s", response.State))
		return nil, err
	}

	return &response, nil
}

func UploadPackageInspectionData(logId, logKey string, req manifest.PackageInspectionSaveRequest) (resp *constants.AnswerState, err error) {
	remoteBaseUrl := specials.GetEnvVariable("REMOTE_SERVER_URL", "http://localhost:8000")
	client := base_api.NewAPI(remoteBaseUrl)
	headers := []baseapi.Header{
		{Key: "Log-ID", Value: logId},
		{Key: "Log-Key", Value: logKey},
	}
	var response constants.AnswerState
	url := "/admin/save/package/inspection"
	respx, err := client.Post(url, req, &response, headers)
	if err != nil {
		return nil, err
	}

	if respx.StatusCode != http.StatusOK {
		slog.Error(fmt.Sprintf("Failed to upload package inspection data, status code: %d and error %v", respx.StatusCode, err))
		return nil, fmt.Errorf("failed to upload package inspection data, status code: %d", respx.StatusCode)
	}

	if response.State != constants.SuccessState {
		slog.Error(fmt.Sprintf("Failed to upload package inspection data, state: %s", response.State))
		return nil, err
	}

	return &response, nil
}
