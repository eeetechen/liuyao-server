package divination

import (
	"fmt"
	"strings"
)

// YaoValue 爻值（传统六爻记法）
//
//	6 = 老阴（动，阴变阳）  7 = 少阳（静）
//	8 = 少阴（静）          9 = 老阳（动，阳变阴）
type YaoValue int

const (
	LaoYin   YaoValue = 6
	ShaoYang YaoValue = 7
	ShaoYin  YaoValue = 8
	LaoYang  YaoValue = 9
)

// IsChanging 该爻是否为动爻
func (v YaoValue) IsChanging() bool { return v == LaoYin || v == LaoYang }

// YinYang 该爻的阴阳
func (v YaoValue) YinYang() YinYang {
	if v == ShaoYang || v == LaoYang {
		return Yang
	}
	return Yin
}

// ---------- 输入 ----------

// CastMethod 起卦方式
type CastMethod string

const (
	MethodCoin   CastMethod = "coin"   // 铜钱摇卦（爻值 6/7/8/9）
	MethodManual CastMethod = "manual" // 手动指定阴阳
)

// ChartInput 排盘输入
type ChartInput struct {
	YaoValues []YaoValue // 六爻，自下而上，长度必须为 6
	Question  string     // 所问之事
	Method    CastMethod
	// YongShen 指定用神六亲（父母/兄弟/子孙/妻财/官鬼）。
	// 留空则由 Question 自动推断。
	YongShen string
}

// ---------- 输出 ----------

// YaoInfo 单爻完整信息
type YaoInfo struct {
	Position   int      `json:"position"`   // 爻位 1-6（自下而上）
	Name       string   `json:"name"`       // 爻位名（初/二/三/四/五/上）
	Value      YaoValue `json:"value"`      // 原始爻值
	YinYang    YinYang  `json:"yinYang"`    // 阴阳
	Symbol     string   `json:"symbol"`     // 爻符号
	IsChanging bool     `json:"isChanging"` // 是否明动

	GanZhi string `json:"ganZhi"` // 纳甲干支
	Gan    string `json:"gan"`
	Zhi    string `json:"zhi"`
	WuXing WuXing `json:"wuXing"`

	LiuQin  string `json:"liuQin"`  // 六亲
	LiuShen string `json:"liuShen"` // 六神
	IsShi   bool   `json:"isShi"`   // 是否世爻
	IsYing  bool   `json:"isYing"`  // 是否应爻
	IsKong  bool   `json:"isKong"`  // 是否旬空

	WangShuai     WangShuaiState `json:"wangShuai"`     // 旺相休囚死
	WangShuaiDesc string         `json:"wangShuaiDesc"` // 释义
	Special       SpecialStatus  `json:"special"`       // 暗动/日破/月破
	SpecialLabel  string         `json:"specialLabel"`  // 中文标签
	MoveState     MoveState      `json:"moveState"`     // 动静
	MoveLabel     string         `json:"moveLabel"`     // 动静中文
	Strength      StrengthLevel  `json:"strength"`      // 综合力量
	IsActive      bool           `json:"isActive"`      // 是否发动（明动或暗动）
	Evidence      []string       `json:"evidence"`      // 判定依据

	ChangedYinYang YinYang `json:"changedYinYang,omitempty"` // 变卦阴阳
	ChangedGanZhi  string  `json:"changedGanZhi,omitempty"`  // 变爻纳甲
	ChangedLiuQin  string  `json:"changedLiuQin,omitempty"`  // 变爻六亲
	HuaType        string  `json:"huaType,omitempty"`        // 化进化退/回头生克
}

// GuaInfo 卦信息
type GuaInfo struct {
	Mark     string `json:"mark"`           // 六爻二进制码
	Name     string `json:"name"`           // 卦名
	Gong     string `json:"gong"`           // 卦宫
	Element  WuXing `json:"element"`        // 卦宫五行
	Type     string `json:"type,omitempty"` // 六冲/六合/游魂/归魂
	ShiPos   int    `json:"shiPos"`         // 世爻位
	YingPos  int    `json:"yingPos"`        // 应爻位
	UpperGua string `json:"upperGua"`       // 上卦
	LowerGua string `json:"lowerGua"`       // 下卦
}

