package http

import (
	"habit-tracker/internal/middleware"
	"habit-tracker/internal/usecase"
	"habit-tracker/pkg/jwtutil"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Handlers struct {
	Auth  *AuthHandler
	Habit *HabitHandler
}

func NewRouter(h *Handlers, jwtManager *jwtutil.Manager) *gin.Engine {
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Swagger UI (после `swag init` в корне проекта)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := r.Group("/api/v1")

	auth := api.Group("/auth")
	{
		auth.POST("/register", h.Auth.Register)
		auth.POST("/login", h.Auth.Login)
	}

	protected := api.Group("")
	protected.Use(middleware.JWT(jwtManager))
	{
		protected.GET("/me", h.Auth.Me)
		protected.GET("/habits", h.Habit.List)
		protected.POST("/habits", h.Habit.Create)
		protected.GET("/habits/:id", h.Habit.Get)
		protected.PUT("/habits/:id", h.Habit.Update)
		protected.DELETE("/habits/:id", h.Habit.Delete)
		protected.POST("/habits/:id/check-ins", h.Habit.CheckIn)
		protected.GET("/habits/:id/stats", h.Habit.Stats)
	}

	return r
}

func NewHandlers(auth *usecase.AuthService, habits *usecase.HabitService) *Handlers {
	return &Handlers{
		Auth:  NewAuthHandler(auth),
		Habit: NewHabitHandler(habits),
	}
}
