package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"gully-backend/services"
)

type PlayerHandler struct {
	playerService *services.PlayerService
	groupService  *services.GroupService
}

func NewPlayerHandler(playerService *services.PlayerService, groupService *services.GroupService) *PlayerHandler {
	return &PlayerHandler{playerService: playerService, groupService: groupService}
}

type createPlayerRequest struct {
	Name string `json:"name" binding:"required"`
}

func (h *PlayerHandler) CreatePlayer(c *gin.Context) {
	groupID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group id"})
		return
	}

	var req createPlayerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	player, err := h.playerService.CreatePlayer(c.Request.Context(), req.Name, groupID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"player": player})
}

func (h *PlayerHandler) GetPlayers(c *gin.Context) {
	groupID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group id"})
		return
	}

	players, err := h.playerService.GetPlayers(c.Request.Context(), groupID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"players": players})
}

// DeletePlayer removes a player from the group (creator-only).
func (h *PlayerHandler) DeletePlayer(c *gin.Context) {
	groupID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group id"})
		return
	}

	playerID, err := primitive.ObjectIDFromHex(c.Param("playerId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid player id"})
		return
	}

	// Check creator
	userIDStr, _ := c.Get("user_id")
	userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	group, err := h.groupService.GetGroup(c.Request.Context(), groupID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "group not found"})
		return
	}

	if group.CreatedBy != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "only group creator can remove players"})
		return
	}

	if err := h.playerService.DeletePlayer(c.Request.Context(), playerID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "player deleted"})
}

// ── Merge Player (creator-only) ──

type mergePlayerRequest struct {
	TargetPlayerID string `json:"target_player_id" binding:"required"`
	SourcePlayerID string `json:"source_player_id" binding:"required"`
}

func (h *PlayerHandler) MergePlayer(c *gin.Context) {
	groupID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group id"})
		return
	}

	// Check creator
	userIDStr, _ := c.Get("user_id")
	userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	group, err := h.groupService.GetGroup(c.Request.Context(), groupID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "group not found"})
		return
	}
	if group.CreatedBy != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "only group creator can merge players"})
		return
	}

	var req mergePlayerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	targetID, err := primitive.ObjectIDFromHex(req.TargetPlayerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid target_player_id"})
		return
	}
	sourceID, err := primitive.ObjectIDFromHex(req.SourcePlayerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid source_player_id"})
		return
	}

	if err := h.playerService.MergePlayer(c.Request.Context(), groupID, targetID, sourceID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "players merged successfully"})
}
