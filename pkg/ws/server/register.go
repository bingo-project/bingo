package ws

import (
	"encoding/json"
	"log"
	"sync"

	"bingo/pkg/ws/common"
	"bingo/pkg/ws/model"
)

// DisposeFunc 处理函数
type DisposeFunc func(client *Client, seq string, message []byte) (code uint32, msg string, data any)

var (
	handlers        = make(map[string]DisposeFunc)
	handlersRWMutex sync.RWMutex
)

// Register 注册
func Register(key string, value DisposeFunc) {
	handlersRWMutex.Lock()
	defer handlersRWMutex.Unlock()

	handlers[key] = value

	return
}

func getHandlers(key string) (value DisposeFunc, ok bool) {
	handlersRWMutex.RLock()
	defer handlersRWMutex.RUnlock()

	value, ok = handlers[key]

	return
}

// ProcessData 处理数据
func ProcessData(client *Client, message []byte) {
	log.Println("处理数据", client.Addr, string(message))

	defer func() {
		if r := recover(); r != nil {
			log.Println("处理数据 stop", r)
		}
	}()

	request := &model.Request{}
	if err := json.Unmarshal(message, request); err != nil {
		log.Println("处理数据 json Unmarshal", err)

		client.SendMsg([]byte("数据不合法"))

		return
	}

	requestData, err := json.Marshal(request.Data)
	if err != nil {
		log.Println("处理数据 json Marshal", err)
		client.SendMsg([]byte("处理数据失败"))

		return
	}

	seq := request.Seq
	cmd := request.Cmd

	var (
		code uint32
		msg  string
		data any
	)

	// request
	log.Println("acc_request", cmd, client.Addr)

	// 采用 map 注册的方式
	if value, ok := getHandlers(cmd); ok {
		code, msg, data = value(client, seq, requestData)
	} else {
		code = common.RoutingNotExist
		log.Println("处理数据 路由不存在", client.Addr, "cmd", cmd)
	}

	msg = common.GetErrorMessage(code, msg)
	responseHead := model.NewResponseHead(seq, cmd, code, msg, data)
	headByte, err := json.Marshal(responseHead)
	if err != nil {
		log.Println("处理数据 json Marshal", err)

		return
	}

	client.SendMsg(headByte)
	log.Println("acc_response send", client.Addr, client.AppID, client.UserID, "cmd", cmd, "code", code)

	return
}
