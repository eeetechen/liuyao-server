package divination

import (
	"fmt"
	"sort"
	"strings"
)

// ============================================================
// 五行生克 · 三合 · 六合 · 六冲
// ============================================================

// BranchRef 参与生合冲克关系的一个地支来源。
type BranchRef struct {
	Scope    string `json:"scope"`              // yao / day / month
	Position int    `json:"position,omitempty"` // 爻位（1-6），仅 scope=yao
	Label    string `json:"label"`              // 初爻…上爻 / 日辰 / 月建
	Zhi      string `json:"zhi"`
	WuXing   WuXing `json:"wuXing"`
	Moving   bool   `json:"moving,omitempty"` // 明动
	Active   bool   `json:"active,omitempty"` // 发动（含暗动）
}

// Influence 某来源（月建 / 日辰）对一爻的五行作用。
type Influence struct {
	Source       string `json:"source"` // 月建 / 日辰
	SourceWuXing WuXing `json:"sourceWuXing"`
	Kind         string `json:"kind"`  // 生 / 克 / 比和 / 泄 / 耗
	Label        string `json:"label"` // 得生 / 受克 / 扶 / 泄气 / 耗力
	Desc         string `json:"desc"`
}

// YaoInfluence 单爻所受月建与日辰的五行作用。
type YaoInfluence struct {
	Position int       `json:"position"`
	Label    string    `json:"label"`
	Zhi      string    `json:"zhi"`
	WuXing   WuXing    `json:"wuXing"`
	LiuQin   string    `json:"liuQin"`
	Month    Influence `json:"month"`
	Day      Influence `json:"day"`
}

// ShiYingRelation 世爻与应爻的五行关系。
type ShiYingRelation struct {
	ShiPosition  int    `json:"shiPosition"`
	YingPosition int    `json:"yingPosition"`
	ShiWuXing    WuXing `json:"shiWuXing"`
	YingWuXing   WuXing `json:"yingWuXing"`
	Kind         string `json:"kind"`  // 生 / 克 / 比和 / 泄 / 耗
	Label        string `json:"label"` // 应生世 / 应克世 / 比和 …
	Desc         string `json:"desc"`
}

// BranchPair 地支之间的六合或六冲关系。
type BranchPair struct {
	A    BranchRef `json:"a"`
	B    BranchRef `json:"b"`
	Kind string    `json:"kind"`          // yao-yao / yao-day / yao-month
	Hua  WuXing    `json:"hua,omitempty"` // 六合化气（仅六合有）
	Note string    `json:"note,omitempty"`
}

// SanHeJu 三合局或半合。
type SanHeJu struct {
	Name      string   `json:"name"`      // 申子辰
	Element   WuXing   `json:"element"`   // 水
	Branches  []string `json:"branches"`  // 参与的地支
	Positions []int    `json:"positions"` // 参与的爻位
	WithDay   bool     `json:"withDay"`   // 有日辰参与
	WithMonth bool     `json:"withMonth"` // 有月建参与
	Complete  bool     `json:"complete"`  // 三支俱全
	IsBanHe   bool     `json:"isBanHe"`   // 半合
	HasActive bool     `json:"hasActive"` // 有动爻（含暗动）参与
	Desc      string   `json:"desc"`
}

// Relations 五行生克、三合、六合、六冲之汇总。
type Relations struct {
	ShengKe    []YaoInfluence   `json:"shengKe"`
	ShiYing    *ShiYingRelation `json:"shiYing,omitempty"`
	LiuHe      []BranchPair     `json:"liuHe"`
	LiuChong   []BranchPair     `json:"liuChong"`
	SanHe      []SanHeJu        `json:"sanHe"`
	BanHe      []SanHeJu        `json:"banHe"`
	Xing       []XingPair       `json:"xing"`
	Hai        []HaiPair        `json:"hai"`
	Highlights []string         `json:"highlights"`
}

// ---------- 表 ----------

