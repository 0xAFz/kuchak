package api

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,password"`
}

type RegisterRequest struct {
	Email          string `json:"email" validate:"required,email"`
	Password       string `json:"password" validate:"required,password"`
	PasswordRepeat string `json:"password_repeat" validate:"required,password"`
}

type PasswordUpdateRequest struct {
	OldPassword    string `json:"old_password" validate:"required,password"`
	Password       string `json:"password" validate:"required,password"`
	PasswordRepeat string `json:"password_repeat" validate:"required,password"`
}

type PasswordResetRequest struct {
	Token          string `json:"token" validate:"required"`
	Password       string `json:"password" validate:"required,password"`
	PasswordRepeat string `json:"password_repeat" validate:"required,password"`
}

type EmailRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type AuthTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type URLRequest struct {
	OriginalURL string `json:"original_url"`
}

type ErrMessage struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

type ResponseOk struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
}
