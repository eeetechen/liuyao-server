// Package calendar 封装农历/干支历计算，对外只暴露排盘所需的最小信息。
package calendar

import (
	"fmt"
	"time"

	"github.com/6tail/lunar-go/calendar"

	"github.com/apple/liuyao/internal/divination"
)

// Info 干支历信息
type Info struct {
	Year       string // 年柱
	Month      string // 月柱（节气月）
	Day        string // 日柱
	Hour       string // 时柱
	YearGan    string // 年干
	YearZhi    string // 年支
	DayGan     string // 日干
	DayZhi     string // 日支（日辰）
	MonthZhi   string // 月支（月建）
	YearLabel  string // 年份标注，如「丙午年（马年）」
	MonthLabel string // 月柱标注，如「月柱 甲午（月建午）」
	DayLabel   string // 日柱标注，如「日柱 庚申（日辰申）」
	HourLabel  string // 时柱标注
	XunKong    string // 日柱旬空
	SolarText  string // 公历文本
	LunarText  string // 农历文本
}

// FromSolar 由公历时间推算四柱与旬空。
// DayBoundary 日干支的换日口径。
//
// 子时跨日历来有两说，六爻排盘中直接影响「日辰」，进而影响暗动与日破的判定：
//
//   - BoundaryZi（晚子时算次日，23:00 换日）：较常见于传统命理排盘。
//   - BoundaryMidnight（00:00 换日，晚子时仍属当日）：即所谓「子正换日」。
//
// 两者仅在 23:00–23:59 起卦时结果不同。为免歧义，本服务将其显式化，
// 由请求参数决定，默认采用 BoundaryZi。
type DayBoundary string

const (
	BoundaryZi       DayBoundary = "zi"       // 23:00 换日（晚子时属次日）
	BoundaryMidnight DayBoundary = "midnight" // 00:00 换日（晚子时属当日）
)

// ParseDayBoundary 解析换日口径，空值取默认 BoundaryZi
func ParseDayBoundary(s string) DayBoundary {
	if DayBoundary(s) == BoundaryMidnight {
		return BoundaryMidnight
	}
	return BoundaryZi
}

// FromSolar 以默认换日口径（23:00 换日）推算四柱与旬空。
func FromSolar(t time.Time) Info {
	return FromSolarWithBoundary(t, BoundaryZi)
}

// FromSolarWithBoundary 由公历时间推算四柱与旬空，可指定换日口径。
//
// 采用「节气」口径：
//   - 年柱以立春为界（GetYearInGanZhiExact）
//   - 月柱以节气为界（GetMonthInGanZhiExact）
//   - 日柱按换日口径取 GetDayInGanZhiExact（23:00 换日）或 Exact2（00:00 换日）
//   - 时柱为时辰（GetTimeInGanZhi）
func FromSolarWithBoundary(t time.Time, b DayBoundary) Info {
	solar := calendar.NewSolar(t.Year(), int(t.Month()), t.Day(), t.Hour(), t.Minute(), t.Second())
	lunar := solar.GetLunar()

	var day, xunKong string
	if b == BoundaryMidnight {
		day = lunar.GetDayInGanZhiExact2()
		xunKong = lunar.GetDayXunKongExact2()
	} else {
		day = lunar.GetDayInGanZhiExact()
		xunKong = lunar.GetDayXunKongExact()
	}

	year := lunar.GetYearInGanZhiExact()
	month := lunar.GetMonthInGanZhiExact()
	hour := lunar.GetTimeInGanZhi()

	yGan, yZhi := firstRune(year), secondRune(year)
	dGan, dZhi := firstRune(day), secondRune(day)
	mZhi := secondRune(month)

	// 生肖以立春为界（Exact 口径），与年柱保持一致
	sx := lunar.GetYearShengXiaoExact()

	return Info{
		Year:       year,
		Month:      month,
		Day:        day,
		Hour:       hour,
		YearGan:    yGan,
		YearZhi:    yZhi,
		DayGan:     dGan,
		DayZhi:     dZhi,
		MonthZhi:   mZhi,
		YearLabel:  fmt.Sprintf("%s年（%s年）", year, sx),
		MonthLabel: fmt.Sprintf("月柱 %s（月建%s）", month, mZhi),
		DayLabel:   fmt.Sprintf("日柱 %s（日辰%s）", day, dZhi),
		HourLabel:  fmt.Sprintf("时柱 %s", hour),
		XunKong:    xunKong,
		SolarText:  fmt.Sprintf("%d年%d月%d日 %02d:%02d", t.Year(), int(t.Month()), t.Day(), t.Hour(), t.Minute()),
		LunarText:  lunar.String(),
	}
}

// ToGanZhiTime 转换为排盘引擎所需的格式
func (i Info) ToGanZhiTime() divination.GanZhiTime {
	return divination.GanZhiTime{
		Year:       i.Year,
		Month:      i.Month,
		Day:        i.Day,
		Hour:       i.Hour,
		YearGan:    i.YearGan,
		YearZhi:    i.YearZhi,
		DayGan:     i.DayGan,
		DayZhi:     i.DayZhi,
		MonthZhi:   i.MonthZhi,
		YearLabel:  i.YearLabel,
		MonthLabel: i.MonthLabel,
		DayLabel:   i.DayLabel,
		HourLabel:  i.HourLabel,
		XunKong:    i.XunKong,
		SolarText:  i.SolarText,
		LunarText:  i.LunarText,
	}
}

func firstRune(s string) string {
	r := []rune(s)
	if len(r) == 0 {
		return ""
	}
	return string(r[0])
}

func secondRune(s string) string {
	r := []rune(s)
	if len(r) < 2 {
		return ""
	}
	return string(r[1])
}
