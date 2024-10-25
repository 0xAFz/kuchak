package api

type LoginUserData struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,password"`
}

type RegisterUserData struct {
	Email          string `json:"email" validate:"required,email"`
	Password       string `json:"password" validate:"required,password"`
	PasswordRepeat string `json:"password_repeat" validate:"required,password"`
}

type UpdatePasswordData struct {
	OldPassword    string `json:"old_password" validate:"required,password"`
	Password       string `json:"password" validate:"required,password"`
	PasswordRepeat string `json:"password_repeat" validate:"required,password"`
}

type ResetPasswordData struct {
	Token          string `json:"token" validate:"required"`
	Password       string `json:"password" validate:"required,password"`
	PasswordRepeat string `json:"password_repeat" validate:"required,password"`
}

type EmailData struct {
	Email string `json:"email" validate:"required,email"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