// SanHeGroups 三合局
var SanHeGroups = []struct {
	Name     string
	Branches [3]string
	Element  WuXing
}{
	{"申子辰", [3]string{"申", "子", "辰"}, Shui},
	{"亥卯未", [3]string{"亥", "卯", "未"}, Mu},
	{"寅午戌", [3]string{"寅", "午", "戌"}, Huo},
	{"巳酉丑", [3]string{"巳", "酉", "丑"}, Jin},
}

// BanHePairs 半合（生半合与墓半合）。
//
// 注意：三合局内另两组（申辰、亥未、寅戌、巳丑）传统称「拱合」，
// 不作半合论，故不列入。
var BanHePairs = []struct {
	A, B, Missing string
	Element       WuXing
}{
	{"申", "子", "辰", Shui}, {"子", "辰", "申", Shui},
	{"亥", "卯", "未", Mu}, {"卯", "未", "亥", Mu},
	{"寅", "午", "戌", Huo}, {"午", "戌", "寅", Huo},
	{"巳", "酉", "丑", Jin}, {"酉", "丑", "巳", Jin},
}

// ---------- 五行生克 ----------

// influenceOf 判断来源五行对目标五行的作用。
func influenceOf(sourceLabel string, sourceWX, targetWX WuXing) Influence {
	in := Influence{Source: sourceLabel, SourceWuXing: sourceWX}
	switch {
	case sourceWX == targetWX:
		in.Kind, in.Label = "比和", "扶"
		in.Desc = fmt.Sprintf("%s%s与爻同为%s，比和相扶", sourceLabel, sourceWX, sourceWX)
	case sheng[sourceWX] == targetWX:
		in.Kind, in.Label = "生", "得生"
		in.Desc = fmt.Sprintf("%s%s生爻%s，得生扶", sourceLabel, sourceWX, targetWX)
	case ke[sourceWX] == targetWX:
		in.Kind, in.Label = "克", "受克"
		in.Desc = fmt.Sprintf("%s%s克爻%s，受克制", sourceLabel, sourceWX, targetWX)
	case sheng[targetWX] == sourceWX:
		in.Kind, in.Label = "泄", "泄气"
		in.Desc = fmt.Sprintf("爻%s生%s%s，泄气", targetWX, sourceLabel, sourceWX)
	default:
		in.Kind, in.Label = "耗", "耗力"
		in.Desc = fmt.Sprintf("爻%s克%s%s，耗力", targetWX, sourceLabel, sourceWX)
	}
	return in
}

// buildShengKe 逐爻给出月建与日辰的五行生克。
func buildShengKe(yaos []YaoInfo, t GanZhiTime) []YaoInfluence {
	monthWX := WuXingOfZhi(t.MonthZhi)
	dayWX := WuXingOfZhi(t.DayZhi)

	out := make([]YaoInfluence, 0, len(yaos))
	for _, y := range yaos {
		out = append(out, YaoInfluence{
			Position: y.Position,
			Label:    y.Name,
			Zhi:      y.Zhi,
			WuXing:   y.WuXing,
			LiuQin:   y.LiuQin,
			Month:    influenceOf("月建", monthWX, y.WuXing),
			Day:      influenceOf("日辰", dayWX, y.WuXing),
		})
	}
	return out
}

// buildShiYing 世应五行关系。
//
// 以「应爻对我（世爻）的作用」立论：应生世为得助，应克世为受制，
// 比和为同心；世生应为我付出，世克应为我制彼。
func buildShiYing(shi, ying YaoInfo) *ShiYingRelation {
	// kind 为机器可读枚举，label 为人读简称
	var kind, label, desc string
	switch {
	case ying.WuXing == shi.WuXing:
		kind, label = "比和", "世应比和"
		desc = fmt.Sprintf("世应皆为%s，彼此比和，主同心协力、事情顺遂", shi.WuXing)
	case sheng[ying.WuXing] == shi.WuXing:
		kind, label = "生", "应生世"
		desc = fmt.Sprintf("应%s生世%s，为得助之象，主对方来就我、事有助力", ying.WuXing, shi.WuXing)
	case ke[ying.WuXing] == shi.WuXing:
		kind, label = "克", "应克世"
		desc = fmt.Sprintf("应%s克世%s，为受制之象，主对方制约于我，须防受阻", ying.WuXing, shi.WuXing)
	case sheng[shi.WuXing] == ying.WuXing:
		kind, label = "泄", "世生应"
		desc = fmt.Sprintf("世%s生应%s，为我付出之象，主我方耗力于彼", shi.WuXing, ying.WuXing)
	default:
		kind, label = "耗", "世克应"
		desc = fmt.Sprintf("世%s克应%s，为我制彼之象，事在掌控", shi.WuXing, ying.WuXing)
	}

	return &ShiYingRelation{
		ShiPosition:  shi.Position,
		YingPosition: ying.Position,
		ShiWuXing:    shi.WuXing,
		YingWuXing:   ying.WuXing,
		Kind:         kind,
		Label:        label,
		Desc:         desc,
	}
}

