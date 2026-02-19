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
	groupService *services.GroupService
	hub          *ws.Hub
}

func NewMatchHandler(matchService *services.MatchService, groupService *services.GroupService, hub *ws.Hub) *MatchHandler {
	return &MatchHandler{matchService: matchService, groupService: groupService, hub: hub}
}

// isGroupCreator checks if the current user is the creator of the group that owns the match.
func (h *MatchHandler) isGroupCreator(c *gin.Context, groupID primitive.ObjectID) bool {
	userIDStr, _ := c.Get("user_id")
	userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		return false
	}
	group, err := h.groupService.GetGroup(c.Request.Context(), groupID)
	if err != nil {
		return false
	}
	return group.CreatedBy == userID
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

// DeleteMatch deletes a match (creator-only).
func (h *MatchHandler) DeleteMatch(c *gin.Context) {
	matchID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid match id"})
		return
	}

	// Get match to find its group
	match, err := h.matchService.GetMatch(c.Request.Context(), matchID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "match not found"})
		return
	}

	if !h.isGroupCreator(c, match.GroupID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "only group creator can delete matches"})
		return
	}

	if err := h.matchService.DeleteMatch(c.Request.Context(), matchID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.hub.BroadcastToGroup(match.GroupID.Hex(), gin.H{"type": "match_deleted", "match_id": matchID.Hex()})
	c.JSON(http.StatusOK, gin.H{"message": "match deleted"})
}

type editScoreRequest struct {
	Score1 int `json:"score1"`
	Score2 int `json:"score2"`
}

// EditScore directly sets scores (creator-only).
func (h *MatchHandler) EditScore(c *gin.Context) {
	matchID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid match id"})
		return
	}

	match, err := h.matchService.GetMatch(c.Request.Context(), matchID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "match not found"})
		return
	}

	if !h.isGroupCreator(c, match.GroupID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "only group creator can edit scores"})
		return
	}

	var req editScoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updated, err := h.matchService.EditScore(c.Request.Context(), matchID, req.Score1, req.Score2)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.hub.BroadcastToGroup(updated.GroupID.Hex(), gin.H{"type": "score_update", "match": updated})
	c.JSON(http.StatusOK, gin.H{"match": updated})
}

type addResultRequest struct {
	GroupID   string `json:"group_id" binding:"required"`
	Player1ID string `json:"player1_id" binding:"required"`
	Player2ID string `json:"player2_id" binding:"required"`
	Score1    int    `json:"score1"`
	Score2    int    `json:"score2"`
}

// AddResult creates a finished match directly (past match).
func (h *MatchHandler) AddResult(c *gin.Context) {
	var req addResultRequest
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

	match, err := h.matchService.AddResult(c.Request.Context(), groupID, p1, p2, req.Score1, req.Score2)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"match": match})
}
