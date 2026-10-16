package divination

import "testing"

// ============ 旺相休囚死 ============

func TestWangShuaiWuXing(t *testing.T) {
	// 月建为寅（木）：木旺、火相、水休、金囚、土死
	cases := []struct {
		yao  WuXing
		want WangShuaiState
	}{
		{Mu, Wang},   // 同月令
		{Huo, Xiang}, // 月令生
		{Shui, Xiu},  // 生月令
		{Jin, Qiu},   // 克月令
		{Tu, Si},     // 被月令克
	}
	for _, c := range cases {
		if got := GetWangShuai(c.yao, "寅"); got != c.want {
			t.Errorf("GetWangShuai(%s, 寅) = %s, want %s", c.yao, got, c.want)
		}
	}
}

func TestWangShuaiAllMonths(t *testing.T) {
	// 逐月校验：月建自身五行必旺，月令所生必相
	for _, mz := range DiZhi {
		ling := WuXingOfZhi(mz)
		if got := GetWangShuai(ling, mz); got != Wang {
			t.Errorf("月建 %s(%s) 应旺，实际 %s", mz, ling, got)
		}
		if got := GetWangShuai(sheng[ling], mz); got != Xiang {
			t.Errorf("月建 %s 所生 %s 应相，实际 %s", mz, sheng[ling], got)
		}
		if got := GetWangShuai(ke[ling], mz); got != Si {
			t.Errorf("月建 %s 所克 %s 应死，实际 %s", mz, ke[ling], got)
		}
	}
}

func TestIsWangXiang(t *testing.T) {
	if !IsWangXiang(Wang) || !IsWangXiang(Xiang) {
		t.Error("旺、相应判定为有力")
	}
	for _, s := range []WangShuaiState{Xiu, Qiu, Si} {
		if IsWangXiang(s) {
			t.Errorf("%s 不应判定为旺相", s)
		}
	}
}

// ============ 暗动 / 日破 / 月破 ============

func TestAnDongWangXiang(t *testing.T) {
	// 旺相逢冲 → 暗动
	// 构造：爻支辰(土)，日辰戌冲之；月建丑(土)使辰土得「旺」，
	// 且月建丑不冲辰（避免落入月破优先判定）。
	ws := GetWangShuai(Tu, "丑")
	if ws != Wang {
		t.Fatalf("前置条件有误：辰土在丑月应为旺，实际 %s", ws)
	}
	sp, ev := JudgeSpecialStatus("辰", "丑", "戌", false, ws)
	if sp != StatusAnDong {
		t.Fatalf("旺相逢冲应为暗动，实际 %s (%v)", sp, ev)
	}
	if len(ev) == 0 {
		t.Error("暗动应给出判定依据")
	}
	// 月建巳(火)生土 → 相，亦应暗动
	sp2, _ := JudgeSpecialStatus("辰", "巳", "戌", false, GetWangShuai(Tu, "巳"))
	if sp2 != StatusAnDong {
		t.Errorf("相而逢冲亦应暗动，实际 %s", sp2)
	}
}

func TestAnDongVsRiPo(t *testing.T) {
	// 同一爻支、同一日冲，仅因月令旺衰不同而区分暗动与日破
	// 爻支辰(土)，日辰戌冲之
	anDong, _ := JudgeSpecialStatus("辰", "丑", "戌", false, GetWangShuai(Tu, "丑")) // 丑月土旺
	if anDong != StatusAnDong {
		t.Errorf("旺相逢冲应为暗动，实际 %s", anDong)
	}
	riPo, _ := JudgeSpecialStatus("辰", "寅", "戌", false, GetWangShuai(Tu, "寅")) // 寅月木旺克土，土死
	if riPo != StatusRiPo {
		t.Errorf("休囚逢冲应为日破，实际 %s", riPo)
	}
}

func TestYuePoPriorityOverRiChong(t *testing.T) {
	// 月破优先：爻支辰，月建戌冲之，日辰亦戌冲之 → 判月破而非暗动
	sp, ev := JudgeSpecialStatus("辰", "戌", "戌", false, GetWangShuai(Tu, "戌"))
	if sp != StatusYuePo {
		t.Fatalf("月建冲爻应优先判月破，实际 %s (%v)", sp, ev)
	}
	// 纯月破：爻支子，月建午冲，日辰寅不冲子
	sp2, ev2 := JudgeSpecialStatus("子", "午", "寅", false, GetWangShuai(Shui, "午"))
	if sp2 != StatusYuePo {
		t.Fatalf("月建午冲子应为月破，实际 %s (%v)", sp2, ev2)
	}
	// 静而未受冲 → 无特殊状态
	sp3, _ := JudgeSpecialStatus("寅", "子", "丑", false, GetWangShuai(Mu, "子"))
	if sp3 != StatusNone {
		t.Errorf("未受月日冲应无特殊状态，实际 %s", sp3)
	}
}

func TestChangingYaoNotBroken(t *testing.T) {
	// 明动之爻不参与日破/暗动判定
	sp, _ := JudgeSpecialStatus("辰", "寅", "戌", true, GetWangShuai(Tu, "寅"))
	if sp != StatusNone {
		t.Errorf("明动之爻不应判为日破/暗动，实际 %s", sp)
	}
}

