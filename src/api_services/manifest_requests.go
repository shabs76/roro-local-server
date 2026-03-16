package apiservices

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/shabs76/roro-local-server/base_api"
	"github.com/shabs76/roro-local-server/constants"
	baseapi "github.com/shabs76/roro-local-server/constants/modules/base_api"
	"github.com/shabs76/roro-local-server/constants/modules/manifest"
	"github.com/shabs76/roro-local-server/specials"
)

func GetManifestLists(logId, logKey, query string, page, limit int) (data []manifest.ManifestToshow, err error) {
	remoteBaseUrl := specials.GetEnvVariable("REMOTE_SERVER_URL", "http://localhost:8000")
	client := base_api.NewAPI(remoteBaseUrl)
	type ManifestListResponse struct {
		State string                    `json:"state"`
		Data  []manifest.ManifestToshow `json:"data"`
	}
	var manifestListResponse ManifestListResponse
	headers := []baseapi.Header{
		{Key: "Log-ID", Value: logId},
		{Key: "Log-Key", Value: logKey},
	}
	url := fmt.Sprintf("/admin/get/manifests?query=%s&page=%d&limit=%d", query, page, limit)
	resp, err := client.Get(url, &manifestListResponse, headers)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, err
	}

	if manifestListResponse.State != constants.SuccessState {
		return nil, err
	}

	return manifestListResponse.Data, nil
}

func UploadSingleVehicleToRemote(logId, logKey string, req manifest.AddVehicleRequest) (newVehicleId string, err error) {
	remoteBaseUrl := specials.GetEnvVariable("REMOTE_SERVER_URL", "http://localhost:8000")
	client := base_api.NewAPI(remoteBaseUrl)
	headers := []baseapi.Header{
		{Key: "Log-ID", Value: logId},
		{Key: "Log-Key", Value: logKey},
	}
	var response constants.AnswerState
	url := "/admin/save/vehicle/in/manifest/manual"
	respx, err := client.Post(url, req, &response, headers)
	if err != nil {
		return "", err
	}

	if respx.StatusCode != http.StatusOK {
		slog.Error(fmt.Sprintf("Failed to upload vehicle, status code: %d and error %v", respx.StatusCode, err))
		return "", fmt.Errorf("failed to upload vehicle, status code: %d", respx.StatusCode)
	}

	if response.State != constants.SuccessState {
		slog.Error(fmt.Sprintf("Failed to upload vehicle, state: %s", response.State))
		return "", err
	}

	return response.Adv, nil
}

func UploadSinglePackageToRemote(logId, logKey string, req manifest.AddPackageRequest) (newPackageId string, err error) {
	remoteBaseUrl := specials.GetEnvVariable("REMOTE_SERVER_URL", "http://localhost:8000")
	client := base_api.NewAPI(remoteBaseUrl)
	headers := []baseapi.Header{
		{Key: "Log-ID", Value: logId},
		{Key: "Log-Key", Value: logKey},
	}
	var response constants.AnswerState
	url := "/admin/save/manifest/packages"
	respx, err := client.Post(url, req, &response, headers)
	if err != nil {
		return "", err
	}

	if respx.StatusCode != http.StatusOK {
		slog.Error(fmt.Sprintf("Failed to upload package, status code: %d and error %v", respx.StatusCode, err))
		return "", fmt.Errorf("failed to upload package, status code: %d", respx.StatusCode)
	}

	if response.State != constants.SuccessState {
		slog.Error(fmt.Sprintf("Failed to upload package, state: %s", response.State))
		return "", err
	}

	return response.Adv, nil
}

func UploadImage(logId, logKey, filePath string) (url string, err error) {
	remoteBaseUrl := specials.GetEnvVariable("REMOTE_SERVER_URL", "http://localhost:8000")
	client := base_api.NewAPI(remoteBaseUrl)

	type UploadResponse struct {
		Data  string `json:"data"`
		State string `json:"state"`
	}
	var uploadResp UploadResponse
	headers := []baseapi.Header{
		{Key: "Log-ID", Value: logId},
		{Key: "Log-Key", Value: logKey},
	}

	respx, err := client.SendFile("/media/upload/image", filePath, "image", nil, &uploadResp, headers)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to upload image, error %v", err))
		return "", err
	}

	if respx.StatusCode != http.StatusOK {
		slog.Error(fmt.Sprintf("Failed to upload image, status code: %d", respx.StatusCode))
		return "", fmt.Errorf("failed to upload image, status code: %d", respx.StatusCode)
	}

	if uploadResp.State != constants.SuccessState {
		slog.Error(fmt.Sprintf("Failed to upload image, state: %s", uploadResp.State))
		return "", fmt.Errorf("failed to upload image, state: %s", uploadResp.State)
	}

	return uploadResp.Data, nil
}

func UploadVideo(logId, logKey, filePath string) (url string, err error) {
	remoteBaseUrl := specials.GetEnvVariable("REMOTE_SERVER_URL", "http://localhost:8000")
	client := base_api.NewAPI(remoteBaseUrl)

	type UploadResponse struct {
		Data  string `json:"data"`
		State string `json:"state"`
	}
	var uploadResp UploadResponse
	headers := []baseapi.Header{
		{Key: "Log-ID", Value: logId},
		{Key: "Log-Key", Value: logKey},
	}

	respx, err := client.SendFile("/media/upload/video", filePath, "video", nil, &uploadResp, headers)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to upload video, error %v", err))
		return "", err
	}

	if respx.StatusCode != http.StatusOK {
		slog.Error(fmt.Sprintf("Failed to upload video, status code: %d", respx.StatusCode))
		return "", fmt.Errorf("failed to upload video, status code: %d", respx.StatusCode)
	}

	if uploadResp.State != constants.SuccessState {
		slog.Error(fmt.Sprintf("Failed to upload video, state: %s", uploadResp.State))
		return "", fmt.Errorf("failed to upload video, state: %s", uploadResp.State)
	}

	return uploadResp.Data, nil
}

func UploadPdf(logId, logKey, filePath string) (url string, err error) {
	remoteBaseUrl := specials.GetEnvVariable("REMOTE_SERVER_URL", "http://localhost:8000")
	client := base_api.NewAPI(remoteBaseUrl)

	type UploadResponse struct {
		Data  string `json:"data"`
		State string `json:"state"`
	}
	var uploadResp UploadResponse
	headers := []baseapi.Header{
		{Key: "Log-ID", Value: logId},
		{Key: "Log-Key", Value: logKey},
	}

	respx, err := client.SendFile("/media/upload/pdf", filePath, "pdf", nil, &uploadResp, headers)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to upload pdf, error %v", err))
		return "", err
	}

	if respx.StatusCode != http.StatusOK {
		slog.Error(fmt.Sprintf("Failed to upload pdf, status code: %d", respx.StatusCode))
		return "", fmt.Errorf("failed to upload pdf, status code: %d", respx.StatusCode)
	}

	if uploadResp.State != constants.SuccessState {
		slog.Error(fmt.Sprintf("Failed to upload pdf, state: %s", uploadResp.State))
		return "", fmt.Errorf("failed to upload pdf, state: %s", uploadResp.State)
	}

	return uploadResp.Data, nil
}
