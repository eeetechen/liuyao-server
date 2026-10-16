package divination

import (
	"encoding/json"
	"os"
	"testing"
)

// 交叉验证：使用外部独立实现（bopo/najia，MIT）为全部 64 卦生成的基准数据，
// 校验本引擎的卦名、卦宫、世应、六亲、纳甲五行、六神是否完全一致。
//
// 基准文件 /tmp/najia_ref.json 由 najia 引擎实际运行生成。
// 若文件不存在则跳过（保证测试在无基准文件时仍可运行）。
type refEntry struct {
	Name string   `json:"name"`
	Gong string   `json:"gong"`
	Shi  int      `json:"shi"`
	Ying int      `json:"ying"`
	Qin6 []string `json:"qin6"`
	Qinx []string `json:"qinx"`
	God6 []string `json:"god6"`
}

func loadRef(t *testing.T) map[string]refEntry {
	t.Helper()
	b, err := os.ReadFile("/tmp/najia_ref.json")
	if err != nil {
		t.Skipf("跳过交叉验证：基准文件不可用 (%v)", err)
	}
	var m map[string]refEntry
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("基准文件解析失败: %v", err)
	}
	return m
}

func TestCrossValidate64Gua(t *testing.T) {
	ref := loadRef(t)
	if len(ref) != 64 {
		t.Fatalf("基准数据应有 64 卦，实际 %d", len(ref))
	}

	// 固定基准时间：2024-01-01 12:00 => 甲子日，青龙起
	dayGan := "甲"
	shen := GetLiuShen(dayGan)

	checked := 0
	for mark, want := range ref {
		// 卦名
		if got := Gua64[mark]; got != want.Name {
			t.Errorf("卦名不符 mark=%s: got %q want %q", mark, got, want.Name)
		}
		// 卦宫
		shi := setShiYao(mark)
		gong := palace(mark, shi.Shi)
		if gong != want.Gong {
			t.Errorf("卦宫不符 %s(%s): got %q want %q", want.Name, mark, gong, want.Gong)
		}
		// 世应
		if shi.Shi != want.Shi || shi.Ying != want.Ying {
			t.Errorf("世应不符 %s: got 世%d应%d want 世%d应%d",
				want.Name, shi.Shi, shi.Ying, want.Shi, want.Ying)
		}
		// 纳甲五行 + 六亲
		gongIdx := indexOf(BaGua[:], gong)
		gongWuXing := BaGuaWuXing[gongIdx]
		naJia := getNaJia(mark)
		for i := 0; i < 6; i++ {
			gotQinx := naJia[i].GanZhi + string(naJia[i].WuXing)
			if gotQinx != want.Qinx[i] {
				t.Errorf("纳甲不符 %s 第%d爻: got %q want %q", want.Name, i+1, gotQinx, want.Qinx[i])
			}
			gotQin := GetLiuQin(gongWuXing, naJia[i].WuXing)
			if gotQin != want.Qin6[i] {
				t.Errorf("六亲不符 %s 第%d爻: got %q want %q", want.Name, i+1, gotQin, want.Qin6[i])
			}
		}
		// 六神
		for i := 0; i < 6; i++ {
			if shen[i] != want.God6[i] {
				t.Errorf("六神不符 %s 第%d爻: got %q want %q", want.Name, i+1, shen[i], want.God6[i])
			}
		}
		checked++
	}
	t.Logf("交叉验证通过：%d 卦", checked)
}
