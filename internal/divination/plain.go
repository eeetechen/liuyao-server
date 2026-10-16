package divination

import (
	"fmt"
	"strings"
)

// ============================================================
// 白话解析：把排盘结果转成通俗易懂的说明
//
// 全部由确定性规则生成，不使用大模型，故结果可复现、可核验。
// ============================================================

// PlainSection 白话解析的一个小节
type PlainSection struct {
	Title string   `json:"title"`
	Lines []string `json:"lines"`
}

// Plain 白话解析
type Plain struct {
	Overview    string         `json:"overview"`    // 一句话总览
	Sections    []PlainSection `json:"sections"`    // 分节说明
	Suggestions []string       `json:"suggestions"` // 行动建议
}

// 爻位对应的生活含义（传统取象，便于入门者理解）
var positionMeaning = [6]string{
	"初爻——事情刚刚起步、根基与开端",
	"二爻——事情的发展过程、内部情况",
	"三爻——事情的转折处、内卦之终",
	"四爻——事情接近核心、外部压力",
	"五爻——事情的主导方、关键决策者",
	"上爻——事情的结果、收尾与结局",
}

// BuildPlain 生成白话解析。
func BuildPlain(c *Chart) Plain {
	p := Plain{Sections: []PlainSection{}, Suggestions: []string{}}

	p.Overview = plainOverview(c)

	for _, s := range []PlainSection{
		plainTime(c),
		plainGua(c),
		plainYongShen(c),
		plainMoving(c),
		plainKeyYao(c),
		plainRelations(c),
	} {
		if len(s.Lines) > 0 {
			p.Sections = append(p.Sections, s)
		}
	}

	p.Suggestions = plainSuggestions(c)
	return p
}

// ---------- 总览 ----------

func plainOverview(c *Chart) string {
	var b strings.Builder

	// 所问之事为可选；未填写时省略该分句，避免出现「问的是『…』」的空话
	q := strings.TrimSpace(c.Question)
	if q != "" {
		b.WriteString(fmt.Sprintf("你在「%s」起卦，问的是「%s」，得到【%s】卦。",
			c.Time.SolarText, q, c.BenGua.Name))
	} else {
		b.WriteString(fmt.Sprintf("你在「%s」起卦，得到【%s】卦。",
			c.Time.SolarText, c.BenGua.Name))
	}

	if c.BianGua != nil {
		b.WriteString(fmt.Sprintf("其中 %s 发动，事情会往【%s】卦的方向演变。",
			plainPositions(changingPositions(c)), c.BianGua.Name))
	} else {
		b.WriteString("六爻都没有发动，说明目前局势比较稳定，暂时不会有大的变化。")
	}

	return b.String()
}

// ---------- 时间 ----------

func plainTime(c *Chart) PlainSection {
	t := c.Time
	lines := []string{
		fmt.Sprintf("公历：%s（%s）", t.SolarText, t.LunarText),
		fmt.Sprintf("年份：%s。占卦用的年份以「立春」为界，不是正月初一，所以立春前仍算上一年。", t.YearLabel),
		fmt.Sprintf("四柱：年柱 %s、%s、%s、%s。", t.Year, t.MonthLabel, t.DayLabel, t.HourLabel),
		fmt.Sprintf("月建是「%s」——它决定这个月里五行的旺衰；日辰是「%s」——它决定哪一爻被冲动。月建和日辰是断卦最重要的两个参照。",
			t.MonthZhi, t.DayZhi),
	}
	if t.XunKong != "" {
		lines = append(lines, fmt.Sprintf("本旬旬空为「%s」。落在旬空里的爻，眼下使不上力，要等出了这一旬、或者被冲实了才起作用。", t.XunKong))
	}
	return PlainSection{Title: "一、时间信息", Lines: lines}
}

// ---------- 卦体 ----------

