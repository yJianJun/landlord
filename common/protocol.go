package common

const (

	// ReqCheat 表示一个常量值，用于在用户想要作弊时处理 websocket 请求。
	// ReqCheat 的用法示例：
	// 如果 len(数据) < 2 {
	//logs.Error(“用户[%d]请求ReqCheat，但缺少用户id”，client.UserInfo.UserId)
	// 			返回
	// }
	ReqCheat = 1
	ResCheat = 2

	// ReqLogin 是一个常量，值为 11，表示登录请求。它被用来
	// 在wsRequest函数中处理websocket请求。
	// 使用示例：
	// ReqLogin 用于向客户端发送包含登录响应的响应，
	// 包括用户 ID 和用户名。
	ReqLogin = 11

	// ResLogin 是一个常量，表示登录请求的响应代码。
	ResLogin = 12

	// ReqRoomList is a constant representing the request code for getting the list of rooms.
	// Example usage of ReqRoomList:
	//   client.sendMsg([]interface{}{common.ResRoomList})
	ReqRoomList = 13

	// ResRoomList represents the constant value for the response code of room list request.
	ResRoomList = 14

	// ReqTableList constant represents the request for the list of tables in a room.
	ReqTableList = 15

	// ResTableList represents the response constant for sending existing table information within a room.
	ResTableList = 16

	// ReqJoinRoom is a constant representing the request to join a room.
	ReqJoinRoom = 17
	ResJoinRoom = 18

	// ReqJoinTable is a constant representing the request to join a table.
	ReqJoinTable = 19

	// ResJoinTable represents the constant value for the response of joining a table.
	ResJoinTable = 20

	// ReqNewTable represents the constant value for the request to create a new table.
	ReqNewTable = 21

	// ResNewTable is a constant representing the response code for creating a new table.
	// Its value is 22.
	ResNewTable = 22

	// ReqDealPoker represents a constant for the request to deal poker in a game.
	ReqDealPoker = 31

	// ResDealPoker represents the response code for dealing poker in the poker game.
	ResDealPoker = 32

	// ReqCallScore represents the constant value for the request to call a score in a game.
	// It is used to handle the websocket request and determine the turn of the player calling the score.
	// The score should be between 0 and 3, and it will be checked against the maximum call score set in the game.
	// If the score is valid, it will be broadcasted to all players in the table.
	ReqCallScore = 33

	// ResCallScore represents the constant value for the response call score protocol code.
	ResCallScore = 34

	ReqShowPoker = 35

	// ResShowPoker 是响应的常量表示形式，指示轮到显示扑克了。
	//
	// 用法示例：
	// 案例 common.ResShowPoker:
	// 时间.睡眠(时间.秒)
	// //logs.Debug("机器人[%v]角色[%v]接收消息ResShowPoker转:%v", c.UserInfo.Username, c.UserInfo.Role, c.Table.GameManage.Turn.UserInfo.用户名）
	// c.Table.Lock.RLock()
	// 如果 c.Table.GameManage.Turn == c || (c.Table.GameManage.Turn == nil && c.UserInfo.Role == RoleLandlord) {
	// c.autoShotPoker()
	// }
	// c.Table.Lock.RUNlock()
	ResShowPoker = 36

	// ReqShotPoker 是一个常量，表示在游戏中射击扑克的请求。
	// 用作websocket请求的请求码。
	ReqShotPoker = 37
	ResShotPoker = 38

	ReqGameOver = 41

	// ResGameOver 是一个常量，表示游戏结束的响应代码。
	ResGameOver = 42

	// ReqChat 表示请求类型“Chat”的常量值。
	// 它用于处理 websocket 通信中的聊天请求。
	// 用法示例：
	// wsRequest(data []interface{}, 客户端 *Client)
	// 切换请求 {
	// 常见情况.ReqChat:
	// 如果 len(数据) > 1 {
	// 切换数据[1].(type) {
	// 大小写字符串：
	// client.Table.chat(client, data[1].(string))
	// }
	// }
	// }
	ReqChat = 43

	// ResChat 是一个常量，用于表示聊天操作的响应代码。
	ResChat = 44

	// ReqRestart 表示“ReqRestart”请求的常量值。
	// 该常量用于处理 WebSocket 服务器中的重启请求。
	// 如果客户端发送“ReqRestart”请求，服务器将检查表中的所有客户端是否准备好重新启动游戏。
	// 如果所有客户端都准备好，服务器将重置桌子并开始新游戏。
	// 如果表不处于“GameEnd”状态，则忽略该请求。
	ReqRestart = 45

	// ResRestart 表示重新开始游戏的响应常量。
	ResRestart = 46
)
