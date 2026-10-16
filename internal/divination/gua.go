package divination

import "strings"

// ---------- 五行生克 ----------

// sheng 五行相生：木→火→土→金→水→木
var sheng = map[WuXing]WuXing{Mu: Huo, Huo: Tu, Tu: Jin, Jin: Shui, Shui: Mu}

// ke 五行相克：木→土→水→火→金→木
var ke = map[WuXing]WuXing{Mu: Tu, Tu: Shui, Shui: Huo, Huo: Jin, Jin: Mu}

// LiuChongZhi 地支六冲
var LiuChongZhi = map[string]string{
	"子": "午", "午": "子", "丑": "未", "未": "丑",
	"寅": "申", "申": "寅", "卯": "酉", "酉": "卯",
	"辰": "戌", "戌": "辰", "巳": "亥", "亥": "巳",
}

// LiuHeZhi 地支六合及其合化五行。
//
// 合化口诀：子丑合化土、寅亥合化木、卯戌合化火、
//
//	辰酉合化金、巳申合化水、午未合化土。
//
// 关于「午未」有异说：多数命理典籍及严谨实现（如 taibu）作合化「土」
// （未为土之本气，且午火生未土），亦有作合化「火」者，另有一派主张
// 午未为日月之合、合而不化。此处采合化土。若所需口径不同，改本表即可。
var LiuHeZhi = map[string]struct {
	Partner string
	Result  WuXing
}{
	"子": {"丑", Tu}, "丑": {"子", Tu},
	"寅": {"亥", Mu}, "亥": {"寅", Mu},
	"卯": {"戌", Huo}, "戌": {"卯", Huo},
	"辰": {"酉", Jin}, "酉": {"辰", Jin},
	"巳": {"申", Shui}, "申": {"巳", Shui},
	"午": {"未", Tu}, "未": {"午", Tu},
}

// ---------- 索引辅助 ----------

func indexOf(arr []string, v string) int {
	for i, s := range arr {
		if s == v {
			return i
		}
	}
	return -1
}

func zhiIndex(z string) int { return indexOf(DiZhi[:], z) }
func ganIndex(g string) int { return indexOf(TianGan[:], g) }

// BaGuaIndexByBinary 由三爻二进制码求八卦索引
func BaGuaIndexByBinary(bin string) int { return indexOf(BaGuaBinary[:], bin) }

// BaGuaByBinary 由三爻二进制码求卦名
func BaGuaByBinary(bin string) string {
	if i := BaGuaIndexByBinary(bin); i >= 0 {
		return BaGua[i]
	}
	return ""
}

// WuXingOfZhi 地支五行
func WuXingOfZhi(z string) WuXing {
	if i := zhiIndex(z); i >= 0 {
		return ZhiWuXing[i]
	}
	return ""
}

// ---------- 卦宫 / 世应 ----------

// soul 判断游魂、归魂（"天同人地不同为游魂，人不同天地同为归魂"）
func soul(mark string) string {
	nei, wai := mark[:3], mark[3:]
	if wai[1] == nei[1] {
		if wai[0] != nei[0] && wai[2] != nei[2] {
			return "游魂"
		}
	} else if wai[0] == nei[0] && wai[2] == nei[2] {
		return "归魂"
	}
	return ""
}

type shiYaoInfo struct {
	Shi  int // 世爻位 1-6
	Ying int // 应爻位 1-6
}

// setShiYao 依《寻世诀》定世应：
//
//	天同二世天变五，地同四世地变初。
//	本宫六世三世异，人同游魂人变归。
func setShiYao(mark string) shiYaoInfo {
	nei, wai := mark[:3], mark[3:]
	mk := func(shi int) shiYaoInfo {
		ying := shi + 3
		if shi > 3 {
			ying = shi - 3
		}
		return shiYaoInfo{Shi: shi, Ying: ying}
	}

	// 天同二世，天变五（天 = 三爻 / 六爻）
	if wai[2] == nei[2] {
		if wai[1] != nei[1] && wai[0] != nei[0] {
			return mk(2)
		}
	} else if wai[1] == nei[1] && wai[0] == nei[0] {
		return mk(5)
	}

	// 人同游魂，人变归（人 = 二爻 / 五爻）
	if wai[1] == nei[1] {
		if wai[0] != nei[0] && wai[2] != nei[2] {
			return mk(4)
		}
	} else if wai[0] == nei[0] && wai[2] == nei[2] {
		return mk(3)
	}

	// 地同四世，地变初（地 = 初爻 / 四爻）
	if wai[0] == nei[0] {
		if wai[1] != nei[1] && wai[2] != nei[2] {
			return mk(4)
		}
	} else if wai[1] == nei[1] && wai[2] == nei[2] {
		return mk(1)
	}

	// 本宫六世
	if wai == nei {
		return mk(6)
	}
	// 三世异
	return mk(3)
}

// invertBinary 三爻二进制取反（认宫诀"四五游魂内变更"）
func invertBinary(bin string) string {
	var b strings.Builder
	for _, c := range bin {
		if c == '1' {
			b.WriteByte('0')
		} else {
			b.WriteByte('1')
		}
	}
	return b.String()
}

