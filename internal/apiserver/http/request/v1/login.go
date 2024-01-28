package v1

type LoginRequest struct {
	Username string `json:"username" binding:"required,alphanum,min=2,max=255"`
	Password string `json:"password" binding:"required,min=6,max=18"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type ChangePasswordRequest struct {
	PasswordOld string `json:"passwordOld" binding:"required,min=6,max=18"`
	PasswordNew string `json:"passwordNew" binding:"required,min=6,max=18"`
}
