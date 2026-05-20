package http

import (
	"net/http"
	"strconv"
	"time"

	"habit-tracker/internal/domain"
	"habit-tracker/internal/middleware"
	"habit-tracker/internal/usecase"

	"github.com/gin-gonic/gin"
)

type HabitHandler struct {
	habits *usecase.HabitService
}

func NewHabitHandler(habits *usecase.HabitService) *HabitHandler {
	return &HabitHandler{habits: habits}
}

// Ссылки на типы для swag (комментарии @Success) и компилятора.
var (
	_ = domain.Habit{}
	_ = domain.CheckIn{}
	_ = usecase.CheckInResult{}
)

// List godoc
// @Summary      Список привычек пользователя
// @Description  Без параметров — все привычки. `today=true` — только те, что нужно выполнить сегодня и ещё не отмечены.
// @Tags         habits
// @Security     BearerAuth
// @Produce      json
// @Param        today  query     bool  false  "Фильтр «на сегодня»"
// @Success      200    {array}   domain.Habit
// @Failure      401    {object}  errorResponse
// @Router       /habits [get]
func (h *HabitHandler) List(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
		return
	}

	today := c.Query("today") == "true"
	list, err := h.habits.List(c.Request.Context(), userID, usecase.ListHabitsInput{Today: today})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, list)
}

// Create godoc
// @Summary      Создать привычку
// @Tags         habits
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body      createHabitRequest  true  "Данные привычки"
// @Success      201   {object}  domain.Habit
// @Failure      400   {object}  errorResponse
// @Router       /habits [post]
func (h *HabitHandler) Create(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
		return
	}

	var req createHabitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	habit, err := h.habits.Create(c.Request.Context(), userID, usecase.CreateHabitInput{
		Title:       req.Title,
		Description: req.Description,
		Frequency:   req.Frequency,
		Weekday:     req.Weekday,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, habit)
}

// Get godoc
// @Summary      Получить привычку по ID
// @Tags         habits
// @Security     BearerAuth
// @Produce      json
// @Param        id   path      int  true  "Habit ID"
// @Success      200  {object}  domain.Habit
// @Failure      404  {object}  errorResponse
// @Router       /habits/{id} [get]
func (h *HabitHandler) Get(c *gin.Context) {
	userID, habitID, ok := parseHabitID(c)
	if !ok {
		return
	}

	habit, err := h.habits.Get(c.Request.Context(), userID, habitID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, habit)
}

// Update godoc
// @Summary      Обновить привычку
// @Tags         habits
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id    path      int                 true  "Habit ID"
// @Param        body  body      updateHabitRequest  true  "Данные"
// @Success      200   {object}  domain.Habit
// @Router       /habits/{id} [put]
func (h *HabitHandler) Update(c *gin.Context) {
	userID, habitID, ok := parseHabitID(c)
	if !ok {
		return
	}

	var req updateHabitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	habit, err := h.habits.Update(c.Request.Context(), userID, habitID, usecase.UpdateHabitInput{
		Title:       req.Title,
		Description: req.Description,
		Frequency:   req.Frequency,
		Weekday:     req.Weekday,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, habit)
}

// Delete godoc
// @Summary      Удалить привычку
// @Tags         habits
// @Security     BearerAuth
// @Param        id  path  int  true  "Habit ID"
// @Success      204
// @Router       /habits/{id} [delete]
func (h *HabitHandler) Delete(c *gin.Context) {
	userID, habitID, ok := parseHabitID(c)
	if !ok {
		return
	}

	if err := h.habits.Delete(c.Request.Context(), userID, habitID); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// CheckIn godoc
// @Summary      Отметить выполнение привычки
// @Tags         habits
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id    path      int              true  "Habit ID"
// @Param        body  body      checkInRequest   false  "Дата (опционально)"
// @Success      201   {object}  usecase.CheckInResult
// @Router       /habits/{id}/check-ins [post]
func (h *HabitHandler) CheckIn(c *gin.Context) {
	userID, habitID, ok := parseHabitID(c)
	if !ok {
		return
	}

	var req checkInRequest
	_ = c.ShouldBindJSON(&req)

	date := time.Now().UTC()
	if req.Date != "" {
		parsed, err := time.Parse("2006-01-02", req.Date)
		if err != nil {
			c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid date format, use YYYY-MM-DD"})
			return
		}
		date = parsed
	}

	result, err := h.habits.CheckIn(c.Request.Context(), userID, habitID, date)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, result)
}

// Stats godoc
// @Summary      Статистика привычки (streak, всего отметок)
// @Tags         habits
// @Security     BearerAuth
// @Produce      json
// @Param        id  path  int  true  "Habit ID"
// @Success      200  {object}  usecase.HabitStats
// @Router       /habits/{id}/stats [get]
func (h *HabitHandler) Stats(c *gin.Context) {
	userID, habitID, ok := parseHabitID(c)
	if !ok {
		return
	}

	stats, err := h.habits.Stats(c.Request.Context(), userID, habitID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, stats)
}

func parseHabitID(c *gin.Context) (userID, habitID uint, ok bool) {
	uid, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
		return 0, 0, false
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid habit id"})
		return 0, 0, false
	}

	return uid, uint(id), true
}
