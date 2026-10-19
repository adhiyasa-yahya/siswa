// Package errs defines the application error catalogue.
//
// Every error that leaves a service is translated into an *AppError carrying a
// stable business code (0 = success, 99 = generic error, 1xxx = client faults,
// 2xxx = infrastructure faults) plus the HTTP status used by the transport layer.
package errs

import (
	"errors"
	"fmt"
	"net/http"
)

// Code is the business error code returned in every JSON envelope.
type Code int

const (
	CodeSuccess Code = 0  // 0  : sukses
	CodeError   Code = 99 // 99 : error (generic / unclassified)

	// 1xxx - client / validation faults.
	CodeValidation      Code = 1001
	CodeNotFound        Code = 1002
	CodeDuplicateSKU    Code = 1003
	CodeNegativeStock   Code = 1004
	CodeInvalidCategory Code = 1005
	CodeInvalidLocation Code = 1006
	CodeItemDeleted     Code = 1007
	CodeInvalidPayload  Code = 1008
	CodeInvalidQuantity Code = 1009
	CodeRouteNotFound   Code = 1010

	// 2xxx - infrastructure faults.
	CodeDatabaseError Code = 2001
	CodeCacheError    Code = 2002
	CodeInternalError Code = 2003
)

// CodeInfo documents a single entry of the error catalogue.
type CodeInfo struct {
	Code        Code   `json:"code"`
	Name        string `json:"name"`
	HTTPStatus  int    `json:"http_status"`
	Description string `json:"description"`
}

// Catalogue is the human readable list of every code the API can return.
// It is exposed through GET /api/v1/error-codes and documented in Swagger.
var Catalogue = []CodeInfo{
	{Code: CodeSuccess, Name: "SUCCESS", HTTPStatus: http.StatusOK, Description: "Request processed successfully"},
	{Code: CodeError, Name: "ERROR", HTTPStatus: http.StatusInternalServerError, Description: "Generic / unclassified error"},
	{Code: CodeValidation, Name: "VALIDATION_ERROR", HTTPStatus: http.StatusBadRequest, Description: "One or more fields failed validation"},
	{Code: CodeNotFound, Name: "NOT_FOUND", HTTPStatus: http.StatusNotFound, Description: "Requested resource does not exist"},
	{Code: CodeDuplicateSKU, Name: "DUPLICATE_SKU", HTTPStatus: http.StatusConflict, Description: "SKU already exists"},
	{Code: CodeNegativeStock, Name: "NEGATIVE_STOCK", HTTPStatus: http.StatusUnprocessableEntity, Description: "Mutation would make the stock negative"},
	{Code: CodeInvalidCategory, Name: "INVALID_CATEGORY", HTTPStatus: http.StatusUnprocessableEntity, Description: "Category does not exist or is inactive"},
	{Code: CodeInvalidLocation, Name: "INVALID_LOCATION", HTTPStatus: http.StatusUnprocessableEntity, Description: "Warehouse location does not exist"},
	{Code: CodeItemDeleted, Name: "ITEM_DELETED", HTTPStatus: http.StatusGone, Description: "Item has been soft deleted"},
	{Code: CodeInvalidPayload, Name: "INVALID_PAYLOAD", HTTPStatus: http.StatusBadRequest, Description: "Request body is not valid JSON"},
	{Code: CodeInvalidQuantity, Name: "INVALID_QUANTITY", HTTPStatus: http.StatusBadRequest, Description: "Quantity must be a positive integer"},
	{Code: CodeRouteNotFound, Name: "ROUTE_NOT_FOUND", HTTPStatus: http.StatusNotFound, Description: "Endpoint or method is not registered"},
	{Code: CodeDatabaseError, Name: "DATABASE_ERROR", HTTPStatus: http.StatusInternalServerError, Description: "Database command failed"},
	{Code: CodeCacheError, Name: "CACHE_ERROR", HTTPStatus: http.StatusInternalServerError, Description: "Redis command failed"},
	{Code: CodeInternalError, Name: "INTERNAL_ERROR", HTTPStatus: http.StatusInternalServerError, Description: "Unexpected server error"},
}

// FieldError describes a single invalid field of the request.
type FieldError struct {
	Field  string `json:"field"`
	Reason string `json:"reason"`
}

