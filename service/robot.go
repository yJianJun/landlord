package service

import (
	"github.com/astaxie/beego/logs"
	"landlord/common"
	"time"
)

// runRobot 运行玩游戏的机器人逻辑。
// 它监听两个通道：`c.toServer` 和 `c.toRobot`。
// 如果 `c.toServer` 有消息，它会使用 `wsRequest` 将消息发送到服务器。
// 如果 `c.toRobot` 有消息，它会根据协议代码处理消息。
// - 如果代码是 `common.ResDealPoker`，机器人将调用 autoCallScore 函数。
// - 如果代码是 `common.ResCallScore`，机器人会检查是否需要调用分数，并在必要时调用 autoCallScore 函数。
// - 如果代码是 `common.ResShotPoker`，机器人将调用 autoShotPoker 函数。
// - 如果代码是 `common.ResShowPoker`，如果轮到或者没有人轮到并且机器人是地主，机器人将调用 autoShotPoker 函数。
// - 如果代码是 `common.ResGameOver`，机器人将 `c.Ready` 设置为 true。
// 该函数无限循环运行，直到“c.toServer”或“c.toRobot”通道关闭。
func (c *Client) runRobot() {
	for {
		select {
		case msg, ok := <-c.toServer:
			if !ok {
				return
			}
			wsRequest(msg, c)
		case msg, ok := <-c.toRobot:
			if !ok {
				return
			}
			logs.Debug("robot [%v] receive  message %v ", c.UserInfo.Username, msg)
			if len(msg) < 1 {
				logs.Error("send to robot [%v],message err ,%v", c.UserInfo.Username, msg)
				return
			}
			if act, ok := msg[0].(int); ok {
				protocolCode := int(act)
				switch protocolCode {
				case common.ResDealPoker:
					time.Sleep(time.Second)
					c.Table.Lock.RLock()
					if c.Table.GameManage.FirstCallScore == c {
						c.autoCallScore()
					}
					c.Table.Lock.RUnlock()

				case common.ResCallScore:
					if len(msg) < 4 {
						logs.Error("ResCallScore msg err:%v", msg)
						return
					}
					time.Sleep(time.Second)
					c.Table.Lock.RLock()
					if c.Table.GameManage.Turn == c && !c.IsCalled {
						var callEnd bool
						logs.Debug("ResCallScore %t", msg[3])
						if res, ok := msg[3].(bool); ok {
							callEnd = bool(res)
						}
						if !callEnd {
							c.autoCallScore()
						}
					}
					c.Table.Lock.RUnlock()

				case common.ResShotPoker:
					time.Sleep(time.Second)
					c.Table.Lock.RLock()
					if c.Table.GameManage.Turn == c {
						c.autoShotPoker()
					}
					c.Table.Lock.RUnlock()

				case common.ResShowPoker:
					time.Sleep(time.Second)
					//logs.Debug("robot [%v] role [%v] receive message ResShowPoker turn :%v", c.UserInfo.Username, c.UserInfo.Role, c.Table.GameManage.Turn.UserInfo.Username)
					c.Table.Lock.RLock()
					if c.Table.GameManage.Turn == c || (c.Table.GameManage.Turn == nil && c.UserInfo.Role == RoleLandlord) {
						c.autoShotPoker()
					}
					c.Table.Lock.RUnlock()
				case common.ResGameOver:
					c.Ready = true
				}
			}
		}
	}
}

// autoShotPoker 自动出牌
// 该方法判断机器人是否需要出牌，并根据规则选择要出的牌。
// 如果机器人的出牌队列为空或者上一次出牌的玩家是机器人自己，
// 机器人将选择手中的第一张牌进行出牌；否则，机器人将选择
// 手中大于上一次出牌的牌进行出牌。
// 出牌时，将牌转换为 float64 类型的数值，并将它们作为参数请求服务器。
// 该方法可以捕获异常并在异常发生时进行日志记录。
func (c *Client) autoShotPoker() {
	//因为机器人休眠一秒后才出牌，有可能因用户退出而关闭chan
	defer func() {
		err := recover()
		if err != nil {
			logs.Warn("autoShotPoker err : %v", err)
		}
	}()
	logs.Debug("robot [%v] auto-shot poker", c.UserInfo.Username)
	shotPokers := make([]int, 0)
	if len(c.Table.GameManage.LastShotPoker) == 0 || c.Table.GameManage.LastShotClient == c {
		shotPokers = append(shotPokers, c.HandPokers[0])
	} else {
		shotPokers = common.CardsAbove(c.HandPokers, c.Table.GameManage.LastShotPoker)
	}
	float64Pokers := make([]interface{}, 0)
	for _, poker := range shotPokers {
		float64Pokers = append(float64Pokers, float64(poker))
	}
	req := []interface{}{float64(common.ReqShotPoker)}
	req = append(req, float64Pokers)
	logs.Debug("robot [%v] autoShotPoker %v", c.UserInfo.Username, float64Pokers)
	c.toServer <- req
}

// 自动叫分
func (c *Client) autoCallScore() {
	defer func() {
		err := recover()
		if err != nil {
			logs.Warn("autoCallScore err : %v", err)
		}
	}()
	logs.Debug("robot [%v] autoCallScore", c.UserInfo.Username)
	c.toServer <- []interface{}{float64(common.ReqCallScore), float64(3)}
}
