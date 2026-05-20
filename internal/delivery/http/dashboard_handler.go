package http

import (
	"net/http"

	"habit-tracker/internal/middleware"
	"habit-tracker/internal/usecase"

	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	habits *usecase.HabitService
}

func NewDashboardHandler(habits *usecase.HabitService) *DashboardHandler {
	return &DashboardHandler{habits: habits}
}

// Get godoc
// @Summary      Дашборд пользователя (аналитика)
// @Description  XP, уровень, привычки на сегодня, отметки за сегодня, активные streak.
// @Tags         dashboard
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  usecase.Dashboard
// @Failure      401  {object}  errorResponse
// @Router       /dashboard [get]
func (h *DashboardHandler) Get(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
		return
	}

	dash, err := h.habits.Dashboard(c.Request.Context(), userID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dash)
}
