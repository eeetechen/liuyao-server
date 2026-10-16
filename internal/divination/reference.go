package divination

// ============================================================
// 五行生克参考表（供使用者查询印证）
// ============================================================

// TianGanWuXing 天干五行：甲乙木、丙丁火、戊己土、庚辛金、壬癸水
var TianGanWuXing = map[string]WuXing{
	"甲": Mu, "乙": Mu,
	"丙": Huo, "丁": Huo,
	"戊": Tu, "己": Tu,
	"庚": Jin, "辛": Jin,
	"壬": Shui, "癸": Shui,
}

// TianGanYinYang 天干阴阳：甲丙戊庚壬为阳，乙丁己辛癸为阴
var TianGanYinYang = map[string]YinYang{
	"甲": Yang, "乙": Yin,
	"丙": Yang, "丁": Yin,
	"戊": Yang, "己": Yin,
	"庚": Yang, "辛": Yin,
	"壬": Yang, "癸": Yin,
}

// ZhiYinYang 地支阴阳（以子为阳起，顺次交替）
var ZhiYinYang = map[string]YinYang{
	"子": Yang, "丑": Yin, "寅": Yang, "卯": Yin,
	"辰": Yang, "巳": Yin, "午": Yang, "未": Yin,
	"申": Yang, "酉": Yin, "戌": Yang, "亥": Yin,
}

// ShengXiao 十二生肖，与 DiZhi 同序
var ShengXiao = [12]string{"鼠", "牛", "虎", "兔", "龙", "蛇", "马", "羊", "猴", "鸡", "狗", "猪"}

// WuXingColor 五行对应的传统五色
var WuXingColor = map[WuXing]string{
	Mu: "青", Huo: "赤", Tu: "黄", Jin: "白", Shui: "黑",
}

// WuXingDirection 五行方位
var WuXingDirection = map[WuXing]string{
	Mu: "东", Huo: "南", Tu: "中央", Jin: "西", Shui: "北",
}

// WuXingSeason 五行主令季节。
//
// 土旺于四季之末（辰未戌丑四月），故称「四季月」。
var WuXingSeason = map[WuXing]string{
	Mu: "春季", Huo: "夏季", Tu: "四季月", Jin: "秋季", Shui: "冬季",
}

// ---------- 参考表数据结构 ----------

// RefWuXing 五行条目
type RefWuXing struct {
	WuXing    WuXing `json:"wuXing"`
	Color     string `json:"color"`     // 青赤黄白黑
	Direction string `json:"direction"` // 东南中西北
	Season    string `json:"season"`    // 春/夏/四季月/秋/冬
	Sheng     WuXing `json:"sheng"`     // 我生者
	Ke        WuXing `json:"ke"`        // 我克者
	BeiSheng  WuXing `json:"beiSheng"`  // 生我者
	BeiKe     WuXing `json:"beiKe"`     // 克我者
}

// RefGanZhi 干支条目
type RefGanZhi struct {
	Name      string `json:"name"` // 甲 / 子
	Kind      string `json:"kind"` // gan / zhi
	WuXing    WuXing `json:"wuXing"`
	YinYang   string `json:"yinYang"`             // 阳 / 阴
	YinYangEn string `json:"yinYangEn"`           // yang / yin，供程序判断
	ShengXiao string `json:"shengXiao,omitempty"` // 仅地支
}

// yinYangCn 将内部枚举转为中文
func yinYangCn(y YinYang) string {
	if y == Yang {
		return "阳"
	}
	return "阴"
}

// RefPair 地支两两关系
type RefPair struct {
	A     string `json:"a"`
	B     string `json:"b"`
	Hua   WuXing `json:"hua,omitempty"` // 六合化气
	Label string `json:"label"`         // 如「子午相冲」「子丑合土」
}

// RefSanHe 三合局条目
type RefSanHe struct {
	Name     string   `json:"name"`
	Branches []string `json:"branches"`
	Element  WuXing   `json:"element"`
	Colors   []string `json:"colors"` // 三支各自五行色
}

// RefShengKe 五行生克条目（一句话白话）
type RefShengKe struct {
	From  WuXing `json:"from"`
	To    WuXing `json:"to"`
	Kind  string `json:"kind"`  // 生 / 克
	Text  string `json:"text"`  // 如「水生木」「水克火」
	Plain string `json:"plain"` // 白话解释
}