// AppError is the only error type the HTTP layer needs to understand.
type AppError struct {
	Code       Code
	HTTPStatus int
	Message    string
	Fields     []FieldError
	cause      error
}

func (e *AppError) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.cause)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error { return e.cause }

// WithCause attaches the technical error, kept for the server side log only.
func (e *AppError) WithCause(err error) *AppError {
	e.cause = err
	return e
}

// WithFields attaches field level details to the error.
func (e *AppError) WithFields(fields ...FieldError) *AppError {
	e.Fields = append(e.Fields, fields...)
	return e
}

func newf(code Code, status int, format string, args ...any) *AppError {
	return &AppError{Code: code, HTTPStatus: status, Message: fmt.Sprintf(format, args...)}
}

// --- constructors -----------------------------------------------------------

func Validation(message string, fields ...FieldError) *AppError {
	return &AppError{Code: CodeValidation, HTTPStatus: http.StatusBadRequest, Message: message, Fields: fields}
}

func InvalidPayload(cause error) *AppError {
	return newf(CodeInvalidPayload, http.StatusBadRequest, "Invalid request payload").WithCause(cause)
}

func NotFound(entity string, id any) *AppError {
	return newf(CodeNotFound, http.StatusNotFound, "%s with id %v was not found", entity, id)
}

func NotFoundMsg(message string) *AppError {
	return newf(CodeNotFound, http.StatusNotFound, "%s", message)
}

func DuplicateSKU(sku string) *AppError {
	return (&AppError{
		Code:       CodeDuplicateSKU,
		HTTPStatus: http.StatusConflict,
		Message:    "SKU already exists",
	}).WithFields(FieldError{Field: "sku", Reason: fmt.Sprintf("Duplicate entry: %s", sku)})
}

func NegativeStock(itemID, locationID int64, current, delta int64) *AppError {
	return (&AppError{
		Code:       CodeNegativeStock,
		HTTPStatus: http.StatusUnprocessableEntity,
		Message:    "Stock cannot be negative",
	}).WithFields(FieldError{
		Field: "qty",
		Reason: fmt.Sprintf(
			"current stock of item %d at location %d is %d, mutation %d would result in %d",
			itemID, locationID, current, delta, current+delta,
		),
	})
}

func InvalidCategory(categoryID int64, reason string) *AppError {
	return (&AppError{
		Code:       CodeInvalidCategory,
		HTTPStatus: http.StatusUnprocessableEntity,
		Message:    fmt.Sprintf("Category %d is not usable", categoryID),
	}).WithFields(FieldError{Field: "category_id", Reason: reason})
}

func InvalidLocation(locationID int64) *AppError {
	return (&AppError{
		Code:       CodeInvalidLocation,
		HTTPStatus: http.StatusUnprocessableEntity,
		Message:    "Location does not exist",
	}).WithFields(FieldError{Field: "location_id", Reason: fmt.Sprintf("Location %d is not registered", locationID)})
}

func InvalidQuantity(reason string) *AppError {
	return (&AppError{
		Code:       CodeInvalidQuantity,
		HTTPStatus: http.StatusBadRequest,
		Message:    "Quantity is invalid",
	}).WithFields(FieldError{Field: "qty", Reason: reason})
}

func RouteNotFound(method, path string) *AppError {
	return newf(CodeRouteNotFound, http.StatusNotFound, "Route %s %s is not registered", method, path)
}

func Database(cause error) *AppError {
	return newf(CodeDatabaseError, http.StatusInternalServerError, "Database error").WithCause(cause)
}

func Cache(cause error) *AppError {
	return newf(CodeCacheError, http.StatusInternalServerError, "Cache error").WithCause(cause)
}

func Internal(cause error) *AppError {
	return newf(CodeInternalError, http.StatusInternalServerError, "Internal server error").WithCause(cause)
}

// As reports whether err is an *AppError and returns it.
func As(err error) (*AppError, bool) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}

// Filter normalises any error into an *AppError so the HTTP layer never leaks
// driver specific messages to the client. This is the single error filter used
// by every handler.
func Filter(err error) *AppError {
	if err == nil {
		return nil
	}
	if appErr, ok := As(err); ok {
		return appErr
	}
	return Internal(err)
}
