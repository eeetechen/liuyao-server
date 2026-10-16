package divination

import (
	"fmt"
	"strings"
)

// ============================================================
// 用神 / 忌神 / 原神 / 仇神 自动选取
//
// 取用原则（传统纳甲筮法）：
//   - 用神：所占之事的「主爻」，按所问主题取对应六亲
//   - 原神：生用神之六亲，为助用神者
//   - 忌神：克用神之六亲，为伤用神者
//   - 仇神：生忌神而克原神者，为助忌神者
//
// 取用神之次序：本卦现则取本卦 → 本卦无则取变爻所出 → 变卦亦无则取伏神。
// ============================================================

// YongShenCandidate 用神候选（某一爻）
type YongShenCandidate struct {
	LiuQin    string         `json:"liuQin"`
	Position  int            `json:"position"`
	Label     string         `json:"label"` // 初爻…上爻
	GanZhi    string         `json:"ganZhi"`
	Zhi       string         `json:"zhi"`
	WuXing    WuXing         `json:"wuXing"`
	WangShuai WangShuaiState `json:"wangShuai"`
	Strength  StrengthLevel  `json:"strength"`
	Moving    bool           `json:"moving"`
	Special   string         `json:"special"` // 暗动/日破/月破
	IsKong    bool           `json:"isKong"`
	IsShi     bool           `json:"isShi"`
	IsYing    bool           `json:"isYing"`
	IsFuShen  bool           `json:"isFuShen"` // 是否伏神
}

// YongShenSelection 取用神的来源
type YongShenSelection string

const (
	SelectVisible  YongShenSelection = "visible"  // 取本卦明现之爻
	SelectChanged  YongShenSelection = "changed"  // 本卦无，取变爻所出
	SelectFuShen   YongShenSelection = "fushen"   // 变卦亦无，取伏神
	SelectTemporal YongShenSelection = "temporal" // 卦与伏神皆无，以日月代之
	SelectMissing  YongShenSelection = "missing"  // 遍寻不得
)

// ShenSystem 用神体系
type ShenSystem struct {
	Topic      string              `json:"topic"`             // 所问主题（能推断时）
	YongShen   string              `json:"yongShen"`          // 用神六亲
	YuanShen   string              `json:"yuanShen"`          // 原神（生用神）
	JiShen     string              `json:"jiShen"`            // 忌神（克用神）
	ChouShen   string              `json:"chouShen"`          // 仇神（生忌神、克原神）
	Inferred   bool                `json:"inferred"`          // 用神是否由问题推断而得
	Selection  YongShenSelection   `json:"selection"`         // 取用来源
	SelectNote string              `json:"selectNote"`        // 取用说明
	YongYao    *YongShenCandidate  `json:"yongYao,omitempty"` // 主用神
	Candidates []YongShenCandidate `json:"candidates"`        // 同六亲并见者
	Ambiguous  bool                `json:"ambiguous"`         // 用神并见，须兼看
	Advice     []string            `json:"advice"`            // 断语
}

// 四神的推法（关键：六亲一律相对「卦宫五行」而定，不可相对用神）
//
// 设卦宫五行为 G，用神五行为 Y：
//
//	原神 = 生用神者 → 五行 = 生 Y 者，再以 G 为我取六亲
//	忌神 = 克用神者 → 五行 = 克 Y 者，再以 G 为我取六亲
//	仇神 = 生忌神者 → 五行 = 生(克 Y 者)，再以 G 为我取六亲
//
// 例：乾宫属金，用神取妻财（木）：
//
//	原神：生木者水，水在乾宫为子孙 → 原神子孙
//	忌神：克木者金，金在乾宫为兄弟 → 忌神兄弟
//	仇神：生金者土，土在乾宫为父母 → 仇神父母
//	（校验：父母土克子孙水＝仇神克原神；土生金＝仇神生忌神 ✓）
func shenOf(gongWuXing, yongWuXing WuXing) (yuan, ji, chou string) {
	yuanWX := inverseOf(sheng, yongWuXing) // 生用神者
	jiWX := inverseOf(ke, yongWuXing)      // 克用神者
	chouWX := inverseOf(sheng, jiWX)       // 生忌神者（注意：是「生忌神」，非「忌神所生」）

	return GetLiuQin(gongWuXing, yuanWX),
		GetLiuQin(gongWuXing, jiWX),
		GetLiuQin(gongWuXing, chouWX)
}