// Reference 五行生克参考表
type Reference struct {
	WuXing   []RefWuXing  `json:"wuXing"`
	ShengKe  []RefShengKe `json:"shengKe"`
	TianGan  []RefGanZhi  `json:"tianGan"`
	DiZhi    []RefGanZhi  `json:"diZhi"`
	LiuChong []RefPair    `json:"liuChong"`
	LiuHe    []RefPair    `json:"liuHe"`
	SanHe    []RefSanHe   `json:"sanHe"`
	BanHe    []RefPair    `json:"banHe"`
	Xing     []RefXing    `json:"xing"`
	Hai      []RefPair    `json:"hai"`
	LiuShen  []RefLiuShen `json:"liuShen"`
	Notes    []string     `json:"notes"`
}

// RefLiuShen 六神条目
type RefLiuShen struct {
	Name     string   `json:"name"`     // 青龙
	Alias    string   `json:"alias"`    // 青龙（六兽）
	WuXing   WuXing   `json:"wuXing"`   // 所属五行
	Color    string   `json:"color"`    // 对应色
	StartGan []string `json:"startGan"` // 起初爻的日干
	Nature   string   `json:"nature"`   // 吉凶倾向：吉 / 凶 / 中
	Image    []string `json:"image"`    // 意象（人物、事物、性情等）
	Meaning  string   `json:"meaning"`  // 一句话总括
	Note     string   `json:"note"`     // 取象提示
}

// RefXing 三刑条目
type RefXing struct {
	Name     string   `json:"name"`
	Branches []string `json:"branches"`
	SelfXing bool     `json:"selfXing"`
	Label    string   `json:"label"`
}

// 五行相生顺序表（木→火→土→金→水）
var wuXingShengOrder = []WuXing{Mu, Huo, Tu, Jin, Shui}

