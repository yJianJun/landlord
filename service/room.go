package service

import (
	"github.com/astaxie/beego/logs"
	"sync"
)

// roomManager 是 RoomManager 的一个实例，用于管理多个房间及其桌子。
// 它初始化为两个房间，每个房间都有一个唯一的 RoomId。每个房间都有自己的一套桌子。
// `AllowRobot` 指定是否允许机器人进入房间。
// `EntranceFee` 指定玩家进入房间需要支付的费用。
// `Tables` 是 TableId 到 Table 实例的映射，代表房间中的桌子。
// 每次创建新表时，`TableId` 都会递增。
// 应使用适当的锁定来访问 roomManager 实例，以确保线程安全。
var (
	roomManager = RoomManager{
		Rooms: map[int]*Room{
			1: {
				RoomId:      1,
				AllowRobot:  true,
				EntranceFee: 200,
				Tables:      make(map[TableId]*Table),
			},
			2: {
				RoomId:      2,
				AllowRobot:  false,
				EntranceFee: 200,
				Tables:      make(map[TableId]*Table),
			},
		},
	}
)

// RoomId 代表房间的唯一标识符。它的类型是“int”。
type RoomId int

// RoomManager维护了一个互斥锁（Lock），用于保护房间和牌桌的访问。
// Rooms是一个map，将房间ID映射到对应的房间实例。
// TableIdInc是一个递增的计数器，用于生成唯一的牌桌ID。
type RoomManager struct {
	Lock       sync.RWMutex
	Rooms      map[int]*Room
	TableIdInc TableId
}

// Room 表示游戏中的房间，包含以下字段：
// - RoomId: 房间的唯一标识符，类型为 RoomId。
// - Lock: 读写锁，用于保护对房间的并发访问。
// - AllowRobot: 是否允许加入机器人，类型为 bool。
// - Tables: 存储该房间中的牌桌，类型为 map[TableId]*Table。
// - EntranceFee: 加入房间需要支付的入场费，类型为 int。
type Room struct {
	RoomId      RoomId
	Lock        sync.RWMutex
	AllowRobot  bool
	Tables      map[TableId]*Table
	EntranceFee int
}

// newTable 在房间中创建一张新桌子。
func (r *Room) newTable(client *Client) (table *Table) {
	roomManager.Lock.Lock()
	defer roomManager.Lock.Unlock()

	r.Lock.Lock()
	defer r.Lock.Unlock()
	roomManager.TableIdInc = roomManager.TableIdInc + 1
	table = &Table{
		TableId:      roomManager.TableIdInc,
		Creator:      client,
		TableClients: make(map[UserId]*Client, 3),
		GameManage: &GameManage{
			FirstCallScore: client,
			Multiple:       1,
			LastShotPoker:  make([]int, 0),
			Pokers:         make([]int, 0, 54),
		},
	}
	r.Tables[table.TableId] = table
	logs.Debug("create new table ok! allow robot :%v", r.AllowRobot)
	return
}

//func init()  {
//	go func() {		//压测
//		time.Sleep(time.Second * 3)
//		for i:=0;i<1;i++{
//			client := &Client{
//				Room:       roomManager.Rooms[1],
//				HandPokers: make([]int, 0, 21),
//				UserInfo: &UserInfo{
//					UserId:   UserId(rand.Intn(10000)),
//					Username: "ROBOT-0",
//					Coin:     10000,
//				},
//				IsRobot:  true,
//				toRobot: make(chan []interface{}, 3),
//				toServer: make(chan []interface{}, 3),
//			}
//			go client.runRobot()
//			table := client.Room.newTable(client)
//			table.joinTable(client)
//		}
//	}()
//}
