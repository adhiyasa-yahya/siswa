// Package handler contains the Gin HTTP handlers (transport layer only: bind,
// delegate to a service, render the standard envelope).
package handler

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/nugroho/siswa/backend/internal/dto"
	"github.com/nugroho/siswa/backend/internal/errs"
	"github.com/nugroho/siswa/backend/internal/middleware"
	"github.com/nugroho/siswa/backend/internal/response"
	"github.com/nugroho/siswa/backend/internal/service"
)

// ItemHandler serves /api/v1/items and /api/v1/categories.
type ItemHandler struct {
	items *service.ItemService
}

func NewItemHandler(items *service.ItemService) *ItemHandler {
	return &ItemHandler{items: items}
}

// Create handles POST /api/v1/items.
func (h *ItemHandler) Create(c *gin.Context) {
	var req dto.CreateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	item, err := h.items.Create(c.Request.Context(), req, c.GetString(middleware.ContextRequestID))
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Created(c, "Item created successfully", dto.FromItem(*item))
}

// List handles GET /api/v1/items (search, category filter, pagination, sorting).
func (h *ItemHandler) List(c *gin.Context) {
	q := dto.ListItemQuery{
		Search:  firstNonEmpty(c.Query("search"), c.Query("q")),
		SortBy:  c.Query("sort_by"),
		SortDir: c.Query("sort_dir"),
		Page:    queryInt(c, "page", 1),
		Limit:   queryInt(c, "limit", 10),
	}

	if raw := strings.TrimSpace(c.Query("category_id")); raw != "" {
		categoryID, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || categoryID <= 0 {
			response.Error(c, errs.Validation("Invalid query parameter",
				errs.FieldError{Field: "category_id", Reason: "Must be a positive integer"}))
			return
		}
		q.CategoryID = &categoryID
	}

	q.Sanitize()

	items, total, err := h.items.List(c.Request.Context(), q)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OKWithMeta(c, "Items retrieved successfully", dto.FromItems(items),
		response.NewMeta(q.Page, q.Limit, total))
}

// Get handles GET /api/v1/items/:id.
func (h *ItemHandler) Get(c *gin.Context) {
	id, err := pathID(c, "id")
	if err != nil {
		response.Error(c, err)
		return
	}

	item, err := h.items.Get(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, "Item retrieved successfully", dto.FromItem(*item))
}

// Update handles PUT /api/v1/items/:id.
func (h *ItemHandler) Update(c *gin.Context) {
	id, err := pathID(c, "id")
	if err != nil {
		response.Error(c, err)
		return
	}

	var req dto.UpdateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	item, err := h.items.Update(c.Request.Context(), id, req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, "Item updated successfully", dto.FromItem(*item))
}

// Delete handles DELETE /api/v1/items/:id (soft delete).
func (h *ItemHandler) Delete(c *gin.Context) {
	id, err := pathID(c, "id")
	if err != nil {
		response.Error(c, err)
		return
	}

	if err := h.items.Delete(c.Request.Context(), id); err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, "Item deleted successfully", gin.H{"id": id, "soft_deleted": true})
}

// CreateCategory handles POST /api/v1/categories.
func (h *ItemHandler) CreateCategory(c *gin.Context) {
	var req dto.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	category, err := h.items.CreateCategory(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Created(c, "Category created successfully", category)
}

// Categories handles GET /api/v1/categories.
func (h *ItemHandler) Categories(c *gin.Context) {
	onlyActive := strings.EqualFold(c.DefaultQuery("status", "active"), "active")

	categories, err := h.items.Categories(c.Request.Context(), onlyActive)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, "Categories retrieved successfully", categories)
}

// --- helpers ---------------------------------------------------------------

func pathID(c *gin.Context, name string) (int64, error) {
	raw := c.Param(name)
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, errs.Validation("Invalid path parameter",
			errs.FieldError{Field: name, Reason: "Must be a positive integer"})
	}
	return id, nil
}

func queryInt(c *gin.Context, name string, fallback int) int {
	raw := strings.TrimSpace(c.Query(name))
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return value
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if trimmed := strings.TrimSpace(v); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
