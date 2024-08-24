package service

import (
	"fmt"
	"github.com/astaxie/beego/logs"
	"landlord/common"
	"math/rand"
	"sort"
	"sync"
	"time"
)

// TableId 代表房间中牌桌的唯一标识符。
// 它用于映射 Room 结构中的表。
type TableId int

const (

	// GameWaitting represents the state of the game when the players are waiting to start playing.
	GameWaitting = iota

	// GameCallScore 表示游戏状态“Call Score”的常量值
	GameCallScore

	// GamePlaying 表示玩游戏时的状态。
	GamePlaying

	// GameEnd 是一个常量，表示游戏结束时的状态。
	GameEnd
)

// Table 代表房间中的牌桌，包含牌桌的基本信息和状态。
type Table struct {
	Lock         sync.RWMutex
	TableId      TableId
	State        int
	Creator      *Client
	TableClients map[UserId]*Client
	GameManage   *GameManage
}

// GameManage 表示游戏管理的数据结构。
// 它包含了游戏中的各种状态和变量。
//
// GameManage 有以下字段:
//
// - Turn: 目前轮到的玩家。
// - FirstCallScore: 每局轮转的玩家。
// - MaxCallScore: 最大叫分值。
// - MaxCallScoreTurn: 叫分最高的玩家。
// - LastShotClient: 上一次出牌的玩家。
// - Pokers: 当前玩家手中的扑克牌。
// - LastShotPoker: 上一次出牌的牌。
// - Multiple: 加倍倍数。
type GameManage struct {
	Turn             *Client
	FirstCallScore   *Client //每局轮转
	MaxCallScore     int     //最大叫分
	MaxCallScoreTurn *Client
	LastShotClient   *Client
	Pokers           []int
	LastShotPoker    []int
	Multiple         int //加倍
}

// allCalled 检查是否已调用牌桌中的所有客户端。
func (table *Table) allCalled() bool {
	for _, client := range table.TableClients {
		if !client.IsCalled {
			return false
		}
	}
	return true
}

// gameOver 一局结束，计算玩家的输赢情况并发送消息给每个玩家。
// 接收参数：
//   - table *Table: 牌桌对象
//   - client *Client: 当前玩家对象
//
// 逻辑：
//  1. 计算玩家需要支付的筹码数量：入场费 * 最大叫分 * 当前局的倍数
//  2. 设置牌桌状态为游戏结束状态
//  3. 遍历牌桌的每一个客户端玩家
//     - 初始化一个结果切片并将消息类型(common.ResGameOver)和当前玩家的ID添加到结果切片中
//     - 如果当前玩家是比赛的胜利者，则将该玩家需要支付的筹码数量加倍再减100，并将结果添加到结果切片中
//     - 否则，将该玩家需要支付的筹码数量添加到结果切片中
//     - 遍历其他客户端玩家，将其他客户端玩家的ID和手牌添加到结果切片中
//     - 将结果切片发送给当前玩家
//  4. 记录一条调试日志，表示牌桌已结束游戏
func (table *Table) gameOver(client *Client) {
	coin := table.Creator.Room.EntranceFee * table.GameManage.MaxCallScore * table.GameManage.Multiple
	table.State = GameEnd
	for _, c := range table.TableClients {
		res := []interface{}{common.ResGameOver, client.UserInfo.UserId}
		if client == c {
			res = append(res, coin*2-100)
		} else {
			res = append(res, coin)
		}
		for _, cc := range table.TableClients {
			if cc != c {
				userPokers := make([]int, 0, len(cc.HandPokers)+1)
				userPokers = append(append(userPokers, int(cc.UserInfo.UserId)), cc.HandPokers...)
				res = append(res, userPokers)
			}
		}
		c.sendMsg(res)
	}
	logs.Debug("table[%d] game over", table.TableId)
}

