package http

import (
	"net/http"

	"habit-tracker/internal/middleware"
	"habit-tracker/internal/usecase"

	"github.com/gin-gonic/gin"
)

type AchievementHandler struct {
	achievements *usecase.AchievementService
}

func NewAchievementHandler(achievements *usecase.AchievementService) *AchievementHandler {
	return &AchievementHandler{achievements: achievements}
}

// List godoc
// @Summary      Достижения пользователя
// @Tags         achievements
// @Security     BearerAuth
// @Produce      json
// @Success      200  {array}   usecase.UserAchievementView
// @Failure      401  {object}  errorResponse
// @Router       /achievements [get]
func (h *AchievementHandler) List(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
		return
	}

	list, err := h.achievements.ListForUser(c.Request.Context(), userID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, list)
}
