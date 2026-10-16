package divination

import "testing"

// ============================================================
// 三刑 · 六害
// ============================================================

// 六害表自洽：互为相害、无自害
func TestLiuHaiTable(t *testing.T) {
	if len(LiuHaiZhi) != 12 {
		t.Fatalf("六害表应有 12 项，实际 %d", len(LiuHaiZhi))
	}
	for zhi, opp := range LiuHaiZhi {
		if zhi == opp {
			t.Errorf("%s 不应自害", zhi)
		}
		if back, ok := LiuHaiZhi[opp]; !ok || back != zhi {
			t.Errorf("%s 害 %s，但反向不成立", zhi, opp)
		}
	}
	// 传统六害六组
	want := [][2]string{{"子", "未"}, {"丑", "午"}, {"寅", "巳"}, {"卯", "辰"}, {"申", "亥"}, {"酉", "戌"}}
	for _, w := range want {
		if LiuHaiZhi[w[0]] != w[1] {
			t.Errorf("%s%s 应为相害", w[0], w[1])
		}
	}
}

// 三刑组表正确
func TestSanXingGroupsTable(t *testing.T) {
	if len(SanXingGroups) != 7 {
		t.Fatalf("三刑组应有 7 组（3 组三支 + 4 自刑），实际 %d", len(SanXingGroups))
	}
	byName := map[string]XingGroup{}
	for _, g := range SanXingGroups {
		byName[g.Name] = g
	}
	for _, want := range []struct {
		name     string
		branches string
		self     bool
	}{
		{"无恩之刑", "寅巳申", false},
		{"恃势之刑", "丑戌未", false},
		{"无礼之刑", "子卯", false},
	} {
		g, ok := byName[want.name]
		if !ok {
			t.Errorf("缺 %s", want.name)
			continue
		}
		if joinBranches(g.Branches) != want.branches {
			t.Errorf("%s 分支 = %s, want %s", want.name, joinBranches(g.Branches), want.branches)
		}
		if g.SelfXing != want.self {
			t.Errorf("%s 自刑标记错误", want.name)
		}
	}
	// 四个自刑
	selfCount := 0
	for _, g := range SanXingGroups {
		if g.SelfXing {
			selfCount++
			if len(g.Branches) != 1 {
				t.Errorf("自刑 %s 应只含一支", g.Name)
			}
			if zhiIndex(g.Branches[0]) < 0 {
				t.Errorf("自刑 %s 的地支非法", g.Name)
			}
		}
	}
	if selfCount != 4 {
		t.Errorf("自刑应有 4 组，实际 %d", selfCount)
	}
}

// 三刑检出：乾为天含 子寅辰午申戌，丑戌未不全；
// 用含丑戌未的卦验证恃势之刑。壬戌、癸丑 + 未
func TestXingDetection(t *testing.T) {
	// 坤为地（未巳卯丑亥酉）无戌 → 恃势之刑不成；
	// 但月建/日辰若为戌，则丑未戌齐 → 成刑
	c := mustChart(t, []YaoValue{8, 8, 8, 8, 8, 8}, fixedTime("戌", "寅", "甲", "戌亥"))
	found := false
	for _, x := range c.Relations.Xing {
		if x.Name == "恃势之刑" {
			found = true
			if !x.Complete {
				t.Error("三支俱全应标记 Complete")
			}
		}
	}
	if !found {
		t.Errorf("坤为地(丑未)遇月建戌应成恃势之刑，实际: %+v", c.Relations.Xing)
	}
}

// 自刑检出：同一地支出现两次
func TestSelfXingDetection(t *testing.T) {
	// 乾为天有 辰(三爻)、戌(上爻)... 无重复。
	// 离为火：卯丑亥酉未巳 → 无重复。
	// 用月建与爻支同为辰来构造：乾为天三爻为辰，月建取辰
	c := mustChart(t, []YaoValue{7, 7, 7, 7, 7, 7}, fixedTime("辰", "寅", "甲", "戌亥"))
	found := false
	for _, x := range c.Relations.Xing {
		if x.SelfXing {
			found = true
			if x.Branches[0] != "辰" {
				t.Errorf("应为辰自刑，实际 %s", x.Branches[0])
			}
			if len(x.Members) < 2 {
				t.Errorf("自刑至少需两方参与，实际 %d", len(x.Members))
			}
		}
	}
	if !found {
		t.Errorf("乾为天三爻辰 + 月建辰 应成辰自刑，实际: %+v", c.Relations.Xing)
	}
}

