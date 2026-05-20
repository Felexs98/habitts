package http

type registerRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type createHabitRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
}

type updateHabitRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
}

type checkInRequest struct {
	Date string `json:"date"` // YYYY-MM-DD, optional (default: today)
}

type errorResponse struct {
	Error string `json:"error"`
}
