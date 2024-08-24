package common

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

var (

	// Pokers 是一个存储扑克牌组合的映射，由它们的配置表示为键，
	// 及其相应的组合结构作为值。组合结构指定类型、分数、
	// 以及每个组合的配置。
	Pokers = make(map[string]*Combination, 16384)

	// TypeToPokers 是一个将扑克类型映射到组合指针数组的映射。
	TypeToPokers = make(map[string][]*Combination, 38)
)

// 组合表示具有特定类型、分数和扑克配置的扑克牌的组合。
//
// “类型”字段指示组合的类型（例如，单个、对、三重奏、炸弹等）。
//
// “分数”字段表示组合的分数或排名。
//
// “Poker”字段表示组合中扑克牌的实际配置。
type Combination struct {
	Type  string
	Score int
	Poker string
}

// init 是一个 Go 函数，在程序启动时自动调用。
// 它初始化rule.json文件，读取其内容，并填充
// Pokers 和 TypeToPokers 与解析后的数据进行映射。
//
// 如果rule.json文件不存在，则调用write函数生成。
// 如果打开或读取文件时发生错误，则会引发恐慌。
//
// 该函数以 1024 字节为单位读取文件并附加内容
// 到 jsonStrByte 字节片，直到读取整个文件。
//
// 然后将 jsonStrByte 字节切片解析并解组到规则映射中。
// 如果在解组过程中发生错误，则打印错误消息。
//
// 最后，该函数迭代规则映射，创建组合结构
// 对于每种扑克类型、分数和配置。这些结构体存储在
// Pokers 映射，它们的引用也存储在 TypeToPokers 映射中。
func init() {
	path := "rule.json"
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		write()
	}
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	var jsonStrByte []byte
	for {
		buf := make([]byte, 1024)
		readNum, err := file.Read(buf)
		if err != nil && err != io.EOF {
			panic(err)
		}
		for i := 0; i < readNum; i++ {
			jsonStrByte = append(jsonStrByte, buf[i])
		}
		if 0 == readNum {
			break
		}
	}
	var rule = make(map[string][]string)
	err = json.Unmarshal(jsonStrByte, &rule)
	if err != nil {
		fmt.Printf("json unmarsha1 err:%v \n", err)
		return
	}
	for pokerType, pokers := range rule {
		for score, poker := range pokers {
			cards := SortStr(poker)
			p := &Combination{
				Type:  pokerType,
				Score: score,
				Poker: cards,
			}
			Pokers[cards] = p
			TypeToPokers[pokerType] = append(TypeToPokers[pokerType], p)
		}
	}
}
