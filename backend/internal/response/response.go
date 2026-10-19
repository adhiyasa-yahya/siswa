// Package response renders the single JSON envelope used by every endpoint.
//
// Success:
//
//	{"success":true,"code":0,"message":"...","data":{...},"meta":{...}}
//
// Error:
//
//	{"success":false,"code":1003,"message":"SKU already exists",
//	 "errors":[{"field":"sku","reason":"Duplicate entry"}]}
package response

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/nugroho/siswa/backend/internal/errs"
)

// Meta carries pagination information for list endpoints.
type Meta struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// NewMeta builds pagination metadata, deriving the page count from the total.
func NewMeta(page, limit int, total int64) *Meta {
	totalPages := 0
	if limit > 0 {
		totalPages = int((total + int64(limit) - 1) / int64(limit))
	}
	return &Meta{Page: page, Limit: limit, Total: total, TotalPages: totalPages}
}

// Envelope is the base JSON body returned by the API.
type Envelope struct {
	Success bool              `json:"success"`
	Code    errs.Code         `json:"code"`
	Message string            `json:"message"`
	Data    any               `json:"data,omitempty"`
	Meta    *Meta             `json:"meta,omitempty"`
	Errors  []errs.FieldError `json:"errors,omitempty"`
}

// OK writes 200 with a data payload.
func OK(c *gin.Context, message string, data any) {
	write(c, http.StatusOK, Envelope{Success: true, Code: errs.CodeSuccess, Message: message, Data: data})
}

// OKWithMeta writes 200 with a data payload and pagination metadata.
func OKWithMeta(c *gin.Context, message string, data any, meta *Meta) {
	write(c, http.StatusOK, Envelope{Success: true, Code: errs.CodeSuccess, Message: message, Data: data, Meta: meta})
}

// Created writes 201 with the newly created resource.
func Created(c *gin.Context, message string, data any) {
	write(c, http.StatusCreated, Envelope{Success: true, Code: errs.CodeSuccess, Message: message, Data: data})
}

// Error filters any error into the standard error envelope.
func Error(c *gin.Context, err error) {
	appErr := translate(err)

	// The technical cause stays server side; the client only sees Message/Errors.
	if appErr.HTTPStatus >= http.StatusInternalServerError {
		slog.Error("request failed",
			"code", int(appErr.Code),
			"status", appErr.HTTPStatus,
			"path", c.FullPath(),
			"error", appErr.Error(),
		)
	}

	c.Set("app_error_code", int(appErr.Code))
	write(c, appErr.HTTPStatus, Envelope{
		Success: false,
		Code:    appErr.Code,
		Message: appErr.Message,
		Errors:  appErr.Fields,
	})
}

func write(c *gin.Context, status int, body Envelope) {
	c.Set("response_code", int(body.Code))
	c.AbortWithStatusJSON(status, body)
}

// translate converts binding/validation errors into field level AppErrors and
// everything else through the generic error filter.
func translate(err error) *errs.AppError {
	var validationErrs validator.ValidationErrors
	if errors.As(err, &validationErrs) {
		fields := make([]errs.FieldError, 0, len(validationErrs))
		for _, fe := range validationErrs {
			fields = append(fields, errs.FieldError{
				Field:  jsonFieldName(fe),
				Reason: validationReason(fe),
			})
		}
		return errs.Validation("Validation failed", fields...)
	}
	return errs.Filter(err)
}

// jsonFieldName returns the JSON name of the offending field. The validator is
// configured (internal/validation) to report JSON tag names, so this is simply
// the reported field with a lower-cased fallback.
func jsonFieldName(fe validator.FieldError) string {
	if name := fe.Field(); name != "" {
		return name
	}
	return strings.ToLower(fe.StructField())
}

func validationReason(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "This field is required"
	case "min":
		return "Minimum value/length is " + fe.Param()
	case "max":
		return "Maximum value/length is " + fe.Param()
	case "gt":
		return "Must be greater than " + fe.Param()
	case "gte":
		return "Must be greater than or equal to " + fe.Param()
	case "oneof":
		return "Must be one of: " + fe.Param()
	case "email":
		return "Must be a valid email address"
	case "alphanumunicode", "sku":
		return "Only letters, digits, dash and underscore are allowed"
	default:
		return "Failed validation rule: " + fe.Tag()
	}
}