// GanZhiTime 四柱与旬空
type GanZhiTime struct {
	Year     string `json:"year"`  // 年柱，如「丙午」
	Month    string `json:"month"` // 月柱（节气月）
	Day      string `json:"day"`   // 日柱
	Hour     string `json:"hour"`  // 时柱
	YearGan  string `json:"yearGan"`
	YearZhi  string `json:"yearZhi"`
	DayGan   string `json:"dayGan"`
	DayZhi   string `json:"dayZhi"`
	MonthZhi string `json:"monthZhi"`
	// YearLabel 便于阅读的年份标注，如「丙午年（马年）」
	YearLabel string `json:"yearLabel"`
	// MonthLabel / DayLabel / HourLabel 带柱别说明，如「月柱 甲午（月建午）」
	MonthLabel string `json:"monthLabel"`
	DayLabel   string `json:"dayLabel"`
	HourLabel  string `json:"hourLabel"`
	XunKong    string `json:"xunKong"`
	SolarText  string `json:"solarText"`
	LunarText  string `json:"lunarText"`
}

// Summary 结果摘要（便于前端突出显示）
type Summary struct {
	BenGuaName  string   `json:"benGuaName"`
	BianGuaName string   `json:"bianGuaName,omitempty"`
	Gong        string   `json:"gong"`
	GuaType     string   `json:"guaType,omitempty"`     // 本卦性质：六冲/六合
	BianGuaType string   `json:"bianGuaType,omitempty"` // 变卦性质
	IsLiuChong  bool     `json:"isLiuChong"`            // 本卦是否六冲
	IsLiuHe     bool     `json:"isLiuHe"`               // 本卦是否六合
	MovingCount int      `json:"movingCount"`           // 明动爻数
	AnDongCount int      `json:"anDongCount"`           // 暗动爻数
	AnDongYaos  []int    `json:"anDongYaos"`            // 暗动爻位
	BrokenYaos  []int    `json:"brokenYaos"`            // 日破/月破爻位
	KongYaos    []int    `json:"kongYaos"`              // 旬空爻位
	StrongYaos  []int    `json:"strongYaos"`            // 有力爻位
	Highlights  []string `json:"highlights"`            // 重点提示（含六冲/暗动/旺相）
}

// Chart 完整排盘结果
type Chart struct {
	Question  string       `json:"question"`
	Method    CastMethod   `json:"method"`
	Time      GanZhiTime   `json:"time"`
	BenGua    GuaInfo      `json:"benGua"`
	BianGua   *BianGua     `json:"bianGua,omitempty"`
	Yaos      []YaoInfo    `json:"yaos"` // 自下而上 6 爻
	FuShen    []FuShenInfo `json:"fuShen,omitempty"`
	Relations Relations    `json:"relations"`
	Shen      ShenSystem   `json:"shen"`
	Plain     Plain        `json:"plain"`
	Summary   Summary      `json:"summary"`
}

// ---------- 排盘主逻辑 ----------

var yaoPosName = [6]string{"初爻", "二爻", "三爻", "四爻", "五爻", "上爻"}

