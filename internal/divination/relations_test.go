package divination

import "testing"

// ---------- 五行生克 ----------

// 生克判定：以五行关系逐一校验
func TestInfluenceOf(t *testing.T) {
	cases := []struct {
		src, tgt WuXing
		kind     string
		label    string
	}{
		{Mu, Huo, "生", "得生"},  // 木生火
		{Huo, Mu, "泄", "泄气"},  // 火泄木
		{Mu, Tu, "克", "受克"},   // 木克土
		{Tu, Mu, "耗", "耗力"},   // 土耗木
		{Jin, Jin, "比和", "扶"}, // 同五行
		{Shui, Mu, "生", "得生"}, // 水生木
		{Tu, Shui, "克", "受克"}, // 土克水
	}
	for _, c := range cases {
		got := influenceOf("月建", c.src, c.tgt)
		if got.Kind != c.kind || got.Label != c.label {
			t.Errorf("influenceOf(%s->%s) = %s/%s, want %s/%s",
				c.src, c.tgt, got.Kind, got.Label, c.kind, c.label)
		}
	}
}

// 五行生克覆盖全部 25 种组合，且归类唯一
func TestInfluenceCoversAllPairs(t *testing.T) {
	counts := map[string]int{}
	for _, a := range WuXingOrder {
		for _, b := range WuXingOrder {
			in := influenceOf("月建", a, b)
			if in.Kind == "" {
				t.Fatalf("%s->%s 未归类", a, b)
			}
			counts[in.Kind]++
		}
	}
	// 25 组合中：同五行 5 种（比和），生克泄耗各 5 种
	for kind, want := range map[string]int{"比和": 5, "生": 5, "克": 5, "泄": 5, "耗": 5} {
		if counts[kind] != want {
			t.Errorf("%s 计数 = %d, want %d (全部: %v)", kind, counts[kind], want, counts)
		}
	}
}

// 每爻都应给出月建与日辰的生克
func TestBuildShengKe(t *testing.T) {
	c := mustChart(t, []YaoValue{7, 7, 7, 7, 7, 7}, fixedTime("子", "子", "甲", "戌亥"))
	if len(c.Relations.ShengKe) != 6 {
		t.Fatalf("生克条数 = %d, want 6", len(c.Relations.ShengKe))
	}
	for i, sk := range c.Relations.ShengKe {
		if sk.Position != i+1 {
			t.Errorf("第%d条爻位 = %d", i+1, sk.Position)
		}
		if sk.Month.Source != "月建" || sk.Day.Source != "日辰" {
			t.Errorf("第%d条来源标注错误", i+1)
		}
		if sk.Month.Kind == "" || sk.Day.Kind == "" {
			t.Errorf("第%d条生克类型为空", i+1)
		}
	}
}

// 月建子水：乾卦初爻子水应比和，四爻午火应受克
func TestShengKeAgainstMonth(t *testing.T) {
	c := mustChart(t, []YaoValue{7, 7, 7, 7, 7, 7}, fixedTime("子", "子", "甲", "戌亥"))
	// 初爻子水 vs 月建子水 → 比和
	if k := c.Relations.ShengKe[0].Month.Kind; k != "比和" {
		t.Errorf("初爻子水遇月建子水应为比和，实际 %s", k)
	}
	// 四爻午火 vs 月建子水 → 水克火 → 受克
	if k := c.Relations.ShengKe[3].Month.Kind; k != "克" {
		t.Errorf("四爻午火遇月建子水应为受克，实际 %s", k)
	}
}

// 世应关系
func TestShiYingRelation(t *testing.T) {
	c := mustChart(t, []YaoValue{7, 7, 7, 7, 7, 7}, fixedTime("子", "子", "甲", "戌亥"))
	sr := c.Relations.ShiYing
	if sr == nil {
		t.Fatal("应给出世应关系")
	}
	// 乾卦世6应3：世戌土、应辰土 → 同为土 → 比和
	if sr.Kind != "比和" {
		t.Errorf("乾卦世应同为土应比和，实际 %s (%s)", sr.Kind, sr.Label)
	}
	if sr.ShiPosition != 6 || sr.YingPosition != 3 {
		t.Errorf("世应位置 = %d/%d, want 6/3", sr.ShiPosition, sr.YingPosition)
	}
	if sr.Label == "" || sr.Desc == "" {
		t.Error("世应描述不应为空")
	}
}

