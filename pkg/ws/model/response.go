package model

import "encoding/json"

// Head 响应数据头
type Head struct {
	Seq      string    `json:"seq"`      // 消息的ID
	Cmd      string    `json:"cmd"`      // 消息的cmd 动作
	Response *Response `json:"response"` // 消息体
}

// Response 响应数据体
type Response struct {
	// Code defines the business error code.
	Code uint32 `json:"code"`

	// Message contains the detail of this message.
	// This message is suitable to be exposed to external
	Message string `json:"message"`

	Data any `json:"data"`
}

type PushMessage struct {
	Seq     string `json:"seq"`
	UUID    uint64 `json:"uuid"`
	Type    string `json:"type"`
	Message string `json:"message"`
}

// NewResponseHead 设置返回消息
func NewResponseHead(seq string, cmd string, code uint32, message string, data any) *Head {
	response := &Response{Code: code, Message: message, Data: data}

	return &Head{Seq: seq, Cmd: cmd, Response: response}
}

// String to string
func (h *Head) String() (headStr string) {
	headBytes, _ := json.Marshal(h)
	headStr = string(headBytes)

	return
}
