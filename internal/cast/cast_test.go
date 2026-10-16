package cast

import (
	"testing"

	"github.com/apple/liuyao/internal/divination"
)

// 单次摇卦结果必须合法
func TestTossOnceValid(t *testing.T) {
	for i := 0; i < 500; i++ {
		r := TossOnce()
		if r.Value < 6 || r.Value > 9 {
			t.Fatalf("爻值越界: %d", r.Value)
		}
		if r.Backs < 0 || r.Backs > 3 {
			t.Fatalf("背面数越界: %d", r.Backs)
		}
		// 背面数与爻值必须一一对应
		if coinTable[r.Backs] != r.Value {
			t.Fatalf("背面数 %d 与爻值 %d 不符", r.Backs, r.Value)
		}
		// 动爻判定须与爻值一致
		if r.Moving != r.Value.IsChanging() {
			t.Fatalf("动爻标记与爻值不符: %v vs %d", r.Moving, r.Value)
		}
		// 阴阳标记
		if r.YinYang != string(r.Value.YinYang()) {
			t.Fatalf("阴阳标记不符: %s vs %d", r.YinYang, r.Value)
		}
	}
}

// 三枚铜钱映射：0背=老阴(6)、1背=少阳(7)、2背=少阴(8)、3背=老阳(9)
func TestCoinTableMapping(t *testing.T) {
	want := map[int]divination.YaoValue{
		0: divination.LaoYin, 1: divination.ShaoYang,
		2: divination.ShaoYin, 3: divination.LaoYang,
	}
	for backs, w := range want {
		if coinTable[backs] != w {
			t.Errorf("%d 背应为 %d，实际 %d", backs, w, coinTable[backs])
		}
	}
}

// 六次摇卦应返回 6 爻
func TestTossSix(t *testing.T) {
	for i := 0; i < 100; i++ {
		rs := TossSix()
		if len(rs) != 6 {
			t.Fatalf("应返回 6 爻，实际 %d", len(rs))
		}
		vals := ValuesOf(rs)
		if len(vals) != 6 {
			t.Fatalf("ValuesOf 应返回 6 个，实际 %d", len(vals))
		}
		for j, v := range vals {
			if v < 6 || v > 9 {
				t.Fatalf("第%d爻值越界: %d", j+1, v)
			}
		}
	}
}

// 随机性抽样：500 次摇卦应覆盖到多种爻值（极低概率误报）
func TestTossDistribution(t *testing.T) {
	seen := map[divination.YaoValue]int{}
	for i := 0; i < 500; i++ {
		seen[TossOnce().Value]++
	}
	if len(seen) < 3 {
		t.Errorf("500 次摇卦仅出现 %d 种爻值，随机源可能异常: %v", len(seen), seen)
	}
	// 不应出现 6/7/8/9 之外的值
	for v := range seen {
		if v < 6 || v > 9 {
			t.Errorf("出现非法爻值 %d", v)
		}
	}
}
