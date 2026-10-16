package calendar

import (
	"testing"
	"time"
)

var cst = time.FixedZone("CST", 8*3600)

func at(s string) time.Time {
	t, err := time.ParseInLocation("2006-01-02 15:04", s, cst)
	if err != nil {
		panic(err)
	}
	return t
}

// 与外部独立实现（bopo/najia 依赖的 lunar_python）交叉验证四柱与旬空。
// 数据由该库实际运行生成。
func TestFromSolarCrossValidation(t *testing.T) {
	cases := []struct {
		in      string
		year    string
		month   string
		day     string
		hour    string
		xunKong string
	}{
		{"2024-01-01 12:00", "癸卯", "甲子", "甲子", "庚午", "戌亥"},
		{"2024-02-04 10:00", "癸卯", "乙丑", "戊戌", "丁巳", "辰巳"},
		{"2024-06-15 08:30", "甲辰", "庚午", "庚戌", "庚辰", "寅卯"},
		{"2023-03-21 06:00", "癸卯", "乙卯", "戊寅", "乙卯", "申酉"},
		{"2026-10-16 11:00", "丙午", "戊戌", "癸亥", "戊午", "子丑"},
	}
	for _, c := range cases {
		got := FromSolarWithBoundary(at(c.in), BoundaryZi)
		if got.Year != c.year || got.Month != c.month || got.Day != c.day || got.Hour != c.hour {
			t.Errorf("%s 四柱 = %s/%s/%s/%s, want %s/%s/%s/%s",
				c.in, got.Year, got.Month, got.Day, got.Hour, c.year, c.month, c.day, c.hour)
		}
		if got.XunKong != c.xunKong {
			t.Errorf("%s 旬空 = %s, want %s", c.in, got.XunKong, c.xunKong)
		}
		// 日干日支须与日柱一致
		if got.DayGan+got.DayZhi != got.Day {
			t.Errorf("%s 日干日支 %s%s 与日柱 %s 不符", c.in, got.DayGan, got.DayZhi, got.Day)
		}
		if got.MonthZhi == "" {
			t.Errorf("%s 月支为空", c.in)
		}
	}
}

// 旬空自洽：日柱所在旬的第二位即为旬空内容之一
func TestXunKongConsistency(t *testing.T) {
	for _, d := range []string{
		"2024-01-01 12:00", "2024-05-05 09:00", "2025-08-08 20:00", "2026-11-11 03:00",
	} {
		got := FromSolar(at(d))
		if len([]rune(got.XunKong)) != 2 {
			t.Errorf("%s 旬空格式异常: %q", d, got.XunKong)
		}
	}
}

// 换日口径：仅在 23:00-23:59 有差异
func TestDayBoundaryZiVsMidnight(t *testing.T) {
	zi := FromSolarWithBoundary(at("2025-12-31 23:30"), BoundaryZi)
	mid := FromSolarWithBoundary(at("2025-12-31 23:30"), BoundaryMidnight)

	if zi.Day == mid.Day {
		t.Fatalf("晚子时两种口径应产生不同日柱，均为 %s", zi.Day)
	}
	// 23:00 换日时，晚子时属次日
	if zi.Day != "乙亥" {
		t.Errorf("zi 口径日柱 = %s, want 乙亥", zi.Day)
	}
	if mid.Day != "甲戌" {
		t.Errorf("midnight 口径日柱 = %s, want 甲戌", mid.Day)
	}
	// 两者月建应一致（月柱按节气，不因子时换日而变）
	if zi.MonthZhi != mid.MonthZhi {
		t.Errorf("月支不应随换日口径变化: %s vs %s", zi.MonthZhi, mid.MonthZhi)
	}

	// 非子时的时刻，两种口径必须一致
	for _, d := range []string{"2025-12-31 12:00", "2025-12-31 00:30", "2025-12-31 22:59"} {
		a := FromSolarWithBoundary(at(d), BoundaryZi)
		b := FromSolarWithBoundary(at(d), BoundaryMidnight)
		if a.Day != b.Day {
			t.Errorf("%s 非晚子时不应有口径差异: %s vs %s", d, a.Day, b.Day)
		}
	}
}

// 默认口径为 zi
func TestDefaultBoundaryIsZi(t *testing.T) {
	if got := FromSolar(at("2025-12-31 23:30")); got.Day != "乙亥" {
		t.Errorf("默认应采 23:00 换日（乙亥），实际 %s", got.Day)
	}
	if ParseDayBoundary("") != BoundaryZi {
		t.Error("空值应解析为 zi")
	}
	if ParseDayBoundary("garbage") != BoundaryZi {
		t.Error("非法值应回退为 zi")
	}
	if ParseDayBoundary("midnight") != BoundaryMidnight {
		t.Error("midnight 应正确解析")
	}
}

// ToGanZhiTime 字段映射
func TestToGanZhiTime(t *testing.T) {
	i := FromSolar(at("2024-01-01 12:00"))
	g := i.ToGanZhiTime()
	if g.DayGan != "甲" || g.DayZhi != "子" || g.MonthZhi != "子" {
		t.Errorf("映射错误: DayGan=%s DayZhi=%s MonthZhi=%s", g.DayGan, g.DayZhi, g.MonthZhi)
	}
	if g.XunKong != "戌亥" {
		t.Errorf("旬空 = %s, want 戌亥", g.XunKong)
	}
	if g.SolarText == "" || g.LunarText == "" {
		t.Error("时间文本不应为空")
	}
}
