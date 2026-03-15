package dto

import "net/http"

type ApiErrorResponse struct {
	Description      string   `json:"description"`
	Code             int      `json:"code"`
	ExceptionName    string   `json:"expection_name"`
	ExceptionMessage string   `json:"exception_message"`
	StackTrace       []string `json:"stack_trace"`
}

func NewApiErrorResponse(description string, code int,
	exceptionName string, exceptionMsg string, stackTrace []string) ApiErrorResponse {
	return ApiErrorResponse{description, code, exceptionName, exceptionMsg, stackTrace}
}

func NewRequestParsingError(err error) ApiErrorResponse {
	return NewApiErrorResponse("Error parsing request",
		http.StatusInternalServerError, "Request error", err.Error(), nil)
}

func NewServiceError(desc string, err error, code int) ApiErrorResponse {
	return NewApiErrorResponse(desc,
		code, "Service error", err.Error(), nil)
}