// 六害检出
func TestHaiDetection(t *testing.T) {
	// 乾为天初爻子；取日辰未 → 子未相害
	c := mustChart(t, []YaoValue{7, 7, 7, 7, 7, 7}, fixedTime("寅", "未", "乙", "戌亥"))
	found := false
	for _, h := range c.Relations.Hai {
		if (h.A.Zhi == "子" && h.B.Zhi == "未") || (h.A.Zhi == "未" && h.B.Zhi == "子") {
			found = true
			if h.Kind != "yao-day" {
				t.Errorf("应为爻与日辰相害，实际 %s", h.Kind)
			}
		}
	}
	if !found {
		t.Errorf("乾为天初爻子 遇日辰未 应成子未相害，实际: %+v", c.Relations.Hai)
	}
}

// 日辰与月建之间不构成刑害关系
func TestNoXingHaiBetweenDayAndMonth(t *testing.T) {
	// 构造日辰未、月建子（子未相害），卦中无子无未
	c := mustChart(t, []YaoValue{7, 7, 7, 7, 7, 7}, fixedTime("子", "未", "甲", "戌亥"))
	for _, h := range c.Relations.Hai {
		if h.A.Scope != "yao" && h.B.Scope != "yao" {
			t.Errorf("日辰与月建之间不应论害: %+v", h)
		}
	}
	for _, x := range c.Relations.Xing {
		allTime := true
		for _, m := range x.Members {
			if m.Scope == "yao" {
				allTime = false
			}
		}
		if allTime {
			t.Errorf("日辰与月建之间不应论刑: %+v", x)
		}
	}
}

// 64 卦全量：刑害计算不 panic
func TestXingHaiAllGua(t *testing.T) {
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
		if c.Relations.Xing == nil {
			t.Errorf("%s xing 不应为 nil", c.BenGua.Name)
		}
		if c.Relations.Hai == nil {
			t.Errorf("%s hai 不应为 nil", c.BenGua.Name)
		}
	}
}

// ============================================================
// 用神 / 四神
// ============================================================

// 六亲五行反推：以乾宫金为例逐一校验
func TestInferWuXingOfLiuQin(t *testing.T) {
	// 乾宫属金
	cases := map[string]WuXing{
		"兄弟": Jin,  // 比和
		"子孙": Shui, // 金生水
		"父母": Tu,   // 土生金
		"妻财": Mu,   // 金克木
		"官鬼": Huo,  // 火克金
	}
	for qin, want := range cases {
		if got := inferWuXingOfLiuQin(Jin, qin); got != want {
			t.Errorf("乾宫(金) 的 %s 五行 = %s, want %s", qin, got, want)
		}
	}
	// 反推必须与正向 GetLiuQin 自洽
	for _, gong := range WuXingOrder {
		for _, qin := range LiuQin {
			wx := inferWuXingOfLiuQin(gong, qin)
			if wx == "" {
				t.Errorf("卦宫%s 的 %s 反推失败", gong, qin)
				continue
			}
			if back := GetLiuQin(gong, wx); back != qin {
				t.Errorf("卦宫%s: %s → %s，但反向得 %s", gong, qin, wx, back)
			}
		}
	}
}

// 四神推法：以乾宫金逐一对照手工推演结果
func TestShenOfDerivation(t *testing.T) {
	// 手工推演表（卦宫=金）
	cases := []struct {
		yong, yuan, ji, chou string
	}{
		{"妻财", "子孙", "兄弟", "父母"},
		{"官鬼", "妻财", "子孙", "兄弟"},
		{"父母", "官鬼", "妻财", "子孙"},
		{"子孙", "兄弟", "父母", "官鬼"},
		{"兄弟", "父母", "官鬼", "妻财"},
	}
	for _, c := range cases {
		y := inferWuXingOfLiuQin(Jin, c.yong)
		yuan, ji, chou := shenOf(Jin, y)
		if yuan != c.yuan || ji != c.ji || chou != c.chou {
			t.Errorf("用神%s(%s): 原神=%s 忌神=%s 仇神=%s, want %s/%s/%s",
				c.yong, y, yuan, ji, chou, c.yuan, c.ji, c.chou)
		}
	}
}

