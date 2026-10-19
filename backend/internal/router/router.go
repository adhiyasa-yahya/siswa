// Package router wires middleware, Swagger and the versioned API routes.
package router

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/nugroho/siswa/backend/docs"
	"github.com/nugroho/siswa/backend/internal/config"
	"github.com/nugroho/siswa/backend/internal/errs"
	"github.com/nugroho/siswa/backend/internal/handler"
	"github.com/nugroho/siswa/backend/internal/logstore"
	"github.com/nugroho/siswa/backend/internal/middleware"
	"github.com/nugroho/siswa/backend/internal/response"
)

// Handlers groups everything the router needs to mount.
type Handlers struct {
	Item   *handler.ItemHandler
	Stock  *handler.StockHandler
	Log    *handler.LogHandler
	System *handler.SystemHandler
}

// New builds the HTTP engine.
func New(cfg *config.Config, store *logstore.Store, h Handlers) *gin.Engine {
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.RedirectTrailingSlash = true
	r.HandleMethodNotAllowed = true

	r.Use(middleware.RequestID())
	r.Use(middleware.Recovery())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORSAllowOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", middleware.RequestIDHeader},
		ExposeHeaders:    []string{middleware.RequestIDHeader, "X-Log-Source"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))
	r.Use(middleware.RequestLogger(cfg, store))

	// Unknown route / method still answers with the standard envelope.
	r.NoRoute(func(c *gin.Context) {
		response.Error(c, errs.RouteNotFound(c.Request.Method, c.Request.URL.Path))
	})
	r.NoMethod(func(c *gin.Context) {
		response.Error(c, errs.RouteNotFound(c.Request.Method, c.Request.URL.Path))
	})

	docs.Register(r)
	r.GET("/health", h.System.Health)
	r.GET("/", func(c *gin.Context) {
		response.OK(c, "SISWA API is running", gin.H{
			"app":     cfg.AppName,
			"docs":    "/swagger/index.html",
			"health":  "/health",
			"api":     "/api/v1",
			"version": "1.0.0",
		})
	})

	v1 := r.Group("/api/v1")
	{
		items := v1.Group("/items")
		{
			items.POST("", h.Item.Create)
			items.GET("", h.Item.List)
			items.GET("/:id", h.Item.Get)
			items.PUT("/:id", h.Item.Update)
			items.DELETE("/:id", h.Item.Delete)
		}

		stock := v1.Group("/stock")
		{
			stock.POST("/receive", h.Stock.Receive)
			stock.GET("/:item_id", h.Stock.ByItem)
		}

		categories := v1.Group("/categories")
		{
			categories.GET("", h.Item.Categories)
			categories.POST("", h.Item.CreateCategory)
		}

		locations := v1.Group("/locations")
		{
			locations.GET("", h.Stock.Locations)
			locations.POST("", h.Stock.CreateLocation)
		}

		logs := v1.Group("/logs")
		{
			logs.GET("", h.Log.List)
			logs.GET("/stream", h.Log.Stream)
		}

		v1.GET("/error-codes", h.System.ErrorCodes)
	}

	return r
}
