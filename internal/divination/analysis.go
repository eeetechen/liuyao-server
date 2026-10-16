package divination

import "fmt"

// ---------- 五行旺衰（旺相休囚死） ----------

// WangShuaiState 旺衰状态
type WangShuaiState string

const (
	Wang  WangShuaiState = "旺" // 同月令五行，当令最强
	Xiang WangShuaiState = "相" // 月令所生，得生次强
	Xiu   WangShuaiState = "休" // 生月令，泄气平平
	Qiu   WangShuaiState = "囚" // 克月令，耗力较弱
	Si    WangShuaiState = "死" // 被月令所克，受制最弱
)

// WangShuaiDesc 旺衰释义
var WangShuaiDesc = map[WangShuaiState]string{
	Wang:  "当令，力量最强",
	Xiang: "得月令生扶，力量次强",
	Xiu:   "生月令而泄气，力量平平",
	Qiu:   "克月令而耗力，力量较弱",
	Si:    "被月令克制，力量最弱",
}

// GetWangShuai 依月令地支与爻五行判定旺相休囚死。
//
// 口诀：同我者旺，我生者相，生我者休，克我者囚，我克者死。
// 此处「我」指月令。
func GetWangShuai(yaoWuXing WuXing, monthZhi string) WangShuaiState {
	ling := WuXingOfZhi(monthZhi)
	switch {
	case yaoWuXing == ling:
		return Wang // 与月令同五行
	case sheng[ling] == yaoWuXing:
		return Xiang // 月令生爻
	case sheng[yaoWuXing] == ling:
		return Xiu // 爻生月令（泄气）
	case ke[yaoWuXing] == ling:
		return Qiu // 爻克月令（耗力）
	case ke[ling] == yaoWuXing:
		return Si // 月令克爻（受制）
	}
	return Xiu
}

// IsWangXiang 是否旺或相（有力）
func IsWangXiang(s WangShuaiState) bool { return s == Wang || s == Xiang }

// ---------- 特殊状态：暗动 / 日破 / 月破 ----------

// SpecialStatus 爻的特殊状态
type SpecialStatus string

const (
	StatusNone   SpecialStatus = "none"   // 无特殊状态
	StatusAnDong SpecialStatus = "anDong" // 暗动
	StatusRiPo   SpecialStatus = "riPo"   // 日破
	StatusYuePo  SpecialStatus = "yuePo"  // 月破
)

// SpecialStatusLabel 中文标签
var SpecialStatusLabel = map[SpecialStatus]string{
	StatusNone:   "",
	StatusAnDong: "暗动",
	StatusRiPo:   "日破",
	StatusYuePo:  "月破",
}

// JudgeSpecialStatus 判定暗动 / 日破 / 月破。
//
// 规则（传统纳甲筮法）：
//  1. 月破：静爻地支被月建所冲 —— 不论旺衰，整月无力，比日破更重。
//  2. 暗动：静爻地支被日辰所冲，且该爻「旺相」—— 旺相逢冲为暗动。
//  3. 日破：静爻地支被日辰所冲，但该爻「休囚死」—— 休囚逢冲为日破。
//
// 明动之爻（value 为 3 或 4）不参与日破/暗动判定，
// 因动爻本身已发动，冲之不能成破。
//
// 返回特殊状态与判定依据。
func JudgeSpecialStatus(yaoZhi string, monthZhi, dayZhi string, isChanging bool, ws WangShuaiState) (SpecialStatus, []string) {
	var evidence []string

	if isChanging {
		// 动爻不受冲破约束
		return StatusNone, evidence
	}

	// 1. 月破优先
	if LiuChongZhi[yaoZhi] == monthZhi {
		evidence = append(evidence, fmt.Sprintf("月建%s冲爻支%s，月破（整月无力）", monthZhi, yaoZhi))
		return StatusYuePo, evidence
	}

	// 2. 日冲 → 暗动 或 日破
	if LiuChongZhi[yaoZhi] == dayZhi {
		if IsWangXiang(ws) {
			evidence = append(evidence, fmt.Sprintf("日辰%s冲静爻%s，得月令而旺相，旺相逢冲为暗动", dayZhi, yaoZhi))
			return StatusAnDong, evidence
		}
		evidence = append(evidence, fmt.Sprintf("日辰%s冲静爻%s，休囚逢冲为日破", dayZhi, yaoZhi))
		return StatusRiPo, evidence
	}

	return StatusNone, evidence
}

// ---------- 爻的动静 ----------

// MoveState 爻的动静状态
type MoveState string

const (
	MoveStatic MoveState = "static"        // 静爻
	MoveMing   MoveState = "changing"      // 明动（摇卦得老阴老阳）
	MoveAnDong MoveState = "hidden_moving" // 暗动
	MoveBroken MoveState = "day_break"     // 日破/月破
)

// MoveStateLabel 中文标签
var MoveStateLabel = map[MoveState]string{
	MoveStatic: "静",
	MoveMing:   "明动",
	MoveAnDong: "暗动",
	MoveBroken: "破",
}

// ResolveMoveState 综合阴阳变化与特殊状态得出动静
func ResolveMoveState(isChanging bool, sp SpecialStatus) MoveState {
	if isChanging {
		return MoveMing
	}
	switch sp {
	case StatusAnDong:
		return MoveAnDong
	case StatusRiPo, StatusYuePo:
		return MoveBroken
	}
	return MoveStatic
}

