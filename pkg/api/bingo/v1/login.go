package v1

type LoginRequest struct {
	Username string `json:"username" valid:"alphanum,required,stringlength(1|255)"`
	Password string `json:"password" valid:"required,stringlength(6|18)"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type ChangePasswordRequest struct {
	PasswordOld string `json:"passwordOld" valid:"required,stringlength(6|18)"`
	PasswordNew string `json:"passwordNew" valid:"required,stringlength(6|18)"`
}
