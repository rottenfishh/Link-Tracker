package dto

import "net/http"

type APIErrorResponse struct {
	Description      string   `json:"description"`
	Code             int      `json:"code"`
	ExceptionName    string   `json:"expection_name"`
	ExceptionMessage string   `json:"exception_message"`
	StackTrace       []string `json:"stack_trace"`
}

func NewAPIErrorResponse(description string, code int,
	exceptionName string, exceptionMsg string, stackTrace []string) APIErrorResponse {
	return APIErrorResponse{description, code, exceptionName, exceptionMsg, stackTrace}
}

func NewRequestParsingError(err error) APIErrorResponse {
	return NewAPIErrorResponse("Error parsing request",
		http.StatusInternalServerError, "Request error", err.Error(), nil)
}

func NewServiceError(desc string, err error, code int) APIErrorResponse {
	return NewAPIErrorResponse(desc,
		code, "Service error", err.Error(), nil)
}
