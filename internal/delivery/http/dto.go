package http

type registerRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type logoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type updateProfileRequest struct {
	Email           *string `json:"email"`
	DisplayName     *string `json:"display_name"`
	Password        *string `json:"password"`
	CurrentPassword *string `json:"current_password"`
}

type createHabitRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	Frequency   string `json:"frequency" enums:"daily,weekly" example:"daily"`
	Weekday     *int   `json:"weekday" example:"1"` // 0=Вс … 6=Сб, для weekly
}

type updateHabitRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	Frequency   string `json:"frequency" enums:"daily,weekly"`
	Weekday     *int   `json:"weekday"`
}

type checkInRequest struct {
	Date string `json:"date"` // YYYY-MM-DD, optional (default: today)
}

type errorResponse struct {
	Error string `json:"error"`
}
