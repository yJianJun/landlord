package common

import (
	"github.com/astaxie/beego/logs"
	"sort"
)

// SortStr 根据输入字符串 `pokers` 的 Unicode 代码点按升序对它们进行排序。
// 它将输入字符串转换为符文切片，然后使用对符文切片进行排序
// sort.Ints() 函数。最后，它将排序后的符文切片转换回字符串并返回。
// 输出字符串 `sortPokers` 将包含输入字符串 `pokers` 的排序版本。
//
// 请注意，此函数假定输入字符串包含有效的 Unicode 字符。
// 如果输入字符串包含无效的 Unicode 字符，则行为未定义。
//
// 用法示例：
// sortPokers := SortStr("4321") // sortPokers 将包含“1234”
func SortStr(pokers string) (sortPokers string) {
	runeArr := make([]int, 0)
	for _, s := range pokers {
		runeArr = append(runeArr, int(s))
	}
	sort.Ints(runeArr)
	res := make([]byte, 0)
	for _, v := range runeArr {
		res = append(res, byte(v))
	}
	return string(res)
}

// IsContains 检查字符串 `child` 中的每个字符是否在字符串 `parent` 中存在。
// 它使用嵌套循环遍历 `child` 的每个字符，并在 `parent` 中查找相同的字符。
// 如果找到相同的字符，则将其从 `parent` 中删除（通过将其移动到切片末尾并截断切片）。
// 如果 `child` 中的任何字符在 `parent` 中不存在，则函数将返回 false。
// 如果 `child` 中的所有字符都在 `parent` 中存在，则函数将返回 true。
//
// 注意，此函数假定输入字符串包含有效的 Unicode 字符。未定义的是，函数的行为将是未定义的。
//
// 用法示例：
// result := IsContains("abcdef", "bc") // result 将等于 true
func IsContains(parent, child string) (result bool) {
	for _, childCard := range child {
		inHand := false
		for i, parentCard := range parent {
			if childCard == parentCard {
				inHand = true
				tmp := []byte(parent)
				copy(tmp[i:], tmp[i+1:])
				tmp = tmp[:len(tmp)-1]
				parent = string(tmp)
				break
			}
		}
		if !inHand {
			return
		}
	}
	return true
}

// ToPokers 将牌编号转换为扑克牌。
// 输入参数 num 是一个整数切片，表示要转换的牌的编号。
// 函数首先创建一个字符串 totalCards，其中包含 A234567890JQK，这是扑克牌的所有可能值。
// 然后，它创建一个字节切片 res，用于存储转换后的扑克牌。
// 接下来，函数遍历输入的 num 切片，并根据每个编号的不同情况将相应的扑克牌添加到 res 切片中。
// 如果编号等于 52，则将 'W' 添加到 res 切片中。
// 如果编号等于 53，则将 'w' 添加到 res 切片中。
// 否则，根据编号的计算方式，从 totalCards 字符串中获取相应的扑克牌，并将其添加到 res 切片中。
// 最后，函数将 res 切片转换为字符串并返回。
//
// 请注意，此函数假定输入的 num 切片中的编号是有效的。
// 如果 num 切片中包含无效的编号，则函数的行为是未定义的。
//
// 用法示例：
// pokers := ToPokers([]int{1, 2, 3, 4}) // pokers 将包含 "2345"
func ToPokers(num []int) string {
	totalCards := "A234567890JQK"
	res := make([]byte, 0)
	for _, poker := range num {
		if poker == 52 {
			res = append(res, 'W')
		} else if poker == 53 {
			res = append(res, 'w')
		} else {
			res = append(res, totalCards[poker%13])
		}
	}
	return string(res)
}

// ToPoker 将输入的牌字符 `card` 转换为对应的编号切片。
// 如果 `card` 是 'W'，则返回切片 [52]。
// 如果 `card` 是 'w'，则返回切片 [53]。
// 如果 `card` 是 A234567890JQK` 中的一个字符，则返回对应的切片，切片元素依次为该字符在 "A234567890JQK" 中的索引，
// 该字符在索引加 13, 13*2, 13*3 后的索引。
// 如果 `card` 不在上述范围内，则返回切片 [54]。
//
// 注意，该函数假定输入的牌字符是有效的。
//
// 用法示例：
// poker := ToPoker('A') // poker 将包含[0 13 26 39]
func ToPoker(card byte) (poker []int) {
	if card == 'W' {
		return []int{52}
	}
	if card == 'w' {
		return []int{53}
	}
	cards := "A234567890JQK"
	for i, c := range []byte(cards) {
		if c == card {
			return []int{i, i + 13, i + 13*2, i + 13*3}
		}
	}
	return []int{54}
}

// 将机器人要出的牌转换为编号
// 如果 poker 在 pokers 中，isInResPokers 返回 true，否则返回 false。
func pokersInHand(num []int, findPokers string) (pokers []int) {
	var isInResPokers = func(poker int) bool {
		for _, p := range pokers {
			if p == poker {
				return true
			}
		}
		return false
	}

	for _, poker := range findPokers {
		poker := ToPoker(byte(poker))
	out:
		for _, pItem := range poker {
			for _, n := range num {
				if pItem == n && !isInResPokers(n) {
					pokers = append(pokers, pItem)
					break out
				}
			}
		}
	}
	return
}

