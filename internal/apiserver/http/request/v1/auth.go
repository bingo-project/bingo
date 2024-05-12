package v1

type RegisterRequest struct {
	Nickname string `json:"nickname" binding:"alphanum,min=2,max=255" example:"Peter"`
	Username string `json:"username" binding:"required,alphanum,min=2,max=255" example:"peter"`
	Password string `json:"password" binding:"required,min=6,max=18" example:"123123"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required,alphanum,min=2,max=255" example:"peter"`
	Password string `json:"password" binding:"required,min=6,max=18" example:"123123"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type ChangePasswordRequest struct {
	PasswordOld string `json:"passwordOld" binding:"required,min=6,max=18"`
	PasswordNew string `json:"passwordNew" binding:"required,min=6,max=18"`
}
