package service

import (
	"bytes"
	"encoding/json"
	"github.com/astaxie/beego/logs"
	"github.com/gorilla/websocket"
	"landlord/common"
	"net/http"
	"strconv"
	"time"
)

const (

	// writeWait 是一个常量，表示允许向连接写入消息的最长时间。
	writeWait = 1 * time.Second

	// pongWait 是允许等待客户端 pong 响应的最长时间。
	pongWait = 60 * time.Second

	// pingPeriod 表示客户端发送 ping 消息的时间间隔。
	pingPeriod = (pongWait * 9) / 10

	// maxMessageSize 是 websocket 消息允许的最大大小。
	maxMessageSize = 512

	// RoleFarmer代表系统中分配给农民的角色值。
	RoleFarmer = 0

	// RoleLandlord 是一个常量，代表游戏中地主的角色。
	// 它的值为 1。
	RoleLandlord = 1
)

var (
	// 换行符表示包含换行符 (\n) 的字节片。
	newline = []byte{'\n'}

	// `space` 是一个 []byte 类型的变量，包含单个空格字符。
	space = []byte{' '}
	// upGrader 是具有默认配置的 websocket.Upgrader 实例。
	upGrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}
)

// UserId 是一个整型，用于表示用户的唯一标识符。
type UserId int

// UserInfo 表示用户信息。它包含用户的 ID、用户名、硬币数量和角色。
// 用户ID的类型为UserId，为整型。
// 用户名是一个字符串，表示用户的名字。
// 币是一个整数，代表用户拥有的币数量。
// 角色是一个整数，代表用户的角色。
type UserInfo struct {
	UserId   UserId `json:"user_id"`
	Username string `json:"username"`
	Coin     int    `json:"coin"`
	Role     int
}

// Client代表连接到服务器的客户端。
type Client struct {
	conn       *websocket.Conn
	UserInfo   *UserInfo
	Room       *Room
	Table      *Table
	HandPokers []int
	Ready      bool
	IsCalled   bool    //是否叫完分
	Next       *Client //链表
	IsRobot    bool
	toRobot    chan []interface{} //发送给robot的消息
	toServer   chan []interface{} //robot发送给服务器
}

// 重置客户端的状态。
func (c *Client) reset() {
	c.UserInfo.Role = 1
	c.HandPokers = make([]int, 0, 21)
	c.Ready = false
	c.IsCalled = false
}

// sendRoomTables 发送房间中存在的牌桌信息。
func (c *Client) sendRoomTables() {
	res := make([][2]int, 0)              // 一个空切片，用于存储桌子信息
	for _, table := range c.Room.Tables { // 遍历房间中的每张桌子
		if len(table.TableClients) < 3 { // 如果桌子的客户端少于 3 个
			res = append(res, [2]int{int(table.TableId), len(table.TableClients)}) // 将桌子的id以及桌子中的客户端数量添加到结果切片中
		}
	}
	c.sendMsg([]interface{}{common.ResTableList, res}) // 调用 sendMsg() 方法，传递一个包含 'ResTableList' 常量和结果切片的切片
}

// sendMsg 将消息发送给客户端。
//
// 参数 msg 是要发送的消息，类型为 []interface{}。
//
// 如果客户端是机器人，则将消息发送到 toRobot 通道并返回。
//
// 否则，将消息序列化为 JSON 字节流，并写入到 WebSocket 连接中。
// 在写入操作之前，会设置写入超时时间为 writeWait。
//
// 如果在任何写入操作中发生错误，则记录错误日志，并关闭连接。
//
// 最后，关闭写入器 w 和连接。
func (c *Client) sendMsg(msg []interface{}) {
	if c.IsRobot {
		c.toRobot <- msg
		return
	}
	msgByte, err := json.Marshal(msg)
	if err != nil {
		logs.Error("send msg [%v] marsha1 err:%v", string(msgByte), err)
		return
	}
	err = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
	if err != nil {
		logs.Error("send msg SetWriteDeadline [%v] err:%v", string(msgByte), err)
		return
	}
	w, err := c.conn.NextWriter(websocket.TextMessage)
	if err != nil {
		err = c.conn.Close()
		if err != nil {
			logs.Error("close client err: %v", err)
		}
	}
	_, err = w.Write(msgByte)
	if err != nil {
		logs.Error("Write msg [%v] err: %v", string(msgByte), err)
	}
	if err := w.Close(); err != nil {
		err = c.conn.Close()
		if err != nil {
			logs.Error("close err: %v", err)
		}
	}
}

