package ws

import (
	"fmt"
	"log"
	"sync"
	"time"

	"bingo/pkg/ws/cache"
)

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	// user login handler
	Login chan *Login

	// Registered users
	users map[string]*Client

	// Lock
	clientsLock sync.RWMutex
	userLock    sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte, 1000),
		register:   make(chan *Client, 1000),
		unregister: make(chan *Client, 1000),
		Login:      make(chan *Login, 1000),
		users:      make(map[string]*Client),
	}
}

func GetUserKey(appID uint32, userID string) (key string) {
	key = fmt.Sprintf("%d_%s", appID, userID)

	return
}

func (hub *Hub) InClient(client *Client) (ok bool) {
	hub.clientsLock.RLock()
	defer hub.clientsLock.RUnlock()

	// 连接存在，在添加
	_, ok = hub.clients[client]

	return
}

// GetClients 获取所有客户端
func (hub *Hub) GetClients() (clients map[*Client]bool) {
	clients = make(map[*Client]bool)
	hub.ClientsRange(func(client *Client, value bool) (result bool) {
		clients[client] = value

		return true
	})

	return
}

// ClientsRange 遍历
func (hub *Hub) ClientsRange(f func(client *Client, value bool) (result bool)) {
	hub.clientsLock.RLock()
	defer hub.clientsLock.RUnlock()

	for key, value := range hub.clients {
		result := f(key, value)
		if result == false {
			return
		}
	}

	return
}

// GetClientsLen GetClientsLen
func (hub *Hub) GetClientsLen() (clientsLen int) {
	clientsLen = len(hub.clients)

	return
}

// AddClients 添加客户端
func (hub *Hub) AddClients(client *Client) {
	hub.clientsLock.Lock()
	defer hub.clientsLock.Unlock()

	hub.clients[client] = true
}

// DelClients 删除客户端
func (hub *Hub) DelClients(client *Client) {
	hub.clientsLock.Lock()
	defer hub.clientsLock.Unlock()

	if _, ok := hub.clients[client]; ok {
		delete(hub.clients, client)
	}
}

// GetUserClient 获取用户的连接
func (hub *Hub) GetUserClient(appID uint32, userID string) (client *Client) {
	hub.userLock.RLock()
	defer hub.userLock.RUnlock()

	userKey := GetUserKey(appID, userID)
	if value, ok := hub.users[userKey]; ok {
		client = value
	}

	return
}

// GetUsersLen GetClientsLen
func (hub *Hub) GetUsersLen() (userLen int) {
	userLen = len(hub.users)

	return
}

// AddUsers 添加用户
func (hub *Hub) AddUsers(key string, client *Client) {
	hub.userLock.Lock()
	defer hub.userLock.Unlock()

	hub.users[key] = client
}

// DelUsers 删除用户
func (hub *Hub) DelUsers(client *Client) (result bool) {
	hub.userLock.Lock()
	defer hub.userLock.Unlock()

	key := GetUserKey(client.AppID, client.UserID)
	if value, ok := hub.users[key]; ok {
		// 判断是否为相同的用户
		if value.Addr != client.Addr {
			return
		}

		delete(hub.users, key)

		result = true
	}

	return
}

// GetUserKeys 获取用户的key
func (hub *Hub) GetUserKeys() (userKeys []string) {
	userKeys = make([]string, 0, len(hub.users))
	for key := range hub.users {
		userKeys = append(userKeys, key)
	}

	return
}

// GetUserList 获取用户 list
func (hub *Hub) GetUserList(appID uint32) (userList []string) {
	hub.userLock.RLock()
	defer hub.userLock.RUnlock()

	userList = make([]string, 0)
	for _, v := range hub.users {
		if v.AppID == appID {
			userList = append(userList, v.UserID)
		}
	}

	log.Println("GetUserList len:", len(hub.users))

	return
}

// GetUserClients 获取用户的key
func (hub *Hub) GetUserClients() (clients []*Client) {
	hub.userLock.RLock()
	defer hub.userLock.RUnlock()

	clients = make([]*Client, 0)
	for _, v := range hub.users {
		clients = append(clients, v)
	}

	return
}

// sendAll 向全部成员(除了自己)发送数据
func (hub *Hub) sendAll(message []byte, ignoreClient *Client) {
	clients := hub.GetUserClients()
	for _, conn := range clients {
		if conn != ignoreClient {
			conn.SendMsg(message)
		}
	}
}

// sendAppIDAll 向全部成员(除了自己)发送数据
func (hub *Hub) sendAppIDAll(message []byte, appID uint32, ignoreClient *Client) {
	clients := hub.GetUserClients()
	for _, conn := range clients {
		if conn != ignoreClient && conn.AppID == appID {
			conn.SendMsg(message)
		}
	}
}

