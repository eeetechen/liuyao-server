package divination

import "testing"

// ---------- 干支五行 ----------

func TestTianGanWuXing(t *testing.T) {
	want := map[string]WuXing{
		"甲": Mu, "乙": Mu, "丙": Huo, "丁": Huo, "戊": Tu,
		"己": Tu, "庚": Jin, "辛": Jin, "壬": Shui, "癸": Shui,
	}
	if len(TianGanWuXing) != 10 {
		t.Fatalf("天干五行表应有 10 项，实际 %d", len(TianGanWuXing))
	}
	for g, w := range want {
		if TianGanWuXing[g] != w {
			t.Errorf("天干 %s 应为 %s，实际 %s", g, w, TianGanWuXing[g])
		}
	}
}

// 天干阴阳：甲丙戊庚壬阳，乙丁己辛癸阴
func TestTianGanYinYang(t *testing.T) {
	for i, g := range TianGan {
		want := Yin
		if i%2 == 0 {
			want = Yang
		}
		if TianGanYinYang[g] != want {
			t.Errorf("天干 %s 阴阳应为 %s，实际 %s", g, want, TianGanYinYang[g])
		}
	}
}

// 地支阴阳交替，且与生肖一一对应
func TestDiZhiYinYangAndShengXiao(t *testing.T) {
	if len(ShengXiao) != 12 {
		t.Fatalf("生肖应有 12 项，实际 %d", len(ShengXiao))
	}
	for i, z := range DiZhi {
		want := Yin
		if i%2 == 0 {
			want = Yang
		}
		if ZhiYinYang[z] != want {
			t.Errorf("地支 %s 阴阳应为 %s，实际 %s", z, want, ZhiYinYang[z])
		}
	}
	// 抽查常见对应
	cases := map[string]string{"子": "鼠", "卯": "兔", "午": "马", "酉": "鸡", "亥": "猪"}
	for z, sx := range cases {
		if ShengXiao[zhiIndex(z)] != sx {
			t.Errorf("%s 的生肖应为 %s，实际 %s", z, sx, ShengXiao[zhiIndex(z)])
		}
	}
}

// 参考表中的干支条目须与底层常量一致
func TestReferenceGanZhiConsistency(t *testing.T) {
	ref := BuildReference()
	if len(ref.TianGan) != 10 {
		t.Errorf("参考表天干 = %d, want 10", len(ref.TianGan))
	}
	if len(ref.DiZhi) != 12 {
		t.Errorf("参考表地支 = %d, want 12", len(ref.DiZhi))
	}
	for _, g := range ref.TianGan {
		if g.WuXing != TianGanWuXing[g.Name] {
			t.Errorf("参考表天干 %s 五行不符", g.Name)
		}
		if g.YinYang != yinYangCn(TianGanYinYang[g.Name]) {
			t.Errorf("参考表天干 %s 阴阳不符: %s", g.Name, g.YinYang)
		}
		if g.YinYangEn == "" {
			t.Errorf("参考表天干 %s 缺 yinYangEn", g.Name)
		}
	}
	for _, g := range ref.DiZhi {
		if g.WuXing != WuXingOfZhi(g.Name) {
			t.Errorf("参考表地支 %s 五行不符", g.Name)
		}
		if g.ShengXiao != ShengXiao[zhiIndex(g.Name)] {
			t.Errorf("参考表地支 %s 生肖不符", g.Name)
		}
	}
}

// ---------- 参考表完整性 ----------

func TestReferenceShengKe(t *testing.T) {
	ref := BuildReference()
	// 相生 5 条 + 相克 5 条
	if len(ref.ShengKe) != 10 {
		t.Fatalf("生克条目 = %d, want 10", len(ref.ShengKe))
	}
	// 必须包含「水克火」与「水生木」
	texts := map[string]bool{}
	for _, x := range ref.ShengKe {
		texts[x.Text] = true
		if x.Plain == "" {
			t.Errorf("%s 缺白话解释", x.Text)
		}
		if x.Kind != "生" && x.Kind != "克" {
			t.Errorf("%s 类型非法: %s", x.Text, x.Kind)
		}
	}
	for _, want := range []string{"水克火", "水生木", "金克木", "木生火"} {
		if !texts[want] {
			t.Errorf("参考表缺 %s", want)
		}
	}
}

