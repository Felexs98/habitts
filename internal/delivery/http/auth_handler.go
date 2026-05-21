package http

import (
	"errors"
	"net/http"

	"habit-tracker/internal/domain"
	"habit-tracker/internal/middleware"
	"habit-tracker/internal/usecase"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	auth *usecase.AuthService
}

func NewAuthHandler(auth *usecase.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

// Register godoc
// @Summary      Регистрация пользователя
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      registerRequest  true  "Данные регистрации"
// @Success      201   {object}  usecase.AuthResult
// @Failure      400   {object}  errorResponse
// @Failure      409   {object}  errorResponse
// @Router       /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	result, err := h.auth.Register(c.Request.Context(), usecase.RegisterInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusCreated, result)
}

// Login godoc
// @Summary      Вход пользователя
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      loginRequest  true  "Данные входа"
// @Success      200   {object}  usecase.AuthResult
// @Failure      401   {object}  errorResponse
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	result, err := h.auth.Login(c.Request.Context(), usecase.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// Me godoc
// @Summary      Профиль текущего пользователя
// @Tags         auth
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  domain.User
// @Failure      401  {object}  errorResponse
// @Failure      404  {object}  errorResponse
// @Router       /me [get]
func (h *AuthHandler) Me(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
		return
	}

	user, err := h.auth.GetMe(c.Request.Context(), userID)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, user)
}

// Refresh godoc
// @Summary      Обновление access-токена
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      refreshRequest  true  "Refresh token"
// @Success      200   {object}  usecase.AuthResult
// @Failure      401   {object}  errorResponse
// @Router       /auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	result, err := h.auth.Refresh(c.Request.Context(), usecase.RefreshInput{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// UpdateMe godoc
// @Summary      Обновление профиля
// @Tags         auth
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body      updateProfileRequest  true  "Поля для обновления"
// @Success      200   {object}  domain.User
// @Failure      400   {object}  errorResponse
// @Failure      401   {object}  errorResponse
// @Failure      409   {object}  errorResponse
// @Router       /me [patch]
func (h *AuthHandler) UpdateMe(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
		return
	}

	var req updateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	user, err := h.auth.UpdateProfile(c.Request.Context(), userID, usecase.UpdateProfileInput{
		Email:           req.Email,
		DisplayName:     req.DisplayName,
		Password:        req.Password,
		CurrentPassword: req.CurrentPassword,
	})
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, user)
}

func writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidInput):
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
	case errors.Is(err, domain.ErrAlreadyExists):
		c.JSON(http.StatusConflict, errorResponse{Error: err.Error()})
	case errors.Is(err, domain.ErrUnauthorized):
		c.JSON(http.StatusUnauthorized, errorResponse{Error: err.Error()})
	case errors.Is(err, domain.ErrForbidden):
		c.JSON(http.StatusForbidden, errorResponse{Error: err.Error()})
	case errors.Is(err, domain.ErrNotFound):
		c.JSON(http.StatusNotFound, errorResponse{Error: err.Error()})
	case errors.Is(err, domain.ErrDuplicateCheck):
		c.JSON(http.StatusConflict, errorResponse{Error: err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, errorResponse{Error: "internal server error"})
	}
}