// BuildChart 依输入与时间信息排盘。
// t 由 calendar 层提供（四柱、旬空、月建、日辰）。
func BuildChart(in ChartInput, t GanZhiTime) (*Chart, error) {
	if len(in.YaoValues) != 6 {
		return nil, fmt.Errorf("需要 6 个爻值（自下而上），实际收到 %d 个", len(in.YaoValues))
	}
	for i, v := range in.YaoValues {
		if v != LaoYin && v != ShaoYang && v != ShaoYin && v != LaoYang {
			return nil, fmt.Errorf("第 %d 爻值非法：%d（应为 6/7/8/9）", i+1, v)
		}
	}

	// 1. 本卦二进制码
	var sb strings.Builder
	var changing []int
	for i, v := range in.YaoValues {
		if v.YinYang() == Yang {
			sb.WriteByte('1')
		} else {
			sb.WriteByte('0')
		}
		if v.IsChanging() {
			changing = append(changing, i)
		}
	}
	mark := sb.String()

	// 2. 卦宫、世应、卦名
	shi := setShiYao(mark)
	gong := palace(mark, shi.Shi)
	gongIdx := indexOf(BaGua[:], gong)
	if gongIdx < 0 {
		return nil, fmt.Errorf("无法判定卦宫：%s", mark)
	}
	gongWuXing := BaGuaWuXing[gongIdx]

	benGua := GuaInfo{
		Mark:     mark,
		Name:     Gua64[mark],
		Gong:     gong,
		Element:  gongWuXing,
		ShiPos:   shi.Shi,
		YingPos:  shi.Ying,
		UpperGua: BaGuaByBinary(mark[3:]),
		LowerGua: BaGuaByBinary(mark[:3]),
	}
	switch {
	case IsLiuChong(mark):
		benGua.Type = "六冲"
	case IsLiuHe(mark):
		benGua.Type = "六合"
	default:
		benGua.Type = soul(mark)
	}

	// 3. 纳甲、变卦
	naJia := getNaJia(mark)
	bian := calcBianGua(mark, changing)

	// 变卦纳甲（用于化进化退、回头生克判断）
	var bianNaJia []NaJia
	if bian != nil {
		bianNaJia = bian.NaJia
	}

	// 4. 六神
	shen := GetLiuShen(t.DayGan)

	// 5. 逐爻组装
	yaos := make([]YaoInfo, 0, 6)
	presentQin := map[string]bool{}

	for i := 0; i < 6; i++ {
		v := in.YaoValues[i]
		nj := naJia[i]
		isChanging := v.IsChanging()

		// 六亲（以本卦卦宫五行为我）
		qin := GetLiuQin(gongWuXing, nj.WuXing)
		presentQin[qin] = true

		// 旺衰：以月建为准
		ws := GetWangShuai(nj.WuXing, t.MonthZhi)

		// 旬空
		isKong := IsKong(nj.Zhi, t.XunKong)

		// 暗动 / 日破 / 月破
		sp, evidence := JudgeSpecialStatus(nj.Zhi, t.MonthZhi, t.DayZhi, isChanging, ws)

		// 日辰生扶（用于力量评级）
		daySheng := sheng[WuXingOfZhi(t.DayZhi)] == nj.WuXing

		moveState := ResolveMoveState(isChanging, sp)
		strength := GetStrengthLevel(ws, sp, isKong, daySheng)
		active := IsActive(isChanging, sp)

		yi := YaoInfo{
			Position:      i + 1,
			Name:          yaoPosName[i],
			Value:         v,
			YinYang:       v.YinYang(),
			IsChanging:    isChanging,
			GanZhi:        nj.GanZhi,
			Gan:           nj.Gan,
			Zhi:           nj.Zhi,
			WuXing:        nj.WuXing,
			LiuQin:        qin,
			LiuShen:       shen[i],
			IsShi:         shi.Shi == i+1,
			IsYing:        shi.Ying == i+1,
			IsKong:        isKong,
			WangShuai:     ws,
			WangShuaiDesc: WangShuaiDesc[ws],
			Special:       sp,
			SpecialLabel:  SpecialStatusLabel[sp],
			MoveState:     moveState,
			MoveLabel:     MoveStateLabel[moveState],
			Strength:      strength,
			IsActive:      active,
			Evidence:      evidence,
		}

		// 旬空提示
		if isKong {
			if !isChanging {
				yi.Evidence = append(yi.Evidence, fmt.Sprintf("爻支%s落旬空（%s），静而空亡", nj.Zhi, t.XunKong))
			} else {
				yi.Evidence = append(yi.Evidence, fmt.Sprintf("爻支%s落旬空（%s），动空待冲实", nj.Zhi, t.XunKong))
			}
			// 传统：空爻逢日辰冲，谓之「冲空」，空而不空、反有发动之意
			if LiuChongZhi[nj.Zhi] == t.DayZhi {
				yi.Evidence = append(yi.Evidence, "空爻逢日辰冲，谓之冲空，空而反实，有激发之意")
			}
		}
		if sheng[WuXingOfZhi(t.DayZhi)] == nj.WuXing {
			yi.Evidence = append(yi.Evidence, "得日辰生扶")
		}

		// 爻符号
		if v.YinYang() == Yang {
			yi.Symbol = "▅▅▅▅▅▅"
		} else {
			yi.Symbol = "▅▅　▅▅"
		}

		// 变爻信息
		if isChanging && bian != nil {
			bnj := bianNaJia[i]
			yi.ChangedYinYang = v.YinYang().flip()
			yi.ChangedGanZhi = bnj.GanZhi
			yi.ChangedLiuQin = GetLiuQin(gongWuXing, bnj.WuXing)
			yi.HuaType = judgeHua(nj.WuXing, bnj.WuXing)
		}

		yaos = append(yaos, yi)
	}

	// 6. 伏神
	fuShen := calcFuShen(mark, gong, presentQin)

	// 7. 五行生克 / 三合 / 六合 / 六冲
	relations := BuildRelations(yaos, t, shi.Shi, shi.Ying)

	// 8. 摘要（并入生合冲克要点，六冲提示仍由 summary 统一给出）
	summary := buildSummary(benGua, bian, yaos)
	summary.Highlights = append(summary.Highlights, relations.Highlights...)

	chart := &Chart{
		Question:  in.Question,
		Method:    in.Method,
		Time:      t,
		BenGua:    benGua,
		BianGua:   bian,
		Yaos:      yaos,
		FuShen:    fuShen,
		Relations: relations,
		Summary:   summary,
	}
	// 9. 用神/忌神体系
	chart.Shen = BuildShenSystem(chart, in.Question, in.YongShen)

	// 10. 白话解析（规则驱动，含用神）
	chart.Plain = BuildPlain(chart)

	return chart, nil
}