// ---------- 六合 / 六冲 ----------

func yaoBranchRef(y YaoInfo) BranchRef {
	return BranchRef{
		Scope: "yao", Position: y.Position, Label: y.Name,
		Zhi: y.Zhi, WuXing: y.WuXing, Moving: y.IsChanging, Active: y.IsActive,
	}
}

func timeBranchRef(scope, label, zhi string) BranchRef {
	return BranchRef{Scope: scope, Label: label, Zhi: zhi, WuXing: WuXingOfZhi(zhi)}
}

// buildHeChong 收集六合与六冲关系。
//
// 范围：爻与爻、爻与日辰、爻与月建。日辰与月建之间不论。
func buildHeChong(yaos []YaoInfo, t GanZhiTime) (he []BranchPair, chong []BranchPair) {
	he, chong = []BranchPair{}, []BranchPair{}
	refs := make([]BranchRef, 0, len(yaos)+2)
	for _, y := range yaos {
		refs = append(refs, yaoBranchRef(y))
	}
	refs = append(refs, timeBranchRef("day", "日辰", t.DayZhi))
	refs = append(refs, timeBranchRef("month", "月建", t.MonthZhi))

	kindOf := func(a, b BranchRef) string {
		yaoA, yaoB := a.Scope == "yao", b.Scope == "yao"
		switch {
		case yaoA && yaoB:
			return "yao-yao"
		case yaoA || yaoB:
			other := b
			if yaoB {
				other = a
			}
			return "yao-" + other.Scope
		}
		return ""
	}

	for i := 0; i < len(refs); i++ {
		for j := i + 1; j < len(refs); j++ {
			a, b := refs[i], refs[j]
			k := kindOf(a, b)
			if k == "" {
				continue // 日辰与月建之间不论
			}

			// 六合
			if info, ok := LiuHeZhi[a.Zhi]; ok && info.Partner == b.Zhi {
				note := ""
				switch k {
				case "yao-yao":
					if a.Active || b.Active {
						note = "合处逢动，作合起论"
					} else {
						note = "两静爻相合，作合绊论"
					}
				case "yao-day":
					note = "日辰合爻，谓之合起，爻得日辰维系"
				case "yao-month":
					note = "月建合爻，得月令维系"
				}
				he = append(he, BranchPair{A: a, B: b, Kind: k, Hua: info.Result, Note: note})
			}

			// 六冲
			if LiuChongZhi[a.Zhi] == b.Zhi {
				note := ""
				switch k {
				case "yao-yao":
					if a.Active || b.Active {
						note = "动爻冲静爻，被冲者受伤；皆动则相冲俱伤"
					} else {
						note = "两静爻相冲，静而不动，冲力甚微"
					}
				case "yao-day":
					note = "日辰冲爻：旺相为暗动，休囚为日破"
				case "yao-month":
					note = "月建冲爻，谓之月破，整月无力"
				}
				chong = append(chong, BranchPair{A: a, B: b, Kind: k, Note: note})
			}
		}
	}
	return he, chong
}

// ---------- 三合 / 半合 ----------