// palace 依《认宫诀》定卦宫。shi 为世爻位（1-6）：
//
//	一二三六外卦宫，四五游魂内变更。
//	若问归魂何所取，归魂内卦是本宫。
func palace(mark string, shi int) string {
	nei, wai := mark[:3], mark[3:]
	hun := soul(mark)

	if hun == "归魂" {
		return BaGuaByBinary(nei)
	}
	if shi == 1 || shi == 2 || shi == 3 || shi == 6 {
		return BaGuaByBinary(wai)
	}
	if shi == 4 || shi == 5 || hun == "游魂" {
		return BaGuaByBinary(invertBinary(nei))
	}
	return ""
}

// ---------- 卦象性质 ----------

// IsLiuChong 判断是否六冲卦。
// 八纯卦（内外卦相同）为六冲；另加天雷无妄、雷天大壮两卦。
func IsLiuChong(mark string) bool {
	nei, wai := mark[:3], mark[3:]
	if nei == wai {
		return true
	}
	if (nei == "100" && wai == "111") || (nei == "111" && wai == "100") {
		return true
	}
	return false
}

// IsLiuHe 判断是否六合卦
func IsLiuHe(mark string) bool {
	name := Gua64[mark]
	for _, k := range LiuHeGua {
		if strings.Contains(name, k) {
			return true
		}
	}
	return false
}

// ---------- 纳甲 ----------

// NaJia 单爻纳甲结果
type NaJia struct {
	Gan    string `json:"gan"`
	Zhi    string `json:"zhi"`
	GanZhi string `json:"ganZhi"`
	WuXing WuXing `json:"wuXing"`
}

// getNaJia 求六爻纳甲（自下而上，索引 0 = 初爻）
func getNaJia(mark string) []NaJia {
	neiIdx := BaGuaIndexByBinary(mark[:3]) // 内卦
	waiIdx := BaGuaIndexByBinary(mark[3:]) // 外卦

	nei := naJiaTable[neiIdx]
	wai := naJiaTable[waiIdx]

	out := make([]NaJia, 0, 6)
	for _, z := range nei.InnerZhi {
		out = append(out, NaJia{Gan: nei.Gan, Zhi: z, GanZhi: nei.Gan + z, WuXing: WuXingOfZhi(z)})
	}
	for _, z := range wai.OuterZhi {
		out = append(out, NaJia{Gan: wai.OuterGan, Zhi: z, GanZhi: wai.OuterGan + z, WuXing: WuXingOfZhi(z)})
	}
	return out
}

// IsValidLiuQin 判断是否为合法六亲名
func IsValidLiuQin(q string) bool {
	for _, v := range LiuQin {
		if v == q {
			return true
		}
	}
	return false
}

// ---------- 六亲 ----------

// GetLiuQin 由「卦宫五行」与「爻五行」推六亲。
//
// 生我者父母，我生者子孙，克我者官鬼，我克者妻财，比和者兄弟。
func GetLiuQin(gongWuXing, yaoWuXing WuXing) string {
	w1 := indexOf(wuXingStrings(), string(gongWuXing))
	w2 := indexOf(wuXingStrings(), string(yaoWuXing))
	if w1 < 0 || w2 < 0 {
		return ""
	}
	ws := w1 - w2
	if ws < 0 {
		ws += 5
	}
	return LiuQin[ws]
}

func wuXingStrings() []string {
	out := make([]string, len(WuXingOrder))
	for i, w := range WuXingOrder {
		out[i] = string(w)
	}
	return out
}

// ---------- 六神 ----------

// LiuShenStart 依日干定初爻所起六神：
// 甲乙起青龙、丙丁起朱雀、戊起勾陈、己起螣蛇、庚辛起白虎、壬癸起玄武。
func LiuShenStart(dayGan string) int {
	switch dayGan {
	case "甲", "乙":
		return 0
	case "丙", "丁":
		return 1
	case "戊":
		return 2
	case "己":
		return 3
	case "庚", "辛":
		return 4
	case "壬", "癸":
		return 5
	}
	return 0
}

// GetLiuShen 返回自下而上的六神序列
func GetLiuShen(dayGan string) [6]string {
	start := LiuShenStart(dayGan)
	var out [6]string
	for i := 0; i < 6; i++ {
		out[i] = LiuShen[(start+i)%6]
	}
	return out
}

// ---------- 旬空 ----------

// XunKongOf 由干支求旬空（不依赖外部库，用于自测）
func XunKongOf(ganZhi string) string {
	r := []rune(ganZhi)
	if len(r) < 2 {
		return ""
	}
	gm := ganIndex(string(r[0]))
	zm := zhiIndex(string(r[1]))
	if gm < 0 || zm < 0 {
		return ""
	}
	if gm == zm || zm < gm {
		zm += 12
	}
	xk := (zm-gm)/2 - 1
	if xk < 0 || xk >= len(XunKong) {
		return ""
	}
	return XunKong[xk]
}

// IsKong 判断某地支是否落入旬空
func IsKong(zhi, xunKong string) bool {
	return strings.Contains(xunKong, zhi)
}