func plainGua(c *Chart) PlainSection {
	var lines []string
	b := c.BenGua

	lines = append(lines, fmt.Sprintf("【%s】属%s宫，五行属%s。上卦%s、下卦%s。",
		b.Name, b.Gong, b.Element, b.UpperGua, b.LowerGua))

	switch b.Type {
	case "六冲":
		lines = append(lines, "这一卦是「六冲卦」。六冲的意思是内在冲突大、变化快，事情容易散、容易被冲开；好处是应验快，短期内就会有动静。")
	case "六合":
		lines = append(lines, "这一卦是「六合卦」。六合的意思是事情容易聚合、纠缠、拖延难散，多半能成，但过程会比较黏。")
	case "游魂":
		lines = append(lines, "这一卦是「游魂卦」。游魂主心神不定、事情反复、人易变动，往往念头多而难落实。")
	case "归魂":
		lines = append(lines, "这一卦是「归魂卦」。归魂主回归、回到原处，事情容易兜回原点，或与旧人旧事有关。")
	}

	lines = append(lines, fmt.Sprintf("世爻在%s、应爻在%s。世爻代表你自己，应爻代表对方或所问之事。",
		b.ShiPosName(), b.YingPosName()))

	if c.BianGua != nil {
		lines = append(lines, fmt.Sprintf("变卦是【%s】，属%s宫。变卦代表事情发展的最终走向。",
			c.BianGua.Name, c.BianGua.Gong))
		if c.BianGua.Type != "" {
			lines = append(lines, fmt.Sprintf("变卦是「%s」卦，说明结局带有%s的性质。", c.BianGua.Type, c.BianGua.Type))
		}
	}
	return PlainSection{Title: "二、卦体说明", Lines: lines}
}

// ---------- 用神 ----------

func plainYongShen(c *Chart) PlainSection {
	var lines []string
	sh := c.Shen

	lines = append(lines, fmt.Sprintf("取「%s」为用神（%s）。用神就是这一卦里代表所问之事的那一爻，断卦先看它。",
		sh.YongShen, sh.Topic))
	if sh.YuanShen != "" {
		lines = append(lines, fmt.Sprintf("四神关系：原神为%s（生用神，帮我的一方）、忌神为%s（克用神，伤我的一方）、仇神为%s（生忌神）。",
			sh.YuanShen, sh.JiShen, sh.ChouShen))
	}
	lines = append(lines, sh.Advice...)

	return PlainSection{Title: "三、用神与四神", Lines: lines}
}

// ---------- 动爻 ----------

func plainMoving(c *Chart) PlainSection {
	var lines []string

	var moving, anDong, broken []YaoInfo
	for _, y := range c.Yaos {
		switch {
		case y.IsChanging:
			moving = append(moving, y)
		case y.Special == StatusAnDong:
			anDong = append(anDong, y)
		case y.Special == StatusRiPo || y.Special == StatusYuePo:
			broken = append(broken, y)
		}
	}

	if len(moving) == 0 && len(anDong) == 0 {
		lines = append(lines, "六爻全静，没有动爻，也没有暗动。这说明事情还没有真正动起来，眼下不宜急动，静观其变即可。判断以世爻、应爻的旺衰为主。")
	}

	for _, y := range moving {
		lines = append(lines, fmt.Sprintf("%s 发动（%s%s，%s，此爻%s）。动爻是事情的「变化点」，这一爻所代表的人或事会主动发生变化。",
			y.Name, y.GanZhi, y.WuXing, y.LiuQin, y.WangShuaiDesc))
	}

	for _, y := range anDong {
		lines = append(lines, fmt.Sprintf("%s 是「暗动」（%s%s，%s）。这一爻本身没动，但被日辰「%s」冲了一下，而它又不弱，所以被激起来了。暗动表示暗地里有人或有事在推动，力量比明动小，但比静爻大。",
			y.Name, y.GanZhi, y.WuXing, y.LiuQin, c.Time.DayZhi))
	}

	for _, y := range broken {
		kind := "日破"
		reason := fmt.Sprintf("被日辰「%s」冲，而它本身力量弱", c.Time.DayZhi)
		if y.Special == StatusYuePo {
			kind = "月破"
			reason = fmt.Sprintf("被月建「%s」冲", c.Time.MonthZhi)
		}
		lines = append(lines, fmt.Sprintf("%s 逢「%s」（%s%s，%s）：%s。这一爻所代表的人或事眼下受损、使不上力。%s",
			y.Name, kind, y.GanZhi, y.WuXing, y.LiuQin, reason,
			plainBrokenAdvice(kind)))
	}

	return PlainSection{Title: "四、动与静", Lines: lines}
}

