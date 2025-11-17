package service

import (
	"context"
	"fmt"
	"time"

	"github.com/bipulkrdas/orgmind/backend/internal/models"
	v3 "github.com/getzep/zep-go/v3"
	v3client "github.com/getzep/zep-go/v3/client"
	"github.com/getzep/zep-go/v3/option"
)

// zepService implements the ZepService interface
type zepService struct {
	client *v3client.Client
}

// NewZepService creates a new Zep service instance
func NewZepService(apiKey string) (ZepService, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("zep API key is required")
	}

	opts := option.WithAPIKey(apiKey)
	client := v3client.NewClient(opts)

	return &zepService{
		client: client,
	}, nil
}

// CreateGraph creates a new graph in Zep Cloud with retry logic
func (s *zepService) CreateGraph(ctx context.Context, graphID, name string, description *string) (string, error) {
	const maxRetries = 3
	const baseDelay = 1 * time.Second

	var lastErr error

	for attempt := range maxRetries {
		if attempt > 0 {
			// Exponential backoff
			delay := baseDelay * time.Duration(1<<uint(attempt-1))
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(delay):
			}
		}

		zepGraphID, err := s.createGraphAttempt(ctx, graphID, name, description)
		if err == nil {
			return zepGraphID, nil
		}

		lastErr = err
	}

	return "", fmt.Errorf("failed to create graph after %d attempts: %w", maxRetries, lastErr)
}

// createGraphAttempt performs a single attempt to create a graph in Zep
func (s *zepService) createGraphAttempt(ctx context.Context, graphID, name string, description *string) (string, error) {
	request := &v3.CreateGraphRequest{
		GraphID:     graphID,
		Name:        &name,
		Description: description,
	}

	graph, err := s.client.Graph.Create(ctx, request)
	if err != nil {
		return "", fmt.Errorf("failed to create graph in Zep: %w", err)
	}

	// Return the Zep-assigned graph ID
	if graph.GraphID != nil {
		return *graph.GraphID, nil
	}
	return graphID, nil
}

// DeleteGraph deletes a graph from Zep Cloud
func (s *zepService) DeleteGraph(ctx context.Context, zepGraphID string) error {
	_, err := s.client.Graph.Delete(ctx, zepGraphID)
	if err != nil {
		return fmt.Errorf("failed to delete graph from Zep: %w", err)
	}

	return nil
}

// AddMemory adds document chunks to a specific graph in Zep Cloud with retry logic
func (s *zepService) AddMemory(ctx context.Context, graphID string, chunks []string, metadata map[string]any) error {
	const maxRetries = 3
	const baseDelay = 1 * time.Second

	var lastErr error

	for attempt := range maxRetries {
		if attempt > 0 {
			// Exponential backoff
			delay := baseDelay * time.Duration(1<<uint(attempt-1))
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}

		err := s.addMemoryAttempt(ctx, graphID, chunks, metadata)
		if err == nil {
			return nil
		}

		lastErr = err
	}

	return fmt.Errorf("failed to add memory after %d attempts: %w", maxRetries, lastErr)
}

// addMemoryAttempt performs a single attempt to add memory to Zep
func (s *zepService) addMemoryAttempt(ctx context.Context, graphID string, chunks []string, metadata map[string]any) error {
	// Add each chunk as graph data using the Graph API
	// This will automatically build the knowledge graph through Zep's Grafiti
	for i, chunk := range chunks {
		// Create source description from metadata if available
		sourceDesc := fmt.Sprintf("Document chunk %d", i+1)
		if desc, ok := metadata["source_description"].(string); ok {
			sourceDesc = desc
		}

		request := &v3.AddDataRequest{
			GraphID:           v3.String(graphID),
			Data:              chunk,
			Type:              v3.GraphDataTypeText,
			SourceDescription: &sourceDesc,
		}

		_, err := s.client.Graph.Add(ctx, request)
		if err != nil {
			return fmt.Errorf("failed to add chunk %d to graph: %w", i, err)
		}
	}

	return nil
}

