// Package cast 提供摇卦（三枚铜钱）与自动起卦的随机逻辑。
package cast

import (
	"crypto/rand"
	"math/big"

	"github.com/apple/liuyao/internal/divination"
)

// CoinResult 单次摇卦结果
type CoinResult struct {
	Value   divination.YaoValue `json:"value"`   // 6/7/8/9
	Backs   int                 `json:"backs"`   // 背面数（0-3）
	YinYang string              `json:"yinYang"` // 阴/阳
	Moving  bool                `json:"moving"`  // 是否动爻
}

// 三枚铜钱：以「背」为阳面计数。
//
//	3 背 → 老阳(9)  动
//	2 背 → 少阴(8)  静
//	1 背 → 少阳(7)  静
//	0 背 → 老阴(6)  动
var coinTable = [4]divination.YaoValue{
	divination.LaoYin,   // 0 背
	divination.ShaoYang, // 1 背
	divination.ShaoYin,  // 2 背
	divination.LaoYang,  // 3 背
}

// TossOnce 摇一次（三枚铜钱），返回该爻结果
func TossOnce() CoinResult {
	backs := 0
	for i := 0; i < 3; i++ {
		if secureBool() {
			backs++
		}
	}
	v := coinTable[backs]
	return CoinResult{
		Value:   v,
		Backs:   backs,
		YinYang: string(v.YinYang()),
		Moving:  v.IsChanging(),
	}
}

// TossSix 连摇六次，返回自下而上的六爻
func TossSix() []CoinResult {
	out := make([]CoinResult, 0, 6)
	for i := 0; i < 6; i++ {
		out = append(out, TossOnce())
	}
	return out
}

// ValuesOf 提取爻值数组
func ValuesOf(results []CoinResult) []divination.YaoValue {
	out := make([]divination.YaoValue, 0, len(results))
	for _, r := range results {
		out = append(out, r.Value)
	}
	return out
}

// secureBool 使用密码学随机源，避免可预测性
func secureBool() bool {
	n, err := rand.Int(rand.Reader, big.NewInt(2))
	if err != nil {
		// 极端情况下退化为固定值，不应发生
		return false
	}
	return n.Int64() == 1
}