// TopicRule 主题推断规则
type TopicRule struct {
	Keywords []string // 命中任一关键词即认为是该主题
	YongShen string   // 对应六亲
	Topic    string   // 主题名
}

// TopicRules 由所问之事推断用神的规则表。
//
// 顺序有意义：越具体的主题越靠前，避免被泛化关键词抢先命中。
var TopicRules = []TopicRule{
	{[]string{"考试", "学业", "升学", "面试", "证书", "文凭", "录取", "论文", "答辩", "资格"}, "父母", "学业考试"},
	{[]string{"房", "车", "合同", "协议", "文书", "证件", "手续", "批复", "审批", "邮件", "消息", "通知"}, "父母", "文书房产"},
	{[]string{"父母", "长辈", "老人", "师父", "老师", "上司", "领导", "老板", "公司", "单位"}, "父母", "长辈上级"},

	{[]string{"财", "钱", "收入", "工资", "薪水", "奖金", "利润", "赚", "盈利", "投资", "理财", "债", "欠款", "货款", "回款", "生意", "买卖", "订单"}, "妻财", "求财"},
	{[]string{"妻", "老婆", "妻子", "老婆", "女友", "女朋友", "婚姻", "结婚", "感情", "恋爱", "对象", "配偶", "老婆"}, "妻财", "婚姻感情"},

	{[]string{"工作", "事业", "求职", "职位", "升职", "晋升", "调岗", "跳槽", "offer", "录用", "任职", "考公", "编制"}, "官鬼", "工作事业"},
	{[]string{"官", "官非", "官司", "诉讼", "法律", "仲裁", "纠纷", "处罚", "违章", "上访"}, "官鬼", "官非诉讼"},
	{[]string{"病", "疾病", "健康", "身体", "治疗", "手术", "医生", "医院", "康复", "新冠", "发烧"}, "官鬼", "疾病健康"},
	{[]string{"考", "名次", "功名", "竞争", "对手", "小人", "障碍", "阻碍"}, "官鬼", "功名阻碍"},

	{[]string{"孩子", "子女", "儿子", "女儿", "晚辈", "学生", "徒弟", "生育", "怀孕", "求子"}, "子孙", "子女晚辈"},
	{[]string{"平安", "化解", "消灾", "解忧", "避险", "医药", "医生", "僧道", "出家人"}, "子孙", "平安解忧"},
	{[]string{"宠物", "猫", "狗", "动物", "养殖", "牲畜"}, "子孙", "宠物"},

	{[]string{"朋友", "兄弟", "同事", "同学", "合伙", "合作", "同行", "竞争", "姐妹", "兄长", "弟", "哥"}, "兄弟", "同辈合作"},
	{[]string{"借款", "借钱", "还钱", "破财", "开销", "花费", "亏", "催款"}, "兄弟", "同辈借贷"},
}

// TopicByYongShen 给出某六亲对应的标准主题名。
//
// 当未提供所问之事、或问题未能匹配任何规则时，以所选六亲的主题为准，
// 避免出现「通用（未识别主题）」这类无意义的标注。
func TopicByYongShen(qin string) string {
	switch qin {
	case "妻财":
		return "求财"
	case "官鬼":
		return "工作事业"
	case "父母":
		return "学业文书"
	case "子孙":
		return "子女平安"
	case "兄弟":
		return "同辈合作"
	}
	return "通用"
}

// InferYongShen 由所问之事推断用神六亲。返回六亲与主题名。
//
// 命中第一条规则即返回；全部未命中则默认取「妻财」
// （财为养命之源，古法无主题时多以财论），并标注为默认。
func InferYongShen(question string) (liuQin, topic string, matched bool) {
	q := question
	for _, r := range TopicRules {
		for _, kw := range r.Keywords {
			if strings.Contains(q, kw) {
				return r.YongShen, r.Topic, true
			}
		}
	}
	return "妻财", "通用（未识别主题）", false
}

