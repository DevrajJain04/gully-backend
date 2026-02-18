package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"gully-backend/services"
	ws "gully-backend/websocket"
)

type MatchHandler struct {
	matchService *services.MatchService
	hub          *ws.Hub
}

func NewMatchHandler(matchService *services.MatchService, hub *ws.Hub) *MatchHandler {
	return &MatchHandler{matchService: matchService, hub: hub}
}

type createMatchRequest struct {
	GroupID   string `json:"group_id" binding:"required"`
	Player1ID string `json:"player1_id" binding:"required"`
	Player2ID string `json:"player2_id" binding:"required"`
}

func (h *MatchHandler) CreateMatch(c *gin.Context) {
	var req createMatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	groupID, err := primitive.ObjectIDFromHex(req.GroupID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group_id"})
		return
	}
	p1, err := primitive.ObjectIDFromHex(req.Player1ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid player1_id"})
		return
	}
	p2, err := primitive.ObjectIDFromHex(req.Player2ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid player2_id"})
		return
	}

	match, err := h.matchService.CreateMatch(c.Request.Context(), groupID, p1, p2)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.hub.BroadcastToGroup(req.GroupID, gin.H{"type": "match_created", "match": match})
	c.JSON(http.StatusCreated, gin.H{"match": match})
}

func (h *MatchHandler) GetMatches(c *gin.Context) {
	groupID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group id"})
		return
	}

	matches, err := h.matchService.GetMatches(c.Request.Context(), groupID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"matches": matches})
}

type updateScoreRequest struct {
	Player int `json:"player" binding:"required"` // 1 or 2
}

func (h *MatchHandler) UpdateScore(c *gin.Context) {
	matchID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid match id"})
		return
	}

	var req updateScoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	match, err := h.matchService.UpdateScore(c.Request.Context(), matchID, req.Player)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.hub.BroadcastToGroup(match.GroupID.Hex(), gin.H{"type": "score_update", "match": match})
	c.JSON(http.StatusOK, gin.H{"match": match})
}

func (h *MatchHandler) UndoScore(c *gin.Context) {
	matchID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid match id"})
		return
	}

	match, err := h.matchService.UndoScore(c.Request.Context(), matchID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.hub.BroadcastToGroup(match.GroupID.Hex(), gin.H{"type": "score_update", "match": match})
	c.JSON(http.StatusOK, gin.H{"match": match})
}

func (h *MatchHandler) FinishMatch(c *gin.Context) {
	matchID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid match id"})
		return
	}

	match, err := h.matchService.FinishMatch(c.Request.Context(), matchID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.hub.BroadcastToGroup(match.GroupID.Hex(), gin.H{"type": "match_finished", "match": match})
	c.JSON(http.StatusOK, gin.H{"match": match})
}
