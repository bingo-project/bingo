package model

// Request 通用请求数据格式
type Request struct {
	Seq  string `json:"seq"`            // 消息的唯一ID
	Cmd  string `json:"cmd"`            // 请求命令字
	Data any    `json:"data,omitempty"` // 数据 json
}

// LoginRequest 登录请求数据
type LoginRequest struct {
	ServiceToken string `json:"serviceToken"` // 用户登录 Token
	AppID        uint32 `json:"appID,omitempty"`
	UserID       string `json:"userID,omitempty"`
}

// HeartBeat 心跳请求数据
type HeartBeat struct {
	UserID string `json:"userID,omitempty"`
}