func TestReferenceWuXingAttrs(t *testing.T) {
	ref := BuildReference()
	if len(ref.WuXing) != 5 {
		t.Fatalf("五行条目 = %d, want 5", len(ref.WuXing))
	}
	for _, w := range ref.WuXing {
		if w.Color == "" || w.Direction == "" || w.Season == "" {
			t.Errorf("%s 缺属性", w.WuXing)
		}
		// 生克指向必须自洽
		if sheng[w.WuXing] != w.Sheng {
			t.Errorf("%s 的「生」应为 %s", w.WuXing, sheng[w.WuXing])
		}
		if ke[w.WuXing] != w.Ke {
			t.Errorf("%s 的「克」应为 %s", w.WuXing, ke[w.WuXing])
		}
		if sheng[w.BeiSheng] != w.WuXing {
			t.Errorf("%s 的「生我」%s 不正确", w.WuXing, w.BeiSheng)
		}
		if ke[w.BeiKe] != w.WuXing {
			t.Errorf("%s 的「克我」%s 不正确", w.WuXing, w.BeiKe)
		}
	}
}

func TestReferencePairs(t *testing.T) {
	ref := BuildReference()
	// 六冲 6 对、六合 6 对、三合 4 局
	if len(ref.LiuChong) != 6 {
		t.Errorf("六冲 = %d, want 6", len(ref.LiuChong))
	}
	if len(ref.LiuHe) != 6 {
		t.Errorf("六合 = %d, want 6", len(ref.LiuHe))
	}
	if len(ref.SanHe) != 4 {
		t.Errorf("三合 = %d, want 4", len(ref.SanHe))
	}
	// 半合 8 对，且两两不重复
	if len(ref.BanHe) != 8 {
		t.Errorf("半合 = %d, want 8", len(ref.BanHe))
	}
	seen := map[string]bool{}
	for _, p := range ref.BanHe {
		k1, k2 := p.A+p.B, p.B+p.A
		if seen[k1] || seen[k2] {
			t.Errorf("半合 %s 重复", p.Label)
		}
		seen[k1] = true
	}
	// 六冲两两不重复
	seenC := map[string]bool{}
	for _, p := range ref.LiuChong {
		k1, k2 := p.A+p.B, p.B+p.A
		if seenC[k1] || seenC[k2] {
			t.Errorf("六冲 %s 重复", p.Label)
		}
		seenC[k1] = true
	}
	// 六合必须带合化五行
	for _, p := range ref.LiuHe {
		if p.Hua == "" {
			t.Errorf("六合 %s 缺合化五行", p.Label)
		}
	}
	// 三合三支必须同属该局
	for _, j := range ref.SanHe {
		for _, b := range j.Branches {
			if WuXingOfZhi(b) != j.Element {
				// 三合局的三支并非都属该五行（如申金、子水、辰土合成水局），
				// 故此处仅校验支数与非空
				_ = b
			}
		}
		if len(j.Branches) != 3 {
			t.Errorf("%s 分支数 = %d, want 3", j.Name, len(j.Branches))
		}
		if len(j.Colors) != 3 {
			t.Errorf("%s 颜色数 = %d, want 3", j.Name, len(j.Colors))
		}
	}
}

func TestReferenceNotes(t *testing.T) {
	ref := BuildReference()
	if len(ref.Notes) == 0 {
		t.Error("参考表应有说明文字")
	}
}

// ---------- 白话解析 ----------