// 四神的内在一致性（对全部卦宫、全部用神校验）
func TestShenOfInvariants(t *testing.T) {
	for _, gong := range WuXingOrder {
		for _, qin := range LiuQin {
			y := inferWuXingOfLiuQin(gong, qin)
			yuan, ji, chou := shenOf(gong, y)

			yuanWX := inferWuXingOfLiuQin(gong, yuan)
			jiWX := inferWuXingOfLiuQin(gong, ji)
			chouWX := inferWuXingOfLiuQin(gong, chou)

			// 原神须生用神
			if sheng[yuanWX] != y {
				t.Errorf("卦宫%s 用神%s: 原神%s(%s) 未生用神(%s)", gong, qin, yuan, yuanWX, y)
			}
			// 忌神须克用神
			if ke[jiWX] != y {
				t.Errorf("卦宫%s 用神%s: 忌神%s(%s) 未克用神(%s)", gong, qin, ji, jiWX, y)
			}
			// 仇神须生忌神
			if sheng[chouWX] != jiWX {
				t.Errorf("卦宫%s 用神%s: 仇神%s(%s) 未生忌神%s(%s)", gong, qin, chou, chouWX, ji, jiWX)
			}
			// 仇神须克原神
			if ke[chouWX] != yuanWX {
				t.Errorf("卦宫%s 用神%s: 仇神%s(%s) 未克原神%s(%s)", gong, qin, chou, chouWX, yuan, yuanWX)
			}
			// 四神互不相同（用神本身不应与任何神重合）
			for _, other := range []string{yuan, ji, chou} {
				if other == qin {
					t.Errorf("卦宫%s: 用神%s 与其它神重合", gong, qin)
				}
			}
		}
	}
}

// 主题推断
func TestInferYongShen(t *testing.T) {
	cases := []struct {
		q      string
		want   string
		expect bool
	}{
		{"今年财运怎么样", "妻财", true},
		{"这笔投资能赚钱吗", "妻财", true},
		{"这份工作要不要接", "官鬼", true},
		{"跳槽好不好", "官鬼", true},
		{"考试能不能过", "父母", true},
		{"买房合适吗", "父母", true},
		{"孩子学业如何", "父母", true}, // 「学业」优先命中父母
		{"要不要跟朋友合伙", "兄弟", true},
		{"什么时候能怀孕", "子孙", true},
		{"官司能不能赢", "官鬼", true},
		{"身体检查结果如何", "官鬼", true},
		{"今天天气怎么样", "妻财", false}, // 未识别 → 默认
	}
	for _, c := range cases {
		got, _, matched := InferYongShen(c.q)
		if got != c.want {
			t.Errorf("InferYongShen(%q) = %s, want %s", c.q, got, c.want)
		}
		if matched != c.expect {
			t.Errorf("InferYongShen(%q) matched = %v, want %v", c.q, matched, c.expect)
		}
	}
}

// 自动推断：本卦现则取本卦
func TestShenSystemVisible(t *testing.T) {
	c := mustChart(t, []YaoValue{7, 7, 7, 7, 7, 7}, fixedTime("子", "子", "甲", "戌亥"))
	s := c.Shen
	if s.YongShen != "妻财" { // 默认 question 为空 → 未识别 → 妻财
		t.Errorf("默认用神应为妻财，实际 %s", s.YongShen)
	}
	if s.Selection != SelectVisible {
		t.Errorf("selection = %s, want visible", s.Selection)
	}
	if s.YongYao == nil {
		t.Fatal("应选出用神爻")
	}
	if s.YongYao.LiuQin != s.YongShen {
		t.Errorf("用神爻六亲 = %s, want %s", s.YongYao.LiuQin, s.YongShen)
	}
	if s.SelectNote == "" {
		t.Error("应有取用说明")
	}
	if len(s.Advice) == 0 {
		t.Error("应有断语")
	}
}

// 显式指定用神
func TestShenSystemExplicit(t *testing.T) {
	tm := fixedTime("子", "子", "甲", "戌亥")
	ch, err := BuildChart(ChartInput{
		YaoValues: []YaoValue{7, 7, 7, 7, 7, 7},
		Question:  "随便问问",
		Method:    MethodManual,
		YongShen:  "父母",
	}, tm)
	if err != nil {
		t.Fatal(err)
	}
	if ch.Shen.YongShen != "父母" {
		t.Errorf("用神 = %s, want 父母", ch.Shen.YongShen)
	}
	if ch.Shen.Inferred {
		t.Error("显式指定时 inferred 应为 false")
	}
}