// ---------- 六合 ----------

func TestLiuHeTable(t *testing.T) {
	// 六合配对及其合化
	want := map[string]struct {
		partner string
		hua     WuXing
	}{
		"子": {"丑", Tu}, "丑": {"子", Tu},
		"寅": {"亥", Mu}, "亥": {"寅", Mu},
		"卯": {"戌", Huo}, "戌": {"卯", Huo},
		"辰": {"酉", Jin}, "酉": {"辰", Jin},
		"巳": {"申", Shui}, "申": {"巳", Shui},
		"午": {"未", Tu}, "未": {"午", Tu},
	}
	if len(LiuHeZhi) != 12 {
		t.Fatalf("六合表应有 12 项，实际 %d", len(LiuHeZhi))
	}
	for zhi, w := range want {
		got, ok := LiuHeZhi[zhi]
		if !ok {
			t.Errorf("六合表缺 %s", zhi)
			continue
		}
		if got.Partner != w.partner {
			t.Errorf("%s 的六合对象 = %s, want %s", zhi, got.Partner, w.partner)
		}
		if got.Result != w.hua {
			t.Errorf("%s%s 合化 = %s, want %s", zhi, w.partner, got.Result, w.hua)
		}
	}
}

// 六合关系应能检出涉及日辰者
func TestLiuHeDetection(t *testing.T) {
	// 乾卦初爻子水；取日辰为丑 → 子丑六合
	c := mustChart(t, []YaoValue{7, 7, 7, 7, 7, 7}, fixedTime("子", "丑", "乙", "戌亥"))
	found := false
	for _, p := range c.Relations.LiuHe {
		if (p.A.Zhi == "子" && p.B.Zhi == "丑") || (p.A.Zhi == "丑" && p.B.Zhi == "子") {
			found = true
			if p.Hua != Tu {
				t.Errorf("子丑应合化土，实际 %s", p.Hua)
			}
			if p.Kind != "yao-day" {
				t.Errorf("应为爻与日辰相合，实际 %s", p.Kind)
			}
		}
	}
	if !found {
		t.Errorf("未检出子丑六合，实际: %+v", c.Relations.LiuHe)
	}
}

// ---------- 六冲 ----------

// 六冲表自洽：互为对冲且无自冲
func TestLiuChongTable(t *testing.T) {
	if len(LiuChongZhi) != 12 {
		t.Fatalf("六冲表应有 12 项，实际 %d", len(LiuChongZhi))
	}
	for zhi, opp := range LiuChongZhi {
		if zhi == opp {
			t.Errorf("%s 不应自冲", zhi)
		}
		if back, ok := LiuChongZhi[opp]; !ok || back != zhi {
			t.Errorf("%s 冲 %s，但反向不成立", zhi, opp)
		}
	}
	// 相隔六位
	for _, zhi := range DiZhi {
		a, b := zhiIndex(zhi), zhiIndex(LiuChongZhi[zhi])
		d := (b - a + 12) % 12
		if d != 6 {
			t.Errorf("%s 与 %s 相隔 %d 位，应为 6", zhi, LiuChongZhi[zhi], d)
		}
	}
}

// 日辰冲爻应检出
func TestLiuChongDetection(t *testing.T) {
	// 乾卦四爻午火；日辰子 → 日辰与四爻成子午冲。
	// 注意：乾卦初爻亦为子，故同时存在「初爻子 ↔ 四爻午」的爻间冲，
	// 两者都应检出，此处分别校验。
	c := mustChart(t, []YaoValue{7, 7, 7, 7, 7, 7}, fixedTime("寅", "子", "甲", "戌亥"))
	var yaoDay, yaoYao bool
	for _, p := range c.Relations.LiuChong {
		if !((p.A.Zhi == "午" && p.B.Zhi == "子") || (p.A.Zhi == "子" && p.B.Zhi == "午")) {
			continue
		}
		switch p.Kind {
		case "yao-day":
			yaoDay = true
		case "yao-yao":
			yaoYao = true
		}
	}
	if !yaoDay {
		t.Errorf("应检出日辰子冲四爻午，实际: %+v", c.Relations.LiuChong)
	}
	if !yaoYao {
		t.Errorf("应检出初爻子与四爻午相冲，实际: %+v", c.Relations.LiuChong)
	}
}