func TestPlainBasic(t *testing.T) {
	c := mustChart(t, []YaoValue{7, 7, 7, 7, 7, 7}, fixedTime("子", "子", "甲", "戌亥"))
	p := c.Plain
	if p.Overview == "" {
		t.Fatal("总览不应为空")
	}
	if len(p.Sections) == 0 {
		t.Fatal("白话小节不应为空")
	}
	if len(p.Suggestions) == 0 {
		t.Fatal("建议不应为空")
	}
	// 总览应包含卦名与所问之事
	if !contains(p.Overview, c.BenGua.Name) {
		t.Errorf("总览应含卦名 %s: %s", c.BenGua.Name, p.Overview)
	}
	if !contains(p.Overview, c.Question) {
		t.Errorf("总览应含所问之事: %s", p.Overview)
	}
	// 每节须有标题与内容
	for _, s := range p.Sections {
		if s.Title == "" {
			t.Error("小节缺标题")
		}
		if len(s.Lines) == 0 {
			t.Errorf("小节「%s」无内容", s.Title)
		}
		for _, l := range s.Lines {
			if l == "" {
				t.Errorf("小节「%s」含空行", s.Title)
			}
		}
	}
}

// 白话须提到年份标注
func TestPlainMentionsYear(t *testing.T) {
	tm := fixedTime("子", "子", "甲", "戌亥")
	tm.Year = "丙午"
	tm.YearLabel = "丙午年（马年）"
	tm.MonthLabel = "月柱 甲午（月建子）"
	tm.DayLabel = "日柱 甲子（日辰子）"
	tm.HourLabel = "时柱 庚午"
	c := mustChart(t, []YaoValue{7, 7, 7, 7, 7, 7}, tm)

	joined := ""
	for _, s := range c.Plain.Sections {
		joined += s.Title + "\n"
		for _, l := range s.Lines {
			joined += l + "\n"
		}
	}
	if !contains(joined, "丙午年") {
		t.Error("白话中应出现年份标注 丙午年")
	}
	if !contains(joined, "马年") {
		t.Error("白话中应出现生肖")
	}
	if !contains(joined, "立春") {
		t.Error("白话中应说明年份以立春为界")
	}
}

// 有动爻时总览须说明变卦方向
func TestPlainWithMoving(t *testing.T) {
	c := mustChart(t, []YaoValue{9, 9, 9, 9, 9, 9}, fixedTime("子", "子", "甲", "戌亥"))
	if !contains(c.Plain.Overview, "变") && !contains(c.Plain.Overview, "发动") {
		t.Errorf("有动爻的总览应提到变化: %s", c.Plain.Overview)
	}
	if c.BianGua == nil {
		t.Fatal("全动应有变卦")
	}
	if !contains(c.Plain.Overview, c.BianGua.Name) {
		t.Errorf("总览应提到变卦名 %s: %s", c.BianGua.Name, c.Plain.Overview)
	}
}

// 六冲卦的白话须解释六冲
func TestPlainLiuChongExplanation(t *testing.T) {
	c := mustChart(t, []YaoValue{7, 7, 7, 7, 7, 7}, fixedTime("子", "子", "甲", "戌亥"))
	joined := ""
	for _, s := range c.Plain.Sections {
		for _, l := range s.Lines {
			joined += l
		}
	}
	if !contains(joined, "六冲") {
		t.Error("六冲卦应解释六冲")
	}
	if !contains(joined, "容易散") && !contains(joined, "变化快") {
		t.Error("六冲解释应通俗易懂")
	}
}

// 暗动须在白话中说明
func TestPlainAnDong(t *testing.T) {
	c := mustChart(t, []YaoValue{8, 8, 8, 8, 8, 8}, fixedTime("丑", "未", "己", "戌亥"))
	joined := ""
	for _, s := range c.Plain.Sections {
		for _, l := range s.Lines {
			joined += l
		}
	}
	if c.Summary.AnDongCount > 0 && !contains(joined, "暗动") {
		t.Error("有暗动时白话应说明暗动")
	}
}

// 白话须包含免责声明（面向朋友使用的场景）
func TestPlainDisclaimer(t *testing.T) {
	c := mustChart(t, []YaoValue{7, 7, 7, 7, 7, 7}, fixedTime("子", "子", "甲", "戌亥"))
	last := c.Plain.Suggestions[len(c.Plain.Suggestions)-1]
	if !contains(last, "仅供") && !contains(last, "参考") {
		t.Errorf("末条建议应为免责说明，实际: %s", last)
	}
}