// GetGraph retrieves the knowledge graph data for a specific graph from Zep
func (s *zepService) GetGraph(ctx context.Context, graphID, query string) (*models.GraphData, error) {
	// Use default query if empty
	if query == "" {
		query = "all"
	}

	// Step 1: Search for edges using the query
	// The search API returns edges that match the query
	searchQuery := &v3.GraphSearchQuery{
		GraphID: v3.String(graphID),
		Query:   query,
		Limit:   v3.Int(50), // Max limit is 50
	}

	searchResults, err := s.client.Graph.Search(ctx, searchQuery)
	if err != nil {
		// Log the error for debugging
		fmt.Printf("Error searching graph %s with query '%s': %v\n", graphID, query, err)
		// Return empty graph data instead of failing
		return &models.GraphData{
			Nodes: []models.GraphNode{},
			Edges: []models.GraphEdge{},
		}, nil
	}

	// Step 2: Collect unique node IDs from the edges
	nodeIDs := make(map[string]bool)
	if searchResults != nil && searchResults.Edges != nil {
		for _, edge := range searchResults.Edges {
			if edge != nil {
				nodeIDs[edge.SourceNodeUUID] = true
				nodeIDs[edge.TargetNodeUUID] = true
			}
		}
	}

	// Step 3: Fetch all nodes from the graph
	allNodes, err := s.client.Graph.Node.GetByGraphID(ctx, graphID, &v3.GraphNodesRequest{})
	if err != nil {
		fmt.Printf("Error fetching nodes for graph %s: %v\n", graphID, err)
		// Return empty graph data instead of failing
		return &models.GraphData{
			Nodes: []models.GraphNode{},
			Edges: []models.GraphEdge{},
		}, nil
	}

	// Log the results for debugging
	edgeCount := 0
	if searchResults != nil && searchResults.Edges != nil {
		edgeCount = len(searchResults.Edges)
	}
	fmt.Printf("Retrieved %d edges and %d unique node IDs for graph %s with query '%s'\n", edgeCount, len(nodeIDs), graphID, query)
	fmt.Printf("Total nodes in graph: %d\n", len(allNodes))

	// Step 4: Transform to internal format, filtering nodes to only those referenced by edges
	graphData := transformZepGraphToInternal(searchResults, allNodes, nodeIDs)

	return graphData, nil
}

// transformZepGraphToInternal converts Zep's graph format to our internal format
// preserving all metadata from Zep for rich visualization
func transformZepGraphToInternal(searchResults *v3.GraphSearchResults, allNodes []*v3.EntityNode, nodeIDsToInclude map[string]bool) *models.GraphData {
	// Create a map of nodes that are referenced by edges
	nodeMap := make(map[string]*v3.EntityNode)

	// Filter nodes to only include those referenced by edges
	for _, node := range allNodes {
		if node != nil && nodeIDsToInclude[node.UUID] {
			nodeMap[node.UUID] = node
		}
	}

	// Convert nodes to internal format, preserving all Zep metadata
	nodes := make([]models.GraphNode, 0, len(nodeMap))
	for _, zepNode := range nodeMap {
		node := models.GraphNode{
			ID:         zepNode.UUID,
			Name:       zepNode.Name,
			Summary:    zepNode.Summary,
			Labels:     zepNode.Labels,
			Score:      zepNode.Score,
			Relevance:  zepNode.Relevance,
			Attributes: zepNode.Attributes,
			CreatedAt:  zepNode.CreatedAt,
			// Visualization properties
			Size:  calculateNodeSize(zepNode),
			Color: generateNodeColor(zepNode),
		}
		nodes = append(nodes, node)
	}

	// Convert edges to internal format, preserving all Zep metadata
	// Only include edges where both source and target nodes exist in our node map
	edges := make([]models.GraphEdge, 0)
	if searchResults != nil && searchResults.Edges != nil {
		for _, zepEdge := range searchResults.Edges {
			if zepEdge != nil {
				// Only add edge if both nodes exist in our filtered node map
				if _, sourceExists := nodeMap[zepEdge.SourceNodeUUID]; sourceExists {
					if _, targetExists := nodeMap[zepEdge.TargetNodeUUID]; targetExists {
						// Get source and target node names from our node map
						var sourceNodeName, targetNodeName, sourceNodeSummary, targetNodeSummary *string
						if sourceNode, exists := nodeMap[zepEdge.SourceNodeUUID]; exists {
							sourceNodeName = &sourceNode.Name
							sourceNodeSummary = &sourceNode.Summary
						}
						if targetNode, exists := nodeMap[zepEdge.TargetNodeUUID]; exists {
							targetNodeName = &targetNode.Name
							targetNodeSummary = &targetNode.Summary
						}

						edge := models.GraphEdge{
							ID:                zepEdge.UUID,
							Source:            zepEdge.SourceNodeUUID,
							Target:            zepEdge.TargetNodeUUID,
							Name:              zepEdge.Name,
							Fact:              zepEdge.Fact,
							ValidAt:           zepEdge.ValidAt,
							InvalidAt:         zepEdge.InvalidAt,
							Episodes:          zepEdge.Episodes,
							SourceNodeName:    sourceNodeName,
							TargetNodeName:    targetNodeName,
							SourceNodeSummary: sourceNodeSummary,
							TargetNodeSummary: targetNodeSummary,
							Attributes:        zepEdge.Attributes,
							CreatedAt:         zepEdge.CreatedAt,
						}
						edges = append(edges, edge)
					}
				}
			}
		}
	}

	return &models.GraphData{
		Nodes: nodes,
		Edges: edges,
	}
}