// callEnd 在调用阶段结束后推进游戏状态。
// 它将表状态设置为 GamePlaying 并更新第一次调用得分。
// 如果之前没有最大调用分数，则将创建者设置为最大调用分数回合，并将调用分数设置为1。
// 然后将地主设置为最大呼叫分数回合，并将其角色更改为 RoleLandlord。
// 轮到地主了。
// 地主手牌扑克更新为游戏扑克。
// 最后，它使用 show poker 命令、房东的用户 ID 和游戏的扑克向所有牌桌客户端发送响应。
func (table *Table) callEnd() {
	table.State = GamePlaying
	table.GameManage.FirstCallScore = table.GameManage.FirstCallScore.Next
	if table.GameManage.MaxCallScoreTurn == nil || table.GameManage.MaxCallScore == 0 {
		table.GameManage.MaxCallScoreTurn = table.Creator
		table.GameManage.MaxCallScore = 1
		//return
	}
	landLord := table.GameManage.MaxCallScoreTurn
	landLord.UserInfo.Role = RoleLandlord
	table.GameManage.Turn = landLord
	for _, poker := range table.GameManage.Pokers {
		landLord.HandPokers = append(landLord.HandPokers, poker)
	}
	res := []interface{}{common.ResShowPoker, landLord.UserInfo.UserId, table.GameManage.Pokers}
	for _, c := range table.TableClients {
		c.sendMsg(res)
	}
}

// joinTable 客户端加入牌桌。
// 开始临界区，结束临界区，无论我们如何退出这个函数
// 检查牌桌是否已满，记录错误并返回（如果已满）
// 记录用户请求加入
// 检查用户是否已经在牌桌中，如果用户已在牌桌上，则记录错误并返回
// 将牌桌分配给客户端
// 将客户端标记为就绪
// 将新用户分配给某个现有用户的“下一个”字段
// 将用户添加到牌桌中
// 执行用户同步
// 如果牌桌已满，有 3 名玩家，新用户的“下一个”是牌桌创建者，改变牌桌状态，发牌
// 如果房间允许机器人并且牌桌未满，添加机器人玩家，记录机器人加入成功
func (table *Table) joinTable(c *Client) {
	table.Lock.Lock()                // 开始临界区
	defer table.Lock.Unlock()        // 结束临界区，无论我们如何退出这个函数
	if len(table.TableClients) > 2 { // 检查牌桌是否已满
		logs.Error("Player[%d] JOIN Table[%d] FULL", c.UserInfo.UserId, table.TableId) // 记录错误并返回（如果已满）
		return
	}
	logs.Debug("[%v] user [%v] request join table", c.UserInfo.UserId, c.UserInfo.Username) // 记录用户请求加入
	if _, ok := table.TableClients[c.UserInfo.UserId]; ok {                                 // 检查用户是否已经在牌桌中
		logs.Error("[%v] user [%v] already in this table", c.UserInfo.UserId, c.UserInfo.Username) // 如果用户已在牌桌上，则记录错误并返回
		return
	}
	c.Table = table                             // 将牌桌分配给客户端
	c.Ready = true                              // 将客户端标记为就绪
	for _, client := range table.TableClients { // 将新用户分配给某个现有用户的“下一个”字段
		if client.Next == nil {
			client.Next = c
			break
		}
	}
	table.TableClients[c.UserInfo.UserId] = c // 将用户添加到牌桌中
	table.syncUser()                          // 执行用户同步
	if len(table.TableClients) == 3 {         // 如果牌桌已满，有 3 名玩家
		c.Next = table.Creator      // 新用户的“下一个”是牌桌创建者
		table.State = GameCallScore // 改变牌桌状态
		table.dealPoker()           // 发牌
	} else if c.Room.AllowRobot { // 如果房间允许机器人并且牌桌未满
		go table.addRobot(c.Room)   // 添加机器人玩家
		logs.Debug("robot join ok") // 记录机器人加入成功
	}
}

// addRobot 加入机器人.
// 它创建一个机器人客户端并将其加入到牌桌中，当牌桌中的客户端数量少于3个时。
// 机器人客户端具有以下属性:
// - Room: 牌桌所在的房间.
// - HandPokers: 机器人客户端持有的扑克牌.
// - UserInfo: 机器人客户端的用户信息，包括用户ID、用户名和金币数.
// - IsRobot: 指示客户端是否是机器人.
// - toRobot: 用于接收来自机器人客户端的消息的通道.
// - toServer: 用于向机器人客户端发送消息的通道.
// 如果成功创建并加入机器人客户端，将启动一个 goroutine 来运行机器人客户端的逻辑.
// 使用 `table.joinTable` 方法将机器人客户端加入牌桌.
func (table *Table) addRobot(room *Room) {
	logs.Debug("robot [%v] join table", fmt.Sprintf("ROBOT-%d", len(table.TableClients)))
	if len(table.TableClients) < 3 {
		client := &Client{
			Room:       room,
			HandPokers: make([]int, 0, 21),
			UserInfo: &UserInfo{
				UserId:   table.getRobotID(),
				Username: fmt.Sprintf("ROBOT-%d", len(table.TableClients)),
				Coin:     10000,
			},
			IsRobot:  true,
			toRobot:  make(chan []interface{}, 3),
			toServer: make(chan []interface{}, 3),
		}
		go client.runRobot()
		table.joinTable(client)
	}
}

