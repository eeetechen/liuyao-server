package divination

import (
	"testing"
)

// 构造固定时间信息，避免测试依赖真实历法库
func fixedTime(monthZhi, dayZhi, dayGan, xunKong string) GanZhiTime {
	return GanZhiTime{
		Year: "甲辰", Month: "丙寅", Day: "戊辰", Hour: "戊午",
		DayGan: dayGan, DayZhi: dayZhi, MonthZhi: monthZhi,
		XunKong: xunKong, SolarText: "测试", LunarText: "测试",
	}
}

func mustChart(t *testing.T, vals []YaoValue, tm GanZhiTime) *Chart {
	t.Helper()
	c, err := BuildChart(ChartInput{YaoValues: vals, Question: "测试", Method: MethodManual}, tm)
	if err != nil {
		t.Fatalf("排盘失败: %v", err)
	}
	return c
}

// 乾为天应正确排布：卦名、卦宫、世应、六亲、纳甲
func TestBuildChartQian(t *testing.T) {
	c := mustChart(t, []YaoValue{7, 7, 7, 7, 7, 7}, fixedTime("子", "子", "甲", "戌亥"))

	if c.BenGua.Name != "乾为天" {
		t.Errorf("卦名 = %s, want 乾为天", c.BenGua.Name)
	}
	if c.BenGua.Gong != "乾" {
		t.Errorf("卦宫 = %s, want 乾", c.BenGua.Gong)
	}
	if c.BenGua.ShiPos != 6 || c.BenGua.YingPos != 3 {
		t.Errorf("世应 = 世%d应%d, want 世6应3", c.BenGua.ShiPos, c.BenGua.YingPos)
	}
	if !c.Summary.IsLiuChong {
		t.Error("乾为天应为六冲卦")
	}
	// 纳甲：子寅辰午申戌
	wantZhi := []string{"子", "寅", "辰", "午", "申", "戌"}
	wantQin := []string{"子孙", "妻财", "父母", "官鬼", "兄弟", "父母"}
	for i, y := range c.Yaos {
		if y.Zhi != wantZhi[i] {
			t.Errorf("第%d爻地支 = %s, want %s", i+1, y.Zhi, wantZhi[i])
		}
		if y.LiuQin != wantQin[i] {
			t.Errorf("第%d爻六亲 = %s, want %s", i+1, y.LiuQin, wantQin[i])
		}
	}
	// 世应标记
	if !c.Yaos[5].IsShi {
		t.Error("上爻应为世爻")
	}
	if !c.Yaos[2].IsYing {
		t.Error("三爻应为应爻")
	}
}

// 六亲完整性：乾卦六亲为 子孙/妻财/父母/官鬼/兄弟/父母，五种俱全故无伏神
func TestBuildChartLiuQinFull(t *testing.T) {
	c := mustChart(t, []YaoValue{7, 7, 7, 7, 7, 7}, fixedTime("子", "子", "甲", "戌亥"))
	seen := map[string]bool{}
	for _, y := range c.Yaos {
		seen[y.LiuQin] = true
	}
	if len(seen) != 5 {
		t.Errorf("乾卦六亲应五种俱全，实际 %d 种 (%v)", len(seen), seen)
	}
	if len(c.FuShen) != 0 {
		t.Errorf("六亲俱全之卦不应有伏神，实际 %v", c.FuShen)
	}
}

// 六亲不全之卦应产生伏神，且伏神取自本宫卦
func TestBuildChartFuShen(t *testing.T) {
	// 天水讼（内坎010 外乾111）属离宫，六爻六亲不全，应有伏神
	c := mustChart(t, []YaoValue{8, 7, 8, 7, 7, 7}, fixedTime("子", "子", "甲", "戌亥"))
	if c.BenGua.Name != "天水讼" {
		t.Fatalf("前置条件有误，本卦应为天水讼，实际 %s", c.BenGua.Name)
	}
	if len(c.FuShen) == 0 {
		t.Fatal("天水讼六亲不全，应产生伏神")
	}
	for _, f := range c.FuShen {
		if f.LiuQin == "" || f.GanZhi == "" || f.Position < 1 || f.Position > 6 {
			t.Errorf("伏神字段不完整: %+v", f)
		}
	}
}

// 动爻生成变卦
func TestBuildChartMoving(t *testing.T) {
	c := mustChart(t, []YaoValue{9, 9, 9, 9, 9, 9}, fixedTime("子", "子", "甲", "戌亥"))
	if c.BianGua == nil {
		t.Fatal("全动应产生变卦")
	}
	if c.BianGua.Name != "坤为地" {
		t.Errorf("乾为天全动变卦 = %s, want 坤为地", c.BianGua.Name)
	}
	if c.Summary.MovingCount != 6 {
		t.Errorf("明动爻数 = %d, want 6", c.Summary.MovingCount)
	}
	for _, y := range c.Yaos {
		if y.ChangedGanZhi == "" {
			t.Errorf("第%d动爻缺少变爻纳甲", y.Position)
		}
		if y.ChangedLiuQin == "" {
			t.Errorf("第%d动爻缺少变爻六亲", y.Position)
		}
	}
}

// 静卦不变
func TestBuildChartStaticNoBian(t *testing.T) {
	c := mustChart(t, []YaoValue{7, 8, 7, 8, 7, 8}, fixedTime("子", "子", "甲", "戌亥"))
	if c.BianGua != nil {
		t.Error("无动爻不应产生变卦")
	}
}

