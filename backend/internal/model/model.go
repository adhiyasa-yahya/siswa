// Package model holds the database entities of the SISWA warehouse system.
package model

import "time"

// Category is the master data of item categories.
type Category struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"` // active | inactive
}

const (
	CategoryStatusActive   = "active"
	CategoryStatusInactive = "inactive"
)

// Location is a rack / zone inside the warehouse.
type Location struct {
	ID   int64  `json:"id"`
	Code string `json:"code"`
	Zone string `json:"zone"`
	Type string `json:"type"` // rack | bin | pallet | cold_storage | staging
}

// Item is the master data of a warehouse item.
type Item struct {
	ID         int64      `json:"id"`
	SKU        string     `json:"sku"`
	Name       string     `json:"name"`
	CategoryID int64      `json:"category_id"`
	Unit       string     `json:"unit"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	DeletedAt  *time.Time `json:"deleted_at,omitempty"`

	// Joined, read only fields.
	CategoryName string `json:"category_name,omitempty"`
	TotalQty     int64  `json:"total_qty"`
}

// Stock is the quantity of one item inside one location.
type Stock struct {
	ID         int64     `json:"id"`
	ItemID     int64     `json:"item_id"`
	LocationID int64     `json:"location_id"`
	Qty        int64     `json:"qty"`
	UpdatedAt  time.Time `json:"updated_at"`

	// Joined, read only fields.
	LocationCode string `json:"location_code,omitempty"`
	LocationZone string `json:"location_zone,omitempty"`
	LocationType string `json:"location_type,omitempty"`
}

// RequestLog is one recorded HTTP request lifecycle.
type RequestLog struct {
	ID            string    `json:"id"`
	Method        string    `json:"method"`
	URL           string    `json:"api_url"`
	Payload       string    `json:"payload"`
	ExecutedTime  float64   `json:"executed_time_ms"`
	Status        int       `json:"status"`
	ResponseCode  int       `json:"response_code"`
	ClientIP      string    `json:"client_ip"`
	UserAgent     string    `json:"user_agent"`
	Datetime      time.Time `json:"datetime"`
	ResponseBytes int       `json:"response_bytes"`
}