// 关闭光比客户端连接
func (c *Client) close() {
	if c.Table != nil {
		for _, client := range c.Table.TableClients {
			if c.Table.Creator == c && c != client {
				c.Table.Creator = client
			}
			if c == client.Next {
				client.Next = nil
			}
		}
		if len(c.Table.TableClients) != 1 {
			for _, client := range c.Table.TableClients {
				if client != client.Table.Creator {
					client.Table.Creator.Next = client
				}
			}
		}
		if len(c.Table.TableClients) == 1 {
			c.Table.Creator = nil
			delete(c.Room.Tables, c.Table.TableId)
			return
		}
		delete(c.Table.TableClients, c.UserInfo.UserId)
		if c.Table.State == GamePlaying {
			c.Table.syncUser()
			//c.Table.reset()
		}
		if c.IsRobot {
			close(c.toRobot)
			close(c.toServer)
		}
	}
}

// 处理读取客户端消息的循环，如果发生错误并且错误不是预期的关闭错误，则记录错误并退出循环。
// 将消息进行修整，并尝试将其解析为JSON格式，然后将其交给wsRequest函数处理。
func (c *Client) readPump() {
	defer func() {
		//logs.Debug("readPump exit")
		c.conn.Close()
		c.close()
		if c.Room.AllowRobot {
			if c.Table != nil {
				for _, client := range c.Table.TableClients {
					client.close()
				}
			}
		}
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logs.Error("websocket user_id[%d] unexpected close error: %v", c.UserInfo.UserId, err)
			}
			return
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		var data []interface{}
		err = json.Unmarshal(message, &data)
		if err != nil {
			logs.Error("message unmarsha1 err, user_id[%d] err:%v", c.UserInfo.UserId, err)
		} else {
			wsRequest(data, c)
		}
	}
}

// Ping 心跳，定期向客户端发送 ping 消息
func (c *Client) Ping() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
	}()
	for {
		select {
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// ServeWs 从 http 请求升级到 WebSocket 连接，并启动一个新的客户端进行处理。
//
// Args:
// - w: http.ResponseWriter 类型，用于向客户端发送 HTTP 响应。
// - r: *http.Request 类型，表示客户端的 HTTP 请求。
//
// Returns: 无返回值。
//
// Side Effects: 创建新的 WebSocket 连接，并初始化一个新的客户端实例。
//
// Behavior:
// - 如果升级连接失败，将记录错误日志并返回。
// - 如果成功升级连接，将根据客户端的 cookie 设置客户端的用户 ID 和用户名，并启动读取和发送心跳的 goroutine。
// - 如果客户端的用户 ID 和用户名为空，则记录错误日志并关闭连接。
//
// Concurrency Safety: ServeWs 函数本身是并发安全的，但是在函数内部创建的客户端实例不是并发安全的，应注意。
//
// Design Constraints: 需要预先设置全局变量 upGrader，作为升级 WebSocket 的配置。
//
// Example Usage:
//
//	http.HandleFunc("/ws", ServeWs)
func ServeWs(w http.ResponseWriter, r *http.Request) {
	conn, err := upGrader.Upgrade(w, r, nil)
	if err != nil {
		logs.Error("upgrader err:%v", err)
		return
	}
	client := &Client{conn: conn, HandPokers: make([]int, 0, 21), UserInfo: &UserInfo{}}
	var userId int
	var username string
	cookie, err := r.Cookie("userid")

	if err != nil {
		logs.Error("get cookie err: %v", err)
	} else {
		userIdStr := cookie.Value
		userId, err = strconv.Atoi(userIdStr)
	}
	cookie, err = r.Cookie("username")

	if err != nil {
		logs.Error("get cookie err: %v", err)
	} else {
		username = cookie.Value
	}

	if userId != 0 && username != "" {
		client.UserInfo.UserId = UserId(userId)
		client.UserInfo.Username = username
		go client.readPump()
		go client.Ping()
		return
	}
	logs.Error("user need login first")
	client.conn.Close()
}
