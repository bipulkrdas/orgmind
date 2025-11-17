package router

import (
	"time"

	"github.com/bipulkrdas/orgmind/backend/internal/config"
	"github.com/bipulkrdas/orgmind/backend/internal/handler"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// Router holds all handlers and configuration
type Router struct {
	authHandler     *handler.AuthHandler
	documentHandler *handler.DocumentHandler
	graphHandler    *handler.GraphHandler
	config          *config.Config
}

// NewRouter creates a new router instance with all handlers
func NewRouter(
	authHandler *handler.AuthHandler,
	documentHandler *handler.DocumentHandler,
	graphHandler *handler.GraphHandler,
	config *config.Config,
) *Router {
	return &Router{
		authHandler:     authHandler,
		documentHandler: documentHandler,
		graphHandler:    graphHandler,
		config:          config,
	}
}

// Setup initializes the Gin router with all middleware and routes
func (r *Router) Setup() *gin.Engine {
	// Create Gin router
	router := gin.New()

	// Add recovery middleware to handle panics
	router.Use(gin.Recovery())

	// Add custom logging middleware
	router.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return ""
	}))

	// Configure CORS middleware
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // Configure based on environment in production
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Configure request size limits for file uploads (50MB max)
	router.MaxMultipartMemory = 50 << 20 // 50 MB

	// Add error handling middleware
	router.Use(errorHandler())

	// Health check endpoint (for Cloud Run and monitoring)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"service": "orgmind-backend",
		})
	})

	// Setup route groups
	r.setupPublicRoutes(router)
	r.setupAuthenticatedRoutes(router)

	return router
}

// errorHandler is a middleware that handles errors from handlers
func errorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check if there are any errors
		if len(c.Errors) > 0 {
			// Get the last error
			err := c.Errors.Last()

			// Log the error
			gin.DefaultErrorWriter.Write([]byte(err.Error() + "\n"))

			// If response hasn't been written yet, send error response
			if !c.Writer.Written() {
				c.JSON(c.Writer.Status(), gin.H{
					"code":    "INTERNAL_ERROR",
					"message": "An internal error occurred",
				})
			}
		}
	}
}