// 64 卦全量：白话生成不 panic 且结构完整
func TestPlainAllGua(t *testing.T) {
	tm := fixedTime("寅", "子", "甲", "戌亥")
	tm.Year = "丙午"
	tm.YearLabel = "丙午年（马年）"
	tm.MonthLabel = "月柱 庚寅（月建寅）"
	tm.DayLabel = "日柱 甲子（日辰子）"
	tm.HourLabel = "时柱 甲子"
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
		if c.Plain.Overview == "" {
			t.Errorf("%s 白话总览为空", c.BenGua.Name)
		}
		if len(c.Plain.Sections) == 0 {
			t.Errorf("%s 白话小节为空", c.BenGua.Name)
		}
		if len(c.Plain.Suggestions) == 0 {
			t.Errorf("%s 白话建议为空", c.BenGua.Name)
		}
	}
}

// 世应爻位名称
func TestGuaInfoPosNames(t *testing.T) {
	g := GuaInfo{ShiPos: 6, YingPos: 3}
	if g.ShiPosName() != "上爻" {
		t.Errorf("世位6 应为上爻，实际 %s", g.ShiPosName())
	}
	if g.YingPosName() != "三爻" {
		t.Errorf("应位3 应为三爻，实际 %s", g.YingPosName())
	}
	// 越界不应 panic
	g2 := GuaInfo{ShiPos: 0, YingPos: 99}
	_ = g2.ShiPosName()
	_ = g2.YingPosName()
}

// 总览不应出现「爻 爻」之类的重复字
func TestPlainNoDuplicatedYao(t *testing.T) {
	tm := fixedTime("子", "子", "甲", "戌亥")
	for _, vals := range [][]YaoValue{
		{9, 7, 7, 7, 7, 7}, {7, 7, 7, 7, 7, 7},
		{9, 9, 9, 9, 9, 9}, {8, 8, 8, 8, 8, 8},
	} {
		c := mustChart(t, vals, tm)
		if contains(c.Plain.Overview, "爻 爻") {
			t.Errorf("总览出现重复的「爻」: %s", c.Plain.Overview)
		}
		if contains(c.Plain.Overview, "爻爻") {
			t.Errorf("总览出现重复的「爻」: %s", c.Plain.Overview)
		}
	}
}

// 白话小节标题的序号必须唯一且递增
func TestPlainSectionNumberingUnique(t *testing.T) {
	c := mustChart(t, []YaoValue{7, 7, 7, 7, 7, 7}, fixedTime("子", "子", "甲", "戌亥"))
	seen := map[string]bool{}
	for _, s := range c.Plain.Sections {
		if s.Title == "" {
			continue
		}
		no := string([]rune(s.Title)[0])
		if seen[no] {
			t.Errorf("小节序号「%s」重复: %v", no, titlesOf(c.Plain.Sections))
		}
		seen[no] = true
	}
	// 必须连续（从「一」开始）
	order := "一二三四五六七八九十"
	for i, s := range c.Plain.Sections {
		if s.Title == "" {
			continue
		}
		want := string([]rune(order)[i])
		got := string([]rune(s.Title)[0])
		if got != want {
			t.Errorf("第%d节的序号应为 %s，实际 %s（%s）", i+1, want, got, s.Title)
		}
	}
}

func titlesOf(secs []PlainSection) []string {
	out := make([]string, 0, len(secs))
	for _, s := range secs {
		out = append(out, s.Title)
	}
	return out
}

// ---------- 六神意象参考表 ----------