// 暗动应出现在 summary 与对应爻上
func TestBuildChartAnDong(t *testing.T) {
	// 爻支丑，日辰未冲之；月建丑使丑土旺 → 暗动
	// 取一含丑爻的卦：乾为天二爻为寅(木)... 需选含丑的卦。
	// 用坤为地（乙未、乙巳、乙卯、癸丑、癸亥、癸酉），四爻为丑。
	c := mustChart(t, []YaoValue{8, 8, 8, 8, 8, 8}, fixedTime("丑", "未", "己", "戌亥"))

	found := false
	for _, y := range c.Yaos {
		if y.Zhi == "丑" && y.Special == StatusAnDong {
			found = true
			if y.MoveState != MoveAnDong {
				t.Errorf("暗动爻 moveState = %s, want hidden_moving", y.MoveState)
			}
			if !y.IsActive {
				t.Error("暗动爻应标记为发动")
			}
			if !IsWangXiang(y.WangShuai) {
				t.Errorf("暗动爻应旺相，实际 %s", y.WangShuai)
			}
		}
	}
	if !found {
		for _, y := range c.Yaos {
			t.Logf("爻%d %s %s %s special=%s", y.Position, y.GanZhi, y.WangShuai, y.MoveLabel, y.Special)
		}
		t.Fatal("未检出暗动爻（丑土被未日冲，且丑月得旺）")
	}
	if c.Summary.AnDongCount == 0 {
		t.Error("summary.AnDongCount 应大于 0")
	}
}

// 非法爻值应报错
func TestBuildChartInvalidYao(t *testing.T) {
	_, err := BuildChart(ChartInput{YaoValues: []YaoValue{1, 2, 3, 4, 5, 6}}, fixedTime("子", "子", "甲", "戌亥"))
	if err == nil {
		t.Error("非法爻值（1-5）应报错")
	}
	_, err = BuildChart(ChartInput{YaoValues: []YaoValue{7, 7, 7}}, fixedTime("子", "子", "甲", "戌亥"))
	if err == nil {
		t.Error("爻数不足应报错")
	}
}

// 六神应与日干对应
func TestBuildChartLiuShen(t *testing.T) {
	c := mustChart(t, []YaoValue{7, 7, 7, 7, 7, 7}, fixedTime("子", "子", "丙", "戌亥"))
	if c.Yaos[0].LiuShen != "朱雀" {
		t.Errorf("丙日初爻应起朱雀，实际 %s", c.Yaos[0].LiuShen)
	}
	c2 := mustChart(t, []YaoValue{7, 7, 7, 7, 7, 7}, fixedTime("子", "子", "戊", "戌亥"))
	if c2.Yaos[0].LiuShen != "勾陈" {
		t.Errorf("戊日初爻应起勾陈，实际 %s", c2.Yaos[0].LiuShen)
	}
}

// 旬空标标记
func TestBuildChartXunKong(t *testing.T) {
	// 乾为天上爻为戌，旬空戌亥 → 上爻应为空
	c := mustChart(t, []YaoValue{7, 7, 7, 7, 7, 7}, fixedTime("子", "子", "甲", "戌亥"))
	if !c.Yaos[5].IsKong {
		t.Error("乾为天上爻(戌)在戌亥旬空内，应为旬空")
	}
	if len(c.Summary.KongYaos) == 0 {
		t.Error("summary.KongYaos 应包含上爻")
	}
}

// 六冲卦共 10 个，逐一排盘不 panic 且标记一致
func TestAllLiuChongCharts(t *testing.T) {
	n := 0
	tm := fixedTime("子", "子", "甲", "戌亥")
	for mark := range Gua64 {
		var vals []YaoValue
		for i := 0; i < 6; i++ {
			if mark[i] == '1' {
				vals = append(vals, 7)
			} else {
				vals = append(vals, 8)
			}
		}
		c := mustChart(t, vals, tm)
		if IsLiuChong(mark) {
			n++
			if !c.Summary.IsLiuChong {
				t.Errorf("%s 应为六冲", c.BenGua.Name)
			}
			if c.BenGua.Type != "六冲" {
				t.Errorf("%s 类型应为六冲，实际 %s", c.BenGua.Name, c.BenGua.Type)
			}
		}
	}
	if n != 10 {
		t.Errorf("六冲卦计数 = %d, want 10", n)
	}
}

// 全量 64 卦排盘不 panic，且各字段完整
func TestAllGuaCharts(t *testing.T) {
	tm := fixedTime("寅", "子", "甲", "戌亥")
	for mark := range Gua64 {
		var vals []YaoValue
		for i := 0; i < 6; i++ {
			if mark[i] == '1' {
				vals = append(vals, 7)
			} else {
				vals = append(vals, 8)
			}
		}
		c := mustChart(t, vals, tm)
		if c.BenGua.Name != Gua64[mark] {
			t.Errorf("%s 卦名不符: %s", mark, c.BenGua.Name)
		}
		for _, y := range c.Yaos {
			if y.LiuQin == "" || y.LiuShen == "" || y.WangShuai == "" || y.GanZhi == "" {
				t.Errorf("%s 第%d爻字段缺失", c.BenGua.Name, y.Position)
			}
		}
		if len(c.Summary.Highlights) == 0 {
			t.Errorf("%s 应至少产生一条提示", c.BenGua.Name)
		}
	}
}

// 六爻皆静时应给出提示
func TestAllStaticNote(t *testing.T) {
	c := mustChart(t, []YaoValue{7, 7, 7, 7, 7, 7}, fixedTime("申", "子", "甲", "戌亥"))
	joined := ""
	for _, s := range c.Summary.Highlights {
		joined += s
	}
	if !contains(joined, "六爻皆静") {
		t.Errorf("全静卦应提示六爻皆静，实际提示: %v", c.Summary.Highlights)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (func() bool {
		for i := 0; i+len(sub) <= len(s); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
		return false
	})()
}
