package divination

import "fmt"

// ============================================================
// 三刑 · 六害
// ============================================================

// LiuHaiZhi 地支六害（亦称六穿）。
//
// 子未相害、丑午相害、寅巳相害、卯辰相害、申亥相害、酉戌相害。
var LiuHaiZhi = map[string]string{
	"子": "未", "未": "子",
	"丑": "午", "午": "丑",
	"寅": "巳", "巳": "寅",
	"卯": "辰", "辰": "卯",
	"申": "亥", "亥": "申",
	"酉": "戌", "戌": "酉",
}

// XingGroup 三刑组
type XingGroup struct {
	Branches []string // 参与地支
	Name     string   // 刑名
	SelfXing bool     // 是否自刑
}

// SanXingGroups 三刑。
//
// 传统三刑：
//   - 寅巳申 —— 无恩之刑
//   - 丑戌未 —— 恃势之刑
//   - 子卯   —— 无礼之刑
//   - 辰、午、酉、亥 —— 各自刑（自刑）
//
// 另有「子卯」一说为相刑，故单列为一组两支。
var SanXingGroups = []XingGroup{
	{[]string{"寅", "巳", "申"}, "无恩之刑", false},
	{[]string{"丑", "戌", "未"}, "恃势之刑", false},
	{[]string{"子", "卯"}, "无礼之刑", false},
	{[]string{"辰"}, "辰自刑", true},
	{[]string{"午"}, "午自刑", true},
	{[]string{"酉"}, "酉自刑", true},
	{[]string{"亥"}, "亥自刑", true},
}

// XingPair 检出的相刑关系
type XingPair struct {
	Name     string      `json:"name"`     // 刑名，如「无恩之刑」
	Branches []string    `json:"branches"` // 涉事地支（去重）
	Members  []BranchRef `json:"members"`  // 涉事各方（含位置）
	SelfXing bool        `json:"selfXing"` // 是否自刑
	Complete bool        `json:"complete"` // 三支是否俱全（自刑恒为 true）
	Desc     string      `json:"desc"`
}

// HaiPair 检出的相害关系
type HaiPair struct {
	A    BranchRef `json:"a"`
	B    BranchRef `json:"b"`
	Kind string    `json:"kind"`
	Desc string    `json:"desc"`
}

// BuildXingHai 求三刑与六害。
//
// 参与范围为六爻地支加月建、日辰。日辰与月建之间不论。
func BuildXingHai(yaos []YaoInfo, t GanZhiTime) (xing []XingPair, hai []HaiPair) {
	xing, hai = []XingPair{}, []HaiPair{}

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

	// ---------- 三刑 ----------
	for _, g := range SanXingGroups {
		// 自刑：同一地支出现两次及以上
		if g.SelfXing {
			refs := byZhi[g.Branches[0]]
			if len(refs) < 2 {
				continue
			}
			xing = append(xing, XingPair{
				Name: g.Name, Branches: []string{g.Branches[0]},
				Members: refs, SelfXing: true, Complete: true,
				Desc: fmt.Sprintf("%s（%s）", g.Name, branchRefsLabel(refs)),
			})
			continue
		}

		// 非自刑：须所有地支俱现
		allPresent := true
		for _, b := range g.Branches {
			if len(byZhi[b]) == 0 {
				allPresent = false
				break
			}
		}
		if !allPresent {
			continue
		}

		var members []BranchRef
		var branches []string
		for _, b := range g.Branches {
			branches = append(branches, b)
			members = append(members, byZhi[b]...)
		}
		xing = append(xing, XingPair{
			Name: g.Name, Branches: branches, Members: members,
			SelfXing: false, Complete: true,
			Desc: fmt.Sprintf("%s（%s，%s）", g.Name, joinBranches(branches), branchRefsLabel(members)),
		})
	}

	// ---------- 六害 ----------
	for i := 0; i < len(DiZhi); i++ {
		a := DiZhi[i]
		b, ok := LiuHaiZhi[a]
		if !ok || zhiIndex(b) < i {
			continue // 已由前者列出，避免重复
		}
		if len(byZhi[a]) == 0 || len(byZhi[b]) == 0 {
			continue
		}
		// 逐对组合（同支可能多爻出现）
		for _, ra := range byZhi[a] {
			for _, rb := range byZhi[b] {
				if ra.Scope != "yao" && rb.Scope != "yao" {
					continue // 日辰与月建之间不论
				}
				hai = append(hai, HaiPair{
					A: ra, B: rb, Kind: pairKind(ra, rb),
					Desc: fmt.Sprintf("%s%s 与 %s%s 相害", ra.Label, ra.Zhi, rb.Label, rb.Zhi),
				})
			}
		}
	}

	return xing, hai
}

// pairKind 判断关系双方的类型
func pairKind(a, b BranchRef) string {
	yaoA, yaoB := a.Scope == "yao", b.Scope == "yao"
	switch {
	case yaoA && yaoB:
		return "yao-yao"
	case yaoA:
		return "yao-" + b.Scope
	case yaoB:
		return "yao-" + a.Scope
	}
	return ""
}

func joinBranches(bs []string) string {
	out := ""
	for _, b := range bs {
		out += b
	}
	return out
}

func branchRefsLabel(refs []BranchRef) string {
	parts := make([]string, 0, len(refs))
	for _, r := range refs {
		parts = append(parts, r.Label+r.Zhi)
	}
	return joinWith(parts, "、")
}

func joinWith(parts []string, sep string) string {
	out := ""
	for i, p := range parts {
		if i > 0 {
			out += sep
		}
		out += p
	}
	return out
}

// XingAdvice 三刑断语
func XingAdvice() string {
	return "刑主刑伤、纠纷、官讼、身体损伤，多指事情内部有摩擦或人际失和，其性烈于害。"
}

// HaiAdvice 六害断语
func HaiAdvice() string {
	return "害主暗损、彼此猜忌、小事牵绊，其力弱于刑冲，多主暗中受损而未必显于明处。"
}
