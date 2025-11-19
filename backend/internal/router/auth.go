package router

import (
	"github.com/bipulkrdas/orgmind/backend/internal/middleware"
	"github.com/gin-gonic/gin"
)

// setupAuthenticatedRoutes configures all routes that require JWT authentication
func (r *Router) setupAuthenticatedRoutes(router *gin.Engine) {
	// Create authenticated API group with JWT middleware
	authenticated := router.Group("/api")
	authenticated.Use(middleware.AuthMiddleware(r.config.JWTSecret))

	// Document endpoints
	documents := authenticated.Group("/documents")
	{
		documents.POST("/editor", r.documentHandler.SubmitEditorContent)
		documents.POST("/upload", r.documentHandler.UploadFile)
		documents.GET("", r.documentHandler.ListDocuments)
		documents.GET("/:id", r.documentHandler.GetDocument)
		documents.GET("/:id/content", r.documentHandler.GetDocumentContent)
		documents.PUT("/:id", r.documentHandler.UpdateDocument)
		documents.DELETE("/:id", r.documentHandler.DeleteDocument)
	}

	// Graph management endpoints
	graphs := authenticated.Group("/graphs")
	{
		// Graph CRUD operations
		graphs.POST("", r.graphHandler.CreateGraph)
		graphs.GET("", r.graphHandler.ListGraphs)
		graphs.GET("/:id", r.graphHandler.GetGraph)
		graphs.PUT("/:id", r.graphHandler.UpdateGraph)
		graphs.DELETE("/:id", r.graphHandler.DeleteGraph)

		// Membership management
		graphs.POST("/:id/members", r.graphHandler.AddMember)
		graphs.DELETE("/:id/members/:userId", r.graphHandler.RemoveMember)
		graphs.GET("/:id/members", r.graphHandler.ListMembers)

		// Graph-specific data endpoints
		graphs.GET("/:id/documents", r.graphHandler.ListGraphDocuments)
		graphs.GET("/:id/visualization", r.graphHandler.GetGraphVisualization)

		// Chat endpoints - using :id to match parent graph routes
		chat := graphs.Group("/:id/chat")
		{
			// Thread management
			chat.POST("/threads", r.chatHandler.CreateThread)
			chat.GET("/threads/:threadId/messages", r.chatHandler.GetThreadMessages)
			chat.POST("/threads/:threadId/messages", r.chatHandler.SendMessage)

			// SSE streaming endpoint
			chat.GET("/stream", r.chatHandler.StreamResponse)
		}
	}
}