// 用神并见时须提示
func TestShenSystemAmbiguous(t *testing.T) {
	// 乾为天有 父母 在 三爻、上爻 → 并见
	c := mustChart(t, []YaoValue{7, 7, 7, 7, 7, 7}, fixedTime("子", "子", "甲", "戌亥"))
	if c.Shen.YongShen != "妻财" {
		t.Skip("默认用神非妻财，跳过")
	}
	// 乾为天 妻财 仅二爻一处 → 不应 ambiguous
	if c.Shen.Ambiguous {
		t.Errorf("乾为天妻财仅一处，不应标为并见: %+v", c.Shen.Candidates)
	}

	// 显式取父母 → 三爻与上爻并见
	ch, _ := BuildChart(ChartInput{
		YaoValues: []YaoValue{7, 7, 7, 7, 7, 7},
		Method:    MethodManual, YongShen: "父母",
	}, fixedTime("子", "子", "甲", "戌亥"))
	if len(ch.Shen.Candidates) < 2 {
		t.Fatalf("乾为天父母应有两处，实际 %d", len(ch.Shen.Candidates))
	}
	// 两爻皆静，故按 choosePrimary 取世应所在（上爻为世）
	if ch.Shen.YongYao == nil {
		t.Fatal("应选出主用神")
	}
	joined := ""
	for _, a := range ch.Shen.Advice {
		joined += a
	}
	if !contains(joined, "并见") {
		t.Errorf("父母并见应提示兼看，实际断语: %v", ch.Shen.Advice)
	}
}

// 用神伏藏：本卦与变卦皆无该六亲时取伏神
func TestShenSystemFuShen(t *testing.T) {
	// 天水讼（离宫）六亲不全，必有伏神
	c := mustChart(t, []YaoValue{8, 7, 8, 7, 7, 7}, fixedTime("子", "子", "甲", "戌亥"))
	if c.BenGua.Name != "天水讼" {
		t.Fatalf("前置条件有误: %s", c.BenGua.Name)
	}
	if len(c.FuShen) == 0 {
		t.Fatal("天水讼应有伏神")
	}
	// 对每个伏神六亲，建一个以之为用神的盘，应走 fushen 分支
	for _, f := range c.FuShen {
		ch, err := BuildChart(ChartInput{
			YaoValues: []YaoValue{8, 7, 8, 7, 7, 7},
			Method:    MethodManual, YongShen: f.LiuQin,
		}, fixedTime("子", "子", "甲", "戌亥"))
		if err != nil {
			t.Fatal(err)
		}
		// 若本卦已现该六亲，则 selection 为 visible，否则应取伏神
		present := false
		for _, y := range ch.Yaos {
			if y.LiuQin == f.LiuQin {
				present = true
			}
		}
		if present {
			continue
		}
		if ch.Shen.Selection != SelectFuShen {
			t.Errorf("%s 不上卦，应取伏神，实际 %s", f.LiuQin, ch.Shen.Selection)
		}
		if ch.Shen.YongYao == nil || !ch.Shen.YongYao.IsFuShen {
			t.Errorf("%s 的用神应标记 IsFuShen", f.LiuQin)
		}
	}
}

// 64 卦 × 5 用神：四神体系不 panic 且完整
func TestShenSystemAllGuaAllQin(t *testing.T) {
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
		for _, qin := range LiuQin {
			ch, err := BuildChart(ChartInput{
				YaoValues: vals, Method: MethodManual, YongShen: qin,
			}, tm)
			if err != nil {
				t.Fatalf("%s/%s 排盘失败: %v", mark, qin, err)
			}
			s := ch.Shen
			if s.YongShen != qin {
				t.Errorf("用神 = %s, want %s", s.YongShen, qin)
			}
			// 四神必须齐备且互异
			if s.YuanShen == "" || s.JiShen == "" || s.ChouShen == "" {
				t.Errorf("%s/%s 四神不全", mark, qin)
			}
			if s.SelectNote == "" {
				t.Errorf("%s/%s 缺取用说明", mark, qin)
			}
			// selection 必须为合法值
			switch s.Selection {
			case SelectVisible, SelectChanged, SelectFuShen, SelectMissing:
			default:
				t.Errorf("%s/%s selection 非法: %q", mark, qin, s.Selection)
			}
		}
	}
}

// 用神四神不得与用神自身重合（对全部卦宫）
func TestShenNoOverlapWithYong(t *testing.T) {
	for _, gong := range WuXingOrder {
		for _, qin := range LiuQin {
			y := inferWuXingOfLiuQin(gong, qin)
			yuan, ji, chou := shenOf(gong, y)
			for _, o := range []string{yuan, ji, chou} {
				if o == qin {
					t.Errorf("卦宫%s 用神%s: 四神与用神重合", gong, qin)
				}
			}
			if yuan == ji || yuan == chou || ji == chou {
				t.Errorf("卦宫%s 用神%s: 四神之间存在重合 %s/%s/%s", gong, qin, yuan, ji, chou)
			}
		}
	}
}