// BuildShenSystem 组装用神体系。
//
// yongQin 为空时由 question 自动推断。
func BuildShenSystem(c *Chart, question, yongQin string) ShenSystem {
	inferred := false
	topic := ""

	if yongQin == "" {
		// 未指定用神时才依问题推断；问题为空则回落为妻财
		yongQin, topic, _ = InferYongShen(question)
		inferred = true
	}
	// 主题名：问题能匹配上则用匹配结果，否则用所选六亲的标准主题
	if topic == "" || topic == "通用（未识别主题）" {
		topic = TopicByYongShen(yongQin)
	}

	s := ShenSystem{
		Topic:      topic,
		YongShen:   yongQin,
		Inferred:   inferred,
		Candidates: []YongShenCandidate{},
		Advice:     []string{},
	}

	// 用神五行由卦宫推得（用神六亲相对卦宫而定，其五行即该六亲在卦中所属五行）
	// 取本卦中现身的同六亲爻的五行作为参照；若本卦不见，则由卦宫五行反推。
	yongWuXing := wuXingOfLiuQinInBen(c, yongQin)
	if yongWuXing == "" {
		yongWuXing = inferWuXingOfLiuQin(c.BenGua.Element, yongQin)
	}
	if yongWuXing != "" {
		s.YuanShen, s.JiShen, s.ChouShen = shenOf(c.BenGua.Element, yongWuXing)
	}

	// 1. 本卦现则取本卦
	var visible []YongShenCandidate
	for _, y := range c.Yaos {
		if y.LiuQin != yongQin {
			continue
		}
		visible = append(visible, candidateOf(y))
	}

	if len(visible) > 0 {
		s.Selection = SelectVisible
		s.Candidates = visible
		pick, reason := choosePrimary(visible)
		s.YongYao = &pick
		// 用神多现：同一六亲现身两处以上，须兼看而非只取一爻
		s.Ambiguous = len(visible) > 1
		if s.Ambiguous {
			s.SelectNote = fmt.Sprintf("用神「%s」多现，见 %s。依「取动、取临日月、取持世应、舍空取实、舍破取全」之序，取%s为用神（%s）。其余仍须兼看。",
				yongQin, candidatesLabel(visible), pick.Label, reason)
		} else {
			s.SelectNote = fmt.Sprintf("本卦中 %s 现于%s，取为用神。", yongQin, candidatesLabel(visible))
		}
	} else {
		// 2. 本卦无 → 取变爻所出
		if c.BianGua != nil {
			for i, q := range c.BianGua.LiuQin {
				if q != yongQin {
					continue
				}
				// 仅动爻所变出的用神才算「变出用神」
				if !c.Yaos[i].IsChanging {
					continue
				}
				bnj := c.BianGua.NaJia[i]
				cand := YongShenCandidate{
					LiuQin: yongQin, Position: i + 1, Label: yaoPosName[i],
					GanZhi: bnj.GanZhi, Zhi: bnj.Zhi, WuXing: bnj.WuXing,
					WangShuai: GetWangShuai(bnj.WuXing, c.Time.MonthZhi),
					Moving:    true,
				}
				s.Candidates = append(s.Candidates, cand)
			}
		}
		if len(s.Candidates) > 0 {
			s.Selection = SelectChanged
			pick := s.Candidates[0]
			s.YongYao = &pick
			s.SelectNote = fmt.Sprintf("本卦不见 %s，由%s变出，依「变者显也」先取变爻。", yongQin, pick.Label)
		} else if len(c.FuShen) > 0 {
			// 3. 变卦亦无 → 取伏神
			for _, f := range c.FuShen {
				if f.LiuQin != yongQin {
					continue
				}
				cand := YongShenCandidate{
					LiuQin: yongQin, Position: f.Position, Label: yaoPosName[f.Position-1],
					GanZhi: f.GanZhi, Zhi: f.Zhi, WuXing: f.WuXing,
					WangShuai: GetWangShuai(f.WuXing, c.Time.MonthZhi),
					IsKong:    IsKong(f.Zhi, c.Time.XunKong),
					IsFuShen:  true,
				}
				cand.Strength = GetStrengthLevel(cand.WangShuai, StatusNone, cand.IsKong, false)
				s.Candidates = append(s.Candidates, cand)
			}
			if len(s.Candidates) > 0 {
				s.Selection = SelectFuShen
				pick := s.Candidates[0]
				s.YongYao = &pick
				s.SelectNote = fmt.Sprintf("本卦与变卦皆不见 %s，取伏神：%s伏于%s之下。用神伏藏，主事未显、须待引发。",
					yongQin, fShenName(pick), pick.Label)
			}
		}

		if s.Selection == "" {
			s.Selection = SelectMissing
			s.SelectNote = fmt.Sprintf("%s 不上卦，且无伏神可代，须以日月或卦宫五行权代论之。", yongQin)
		}
	}

	// 4. 逐条断语
	s.Advice = shenAdvice(c, &s)
	return s
}