// pokersValue 根据输入字符串 `pokers` 获取与之关联的扑克牌组合的类型和分数。
// 如果输入字符串 `pokers` 在 Pokers 映射中存在，则返回与之关联的组合类型和分数。
//
// 请注意，此函数假定输入字符串 `pokers` 是有效的。
// 如果输入字符串包含无效的字符，则函数的行为是未定义的。
func pokersValue(pokers string) (cardType string, score int) {
	if combination, ok := Pokers[SortStr(pokers)]; ok {
		cardType = combination.Type
		score = combination.Score
	}
	return
}

// ComparePoker 比较两手牌的大小，并返回是否翻倍。
// 函数接受两个整型切片参数 baseNum 和 comparedNum，分别表示两手牌的牌型。
// 函数首先检查 baseNum 和 comparedNum 的长度，如果其中任一手牌为空，
// 则根据情况返回相应的结果。如果两手牌都为空，则返回 0 和 false。
// 如果 baseNum 为空，而 comparedNum 不为空，则首先判断 comparedNum 的牌型是否是rocket或者bomb，
// 如果是，则返回 1 和 true，否则返回 1 和 false。
// 如果两手牌都不为空，则分别调用 pokersValue 函数计算两手牌的牌型和大小。
// 然后，根据两手牌的牌型进行比较。相同牌型的比较返回比较手牌的大小差值和 false。
// 如果 comparedNum 的牌型是 rocket，则返回 1 和 true。
// 如果 baseNum 的牌型是 rocket，则返回 -1 和 false。
// 如果 comparedNum 的牌型是 bomb，则返回 1 和 true。
// 默认情况下，返回 0 和 false。
func ComparePoker(baseNum, comparedNum []int) (int, bool) {
	logs.Debug("comparedNum %v  %v", baseNum, comparedNum)
	if len(baseNum) == 0 || len(comparedNum) == 0 {
		if len(baseNum) == 0 && len(comparedNum) == 0 {
			return 0, false
		} else {
			if len(baseNum) != 0 {
				return -1, false
			} else {
				comparedType, _ := pokersValue(ToPokers(comparedNum))
				if comparedType == "rocket" || comparedType == "bomb" {
					return 1, true
				}
				return 1, false
			}
		}
	}
	baseType, baseScore := pokersValue(ToPokers(baseNum))
	comparedType, comparedScore := pokersValue(ToPokers(comparedNum))
	logs.Debug("compare poker %v, %v, %v, %v", baseType, baseScore, comparedType, comparedScore)
	if baseType == comparedType {
		return comparedScore - baseScore, false
	}
	if comparedType == "rocket" {
		return 1, true
	}
	if baseType == "rocket" {
		return -1, false
	}
	if comparedType == "bomb" {
		return 1, true
	}
	return 0, false
}

// CardsAbove 根据当前手牌和上家出牌，查找手牌中是否有比被比较牌型大的牌。
//
// 输入参数 handsNum 和 lastShotNum 分别是当前手牌和上家出牌的编号。
// 函数首先调用 ToPokers() 函数将手牌编号转换为扑克牌。
// 然后，使用 pokersValue() 函数获取上家出牌的扑克牌类型和大小。
// 接着，在遍历 TypeToPokers[cardType] 数组时，如果找到比被比较牌型大的组合，并且该组合在手牌中存在，
// 则将该组合对应的编号添加到 aboveNum 切片，并返回。
// 如果比较的牌型不是炸弹和火箭，并且手牌中存在炸弹组合，则将炸弹组合对应的编号添加到 aboveNum 切片，并返回。
// 如果比较的牌型是火箭，并且手牌中存在王炸组合，则将王炸组合对应的编号添加到 aboveNum 切片，并返回。
// 如果以上条件都不满足，则返回一个空切片。
//
// 请注意，此函数假定输入的 handsNum 和 lastShotNum 切片中的编号是有效的。
// 如果切片中包含无效的编号，则函数的行为是未定义的。
//
// 用法示例：
// aboveNum := CardsAbove([]int{1, 2, 3, 4}, []int{5, 6, 7})
// 上面的示例将返回比 [1, 2, 3, 4] 大的合法组合的编号切片。
func CardsAbove(handsNum, lastShotNum []int) (aboveNum []int) {
	handCards := ToPokers(handsNum)
	turnCards := ToPokers(lastShotNum)
	cardType, cardScore := pokersValue(turnCards)
	logs.Debug("CardsAbove handsNum %v ,lastShotNum %v, handCards %v,cardType %v,turnCards %v",
		handsNum, lastShotNum, handCards, cardType, turnCards)
	if len(cardType) == 0 {
		return
	}
	for _, combination := range TypeToPokers[cardType] {
		if combination.Score > cardScore && IsContains(handCards, combination.Poker) {
			aboveNum = pokersInHand(handsNum, combination.Poker)
			return
		}
	}
	if cardType != "boom" && cardType != "rocket" {
		for _, combination := range TypeToPokers["boom"] {
			if IsContains(handCards, combination.Poker) {
				aboveNum = pokersInHand(handsNum, combination.Poker)
				return
			}
		}
	} else if IsContains(handCards, "Ww") {
		aboveNum = pokersInHand(handsNum, "Ww")
		return
	}
	return
}