func plainBrokenAdvice(kind string) string {
	if kind == "月破" {
		return "月破要过了这个月才有转机。"
	}
	return "日破逢生扶或出空，还可以救回来。"
}

// ---------- 关键爻 ----------

func plainKeyYao(c *Chart) PlainSection {
	var lines []string

	for _, y := range c.Yaos {
		if !y.IsShi && !y.IsYing {
			continue
		}
		role := "世爻（你自己）"
		if y.IsYing {
			role = "应爻（对方/所问之事）"
		}
		state := fmt.Sprintf("%s，力量%s", y.WangShuaiDesc, strengthWord(y.Strength))
		if y.IsKong {
			state += "，且落在旬空（暂时使不上力）"
		}
		if y.SpecialLabel != "" {
			state += "，" + y.SpecialLabel
		}
		lines = append(lines, fmt.Sprintf("%s在%s：%s%s，%s，属「%s」。%s",
			role, y.Name, y.GanZhi, y.WuXing, state, y.LiuQin, roleMeaning(y.LiuQin)))
	}

	// 世应关系
	if c.Relations.ShiYing != nil {
		lines = append(lines, "世应之间："+c.Relations.ShiYing.Desc+"。")
	}

	// 提示各爻所主
	var strong []string
	for _, y := range c.Yaos {
		if y.Strength == LevelStrong {
			strong = append(strong, fmt.Sprintf("%s(%s)", y.Name, y.LiuQin))
		}
	}
	if len(strong) > 0 {
		lines = append(lines, fmt.Sprintf("全卦比较有力的是：%s。这些爻所代表的人或事比较占优势。",
			strings.Join(strong, "、")))
	}

	return PlainSection{Title: "五、关键爻位", Lines: lines}
}

// 六亲的口语化含义
func roleMeaning(qin string) string {
	switch qin {
	case "兄弟":
		return "主同辈、竞争者、合作伙伴，也主耗财"
	case "父母":
		return "主长辈、文书、房产、车船，也主辛苦劳碌"
	case "官鬼":
		return "主官方、上司、工作，也主疾病、麻烦与阻碍"
	case "妻财":
		return "主钱财、物品，男占婚姻亦主妻子"
	case "子孙":
		return "主子女、晚辈，也主福气、平安、解忧之神"
	}
	return ""
}

func strengthWord(s StrengthLevel) string {
	switch s {
	case LevelStrong:
		return "强（有力）"
	case LevelModerate:
		return "中等（平平）"
	default:
		return "弱（无力）"
	}
}

// ---------- 生合冲克 ----------

func plainRelations(c *Chart) PlainSection {
	var lines []string
	r := c.Relations

	// 三合
	for _, j := range r.SanHe {
		lines = append(lines, fmt.Sprintf("三合局：%s。三支凑齐成局，代表三方力量汇合、事情有合力，是吉象。%s",
			j.Desc, sanHeAdvice(j)))
	}
	for _, j := range r.BanHe {
		lines = append(lines, fmt.Sprintf("半合：%s。只凑到两支，力量比三合弱，是「有合意但未完全成」的意思。", j.Desc))
	}

	// 六合
	for _, p := range r.LiuHe {
		lines = append(lines, fmt.Sprintf("六合：%s%s 与 %s%s 相合（合化%s）。%s",
			p.A.Label, p.A.Zhi, p.B.Label, p.B.Zhi, p.Hua, p.Note))
	}

	// 六冲
	for _, p := range r.LiuChong {
		lines = append(lines, fmt.Sprintf("六冲：%s%s 与 %s%s 相冲。%s",
			p.A.Label, p.A.Zhi, p.B.Label, p.B.Zhi, p.Note))
	}

	if len(lines) == 0 {
		lines = append(lines, "本次没有明显的三合、六合、六冲、三刑或六害。")
	}
	return PlainSection{Title: "六、生合冲克", Lines: lines}
}

