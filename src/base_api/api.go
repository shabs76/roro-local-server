package base_api

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	baseapi "github.com/shabs76/roro-local-server/constants/modules/base_api"
)

type API struct {
	BaseURL string
	Client  *http.Client
}

func NewAPI(baseURL string) *API {
	return &API{
		BaseURL: baseURL,
		Client:  &http.Client{},
	}
}

func (a *API) doRequest(method, path string, body io.Reader, headers []baseapi.Header) (*http.Response, error) {
	url := a.BaseURL + path
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	for _, h := range headers {
		req.Header.Set(h.Key, h.Value)
	}

	return a.Client.Do(req)
}

func processResponse(resp *http.Response, target any, err error) (*baseapi.APIResponse, error) {
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	apiResp := &baseapi.APIResponse{
		StatusCode: resp.StatusCode,
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		apiResp.Status = "success"
	} else {
		apiResp.Status = "error"
	}

	// Read body to parse
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		apiResp.Message = "Failed to read response body"
		return apiResp, err
	}

	if len(bodyBytes) > 0 {
		// check if state is success first
		var stateCheck struct {
			State string `json:"state"`
		}
		if json.Unmarshal(bodyBytes, &stateCheck) == nil && stateCheck.State != "success" {
			apiResp.Status = "error"
		}
		// If target is provided, unmarshal into it
		if target != nil {
			if jsonErr := json.Unmarshal(bodyBytes, target); jsonErr == nil {
				apiResp.Data = target
			} else if apiResp.Status == "error" {
				var errData baseapi.APIErrorData
				if jsonErr2 := json.Unmarshal(bodyBytes, &errData); jsonErr2 == nil {
					apiResp.Data = errData
					apiResp.Message = errData.Data
				} else {
					apiResp.Data = string(bodyBytes)
				}
			} else {
				// Failed to unmarshal into target, maybe explicitly string?
				apiResp.Data = string(bodyBytes)
			}
		} else {
			// No target, try generic map or keep raw?
			// Let's store raw json in Data if it is valid json, else string
			var generic any
			if json.Unmarshal(bodyBytes, &generic) == nil {
				apiResp.Data = generic
			} else {
				apiResp.Data = string(bodyBytes)
			}
		}

		// Try to extract message if present in JSON structure
		var partial struct {
			Message string `json:"message"`
		}
		if json.Unmarshal(bodyBytes, &partial) == nil && partial.Message != "" {
			apiResp.Message = partial.Message
		} else if apiResp.Status == "error" && apiResp.Message == "" {
			// If error and no JSON message, use body as message
			apiResp.Message = string(bodyBytes)
		}
	}

	if apiResp.Message == "" {
		apiResp.Message = http.StatusText(resp.StatusCode)
	}

	return apiResp, nil
}

func (a *API) Get(path string, target any, headers []baseapi.Header) (*baseapi.APIResponse, error) {
	resp, err := a.doRequest(http.MethodGet, path, nil, headers)
	return processResponse(resp, target, err)
}

func (a *API) GetStream(path string, headers []baseapi.Header) (io.ReadCloser, error) {
	resp, err := a.doRequest(http.MethodGet, path, nil, headers)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		resp.Body.Close()
		return nil, http.ErrMissingFile // Or a custom error
	}
	return resp.Body, nil
}

func prepareAPIRequest(body any, headers []baseapi.Header) (io.Reader, []baseapi.Header, error) {
	if body == nil {
		return nil, headers, nil
	}
	if r, ok := body.(io.Reader); ok {
		return r, headers, nil
	}

	data, err := json.Marshal(body)
	if err != nil {
		return nil, headers, err
	}

	newHeaders := make([]baseapi.Header, len(headers), len(headers)+1)
	copy(newHeaders, headers)

	hasContentType := false
	for _, h := range newHeaders {
		if h.Key == "Content-Type" {
			hasContentType = true
			break
		}
	}
	if !hasContentType {
		newHeaders = append(newHeaders, baseapi.Header{Key: "Content-Type", Value: "application/json"})
	}

	return bytes.NewBuffer(data), newHeaders, nil
}

func (a *API) Post(path string, body any, target any, headers []baseapi.Header) (*baseapi.APIResponse, error) {
	reqBody, reqHeaders, err := prepareAPIRequest(body, headers)
	if err != nil {
		return nil, err
	}
	resp, err := a.doRequest(http.MethodPost, path, reqBody, reqHeaders)
	return processResponse(resp, target, err)
}

func (a *API) Put(path string, body any, target any, headers []baseapi.Header) (*baseapi.APIResponse, error) {
	reqBody, reqHeaders, err := prepareAPIRequest(body, headers)
	if err != nil {
		return nil, err
	}
	resp, err := a.doRequest(http.MethodPut, path, reqBody, reqHeaders)
	return processResponse(resp, target, err)
}

func (a *API) Delete(path string, target any, headers []baseapi.Header) (*baseapi.APIResponse, error) {
	resp, err := a.doRequest(http.MethodDelete, path, nil, headers)
	return processResponse(resp, target, err)
}

func (a *API) Patch(path string, body any, target any, headers []baseapi.Header) (*baseapi.APIResponse, error) {
	reqBody, reqHeaders, err := prepareAPIRequest(body, headers)
	if err != nil {
		return nil, err
	}
	resp, err := a.doRequest(http.MethodPatch, path, reqBody, reqHeaders)
	return processResponse(resp, target, err)
}

func (a *API) SendFile(path string, filePath string, fileParamName string, extraFields map[string]string, target any, headers []baseapi.Header) (*baseapi.APIResponse, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile(fileParamName, filepath.Base(filePath))
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return nil, err
	}

	for key, val := range extraFields {
		_ = writer.WriteField(key, val)
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}

	reqHeaders := make([]baseapi.Header, len(headers))
	copy(reqHeaders, headers)
	reqHeaders = append(reqHeaders, baseapi.Header{Key: "Content-Type", Value: writer.FormDataContentType()})

	return a.Post(path, body, target, reqHeaders)
}
