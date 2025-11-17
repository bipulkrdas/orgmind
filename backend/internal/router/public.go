package router

import (
	"github.com/gin-gonic/gin"
)

// setupPublicRoutes configures all public routes that don't require authentication
func (r *Router) setupPublicRoutes(router *gin.Engine) {
	// Create public API group
	public := router.Group("/api/auth")

	// Authentication endpoints
	public.POST("/signup", r.authHandler.SignUp)
	public.POST("/signin", r.authHandler.SignIn)
	public.POST("/reset-password", r.authHandler.ResetPassword)
	public.POST("/update-password", r.authHandler.UpdatePassword)

	// OAuth endpoints
	public.GET("/oauth/:provider", r.authHandler.InitiateOAuth)
	public.GET("/oauth/:provider/callback", r.authHandler.HandleOAuthCallback)
}