// 六神五行与起法必须与底层常量及《起身例纳甲·六神总论》一致
func TestLiuShenReference(t *testing.T) {
	ref := BuildReference()
	if len(ref.LiuShen) != 6 {
		t.Fatalf("六神应为 6 条，实际 %d", len(ref.LiuShen))
	}

	// 五行归属（据《起身例纳甲·六神总论》）
	wantWuXing := map[string]WuXing{
		"青龙": Mu, "朱雀": Huo, "勾陈": Tu,
		"螣蛇": Tu, "白虎": Jin, "玄武": Shui,
	}
	// 起初爻的日干
	wantGan := map[string][]string{
		"青龙": {"甲", "乙"}, "朱雀": {"丙", "丁"}, "勾陈": {"戊"},
		"螣蛇": {"己"}, "白虎": {"庚", "辛"}, "玄武": {"壬", "癸"},
	}

	seen := map[string]bool{}
	for _, ls := range ref.LiuShen {
		seen[ls.Name] = true

		if w, ok := wantWuXing[ls.Name]; !ok {
			t.Errorf("非预期的六神名: %s", ls.Name)
		} else if ls.WuXing != w {
			t.Errorf("%s 五行 = %s, want %s", ls.Name, ls.WuXing, w)
		}

		// 起法须与 wantGan 一致，且与底层 LiuShenStart 自洽
		want := wantGan[ls.Name]
		if len(ls.StartGan) != len(want) {
			t.Errorf("%s 日干数 = %d, want %d", ls.Name, len(ls.StartGan), len(want))
		}
		for _, g := range ls.StartGan {
			idx := LiuShenStart(g)
			if LiuShen[idx] != ls.Name {
				t.Errorf("%s 声称 %s 日起初爻，但 GetLiuShen(%s)[0] = %s",
					ls.Name, g, g, LiuShen[idx])
			}
		}

		if ls.Image == nil || len(ls.Image) == 0 {
			t.Errorf("%s 缺意象条目", ls.Name)
		}
		if ls.Meaning == "" {
			t.Errorf("%s 缺总括释义", ls.Name)
		}
		if ls.Nature == "" {
			t.Errorf("%s 缺吉凶倾向", ls.Name)
		}
		if ls.Color == "" {
			t.Errorf("%s 缺对应色", ls.Name)
		}
	}
	// 六个名字齐备
	for _, n := range LiuShen {
		if !seen[n] {
			t.Errorf("参考表缺六神 %s", n)
		}
	}
}

// 六神五行不得与「十二将」体系混淆：螣蛇在本体系为土，非火
func TestLiuShenTengSheIsEarth(t *testing.T) {
	ref := BuildReference()
	for _, ls := range ref.LiuShen {
		if ls.Name == "螣蛇" {
			if ls.WuXing != Tu {
				t.Errorf("螣蛇在六爻体系应属土，实际 %s", ls.WuXing)
			}
			// 须在释义中提示与十二将体系的区别
			if !contains(ls.Note, "虚惊") && !contains(ls.Meaning, "虚惊") {
				t.Error("螣蛇释义应点出「虚惊」这一主要意象")
			}
		}
	}
}

// 六神五行须与五行相生相克表自洽（每种五行都必须是合法五行）
func TestLiuShenWuXingValid(t *testing.T) {
	ref := BuildReference()
	for _, ls := range ref.LiuShen {
		if indexOf(wuXingStrings(), string(ls.WuXing)) < 0 {
			t.Errorf("%s 的五行 %s 非法", ls.Name, ls.WuXing)
		}
		// 对应色应与五行色一致
		if want := WuXingColor[ls.WuXing]; ls.Color != want {
			t.Errorf("%s(%s) 的对应色 = %s, want %s", ls.Name, ls.WuXing, ls.Color, want)
		}
	}
}

// 六神吉凶倾向：青龙为吉、白虎玄武为凶
func TestLiuShenNature(t *testing.T) {
	ref := BuildReference()
	byName := map[string]RefLiuShen{}
	for _, ls := range ref.LiuShen {
		byName[ls.Name] = ls
	}
	if byName["青龙"].Nature != "吉" {
		t.Errorf("青龙应为吉，实际 %s", byName["青龙"].Nature)
	}
	for _, n := range []string{"白虎", "玄武", "勾陈", "螣蛇"} {
		if byName[n].Nature != "凶" {
			t.Errorf("%s 应为凶，实际 %s", n, byName[n].Nature)
		}
	}
}