// 月建冲爻应标注月破
func TestLiuChongMonthNote(t *testing.T) {
	c := mustChart(t, []YaoValue{7, 7, 7, 7, 7, 7}, fixedTime("子", "寅", "甲", "戌亥"))
	for _, p := range c.Relations.LiuChong {
		if p.Kind == "yao-month" {
			if p.Note == "" {
				t.Error("月建冲爻应给出说明")
			}
		}
	}
}

// ---------- 三合 / 半合 ----------

func TestSanHeGroupsTable(t *testing.T) {
	if len(SanHeGroups) != 4 {
		t.Fatalf("三合局应有 4 组，实际 %d", len(SanHeGroups))
	}
	want := map[string]WuXing{
		"申子辰": Shui, "亥卯未": Mu, "寅午戌": Huo, "巳酉丑": Jin,
	}
	for _, g := range SanHeGroups {
		if want[g.Name] != g.Element {
			t.Errorf("%s 应为 %s 局，实际 %s", g.Name, want[g.Name], g.Element)
		}
		// 三支的五行应与局五行合乎三合之理（长生+帝旺+墓库）
		if len(g.Branches) != 3 {
			t.Errorf("%s 分支数 = %d", g.Name, len(g.Branches))
		}
	}
}

// 三合局检出：申子辰 需三支俱现
func TestSanHeFullDetection(t *testing.T) {
	// 构造含申、子、辰三支的卦：乾为天含 子(初)、辰(三)、申(五) —— 正是申子辰！
	c := mustChart(t, []YaoValue{7, 7, 7, 7, 7, 7}, fixedTime("寅", "寅", "甲", "戌亥"))
	found := false
	for _, j := range c.Relations.SanHe {
		if j.Name == "申子辰" {
			found = true
			if j.Element != Shui {
				t.Errorf("申子辰应为水局，实际 %s", j.Element)
			}
			if !j.Complete {
				t.Error("三支俱现应标记 Complete")
			}
			if len(j.Positions) != 3 {
				t.Errorf("应有三爻参与，实际 %v", j.Positions)
			}
		}
	}
	if !found {
		t.Fatalf("乾为天含子辰申，应检出水局三合，实际: %+v", c.Relations.SanHe)
	}
}

// 三合未全时应作半合，且不再重复列为三合
func TestBanHeDetection(t *testing.T) {
	// 坤为地：未(初)、巳(二)、卯(三)、丑(四)、亥(五)、酉(上)
	// 亥卯未三支俱全 → 木局三合；巳酉丑三支俱全 → 金局三合
	c := mustChart(t, []YaoValue{8, 8, 8, 8, 8, 8}, fixedTime("寅", "寅", "甲", "戌亥"))
	names := map[string]bool{}
	for _, j := range c.Relations.SanHe {
		names[j.Name] = true
	}
	if !names["亥卯未"] {
		t.Errorf("坤为地应成亥卯未木局，实际 %+v", c.Relations.SanHe)
	}
	if !names["巳酉丑"] {
		t.Errorf("坤为地应成巳酉丑金局，实际 %+v", c.Relations.SanHe)
	}
	// 已成三合局者，其两支不应再列为半合
	for _, j := range c.Relations.BanHe {
		if (j.Name == "亥卯" || j.Name == "卯未") && names["亥卯未"] {
			t.Errorf("%s 已成三合局，不应再列半合", j.Name)
		}
	}
}

