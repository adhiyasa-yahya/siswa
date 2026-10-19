// Package dto holds the request/response contracts of the HTTP layer.
package dto

import (
	"strings"

	"github.com/nugroho/siswa/backend/internal/model"
)

// CreateItemRequest is the body of POST /api/v1/items.
//
// LocationID and InitialQty are optional: when a location is given the item is
// placed there in the same transaction that creates it, with InitialQty (0 by
// default) as its starting quantity.
type CreateItemRequest struct {
	SKU        string `json:"sku" binding:"required,min=3,max=64"`
	Name       string `json:"name" binding:"required,min=2,max=160"`
	CategoryID int64  `json:"category_id" binding:"required,gt=0"`
	Unit       string `json:"unit" binding:"required,min=1,max=24"`
	LocationID *int64 `json:"location_id" binding:"omitempty,gt=0"`
	InitialQty *int64 `json:"initial_qty" binding:"omitempty,gte=0"`
}

// Placement returns the optional initial stock placement of a new item.
func (r *CreateItemRequest) Placement() (locationID *int64, qty int64) {
	if r.LocationID == nil {
		return nil, 0
	}
	if r.InitialQty != nil {
		qty = *r.InitialQty
	}
	return r.LocationID, qty
}

// Normalize trims the user input and upper-cases the SKU.
func (r *CreateItemRequest) Normalize() {
	r.SKU = strings.ToUpper(strings.TrimSpace(r.SKU))
	r.Name = strings.TrimSpace(r.Name)
	r.Unit = strings.TrimSpace(r.Unit)
}

// UpdateItemRequest is the body of PUT /api/v1/items/:id.
type UpdateItemRequest struct {
	SKU        string `json:"sku" binding:"required,min=3,max=64"`
	Name       string `json:"name" binding:"required,min=2,max=160"`
	CategoryID int64  `json:"category_id" binding:"required,gt=0"`
	Unit       string `json:"unit" binding:"required,min=1,max=24"`
}

func (r *UpdateItemRequest) Normalize() {
	r.SKU = strings.ToUpper(strings.TrimSpace(r.SKU))
	r.Name = strings.TrimSpace(r.Name)
	r.Unit = strings.TrimSpace(r.Unit)
}

// CreateCategoryRequest is the body of POST /api/v1/categories.
type CreateCategoryRequest struct {
	Name   string `json:"name" binding:"required,min=2,max=100"`
	Status string `json:"status" binding:"omitempty,oneof=active inactive"`
}

// Normalize trims the input and defaults the status to "active".
func (r *CreateCategoryRequest) Normalize() {
	r.Name = strings.TrimSpace(r.Name)
	r.Status = strings.ToLower(strings.TrimSpace(r.Status))
	if r.Status == "" {
		r.Status = "active"
	}
}

// CreateLocationRequest is the body of POST /api/v1/locations.
type CreateLocationRequest struct {
	Code string `json:"code" binding:"required,min=2,max=32"`
	Zone string `json:"zone" binding:"required,min=1,max=32"`
	Type string `json:"type" binding:"required,oneof=rack bin pallet cold_storage staging"`
}

// Normalize trims the input and upper-cases the location code.
func (r *CreateLocationRequest) Normalize() {
	r.Code = strings.ToUpper(strings.TrimSpace(r.Code))
	r.Zone = strings.ToUpper(strings.TrimSpace(r.Zone))
	r.Type = strings.ToLower(strings.TrimSpace(r.Type))
}

// ListItemQuery holds the filters of GET /api/v1/items.
type ListItemQuery struct {
	Search     string `form:"search"`
	CategoryID *int64 `form:"category_id"`
	Page       int    `form:"page"`
	Limit      int    `form:"limit"`
	SortBy     string `form:"sort_by"`
	SortDir    string `form:"sort_dir"`
}

// Sanitize applies defaults and guards against unbounded queries / SQL injection
// through the sort parameters.
func (q *ListItemQuery) Sanitize() {
	q.Search = strings.TrimSpace(q.Search)
	if q.Page < 1 {
		q.Page = 1
	}
	if q.Limit < 1 {
		q.Limit = 10
	}
	if q.Limit > 100 {
		q.Limit = 100
	}

	switch strings.ToLower(q.SortBy) {
	case "sku", "name", "created_at", "id", "category_id":
		q.SortBy = strings.ToLower(q.SortBy)
	default:
		q.SortBy = "created_at"
	}

	if strings.EqualFold(q.SortDir, "asc") {
		q.SortDir = "ASC"
	} else {
		q.SortDir = "DESC"
	}
}