func TestMoveStateAndActive(t *testing.T) {
	if got := ResolveMoveState(true, StatusNone); got != MoveMing {
		t.Errorf("动爻应为明动，实际 %s", got)
	}
	if got := ResolveMoveState(false, StatusAnDong); got != MoveAnDong {
		t.Errorf("暗动状态应映射为 hidden_moving，实际 %s", got)
	}
	if got := ResolveMoveState(false, StatusRiPo); got != MoveBroken {
		t.Errorf("日破状态应映射为 day_break，实际 %s", got)
	}
	if got := ResolveMoveState(false, StatusNone); got != MoveStatic {
		t.Errorf("静爻应为 static，实际 %s", got)
	}
	// 暗动具备发动作用力
	if !IsActive(false, StatusAnDong) {
		t.Error("暗动应视为发动")
	}
	if IsActive(false, StatusNone) {
		t.Error("静爻不应视为发动")
	}
}

// ============ 六冲 / 六合 ============

func TestLiuChongEightPureGua(t *testing.T) {
	// 八纯卦（内外卦相同）皆为六冲
	for _, bin := range BaGuaBinary {
		mark := bin + bin
		if !IsLiuChong(mark) {
			t.Errorf("八纯卦 %s(%s) 应为六冲", Gua64[mark], mark)
		}
	}
}

func TestLiuChongSpecialCases(t *testing.T) {
	// 天雷无妄（内震100、外乾111）与雷天大壮（内乾111、外震100）亦为六冲
	if !IsLiuChong("100111") {
		t.Error("天雷无妄应为六冲")
	}
	if !IsLiuChong("111100") {
		t.Error("雷天大壮应为六冲")
	}
}

func TestLiuChongNotEveryGua(t *testing.T) {
	// 六冲卦共 10 个：八纯卦 + 无妄 + 大壮
	count := 0
	for mark := range Gua64 {
		if IsLiuChong(mark) {
			count++
		}
	}
	if count != 10 {
		t.Errorf("六冲卦应为 10 个，实际 %d 个", count)
	}
}

func TestLiuHe(t *testing.T) {
	// 地天泰为六合
	if !IsLiuHe("111000") {
		t.Errorf("地天泰应为六合，实际 false")
	}
	// 六合卦共 8 个
	count := 0
	for mark := range Gua64 {
		if IsLiuHe(mark) {
			count++
		}
	}
	if count != 8 {
		t.Errorf("六合卦应为 8 个，实际 %d 个", count)
	}
}

// ============ 六亲 / 六神 / 旬空 ============

func TestLiuQinAllRelations(t *testing.T) {
	// 以乾金为我（卦宫五行 = 金）
	// 金见金=兄弟、金生水=子孙、生金者土=父母、克金者火=官鬼、金克者木=妻财
	cases := []struct {
		yao  WuXing
		want string
	}{
		{Jin, "兄弟"},
		{Shui, "子孙"},
		{Tu, "父母"},
		{Huo, "官鬼"},
		{Mu, "妻财"},
	}
	for _, c := range cases {
		if got := GetLiuQin(Jin, c.yao); got != c.want {
			t.Errorf("乾金见 %s 应为 %s，实际 %s", c.yao, c.want, got)
		}
	}
}

func TestLiuShenStart(t *testing.T) {
	// 甲乙起青龙、丙丁起朱雀、戊起勾陈、己起螣蛇、庚辛起白虎、壬癸起玄武
	cases := map[string]string{
		"甲": "青龙", "乙": "青龙", "丙": "朱雀", "丁": "朱雀",
		"戊": "勾陈", "己": "螣蛇", "庚": "白虎", "辛": "白虎",
		"壬": "玄武", "癸": "玄武",
	}
	for gan, want := range cases {
		got := GetLiuShen(gan)[0]
		if got != want {
			t.Errorf("日干 %s 初爻应起 %s，实际 %s", gan, want, got)
		}
	}
}

func TestXunKong(t *testing.T) {
	cases := map[string]string{
		"甲子": "戌亥", "丙申": "辰巳", "甲午": "辰巳",
		"甲戌": "申酉", "甲申": "午未", "甲辰": "寅卯",
		"甲寅": "子丑", "癸亥": "子丑",
	}
	for gz, want := range cases {
		if got := XunKongOf(gz); got != want {
			t.Errorf("旬空(%s) = %s, want %s", gz, got, want)
		}
	}
}

func TestYaoValues(t *testing.T) {
	cases := []struct {
		v      YaoValue
		isMove bool
		yy     YinYang
	}{
		{6, true, Yin}, {7, false, Yang}, {8, false, Yin}, {9, true, Yang},
	}
	for _, c := range cases {
		if c.v.IsChanging() != c.isMove {
			t.Errorf("%d.IsChanging() = %v, want %v", c.v, c.v.IsChanging(), c.isMove)
		}
		if c.v.YinYang() != c.yy {
			t.Errorf("%d.YinYang() = %s, want %s", c.v, c.v.YinYang(), c.yy)
		}
	}
}