// BuildReference 构造五行生克参考表。
func BuildReference() Reference {
	ref := Reference{
		WuXing:   []RefWuXing{},
		ShengKe:  []RefShengKe{},
		TianGan:  []RefGanZhi{},
		DiZhi:    []RefGanZhi{},
		LiuChong: []RefPair{},
		LiuHe:    []RefPair{},
		SanHe:    []RefSanHe{},
		BanHe:    []RefPair{},
		Xing:     []RefXing{},
		Hai:      []RefPair{},
		LiuShen:  []RefLiuShen{},
		Notes:    []string{},
	}

	// 五行属性
	for _, wx := range WuXingOrder {
		ref.WuXing = append(ref.WuXing, RefWuXing{
			WuXing:    wx,
			Color:     WuXingColor[wx],
			Direction: WuXingDirection[wx],
			Season:    WuXingSeason[wx],
			Sheng:     sheng[wx],
			Ke:        ke[wx],
			BeiSheng:  inverseOf(sheng, wx),
			BeiKe:     inverseOf(ke, wx),
		})
	}

	// 五行相生
	for _, wx := range wuXingShengOrder {
		to := sheng[wx]
		ref.ShengKe = append(ref.ShengKe, RefShengKe{
			From: wx, To: to, Kind: "生",
			Text:  string(wx) + "生" + string(to),
			Plain: string(wx) + "能滋养" + string(to) + "，" + string(wx) + "气旺则" + string(to) + "得助",
		})
	}
	// 五行相克
	for _, wx := range wuXingShengOrder {
		to := ke[wx]
		ref.ShengKe = append(ref.ShengKe, RefShengKe{
			From: wx, To: to, Kind: "克",
			Text:  string(wx) + "克" + string(to),
			Plain: string(wx) + "能制约" + string(to) + "，" + string(wx) + "气旺则" + string(to) + "受制",
		})
	}

	// 十天干
	for _, g := range TianGan {
		ref.TianGan = append(ref.TianGan, RefGanZhi{
			Name: g, Kind: "gan", WuXing: TianGanWuXing[g],
			YinYang: yinYangCn(TianGanYinYang[g]), YinYangEn: string(TianGanYinYang[g]),
		})
	}
	// 十二地支
	for i, z := range DiZhi {
		ref.DiZhi = append(ref.DiZhi, RefGanZhi{
			Name: z, Kind: "zhi", WuXing: ZhiWuXing[i],
			YinYang: yinYangCn(ZhiYinYang[z]), YinYangEn: string(ZhiYinYang[z]),
			ShengXiao: ShengXiao[i],
		})
	}

	// 六冲（每对只列一次）
	for i, z := range DiZhi {
		opp := LiuChongZhi[z]
		if zhiIndex(opp) < i {
			continue // 已由前者列出
		}
		ref.LiuChong = append(ref.LiuChong, RefPair{
			A: z, B: opp, Label: z + opp + "相冲",
		})
	}

	// 六合（每对只列一次）
	for i, z := range DiZhi {
		info := LiuHeZhi[z]
		if zhiIndex(info.Partner) < i {
			continue
		}
		ref.LiuHe = append(ref.LiuHe, RefPair{
			A: z, B: info.Partner, Hua: info.Result,
			Label: z + info.Partner + "合化" + string(info.Result),
		})
	}

	// 三合局
	for _, g := range SanHeGroups {
		colors := make([]string, 0, 3)
		for _, b := range g.Branches {
			colors = append(colors, WuXingColor[WuXingOfZhi(b)])
		}
		ref.SanHe = append(ref.SanHe, RefSanHe{
			Name: g.Name, Branches: []string{g.Branches[0], g.Branches[1], g.Branches[2]},
			Element: g.Element, Colors: colors,
		})
	}

	// 半合（每对只列一次）
	seen := map[string]bool{}
	for _, p := range BanHePairs {
		key := p.A + p.B
		rev := p.B + p.A
		if seen[key] || seen[rev] {
			continue
		}
		seen[key] = true
		ref.BanHe = append(ref.BanHe, RefPair{
			A: p.A, B: p.B, Hua: p.Element,
			Label: p.A + p.B + "半合" + string(p.Element) + "（缺" + p.Missing + "）",
		})
	}

	// 三刑
	for _, g := range SanXingGroups {
		label := ""
		if g.SelfXing {
			label = g.Branches[0] + "自刑"
		} else {
			label = joinBranches(g.Branches) + "相刑（" + g.Name + "）"
		}
		ref.Xing = append(ref.Xing, RefXing{
			Name: g.Name, Branches: g.Branches, SelfXing: g.SelfXing, Label: label,
		})
	}

	// 六害（每对只列一次）
	for i, z := range DiZhi {
		opp := LiuHaiZhi[z]
		if zhiIndex(opp) < i {
			continue
		}
		ref.Hai = append(ref.Hai, RefPair{
			A: z, B: opp, Label: z + opp + "相害",
		})
	}

	// 六神意象
	ref.LiuShen = buildLiuShenRef()

	ref.Notes = []string{
		"地支以子为阳起，顺次交替阴阳；天干以甲为阳起，顺次交替。",
		"六神（六兽）按日干起初爻，顺次上排：甲乙起青龙、丙丁起朱雀、戊起勾陈、己起螣蛇、庚辛起白虎、壬癸起玄武。",
		"六神五行：青龙木、朱雀火、勾陈土、螣蛇土、白虎金、玄武水。（十二将体系中螣蛇属火，与此不同，勿混。）",
		"三刑：寅巳申为无恩之刑，丑戌未为恃势之刑，子卯为无礼之刑，辰午酉亥为自刑。刑主刑伤、纠纷、官讼。",
		"六害：子未、丑午、寅巳、卯辰、申亥、酉戌。害主暗损、猜忌，其力弱于刑冲。",
		"用神取自所问之事：问财取妻财，问工作官讼取官鬼，问学业文书房产取父母，问子女平安取子孙，问同辈合作取兄弟。",
		"六合化气：子丑合土、寅亥合木、卯戌合火、辰酉合金、巳申合水、午未合土。",
		"三合局需三支俱全方能成局，仅两支者为半合（须为生半合或墓半合）。",
		"六冲之中，被日辰冲者：旺相为暗动，休囚为日破；被月建冲者为月破。",
		"五行旺衰以月建为准：同月令者旺，月令所生者相，生月令者休，克月令者囚，被月令克者死。",
	}

	return ref
}

// inverseOf 在映射中反查「谁指向 v」
func inverseOf(m map[WuXing]WuXing, v WuXing) WuXing {
	for k, to := range m {
		if to == v {
			return k
		}
	}
	return ""
}

