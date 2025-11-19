package models

import "time"

// Graph represents a knowledge graph entity stored in Zep Cloud
type Graph struct {
	ID            string    `json:"id" db:"id"`
	CreatorID     string    `json:"creatorId" db:"creator_id"`
	ZepGraphID    string    `json:"zepGraphId" db:"zep_graph_id"`
	Name          string    `json:"name" db:"name"`
	Description   *string   `json:"description" db:"description"`
	DocumentCount int       `json:"documentCount" db:"document_count"`
	GeminiStoreID *string   `json:"geminiStoreId,omitempty" db:"gemini_store_id"`
	CreatedAt     time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt     time.Time `json:"updatedAt" db:"updated_at"`
}

// GraphMembership represents a many-to-many relationship between users and graphs
type GraphMembership struct {
	ID        string    `json:"id" db:"id"`
	GraphID   string    `json:"graphId" db:"graph_id"`
	UserID    string    `json:"userId" db:"user_id"`
	Role      string    `json:"role" db:"role"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
}

// CreateGraphRequest represents the request body for creating a new graph
type CreateGraphRequest struct {
	Name        string  `json:"name" binding:"required,min=1,max=255"`
	Description *string `json:"description" binding:"omitempty,max=1000"`
}

// UpdateGraphRequest represents the request body for updating a graph
type UpdateGraphRequest struct {
	Name        *string `json:"name" binding:"omitempty,min=1,max=255"`
	Description *string `json:"description" binding:"omitempty,max=1000"`
}

// AddMemberRequest represents the request body for adding a member to a graph
type AddMemberRequest struct {
	UserID string `json:"userId" binding:"required"`
	Role   string `json:"role" binding:"omitempty,oneof=member editor viewer"`
}

// GraphData represents the knowledge graph visualization data
type GraphData struct {
	Nodes []GraphNode `json:"nodes"`
	Edges []GraphEdge `json:"edges"`
}

// GraphNode represents a node in the knowledge graph with full Zep metadata
type GraphNode struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Summary    string                 `json:"summary"`
	Labels     []string               `json:"labels,omitempty"`
	Score      *float64               `json:"score,omitempty"`
	Relevance  *float64               `json:"relevance,omitempty"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
	CreatedAt  string                 `json:"createdAt,omitempty"`
	// Visualization properties
	Size  float64 `json:"size"`
	Color string  `json:"color"`
}

// GraphEdge represents an edge in the knowledge graph with full Zep metadata
type GraphEdge struct {
	ID                string                 `json:"id"`
	Source            string                 `json:"source"`
	Target            string                 `json:"target"`
	Name              string                 `json:"name"`
	Fact              string                 `json:"fact"`
	ValidAt           *string                `json:"validAt,omitempty"`
	InvalidAt         *string                `json:"invalidAt,omitempty"`
	Episodes          []string               `json:"episodes,omitempty"`
	SourceNodeName    *string                `json:"sourceNodeName,omitempty"`
	TargetNodeName    *string                `json:"targetNodeName,omitempty"`
	SourceNodeSummary *string                `json:"sourceNodeSummary,omitempty"`
	TargetNodeSummary *string                `json:"targetNodeSummary,omitempty"`
	Attributes        map[string]interface{} `json:"attributes,omitempty"`
	CreatedAt         string                 `json:"createdAt,omitempty"`
}

// MemoryResult represents a search result from Zep memory
type MemoryResult struct {
	Content  string                 `json:"content"`
	Metadata map[string]interface{} `json:"metadata"`
	Score    float64                `json:"score"`
}
