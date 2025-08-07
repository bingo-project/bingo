package common

// JSONResult json 返回结构体
type JSONResult struct {
	Code uint32 `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

// Response 响应数据结构
func Response(code uint32, message string, data any) JSONResult {
	message = GetErrorMessage(code, message)
	jsonMap := grantMap(code, message, data)

	return jsonMap
}

// 按照接口格式生成原数据数组
func grantMap(code uint32, message string, data any) JSONResult {
	jsonMap := JSONResult{
		Code: code,
		Msg:  message,
		Data: data,
	}

	return jsonMap
}