// 半合应有真实可检出的场景：找到含半合对且缺第三支的卦
func TestBanHePositiveCase(t *testing.T) {
	// 扫描 64 卦，找出确实产生半合的卦例并校验其正确性
	tm := fixedTime("寅", "寅", "甲", "戌亥") // 月建日辰皆寅，不干扰
	checked := 0
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
		for _, j := range c.Relations.BanHe {
			checked++
			if len(j.Branches) != 2 {
				t.Errorf("半合 %s 应有 2 支，实际 %v", j.Name, j.Branches)
			}
			// 该半合对应的三合局第三支不得出现在卦中或日月
			for _, p := range BanHePairs {
				if p.A == j.Branches[0] && p.B == j.Branches[1] || p.A == j.Branches[1] && p.B == j.Branches[0] {
					if zhisPresent(c, p.Missing) {
						t.Errorf("%s: 半合 %s 缺支 %s 实际已现，不应列为半合", c.BenGua.Name, j.Name, p.Missing)
					}
				}
			}
			if j.Element == "" {
				t.Errorf("半合 %s 缺五行", j.Name)
			}
		}
	}
	t.Logf("共检出 %d 个半合实例", checked)
}

// zhisPresent 判断某地支是否出现在六爻、日辰或月建中
func zhisPresent(c *Chart, zhi string) bool {
	for _, y := range c.Yaos {
		if y.Zhi == zhi {
			return true
		}
	}
	return c.Time.DayZhi == zhi || c.Time.MonthZhi == zhi
}

// IsBanHeValid 辅助：判断两支是否构成半合（第三支缺席）
func TestBanHePairTable(t *testing.T) {
	if len(BanHePairs) != 8 {
		t.Fatalf("半合对应为 8 组，实际 %d", len(BanHePairs))
	}
	for _, p := range BanHePairs {
		// 三支必须同属一个三合局
		ok := false
		for _, g := range SanHeGroups {
			inA, inB, inM := false, false, false
			for _, b := range g.Branches {
				if b == p.A {
					inA = true
				}
				if b == p.B {
					inB = true
				}
				if b == p.Missing {
					inM = true
				}
			}
			if inA && inB && inM {
				ok = true
				if g.Element != p.Element {
					t.Errorf("%s%s 半合五行 %s 与三合局 %s 不符", p.A, p.B, p.Element, g.Element)
				}
			}
		}
		if !ok {
			t.Errorf("%s%s(缺%s) 不属任何三合局", p.A, p.B, p.Missing)
		}
	}
}

// ---------- 汇总 ----------

// Relations 应产出提示，且六冲由 summary 统一给出
func TestRelationsHighlights(t *testing.T) {
	c := mustChart(t, []YaoValue{7, 7, 7, 7, 7, 7}, fixedTime("子", "子", "甲", "戌亥"))
	if len(c.Summary.Highlights) == 0 {
		t.Fatal("提示不应为空")
	}
	var hasShiYing, hasSanHe bool
	for _, h := range c.Summary.Highlights {
		if contains(h, "世应") {
			hasShiYing = true
		}
		if contains(h, "三合") {
			hasSanHe = true
		}
	}
	if !hasShiYing {
		t.Error("提示中应包含世应关系")
	}
	if !hasSanHe {
		t.Error("乾为天含子辰申，提示中应包含三合局")
	}
}

// 64 卦全量：关系计算不 panic 且字段自洽
func TestRelationsAllGua(t *testing.T) {
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
		if len(c.Relations.ShengKe) != 6 {
			t.Errorf("%s 生克条数 = %d", c.BenGua.Name, len(c.Relations.ShengKe))
		}
		if c.Relations.ShiYing == nil {
			t.Errorf("%s 缺世应关系", c.BenGua.Name)
		}
		// 半合不应与三合重复
		full := map[string]bool{}
		for _, j := range c.Relations.SanHe {
			full[j.Name] = true
		}
		for _, j := range c.Relations.BanHe {
			for _, g := range SanHeGroups {
				if full[g.Name] {
					// 检查该半合的两支是否都在已成三合局中
					if containsZhi(g.Branches, j.Branches) {
						t.Errorf("%s: 半合 %s 与三合局 %s 重复", c.BenGua.Name, j.Name, g.Name)
					}
				}
			}
		}
	}
}

func containsZhi(group [3]string, pair []string) bool {
	if len(pair) != 2 {
		return false
	}
	c := 0
	for _, p := range pair {
		for _, g := range group {
			if p == g {
				c++
			}
		}
	}
	return c == 2
}