func (y YinYang) flip() YinYang {
	if y == Yang {
		return Yin
	}
	return Yang
}

// judgeHua 判断化进化退与回头生克
func judgeHua(benWuXing, bianWuXing WuXing) string {
	if benWuXing == bianWuXing {
		return "伏吟"
	}
	// 化进神：同五行不必论，此处按相生方向判断
	if sheng[bianWuXing] == benWuXing {
		return "回头生"
	}
	if ke[bianWuXing] == benWuXing {
		return "回头克"
	}
	if sheng[benWuXing] == bianWuXing {
		return "化泄气"
	}
	if ke[benWuXing] == bianWuXing {
		return "化耗气"
	}
	return ""
}

// buildSummary 汇总关键结论，供前端高亮显示
func buildSummary(ben GuaInfo, bian *BianGua, yaos []YaoInfo) Summary {
	s := Summary{
		BenGuaName: ben.Name,
		Gong:       ben.Gong,
		GuaType:    ben.Type,
		IsLiuChong: IsLiuChong(ben.Mark),
		IsLiuHe:    IsLiuHe(ben.Mark),
	}
	if bian != nil {
		s.BianGuaName = bian.Name
		s.BianGuaType = bian.Type
	}

	for _, y := range yaos {
		switch y.MoveState {
		case MoveMing:
			s.MovingCount++
		case MoveAnDong:
			s.AnDongCount++
			s.AnDongYaos = append(s.AnDongYaos, y.Position)
		case MoveBroken:
			s.BrokenYaos = append(s.BrokenYaos, y.Position)
		}
		if y.IsKong {
			s.KongYaos = append(s.KongYaos, y.Position)
		}
		if y.Strength == LevelStrong {
			s.StrongYaos = append(s.StrongYaos, y.Position)
		}
	}

	// 重点提示
	if s.IsLiuChong {
		s.Highlights = append(s.Highlights,
			fmt.Sprintf("本卦「%s」为六冲卦：主事情动荡、变化快、易散难成；亦主近期即应。", ben.Name))
	}
	if s.IsLiuHe {
		s.Highlights = append(s.Highlights,
			fmt.Sprintf("本卦「%s」为六合卦：主事情和合、聚合、缠绵难散，多主成。", ben.Name))
	}
	if bian != nil && bian.Type == "六冲" {
		s.Highlights = append(s.Highlights,
			fmt.Sprintf("变卦「%s」为六冲卦：事虽初成，终恐冲散，须防有始无终。", bian.Name))
	}
	if bian != nil && bian.Type == "六合" {
		s.Highlights = append(s.Highlights,
			fmt.Sprintf("变卦「%s」为六合卦：事态趋于和合、牵绊，主结果能聚。", bian.Name))
	}
	if s.AnDongCount > 0 {
		names := positionNames(s.AnDongYaos)
		s.Highlights = append(s.Highlights,
			fmt.Sprintf("有暗动：%s。旺相逢冲为暗动，静中带动、暗中有人/事推动，其力小于明动而大于静爻。", names))
	}
	if len(s.BrokenYaos) > 0 {
		var yuePo, riPo []int
		for _, y := range yaos {
			switch y.Special {
			case StatusYuePo:
				yuePo = append(yuePo, y.Position)
			case StatusRiPo:
				riPo = append(riPo, y.Position)
			}
		}
		if len(yuePo) > 0 {
			s.Highlights = append(s.Highlights,
				fmt.Sprintf("月破：%s。被月建所冲，不论旺衰整月无力，须过月方可论用。", positionNames(yuePo)))
		}
		if len(riPo) > 0 {
			s.Highlights = append(s.Highlights,
				fmt.Sprintf("日破：%s。休囚而逢日辰冲为日破，该爻受伤无力，主其所主之事受损。", positionNames(riPo)))
		}
	}
	if len(s.KongYaos) > 0 {
		names := positionNames(s.KongYaos)
		s.Highlights = append(s.Highlights,
			fmt.Sprintf("旬空：%s。空则暂时无力，出旬或逢冲填实方能应事。", names))
	}
	if s.MovingCount == 0 && s.AnDongCount == 0 {
		s.Highlights = append(s.Highlights, "六爻皆静：事态尚未发动，宜静观其变，以世应旺衰定吉凶。")
	}
	return s
}

func positionNames(positions []int) string {
	parts := make([]string, 0, len(positions))
	for _, p := range positions {
		if p >= 1 && p <= 6 {
			parts = append(parts, yaoPosName[p-1])
		}
	}
	return strings.Join(parts, "、")
}