// calculateNodeSize determines the size of a node based on its properties
func calculateNodeSize(node *v3.EntityNode) float64 {
	// Base size
	size := 10.0

	// Increase size based on number of labels (more labels = more important)
	if node.Labels != nil {
		size += float64(len(node.Labels)) * 2.0
	}

	// Use score if available
	if node.Score != nil {
		size += *node.Score * 10.0
	}

	return size
}

// generateNodeColor generates a color for a node based on its labels
func generateNodeColor(node *v3.EntityNode) string {
	// Default color
	defaultColor := "#3b82f6" // Blue

	if len(node.Labels) == 0 {
		return defaultColor
	}

	// Generate color based on primary label
	primaryLabel := node.Labels[0]

	// Simple color mapping based on common entity types
	colorMap := map[string]string{
		"Person":       "#ef4444", // Red
		"Organization": "#10b981", // Green
		"Location":     "#f59e0b", // Orange
		"Event":        "#8b5cf6", // Purple
		"Concept":      "#06b6d4", // Cyan
		"Document":     "#ec4899", // Pink
	}

	if color, exists := colorMap[primaryLabel]; exists {
		return color
	}

	// Generate a consistent color based on label hash
	hash := 0
	for _, char := range primaryLabel {
		hash = int(char) + ((hash << 5) - hash)
	}

	// Use hash to generate RGB values
	r := (hash & 0xFF0000) >> 16
	g := (hash & 0x00FF00) >> 8
	b := hash & 0x0000FF

	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

// SearchMemory searches for memory in a specific graph
func (s *zepService) SearchMemory(ctx context.Context, graphID, query string) ([]models.MemoryResult, error) {
	// Search for episodes (memory entries) in the graph
	searchQuery := &v3.GraphSearchQuery{
		GraphID: v3.String(graphID),
		Query:   query,
		Scope:   v3.GraphSearchScopeEpisodes.Ptr(),
		Limit:   v3.Int(10), // Retrieve up to 10 results
	}

	results, err := s.client.Graph.Search(ctx, searchQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to search memory in graph: %w", err)
	}

	// Transform results to MemoryResult format
	memoryResults := make([]models.MemoryResult, 0)
	if results.Episodes != nil {
		for _, episode := range results.Episodes {
			if episode != nil && episode.Content != "" {
				result := models.MemoryResult{
					Content:  episode.Content,
					Metadata: make(map[string]interface{}),
					Score:    0.0,
				}

				// Add score if available
				if episode.Score != nil {
					result.Score = *episode.Score
				}

				// Add metadata if available
				if episode.SourceDescription != nil {
					result.Metadata["source_description"] = *episode.SourceDescription
				}

				memoryResults = append(memoryResults, result)
			}
		}
	}

	return memoryResults, nil
}