func candidateOf(y YaoInfo) YongShenCandidate {
	return YongShenCandidate{
		LiuQin: y.LiuQin, Position: y.Position, Label: y.Name,
		GanZhi: y.GanZhi, Zhi: y.Zhi, WuXing: y.WuXing,
		WangShuai: y.WangShuai, Strength: y.Strength,
		Moving: y.IsChanging, Special: string(y.Special),
		IsKong: y.IsKong, IsShi: y.IsShi, IsYing: y.IsYing,
	}
}

// shenRank 用神取舍次序（数值越小越优先）。
//
// 传统原则：取动、取临日月、取持世应、舍空取实、舍破取全。
func shenRank(c YongShenCandidate) int {
	switch {
	case c.Moving:
		return 0 // 明动，事已发动，最能显事
	case c.Special == string(StatusAnDong):
		return 1 // 暗动，暗中有动
	case c.IsShi:
		return 2 // 持世，与我最近
	case c.IsYing:
		return 3 // 临应，与对方最近
	case c.Special == string(StatusRiPo) || c.Special == string(StatusYuePo):
		return 6 // 逢破，受伤最重
	case c.IsKong:
		return 5 // 旬空，暂时无力
	default:
		return 4 // 静而完好
	}
}

func shenRankReason(c YongShenCandidate, rank int) string {
	switch rank {
	case 0:
		return "此爻明动，事已发动"
	case 1:
		return "此爻暗动，暗中有动"
	case 2:
		return "此爻持世，与我最近"
	case 3:
		return "此爻临应，与对方最近"
	case 5:
		return "余者逢空，唯此可用"
	case 6:
		return "余者逢破，唯此可用"
	default:
		return "此爻静而完好"
	}
}

// choosePrimary 从多个候选中择定主用神，并给出取舍理由。
func choosePrimary(cands []YongShenCandidate) (YongShenCandidate, string) {
	best := cands[0]
	bestRank := shenRank(best)
	for _, c := range cands[1:] {
		if r := shenRank(c); r < bestRank {
			best, bestRank = c, r
		}
	}
	return best, shenRankReason(best, bestRank)
}

func candidatesLabel(cands []YongShenCandidate) string {
	parts := make([]string, 0, len(cands))
	for _, c := range cands {
		parts = append(parts, fmt.Sprintf("%s(%s%s)", c.Label, c.GanZhi, c.WuXing))
	}
	return strings.Join(parts, "、")
}

func fShenName(c YongShenCandidate) string {
	return c.GanZhi + string(c.WuXing)
}

// YongWuXing 返回用神五行的对外只读副本（供测试与前端校验）
func (s ShenSystem) YongWuXing() WuXing {
	if s.YongYao != nil {
		return s.YongYao.WuXing
	}
	return ""
}

// wuXingOfLiuQinInBen 取本卦中该六亲爻的五行（同一六亲在卦中五行应一致）
func wuXingOfLiuQinInBen(c *Chart, qin string) WuXing {
	for _, y := range c.Yaos {
		if y.LiuQin == qin {
			return y.WuXing
		}
	}
	for i, q := range c.BianGuaLiuQin() {
		if q == qin {
			return c.BianGua.NaJia[i].WuXing
		}
	}
	for _, f := range c.FuShen {
		if f.LiuQin == qin {
			return f.WuXing
		}
	}
	return ""
}

// BianGuaLiuQin 安全取变卦六亲
func (c *Chart) BianGuaLiuQin() []string {
	if c.BianGua == nil {
		return nil
	}
	return c.BianGua.LiuQin
}