// buildSanHe 求三合局与半合。
//
// 参与范围为六爻地支加月建、日辰。三支俱现为三合局；
// 仅现两支且构成生半合 / 墓半合者为半合。
func buildSanHe(yaos []YaoInfo, t GanZhiTime) (full []SanHeJu, ban []SanHeJu) {
	full, ban = []SanHeJu{}, []SanHeJu{}
	// 按地支聚合来源
	byZhi := map[string][]BranchRef{}
	for _, y := range yaos {
		r := yaoBranchRef(y)
		byZhi[r.Zhi] = append(byZhi[r.Zhi], r)
	}
	for _, tr := range []BranchRef{
		timeBranchRef("day", "日辰", t.DayZhi),
		timeBranchRef("month", "月建", t.MonthZhi),
	} {
		byZhi[tr.Zhi] = append(byZhi[tr.Zhi], tr)
	}

	present := func(z string) bool { return len(byZhi[z]) > 0 }

	// 三合局
	for _, g := range SanHeGroups {
		if !present(g.Branches[0]) || !present(g.Branches[1]) || !present(g.Branches[2]) {
			continue
		}
		ju := SanHeJu{
			Name: g.Name, Element: g.Element, Complete: true,
			Branches: []string{g.Branches[0], g.Branches[1], g.Branches[2]},
		}
		for _, z := range g.Branches {
			for _, r := range byZhi[z] {
				switch r.Scope {
				case "yao":
					ju.Positions = append(ju.Positions, r.Position)
					if r.Active {
						ju.HasActive = true
					}
				case "day":
					ju.WithDay = true
				case "month":
					ju.WithMonth = true
				}
			}
		}
		sort.Ints(ju.Positions)
		ju.Desc = fmt.Sprintf("%s 三合%s局", g.Name, g.Element)
		if len(ju.Positions) > 0 {
			ju.Desc = fmt.Sprintf("%s（%s）三合%s局", positionNames(ju.Positions), g.Name, g.Element)
		}
		if ju.HasActive {
			ju.Desc += "，卦中有动爻"
		} else if ju.WithDay || ju.WithMonth {
			ju.Desc += "，含日月"
		} else {
			ju.Desc += "，六爻皆静"
		}
		full = append(full, ju)
	}

	// 半合：仅在三合局未成时论
	for _, p := range BanHePairs {
		if !present(p.A) || !present(p.B) {
			continue
		}
		if present(p.Missing) {
			continue // 三支俱全，已作三合局，不再论半合
		}
		pa, pb := byZhi[p.A], byZhi[p.B]
		ju := SanHeJu{
			Name: p.A + p.B, Element: p.Element, Complete: false, IsBanHe: true,
			Branches: []string{p.A, p.B},
		}
		for _, r := range append(append([]BranchRef{}, pa...), pb...) {
			switch r.Scope {
			case "yao":
				ju.Positions = append(ju.Positions, r.Position)
				if r.Active {
					ju.HasActive = true
				}
			case "day":
				ju.WithDay = true
			case "month":
				ju.WithMonth = true
			}
		}
		sort.Ints(ju.Positions)
		ju.Desc = fmt.Sprintf("%s%s 半合%s局（缺%s）", p.A, p.B, p.Element, p.Missing)
		if len(ju.Positions) > 0 {
			ju.Desc = fmt.Sprintf("%s（%s%s）半合%s局，缺%s", positionNames(ju.Positions), p.A, p.B, p.Element, p.Missing)
		}
		ban = append(ban, ju)
	}
	return full, ban
}

// ---------- 汇总提示 ----------