// ---------- 建议 ----------

func plainSuggestions(c *Chart) []string {
	var out []string

	// 世爻旺衰定主动被动
	for _, y := range c.Yaos {
		if !y.IsShi {
			continue
		}
		switch y.Strength {
		case LevelStrong:
			out = append(out, "世爻有力，你自己这边状态好、有主动权，可以积极一些。")
		case LevelWeak:
			out = append(out, "世爻偏弱，你自己这边力量不足，宜守不宜攻，先补足条件再动。")
		default:
			out = append(out, "世爻力量中等，事情成败更多取决于外部配合，不必强求。")
		}
		if y.IsKong {
			out = append(out, "世爻落空，眼下自己使不上力或心里没底，等出旬或逢冲之后再行动。")
		}
	}

	if c.Summary.AnDongCount > 0 {
		out = append(out, "有暗动，说明暗地里有变化在推动，要多留意没摆在明面上的信息。")
	}
	if len(c.Summary.BrokenYaos) > 0 {
		out = append(out, "有爻逢破，相关的人或事眼下受损，不宜依赖，宜避其锋。")
	}
	if c.Summary.IsLiuChong {
		out = append(out, "六冲主快、主散，事情来得急去得也快，宜速战速决，不宜久拖。")
	}
	if c.Summary.IsLiuHe {
		out = append(out, "六合主聚、主黏，事情容易拖，要有耐心，也要防止拖着不决。")
	}
	if c.BianGua == nil {
		out = append(out, "无动爻，短期内变化不大，可先按现状推进，不必急于求成。")
	}

	if c.Shen.YongYao != nil {
		y := c.Shen.YongYao
		if y.Strength == LevelWeak {
			out = append(out, fmt.Sprintf("用神「%s」偏弱，所问之事底气不足，宜先补足条件、借助原神（%s）之力。",
				c.Shen.YongShen, c.Shen.YuanShen))
		} else if y.Strength == LevelStrong {
			out = append(out, fmt.Sprintf("用神「%s」有力，所问之事本身占优，可放手推进。", c.Shen.YongShen))
		}
	}
	if len(c.Relations.Xing) > 0 {
		out = append(out, "卦中见三刑，主内部摩擦、人际失和或有官讼之扰，凡事宜留余地、避免硬碰。")
	}
	if len(c.Relations.Hai) > 0 {
		out = append(out, "卦中见六害，多主暗中受损或彼此猜忌，注意小人与细节牵绊。")
	}

	out = append(out, "以上为依据传统纳甲筮法规则所作的客观推演，仅供文化研究与自我参考，重要决策请结合实际情况判断。")
	return out
}

// ---------- 辅助 ----------

func changingPositions(c *Chart) []int {
	var out []int
	for _, y := range c.Yaos {
		if y.IsChanging {
			out = append(out, y.Position)
		}
	}
	return out
}

func plainPositions(ps []int) string {
	names := positionNames(ps)
	if names == "" {
		return "没有"
	}
	return names
}

// ShiPosName / YingPosName 世应爻位的中文名
func (g GuaInfo) ShiPosName() string  { return yaoPosName[clampPos(g.ShiPos)] }
func (g GuaInfo) YingPosName() string { return yaoPosName[clampPos(g.YingPos)] }

func clampPos(p int) int {
	if p < 1 {
		return 0
	}
	if p > 6 {
		return 5
	}
	return p - 1
}
