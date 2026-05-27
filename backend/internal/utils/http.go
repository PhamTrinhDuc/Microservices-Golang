package utils

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// HTTPResponse wraps HTTP responses to be consistent
type HTTPResponse struct {
	Data  interface{} `json:"data,omitempty"`
	Error string      `json:"error,omitempty"`
}

// RespondOK sends 200 OK response
func RespondOK(ctx *gin.Context, data interface{}) {
	ctx.JSON(http.StatusOK, HTTPResponse{Data: data})
}

// RespondCreated sends 201 Created response
func RespondCreated(ctx *gin.Context, data interface{}) {
	ctx.JSON(http.StatusCreated, HTTPResponse{Data: data})
}

// RespondNoContent sends 204 No Content response
func RespondNoContent(ctx *gin.Context) {
	ctx.Status(http.StatusNoContent)
}

// RespondBadRequest sends 400 Bad Request response
func RespondBadRequest(ctx *gin.Context, message string) {
	ctx.JSON(http.StatusBadRequest, HTTPResponse{Error: message})
}

// RespondNotFound sends 404 Not Found response
func RespondNotFound(ctx *gin.Context, message string) {
	ctx.JSON(http.StatusNotFound, HTTPResponse{Error: message})
}

// RespondForbidden sends 403 Forbidden response
func RespondForbidden(ctx *gin.Context, message string) {
	ctx.JSON(http.StatusForbidden, HTTPResponse{Error: message})
}

// RespondUnauthorized sends 401 Unauthorized response
func RespondUnauthorized(ctx *gin.Context, message string) {
	ctx.JSON(http.StatusUnauthorized, HTTPResponse{Error: message})
}

// RespondInternalError sends 500 Internal Server Error response
func RespondInternalError(ctx *gin.Context, message string) {
	ctx.JSON(http.StatusInternalServerError, HTTPResponse{Error: message})
}

// RespondError sends appropriate error response based on error type
func RespondError(ctx *gin.Context, err error) {
	if err == nil {
		return
	}
	RespondInternalError(ctx, err.Error())
}

// GetQueryInt parses query parameter as int with default value
func GetQueryInt(ctx *gin.Context, key string, defaultVal int) int {
	if val := ctx.Query(key); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil && parsed > 0 {
			return parsed
		}
	}
	return defaultVal
}