func relationHighlights(yaos []YaoInfo, r Relations) []string {
	var out []string

	// 三合局（Desc 已含成局状态，此处仅补断语，不重复）
	for _, j := range r.SanHe {
		out = append(out, fmt.Sprintf("三合局：%s。三合成局主力量汇聚、事情有合力，%s。",
			j.Desc, sanHeAdvice(j)))
	}
	for _, j := range r.BanHe {
		out = append(out, fmt.Sprintf("半合：%s。半合力量弱于三合，主有合意而未全成，须待补足第三支方成局。", j.Desc))
	}

	// 三刑六害
	if len(r.Xing) > 0 {
		out = append(out, fmt.Sprintf("三刑：%s。%s", describeXing(r.Xing), XingAdvice()))
	}
	if len(r.Hai) > 0 {
		out = append(out, fmt.Sprintf("六害：%s。%s", describeHai(r.Hai), HaiAdvice()))
	}

	// 世应（Desc 已含结论，不再附加，避免语义重复）
	if r.ShiYing != nil {
		out = append(out, fmt.Sprintf("世应关系：%s（%s）。", r.ShiYing.Desc, r.ShiYing.Label))
	}

	// 六合：日辰/月建合爻为「合起」，具实际意义，全部保留；
	// 爻间相合仅列有动爻参与者，避免两静爻相合之噪音。
	for _, p := range r.LiuHe {
		if p.Kind != "yao-yao" || p.A.Active || p.B.Active {
			out = append(out, fmt.Sprintf("%s六合：%s%s 与 %s%s 相合（合化%s）。%s",
				kindLabel(p.Kind), p.A.Label, p.A.Zhi, p.B.Label, p.B.Zhi, p.Hua, p.Note))
		}
	}

	// 六冲
	//
	// 爻与月建相冲（月破）、爻与日辰相冲（暗动/日破）已由总结中的
	// 旺衰·动静部分给出，此处不再重复，以免同一事实反复陈述。
	// 仅补充爻与爻之间的相冲（且涉及动爻者方有实际冲力）。
	for _, p := range r.LiuChong {
		if p.Kind == "yao-yao" && (p.A.Active || p.B.Active) {
			out = append(out, fmt.Sprintf("爻间六冲：%s%s 与 %s%s 相冲。%s",
				p.A.Label, p.A.Zhi, p.B.Label, p.B.Zhi, p.Note))
		}
	}

	return out
}

// sanHeAdvice 给出成局与否的断语（与 Desc 中的状态描述互补，不重复）。
func sanHeAdvice(j SanHeJu) string {
	if j.HasActive {
		return "卦中有动爻引动，成局最速"
	}
	if j.WithDay || j.WithMonth {
		return "得日月引动，亦可成局"
	}
	return "然六爻皆静，须待岁运冲合之日方成"
}

// describeXing 汇总三刑描述（多个时合并为一句）
func describeXing(list []XingPair) string {
	parts := make([]string, 0, len(list))
	for _, x := range list {
		parts = append(parts, x.Desc)
	}
	return strings.Join(parts, "；")
}

// describeHai 汇总六害描述
func describeHai(list []HaiPair) string {
	parts := make([]string, 0, len(list))
	for _, h := range list {
		parts = append(parts, fmt.Sprintf("%s%s与%s%s", h.A.Label, h.A.Zhi, h.B.Label, h.B.Zhi))
	}
	return strings.Join(parts, "、")
}

func kindLabel(kind string) string {
	switch kind {
	case "yao-day":
		return "日辰"
	case "yao-month":
		return "月建"
	default:
		return "爻间"
	}
}

// ---------- 入口 ----------

// BuildRelations 计算五行生克、三合、六合、六冲。
func BuildRelations(yaos []YaoInfo, t GanZhiTime, shiPos, yingPos int) Relations {
	// 各切片一律初始化为空切片，确保 JSON 序列化为 [] 而非 null
	r := Relations{
		ShengKe:    []YaoInfluence{},
		LiuHe:      []BranchPair{},
		LiuChong:   []BranchPair{},
		SanHe:      []SanHeJu{},
		BanHe:      []SanHeJu{},
		Xing:       []XingPair{},
		Hai:        []HaiPair{},
		Highlights: []string{},
	}

	r.ShengKe = buildShengKe(yaos, t)

	if shiPos >= 1 && shiPos <= len(yaos) && yingPos >= 1 && yingPos <= len(yaos) {
		r.ShiYing = buildShiYing(yaos[shiPos-1], yaos[yingPos-1])
	}

	r.LiuHe, r.LiuChong = buildHeChong(yaos, t)
	r.SanHe, r.BanHe = buildSanHe(yaos, t)
	r.Xing, r.Hai = BuildXingHai(yaos, t)
	r.Highlights = relationHighlights(yaos, r)

	return r
}
