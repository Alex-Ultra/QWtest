package utils

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type ResponseStatus string

const (
	ResponseStatusSuccess ResponseStatus = "success"
	ResponseStatusError   ResponseStatus = "error"
)

type APIResponse struct {
	Status    ResponseStatus `json:"status"`
	Data      interface{}    `json:"data"`
	Error     *APIError      `json:"error"`
	Timestamp time.Time      `json:"timestamp"`
}

type APIError struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

func SuccessResponse(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, APIResponse{
		Status:    ResponseStatusSuccess,
		Data:      data,
		Error:     nil,
		Timestamp: time.Now().UTC(),
	})
}

func ErrorResponse(c *gin.Context, statusCode int, errorCode, message string, details map[string]interface{}) {
	err := &APIError{
		Code:    errorCode,
		Message: message,
		Details: details,
	}

	c.JSON(statusCode, APIResponse{
		Status:    ResponseStatusError,
		Data:      nil,
		Error:     err,
		Timestamp: time.Now().UTC(),
	})
}

func ValidationErrorResponse(c *gin.Context, message string, details map[string]interface{}) {
	ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", message, details)
}

func UnauthorizedResponse(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", message, nil)
}

func ForbiddenResponse(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusForbidden, "FORBIDDEN", message, nil)
}

func NotFoundResponse(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusNotFound, "NOT_FOUND", message, nil)
}

func InternalServerErrorResponse(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", message, nil)
}