// inferWuXingOfLiuQin 由卦宫五行反推某六亲的五行。
//
// 六亲定义：生我者父母、我生者子孙、克我者官鬼、我克者妻财、比和者兄弟。
// 已知「我」（卦宫五行）与六亲名，可反推该六亲五行。
func inferWuXingOfLiuQin(gongWuXing WuXing, qin string) WuXing {
	switch qin {
	case "兄弟":
		return gongWuXing
	case "子孙":
		return sheng[gongWuXing] // 我生者
	case "父母":
		return inverseOf(sheng, gongWuXing) // 生我者
	case "妻财":
		return ke[gongWuXing] // 我克者
	case "官鬼":
		return inverseOf(ke, gongWuXing) // 克我者
	}
	return ""
}

// shenAdvice 生成用神体系的断语
func shenAdvice(c *Chart, s *ShenSystem) []string {
	var out []string

	if s.Inferred {
		out = append(out, fmt.Sprintf("依所问之事推断用神为「%s」（%s）。若与取用习惯不同，可另行指定六亲。",
			s.YongShen, s.Topic))
	}
	if s.SelectNote != "" {
		out = append(out, s.SelectNote)
	}

	if s.YuanShen != "" {
		out = append(out, fmt.Sprintf("用神为%s，原神为%s（生用神，助我者），忌神为%s（克用神，伤我者），仇神为%s（生忌神）。",
			s.YongShen, s.YuanShen, s.JiShen, s.ChouShen))
	}

	y := s.YongYao
	if y == nil {
		out = append(out, "用神不上卦，主所问之事一时难以落实，或须借他人之力。宜另择时日再占，或以日月权代推之。")
		return out
	}

	// 用神旺衰
	switch y.Strength {
	case LevelStrong:
		out = append(out, fmt.Sprintf("用神（%s%s）旺相有力，所问之事本身条件不错。", y.GanZhi, y.WuXing))
	case LevelModerate:
		out = append(out, fmt.Sprintf("用神（%s%s）力量平平，事情成否更多取决于外部配合。", y.GanZhi, y.WuXing))
	default:
		out = append(out, fmt.Sprintf("用神（%s%s）衰弱无力，所问之事本身底气不足，宜先补条件。", y.GanZhi, y.WuXing))
	}

	// 用神动静
	if y.Moving {
		out = append(out, "用神发动，主事情会有明显变化或主动推进。")
	}
	if y.Special == string(StatusAnDong) {
		out = append(out, "用神暗动，主暗地里有推动力量，事情在不知不觉中起变化。")
	}
	if y.Special == string(StatusRiPo) {
		out = append(out, "用神日破，眼下受伤无力，宜等出空、逢生之后再论成败。")
	}
	if y.Special == string(StatusYuePo) {
		out = append(out, "用神月破，整月无力，须过了这个月才有转机。")
	}
	if y.IsKong {
		out = append(out, "用神落旬空，眼下使不上力或事情悬而未决，等出旬或逢冲填实才应事。")
	}
	if y.IsFuShen {
		out = append(out, "用神伏藏，主事未显现、当事人尚未出面，须待冲开飞神或伏神值日方见分晓。")
	}

	// 并见提示
	if s.Ambiguous {
		out = append(out, fmt.Sprintf("用神「%s」在卦中并见（%s），须两爻兼看，不可只取一爻。",
			s.YongShen, candidatesLabel(s.Candidates)))
	}

	// 忌神是否发动
	if s.JiShen != "" {
		for _, yy := range c.Yaos {
			if yy.LiuQin != s.JiShen {
				continue
			}
			if yy.IsChanging || yy.Special == StatusAnDong {
				out = append(out, fmt.Sprintf("忌神「%s」在%s发动，对用神不利，须防阻力、损失或对手动作。", s.JiShen, yy.Name))
			}
		}
	}
	// 原神是否发动
	if s.YuanShen != "" {
		for _, yy := range c.Yaos {
			if yy.LiuQin != s.YuanShen {
				continue
			}
			if yy.IsChanging || yy.Special == StatusAnDong {
				out = append(out, fmt.Sprintf("原神「%s」在%s发动，有力生扶用神，主有贵人、契机相助。", s.YuanShen, yy.Name))
			}
		}
	}

	return out
}
