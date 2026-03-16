package baseapi

type APIResponse struct {
	StatusCode int    `json:"status_code"`
	Status     string `json:"status"` // "success" or "error"
	Message    string `json:"message"`
	Data       any    `json:"data"`
}

type APIErrorData struct {
	State string `json:"state"`
	Data  string `json:"data"`
}