// IsActive 该爻是否具备「发动」的作用力（明动或暗动）
func IsActive(isChanging bool, sp SpecialStatus) bool {
	return isChanging || sp == StatusAnDong
}

// ---------- 力量评级 ----------

// StrengthLevel 综合力量
type StrengthLevel string

const (
	LevelStrong   StrengthLevel = "strong"   // 有力
	LevelModerate StrengthLevel = "moderate" // 平平
	LevelWeak     StrengthLevel = "weak"     // 无力
)

// GetStrengthLevel 综合旺衰、特殊状态、旬空得出力量等级
func GetStrengthLevel(ws WangShuaiState, sp SpecialStatus, isKong bool, daySheng bool) StrengthLevel {
	// 日破/月破/静空 → 无力
	if sp == StatusRiPo || sp == StatusYuePo || isKong {
		return LevelWeak
	}
	// 暗动 → 有力
	if sp == StatusAnDong {
		return LevelStrong
	}
	if IsWangXiang(ws) {
		return LevelStrong
	}
	// 休囚但得日辰生扶，可算平平
	if ws == Xiu || daySheng {
		return LevelModerate
	}
	return LevelWeak
}

// ---------- 伏神 ----------

// FuShenInfo 伏神信息
type FuShenInfo struct {
	Position int    `json:"position"` // 伏于第几爻（1-6）
	LiuQin   string `json:"liuQin"`   // 六亲
	GanZhi   string `json:"ganZhi"`   // 干支
	Zhi      string `json:"zhi"`      // 地支
	WuXing   WuXing `json:"wuXing"`   // 五行
}

// calcFuShen 求伏神。
//
// 规则：若本卦六爻所现六亲不足五种，则取该卦「本宫卦」（八纯卦），
// 找出所缺六亲在本宫卦中的爻位，该爻即为伏神，伏于本卦同爻位之下。
// 伏神即以本宫卦为准，故其五行与六亲皆由本宫卦推出。
func calcFuShen(mark string, gong string, presentQin map[string]bool) []FuShenInfo {
	if len(presentQin) >= 5 {
		return nil
	}
	gongIdx := indexOf(BaGua[:], gong)
	if gongIdx < 0 {
		return nil
	}
	// 本宫卦 = 内外皆该宫之卦
	palaceMark := BaGuaBinary[gongIdx] + BaGuaBinary[gongIdx]
	palaceNaJia := getNaJia(palaceMark)
	palaceGongWuXing := BaGuaWuXing[gongIdx]

	var out []FuShenInfo
	for i := 0; i < 6; i++ {
		qin := GetLiuQin(palaceGongWuXing, palaceNaJia[i].WuXing)
		if presentQin[qin] {
			continue // 本卦已现，不伏
		}
		out = append(out, FuShenInfo{
			Position: i + 1,
			LiuQin:   qin,
			GanZhi:   palaceNaJia[i].GanZhi,
			Zhi:      palaceNaJia[i].Zhi,
			WuXing:   palaceNaJia[i].WuXing,
		})
	}
	return out
}

// ---------- 变卦 ----------

// BianGua 变卦信息
type BianGua struct {
	Mark     string   `json:"mark"`
	Name     string   `json:"name"`
	Gong     string   `json:"gong"`
	Element  WuXing   `json:"element"`
	UpperGua string   `json:"upperGua"`
	LowerGua string   `json:"lowerGua"`
	ShiPos   int      `json:"shiPos"`
	YingPos  int      `json:"yingPos"`
	NaJia    []NaJia  `json:"naJia"`
	LiuQin   []string `json:"liuQin"`
	Type     string   `json:"type,omitempty"`
}

// calcBianGua 由动爻推变卦。
// 动爻：老阳(值9/参数4)变阴，老阴(值6/参数1)变阳；静爻不变。
func calcBianGua(mark string, changing []int) *BianGua {
	if len(changing) == 0 {
		return nil
	}
	b := []byte(mark)
	for _, pos := range changing {
		if pos < 0 || pos >= 6 {
			continue
		}
		if b[pos] == '1' {
			b[pos] = '0'
		} else {
			b[pos] = '1'
		}
	}
	bMark := string(b)

	shi := setShiYao(bMark)
	gong := palace(bMark, shi.Shi)
	gongIdx := indexOf(BaGua[:], gong)
	if gongIdx < 0 {
		return nil
	}
	gongWuXing := BaGuaWuXing[gongIdx]

	naJia := getNaJia(bMark)
	qin := make([]string, 6)
	for i := range naJia {
		qin[i] = GetLiuQin(gongWuXing, naJia[i].WuXing)
	}

	// 变卦六亲按「本卦所伏之宫」论，此处同时给出变卦自身卦宫
	bg := &BianGua{
		Mark:     bMark,
		Name:     Gua64[bMark],
		Gong:     gong,
		Element:  gongWuXing,
		UpperGua: BaGuaByBinary(bMark[3:]),
		LowerGua: BaGuaByBinary(bMark[:3]),
		ShiPos:   shi.Shi,
		YingPos:  shi.Ying,
		NaJia:    naJia,
		LiuQin:   qin,
	}
	switch {
	case IsLiuChong(bMark):
		bg.Type = "六冲"
	case IsLiuHe(bMark):
		bg.Type = "六合"
	}
	return bg
}