// getRobotID 生成随机robotID，确保牌桌中不存在相同的客户端ID。
func (table *Table) getRobotID() (robot UserId) {
	time.Sleep(time.Microsecond * 10)
	rand.Seed(time.Now().UnixNano())
	robot = UserId(rand.Intn(10000))
	table.Lock.RLock()
	defer table.Lock.RUnlock()
	if _, ok := table.TableClients[robot]; ok {
		return table.getRobotID()
	}
	return
}

// dealPoker 发牌，首先将所有的扑克牌加入游戏管理的牌组，然后洗牌，再将扑克牌发给玩家，
// 每个玩家发17张牌，最后将玩家的手牌按升序排列，并发送给客户端。
func (table *Table) dealPoker() {
	logs.Debug("deal poker")
	table.GameManage.Pokers = make([]int, 0)
	for i := 0; i < 54; i++ {
		table.GameManage.Pokers = append(table.GameManage.Pokers, i)
	}
	table.ShufflePokers()
	for i := 0; i < 17; i++ {
		for _, client := range table.TableClients {
			client.HandPokers = append(client.HandPokers, table.GameManage.Pokers[len(table.GameManage.Pokers)-1])
			table.GameManage.Pokers = table.GameManage.Pokers[:len(table.GameManage.Pokers)-1]
		}
	}
	response := make([]interface{}, 0, 3)
	response = append(append(append(response, common.ResDealPoker), table.GameManage.FirstCallScore.UserInfo.UserId), nil)
	for _, client := range table.TableClients {
		sort.Ints(client.HandPokers)
		response[len(response)-1] = client.HandPokers
		client.sendMsg(response)
	}
}

// chat 将消息发送给牌桌上的所有客户端
func (table *Table) chat(client *Client, msg string) {
	res := []interface{}{common.ResChat, client.UserInfo.UserId, msg}
	for _, c := range table.TableClients {
		c.sendMsg(res)
	}
}

// reset 重置牌桌状态和游戏管理信息，发送重新开始消息到创建者客户端，并重置
// 所有牌桌客户端，如果牌桌客户端数量为3，则重新发牌。
func (table *Table) reset() {
	table.GameManage = &GameManage{
		FirstCallScore:   table.GameManage.FirstCallScore,
		Turn:             nil,
		MaxCallScore:     0,
		MaxCallScoreTurn: nil,
		LastShotClient:   nil,
		Pokers:           table.GameManage.Pokers[:0],
		LastShotPoker:    table.GameManage.LastShotPoker[:0],
		Multiple:         1,
	}
	table.State = GameCallScore
	if table.Creator != nil {
		table.Creator.sendMsg([]interface{}{common.ResRestart})
	}
	for _, c := range table.TableClients {
		c.reset()
	}
	if len(table.TableClients) == 3 {
		table.dealPoker()
	}
}

// ShufflePokers 实现对牌局中的扑克进行洗牌操作。
func (table *Table) ShufflePokers() {
	logs.Debug("ShufflePokers")
	r := rand.New(rand.NewSource(time.Now().Unix()))
	i := len(table.GameManage.Pokers)
	for i > 0 {
		randIndex := r.Intn(i)
		table.GameManage.Pokers[i-1], table.GameManage.Pokers[randIndex] = table.GameManage.Pokers[randIndex], table.GameManage.Pokers[i-1]
		i--
	}
}

// syncUser 同步用户信息，将牌桌中的用户信息发送给所有客户端。
func (table *Table) syncUser() {
	logs.Debug("sync user")
	response := make([]interface{}, 0, 3)
	response = append(append(response, common.ResJoinTable), table.TableId)
	tableUsers := make([][2]interface{}, 0, 2)
	current := table.Creator
	for i := 0; i < len(table.TableClients); i++ {
		tableUsers = append(tableUsers, [2]interface{}{current.UserInfo.UserId, current.UserInfo.Username})
		current = current.Next
	}
	response = append(response, tableUsers)
	for _, client := range table.TableClients {
		client.sendMsg(response)
	}
}