// buildLiuShenRef 构造六神意象参考表。
//
// 五行归属据《起身例纳甲·六神总论》：
//
//	青龙木神、朱雀火神、勾陈土神、螣蛇土神、白虎金神、玄武水神。
//
// 注意：六壬十二将体系中「螣蛇火将、勾陈土将」，与本体系不同，不可混用。
func buildLiuShenRef() []RefLiuShen {
	return []RefLiuShen{
		{
			Name: "青龙", Alias: "青龙", WuXing: Mu, Color: "青",
			StartGan: []string{"甲", "乙"}, Nature: "吉",
			Image: []string{
				"人物：贵人、长辈、君子、正派之人、有身份者",
				"事物：喜庆、财喜、升迁、婚嫁、生育、酒色、宴乐",
				"性情：仁厚、端庄、有涵养",
				"场所：东方、正堂、富贵之地",
			},
			Meaning: "主吉庆喜事，为六神中最吉者；旺相临用神主进财、升迁、得贵人助。",
			Note:    "青龙主喜庆亦主酒色，得财多而用之当则为吉，用之不当则主酒色破财。",
		},
		{
			Name: "朱雀", Alias: "朱雀", WuXing: Huo, Color: "赤",
			StartGan: []string{"丙", "丁"}, Nature: "中",
			Image: []string{
				"人物：口舌之人、教师、文人、消息传递者",
				"事物：文书、信息、消息、口舌、是非、争讼、火灾、电讯",
				"性情：急躁、多言、好辩",
				"场所：南方、厅堂、炉灶、文书之地",
			},
			Meaning: "主口舌是非与文书信息，吉则文书之喜、音信至，凶则口舌争讼。",
			Note:    "朱雀临用神，须看所占何事：占文书消息为吉，占讼事争端则主口舌。",
		},
		{
			Name: "勾陈", Alias: "勾陈", WuXing: Tu, Color: "黄",
			StartGan: []string{"戊"}, Nature: "凶",
			Image: []string{
				"人物：田土之人、农人、中介、牵绊者",
				"事物：田土、房屋、产业、拘束、迟滞、勾连、牢狱、跌打损伤",
				"性情：固执、迟缓、纠缠不放",
				"场所：中央、田地、房产、仓库",
			},
			Meaning: "主田土房产与迟滞勾连，多主事情拖延、被牵绊或与土地产业有关。",
			Note:    "勾陈主「勾」主「滞」：问产事宜，问事则主拖沓难了、易生纠缠。",
		},
		{
			Name: "螣蛇", Alias: "螣蛇（腾蛇）", WuXing: Tu, Color: "黄",
			StartGan: []string{"己"}, Nature: "凶",
			Image: []string{
				"人物：虚惊之人、心事重者、怪异之人、术士",
				"事物：虚惊、怪异、梦寐、忧疑、缠绕、惊恐、血脉之疾",
				"性情：多疑、敏感、心神不定",
				"场所：曲折之地、暗处、床上、绳索缠绕之处",
			},
			Meaning: "主虚惊怪异与忧疑缠绵，多主心中不安、事情反复缠绕、或遇怪异之事。",
			Note:    "螣蛇主虚惊，多非实祸，然忧疑缠绕最耗心神；占病主血脉、神经之疾。",
		},
		{
			Name: "白虎", Alias: "白虎", WuXing: Jin, Color: "白",
			StartGan: []string{"庚", "辛"}, Nature: "凶",
			Image: []string{
				"人物：凶恶之人、兵卒、医生、执法者、丧事之人",
				"事物：凶伤、灾祸、疾病、丧事、血光、刀兵、道路、金铁",
				"性情：刚猛、严厉、刚直",
				"场所：西方、道路、丧葬之地、医院",
			},
			Meaning: "主凶伤丧病，为六神中最凶者；旺相临用神主凶事、伤病、丧亡之应。",
			Note:    "白虎亦主刚直与威权，占官职、执法反为得力；占病则主重症。",
		},
		{
			Name: "玄武", Alias: "玄武", WuXing: Shui, Color: "黑",
			StartGan: []string{"壬", "癸"}, Nature: "凶",
			Image: []string{
				"人物：盗贼、小人、骗子、暗昧之人、阴私之人",
				"事物：盗窃、暗昧、欺骗、遗失、阴私、水厄、逃亡",
				"性情：阴险、暗昧、好贪",
				"场所：北方、暗处、水中、厕所、隐蔽之地",
			},
			Meaning: "主暗昧盗窃与欺瞒，多主暗中损失、遭人欺骗、或事有隐情不可告人。",
			Note:    "玄武主「暗」：凡失物、失盗、被欺多属之，亦主事情有隐情未明。",
		},
	}
}