// EventRegister 用户建立连接事件
func (hub *Hub) EventRegister(client *Client) {
	hub.AddClients(client)

	log.Println("EventRegister 用户建立连接", client.Addr)
	// client.Send <- []byte("连接成功")
}

// EventLogin 用户登录
func (hub *Hub) EventLogin(login *Login) {
	client := login.Client

	// 连接存在，在添加
	if hub.InClient(client) {
		userKey := login.GetKey()
		hub.AddUsers(userKey, login.Client)
	}

	log.Println("EventLogin 用户登录", client.Addr, login.AppID, login.UserID)

	// orderID := helper.GetOrderIDTime()
	// _, _ = SendUserMessageAll(login.AppID, login.UserID, orderID, model.MessageCmdEnter, "哈喽~")
}

// EventUnregister 用户断开连接
func (hub *Hub) EventUnregister(client *Client) {
	hub.DelClients(client)

	// 删除用户连接
	deleteResult := hub.DelUsers(client)
	if deleteResult == false {
		// 不是当前连接的客户端
		return
	}

	// 清除redis登录数据
	userOnline, err := cache.GetUserOnlineInfo(client.GetKey())
	if err == nil {
		userOnline.LogOut()
		_ = cache.SetUserOnlineInfo(client.GetKey(), userOnline)
	}

	// 关闭 chan
	// close(client.Send)

	log.Println("EventUnregister 用户断开连接", client.Addr, client.AppID, client.UserID)

	// if client.UserID != "" {
	// 	orderID := helper.GetOrderIDTime()
	// 	_, _ = SendUserMessageAll(client.AppID, client.UserID, orderID, model.MessageCmdExit, "用户已经离开~")
	// }
}

func (hub *Hub) Run() {
	for {
		select {

		// 建立连接事件
		case conn := <-hub.register:
			hub.EventRegister(conn)

		// 用户登录
		case l := <-hub.Login:
			hub.EventLogin(l)

		// 断开连接事件
		case conn := <-hub.unregister:
			hub.EventUnregister(conn)

		// 广播事件
		case message := <-hub.broadcast:
			clients := hub.GetClients()
			for conn := range clients {
				select {
				case conn.send <- message:

				default:
					close(conn.send)
				}
			}
		}
	}
}

// GetManagerInfo 获取管理者信息
func GetManagerInfo(isDebug string) (managerInfo map[string]interface{}) {
	managerInfo = make(map[string]interface{})
	managerInfo["clientsLen"] = ClientManager.GetClientsLen()        // 客户端连接数
	managerInfo["usersLen"] = ClientManager.GetUsersLen()            // 登录用户数
	managerInfo["chanRegisterLen"] = len(ClientManager.register)     // 未处理连接事件数
	managerInfo["chanLoginLen"] = len(ClientManager.Login)           // 未处理登录事件数
	managerInfo["chanUnregisterLen"] = len(ClientManager.unregister) // 未处理退出登录事件数
	managerInfo["chanBroadcastLen"] = len(ClientManager.broadcast)   // 未处理广播事件数

	if isDebug == "true" {
		addrList := make([]string, 0)
		ClientManager.ClientsRange(func(client *Client, value bool) (result bool) {
			addrList = append(addrList, client.Addr)
			return true
		})

		users := ClientManager.GetUserKeys()
		managerInfo["clients"] = addrList // 客户端列表
		managerInfo["users"] = users      // 登录用户列表
	}

	return
}

// GetUserClient 获取用户所在的连接
func GetUserClient(appID uint32, userID string) (client *Client) {
	client = ClientManager.GetUserClient(appID, userID)

	return
}

// ClearTimeoutConnections 定时清理超时连接
func ClearTimeoutConnections() {
	currentTime := uint64(time.Now().Unix())
	clients := ClientManager.GetClients()

	for client := range clients {
		if client.IsHeartbeatTimeout(currentTime) {
			log.Println("心跳时间超时 关闭连接", client.Addr, client.UserID, client.LoginTime, client.HeartbeatTime)

			_ = client.conn.Close()
		}
	}
}

// GetUserList 获取全部用户
func GetUserList(appID uint32) (userList []string) {
	log.Println("获取全部用户", appID)

	userList = ClientManager.GetUserList(appID)

	return
}

// AllSendMessages 全员广播
func AllSendMessages(appID uint32, userID string, data string) {
	log.Println("全员广播", appID, userID, data)

	ignoreClient := ClientManager.GetUserClient(appID, userID)
	ClientManager.sendAppIDAll([]byte(data), appID, ignoreClient)
}
