// ABOUTME: JSON-RPC to Biz layer adapter.
// ABOUTME: Routes JSON-RPC methods to typed handlers with automatic serialization.

package jsonrpc

import (
	"context"
	"encoding/json"
	"reflect"

	"bingo/pkg/errorsx"
)

// TODO(方案B): 迁移到 proto.Message 类型约束
// 当前使用 any 类型以兼容现有 biz 层的 Go struct 请求/响应类型。
// 目标是将 biz 层迁移到 proto.Message，届时修改 HandlerFunc 签名为：
// type HandlerFunc func(ctx context.Context, req proto.Message) (proto.Message, error)

// HandlerFunc is the handler function signature.
type HandlerFunc func(ctx context.Context, req any) (any, error)

// Adapter routes JSON-RPC methods to Biz layer handlers.
type Adapter struct {
	handlers map[string]*handlerInfo
}

type handlerInfo struct {
	handler     HandlerFunc
	requestType reflect.Type
}

// NewAdapter creates a new JSON-RPC adapter.
func NewAdapter() *Adapter {
	return &Adapter{
		handlers: make(map[string]*handlerInfo),
	}
}

// RegisterHandler registers a handler for a method.
func (a *Adapter) RegisterHandler(method string, handler HandlerFunc, reqType any) {
	a.handlers[method] = &handlerInfo{
		handler:     handler,
		requestType: reflect.TypeOf(reqType).Elem(),
	}
}

// Handle processes a JSON-RPC request.
func (a *Adapter) Handle(ctx context.Context, req *Request) *Response {
	info, ok := a.handlers[req.Method]
	if !ok {
		return NewErrorResponse(req.ID,
			errorsx.New(404, "MethodNotFound", "Method not found: %s", req.Method))
	}

	// Create new instance of request type
	reqInstance := reflect.New(info.requestType).Interface()

	// Unmarshal params if present
	if len(req.Params) > 0 {
		if err := json.Unmarshal(req.Params, reqInstance); err != nil {
			return NewErrorResponse(req.ID,
				errorsx.New(400, "InvalidParams", "Invalid params: %s", err.Error()))
		}
	}

	// Call handler
	resp, err := info.handler(ctx, reqInstance)
	if err != nil {
		return NewErrorResponse(req.ID, err)
	}

	// Convert response to map for JSON serialization
	data, err := json.Marshal(resp)
	if err != nil {
		return NewErrorResponse(req.ID,
			errorsx.New(500, "InternalError", "Failed to serialize response: %s", err.Error()))
	}

	var result any
	if err := json.Unmarshal(data, &result); err != nil {
		return NewErrorResponse(req.ID,
			errorsx.New(500, "InternalError", "Failed to process response: %s", err.Error()))
	}

	return NewResponse(req.ID, result)
}
