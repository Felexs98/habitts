package http

import (
	"habit-tracker/internal/config"
	"habit-tracker/internal/middleware"
	"habit-tracker/internal/usecase"
	"habit-tracker/pkg/jwtutil"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Handlers struct {
	Auth        *AuthHandler
	Habit       *HabitHandler
	Achievement *AchievementHandler
	Dashboard   *DashboardHandler
}

func NewRouter(h *Handlers, jwtManager *jwtutil.Manager, cfg *config.Config) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.CORS())

	rl := middleware.RateLimitSet{
		API: middleware.RateLimitOptions{
			Enabled:           cfg.RateLimitEnabled,
			RequestsPerMinute: cfg.RateLimitAPIRPM,
			Burst:             cfg.RateLimitAPIBurst,
		},
		Auth: middleware.RateLimitOptions{
			Enabled:           cfg.RateLimitEnabled,
			RequestsPerMinute: cfg.RateLimitAuthRPM,
			Burst:             cfg.RateLimitAuthBurst,
		},
	}

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Swagger UI (после `swag init` в корне проекта)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := r.Group("/api/v1")
	api.Use(middleware.RateLimit(rl.API))

	auth := api.Group("/auth")
	auth.Use(middleware.RateLimit(rl.Auth))
	{
		auth.POST("/register", h.Auth.Register)
		auth.POST("/login", h.Auth.Login)
		auth.POST("/refresh", h.Auth.Refresh)
		auth.POST("/logout", h.Auth.Logout)
	}

	protected := api.Group("")
	protected.Use(middleware.JWT(jwtManager))
	{
		protected.GET("/me", h.Auth.Me)
		protected.PATCH("/me", h.Auth.UpdateMe)
		protected.GET("/dashboard", h.Dashboard.Get)
		protected.GET("/achievements", h.Achievement.List)
		protected.GET("/habits", h.Habit.List)
		protected.POST("/habits", h.Habit.Create)
		protected.GET("/habits/:id", h.Habit.Get)
		protected.PUT("/habits/:id", h.Habit.Update)
		protected.DELETE("/habits/:id", h.Habit.Delete)
		protected.POST("/habits/:id/restore", h.Habit.Restore)
		protected.GET("/habits/:id/check-ins", h.Habit.ListCheckIns)
		protected.POST("/habits/:id/check-ins", h.Habit.CheckIn)
		protected.GET("/habits/:id/stats", h.Habit.Stats)
	}

	return r
}

func NewHandlers(auth *usecase.AuthService, habits *usecase.HabitService, achievements *usecase.AchievementService) *Handlers {
	return &Handlers{
		Auth:        NewAuthHandler(auth),
		Habit:       NewHabitHandler(habits),
		Achievement: NewAchievementHandler(achievements),
		Dashboard:   NewDashboardHandler(habits),
	}
}
