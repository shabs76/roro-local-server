package constants

type AnswerState struct {
	State string `json:"state" binding:"required"`
	Data  string `json:"data" binding:"required"`
	Adv   string `json:"adv" binding:"required"`
}

var SuccessState string = "success"
var ErrorState string = "error"

type PageNationSelect struct {
	Limit      string
	State      string
	ResNum     int
	TotalPages int
}

type statusesTypes struct {
	Active  string
	Deleted string
	Blocked string
	Expired string
}

var StatusTypes = statusesTypes{
	Active:  "active",
	Deleted: "deleted",
	Blocked: "blocked",
	Expired: "expired",
}

type attendencesTypes struct {
	Notyet   string
	Attended string
}

var AttendenceTypes = attendencesTypes{
	Notyet:   "notyet",
	Attended: "attended",
}

type NormalResponse struct {
	State   string `json:"state" binding:"required"`
	Data    any    `json:"data" binding:"required"`
	Message string `json:"message" binding:"required"`
	Adv     string `json:"adv" binding:"required"`
}

// PaginationInfo represents pagination metadata
type PaginationInfo struct {
	CurrentPage  int  `json:"currentPage"`
	ItemsPerPage int  `json:"itemsPerPage"`
	TotalItems   int  `json:"totalItems"`
	TotalPages   int  `json:"totalPages"`
	HasNextPage  bool `json:"hasNextPage"`
	HasPrevPage  bool `json:"hasPrevPage"`
}

type PaginationResponse struct {
	State      string         `json:"state"`
	Data       any            `json:"data"`
	Message    string         `json:"message"`
	Pagination PaginationInfo `json:"pagination"`
}

const MediaBaseDir = "media_data"
