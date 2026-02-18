package main

import (
	"context"
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"gully-backend/config"
	"gully-backend/handlers"
	"gully-backend/repositories"
	"gully-backend/routes"
	"gully-backend/services"
	ws "gully-backend/websocket"
)

func main() {
	// 1. Load config
	cfg := config.Load()

	// 2. Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		log.Fatalf("MongoDB connect error: %v", err)
	}
	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			log.Printf("MongoDB disconnect error: %v", err)
		}
	}()

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("MongoDB ping error: %v", err)
	}
	log.Println("Connected to MongoDB")

	db := client.Database("gullybadminton")

	// 3. Init repos
	userRepo := repositories.NewUserRepo(db)
	groupRepo := repositories.NewGroupRepo(db)
	playerRepo := repositories.NewPlayerRepo(db)
	matchRepo := repositories.NewMatchRepo(db)

	// 4. Init services
	authService := services.NewAuthService(userRepo, cfg.JWTSecret)
	groupService := services.NewGroupService(groupRepo)
	playerService := services.NewPlayerService(playerRepo)
	matchService := services.NewMatchService(matchRepo, playerRepo)

	// 5. Init WebSocket hub
	hub := ws.NewHub()

	// 6. Init handlers
	authHandler := handlers.NewAuthHandler(authService)
	groupHandler := handlers.NewGroupHandler(groupService)
	playerHandler := handlers.NewPlayerHandler(playerService)
	matchHandler := handlers.NewMatchHandler(matchService, hub)

	// 7. Setup Gin
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	routes.Setup(r, cfg.JWTSecret, authHandler, groupHandler, playerHandler, matchHandler, hub)

	// 8. Start server
	log.Printf("Server starting on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