// Offset is the SQL offset derived from page/limit.
func (q *ListItemQuery) Offset() int {
	return (q.Page - 1) * q.Limit
}

// ReceiveStockRequest is the body of POST /api/v1/stock/receive.
type ReceiveStockRequest struct {
	ItemID     int64  `json:"item_id" binding:"required,gt=0"`
	LocationID int64  `json:"location_id" binding:"required,gt=0"`
	Qty        int64  `json:"qty" binding:"required,gt=0"`
	Reference  string `json:"reference" binding:"omitempty,max=64"`
	Note       string `json:"note" binding:"omitempty,max=255"`
}

// ReceiveStockResponse is returned after a successful stock receipt.
type ReceiveStockResponse struct {
	StockID      int64  `json:"stock_id"`
	ItemID       int64  `json:"item_id"`
	SKU          string `json:"sku"`
	LocationID   int64  `json:"location_id"`
	LocationCode string `json:"location_code"`
	QtyBefore    int64  `json:"qty_before"`
	QtyReceived  int64  `json:"qty_received"`
	QtyAfter     int64  `json:"qty_after"`
	Reference    string `json:"reference,omitempty"`
	UpdatedAt    string `json:"updated_at"`
}

// ItemStockResponse is returned by GET /api/v1/stock/:item_id.
type ItemStockResponse struct {
	ItemID    int64         `json:"item_id"`
	SKU       string        `json:"sku"`
	Name      string        `json:"name"`
	Unit      string        `json:"unit"`
	TotalQty  int64         `json:"total_qty"`
	Locations []StockPerLoc `json:"locations"`
}

// StockPerLoc is the quantity of an item inside a single location.
type StockPerLoc struct {
	StockID      int64  `json:"stock_id"`
	LocationID   int64  `json:"location_id"`
	LocationCode string `json:"location_code"`
	Zone         string `json:"zone"`
	Type         string `json:"type"`
	Qty          int64  `json:"qty"`
	UpdatedAt    string `json:"updated_at"`
}

// ListLogQuery holds the filters of GET /api/v1/logs.
type ListLogQuery struct {
	Page   int    `form:"page"`
	Limit  int    `form:"limit"`
	Method string `form:"method"`
	Status *int   `form:"status"`
	Source string `form:"source"` // redis | db
}

func (q *ListLogQuery) Sanitize() {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.Limit < 1 {
		q.Limit = 25
	}
	if q.Limit > 200 {
		q.Limit = 200
	}
	q.Method = strings.ToUpper(strings.TrimSpace(q.Method))
	q.Source = strings.ToLower(strings.TrimSpace(q.Source))
}

func (q *ListLogQuery) Offset() int {
	return (q.Page - 1) * q.Limit
}

// ItemResponse is the item projection returned to clients.
type ItemResponse struct {
	ID           int64  `json:"id"`
	SKU          string `json:"sku"`
	Name         string `json:"name"`
	CategoryID   int64  `json:"category_id"`
	CategoryName string `json:"category_name"`
	Unit         string `json:"unit"`
	TotalQty     int64  `json:"total_qty"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
	DeletedAt    string `json:"deleted_at,omitempty"`
}

// FromItem maps a model.Item into its API projection.
func FromItem(it model.Item) ItemResponse {
	res := ItemResponse{
		ID:           it.ID,
		SKU:          it.SKU,
		Name:         it.Name,
		CategoryID:   it.CategoryID,
		CategoryName: it.CategoryName,
		Unit:         it.Unit,
		TotalQty:     it.TotalQty,
		CreatedAt:    it.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:    it.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
	}
	if it.DeletedAt != nil {
		res.DeletedAt = it.DeletedAt.UTC().Format("2006-01-02T15:04:05Z07:00")
	}
	return res
}

// FromItems maps a slice of items.
func FromItems(items []model.Item) []ItemResponse {
	out := make([]ItemResponse, 0, len(items))
	for _, it := range items {
		out = append(out, FromItem(it))
	}
	return out
}
