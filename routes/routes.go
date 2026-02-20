package routes

import (
	"github.com/gin-gonic/gin"

	"gully-backend/handlers"
	"gully-backend/middleware"
	ws "gully-backend/websocket"
)

func Setup(
	r *gin.Engine,
	jwtSecret string,
	authHandler *handlers.AuthHandler,
	groupHandler *handlers.GroupHandler,
	playerHandler *handlers.PlayerHandler,
	matchHandler *handlers.MatchHandler,
	hub *ws.Hub,
) {
	// Public routes
	auth := r.Group("/api/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
	}

	// WebSocket (no JWT — clients authenticate via group membership)
	r.GET("/ws/group/:groupId", hub.HandleWebSocket)

	// Protected routes
	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware(jwtSecret))
	{
		// User
		api.GET("/user/groups", groupHandler.GetUserGroups)

		// Groups
		api.POST("/groups", groupHandler.CreateGroup)
		api.POST("/groups/join", groupHandler.JoinGroup)
		api.GET("/groups/:id", groupHandler.GetGroup)

		// Players
		api.POST("/groups/:id/players", playerHandler.CreatePlayer)
		api.GET("/groups/:id/players", playerHandler.GetPlayers)
		api.DELETE("/groups/:id/players/:playerId", playerHandler.DeletePlayer)
		api.POST("/groups/:id/players/merge", playerHandler.MergePlayer)

		// Matches
		api.POST("/matches", matchHandler.CreateMatch)
		api.POST("/matches/result", matchHandler.AddResult)
		api.GET("/groups/:id/matches", matchHandler.GetMatches)
		api.POST("/matches/:id/score", matchHandler.UpdateScore)
		api.PUT("/matches/:id/score", matchHandler.EditScore)
		api.POST("/matches/:id/undo", matchHandler.UndoScore)
		api.POST("/matches/:id/finish", matchHandler.FinishMatch)
		api.DELETE("/matches/:id", matchHandler.DeleteMatch)
	}
}
