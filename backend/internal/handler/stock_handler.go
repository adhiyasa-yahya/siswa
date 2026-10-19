package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/nugroho/siswa/backend/internal/dto"
	"github.com/nugroho/siswa/backend/internal/middleware"
	"github.com/nugroho/siswa/backend/internal/response"
	"github.com/nugroho/siswa/backend/internal/service"
)

// StockHandler serves /api/v1/stock and /api/v1/locations.
type StockHandler struct {
	stocks *service.StockService
}

func NewStockHandler(stocks *service.StockService) *StockHandler {
	return &StockHandler{stocks: stocks}
}

// Receive handles POST /api/v1/stock/receive.
func (h *StockHandler) Receive(c *gin.Context) {
	var req dto.ReceiveStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	result, err := h.stocks.Receive(c.Request.Context(), req, c.GetString(middleware.ContextRequestID))
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Created(c, "Stock received successfully", result)
}

// ByItem handles GET /api/v1/stock/:item_id.
func (h *StockHandler) ByItem(c *gin.Context) {
	itemID, err := pathID(c, "item_id")
	if err != nil {
		response.Error(c, err)
		return
	}

	stock, err := h.stocks.StockByItem(c.Request.Context(), itemID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, "Stock retrieved successfully", stock)
}

// CreateLocation handles POST /api/v1/locations.
func (h *StockHandler) CreateLocation(c *gin.Context) {
	var req dto.CreateLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	location, err := h.stocks.CreateLocation(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Created(c, "Location created successfully", location)
}

// Locations handles GET /api/v1/locations.
func (h *StockHandler) Locations(c *gin.Context) {
	locations, err := h.stocks.Locations(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, "Locations retrieved successfully", locations)
}
