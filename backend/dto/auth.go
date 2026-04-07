package dto

type AuthLoginRequestDTO struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type AuthUpdateProfileRequestDTO struct {
	Username string `json:"username" binding:"required"`
	RealName string `json:"real_name"`
}

type AuthChangePasswordRequestDTO struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}
