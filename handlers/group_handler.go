package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"gully-backend/repositories"
	"gully-backend/services"
)

type GroupHandler struct {
	groupService  *services.GroupService
	playerService *services.PlayerService
	userRepo      *repositories.UserRepo
}

func NewGroupHandler(groupService *services.GroupService, playerService *services.PlayerService, userRepo *repositories.UserRepo) *GroupHandler {
	return &GroupHandler{groupService: groupService, playerService: playerService, userRepo: userRepo}
}

type createGroupRequest struct {
	Name string `json:"name" binding:"required"`
}

func (h *GroupHandler) CreateGroup(c *gin.Context) {
	var req createGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userIDStr, _ := c.Get("user_id")
	userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	group, err := h.groupService.CreateGroup(c.Request.Context(), req.Name, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Auto-create a player record for the creator
	h.autoCreatePlayer(c, userID, group.ID)

	c.JSON(http.StatusCreated, gin.H{"group": group})
}

type joinGroupRequest struct {
	Code string `json:"code" binding:"required"`
}

func (h *GroupHandler) JoinGroup(c *gin.Context) {
	var req joinGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userIDStr, _ := c.Get("user_id")
	userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	group, err := h.groupService.JoinGroup(c.Request.Context(), req.Code, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "group not found"})
		return
	}

	// Auto-create a player record for the joining user
	h.autoCreatePlayer(c, userID, group.ID)

	c.JSON(http.StatusOK, gin.H{"group": group})
}

func (h *GroupHandler) GetGroup(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group id"})
		return
	}

	group, err := h.groupService.GetGroup(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "group not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"group": group})
}

func (h *GroupHandler) GetUserGroups(c *gin.Context) {
	userIDStr, _ := c.Get("user_id")
	userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	groups, err := h.groupService.GetUserGroups(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"groups": groups})
}

// autoCreatePlayer looks up the user's username and creates a player record
// in the group if one with that name doesn't already exist.
func (h *GroupHandler) autoCreatePlayer(c *gin.Context, userID, groupID primitive.ObjectID) {
	user, err := h.userRepo.FindByID(c.Request.Context(), userID)
	if err != nil {
		return // silently skip — user not found
	}
	// CreatePlayerIfNotExists is idempotent
	_, _ = h.playerService.CreatePlayerIfNotExists(c.Request.Context(), user.Username, groupID)
}
