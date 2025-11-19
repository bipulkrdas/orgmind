package models

import "time"

// GeminiFileSearchStore represents a Gemini File Search store in the system
type GeminiFileSearchStore struct {
	ID        string    `json:"id" db:"id"`
	StoreName string    `json:"storeName" db:"store_name"`
	StoreID   string    `json:"storeId" db:"store_id"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
}
