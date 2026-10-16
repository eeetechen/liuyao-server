// Package api 提供 HTTP JSON 接口，与前端完全解耦。
package api

import (
	"encoding/json"
	"errors"
	"io/fs"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/apple/liuyao"
	"github.com/apple/liuyao/internal/calendar"
	"github.com/apple/liuyao/internal/cast"
	"github.com/apple/liuyao/internal/divination"
)

// Server HTTP 服务
type Server struct {
	mux *http.ServeMux
}

// New 构造服务并注册路由
func New() *Server {
	s := &Server{mux: http.NewServeMux()}
	s.routes()
	return s
}

func (s *Server) routes() {
	s.mux.HandleFunc("/api/health", s.handleHealth)
	s.mux.HandleFunc("/api/meta", s.handleMeta)
	s.mux.HandleFunc("/api/cast", s.handleCast)
	s.mux.HandleFunc("/api/chart", s.handleChart)
	s.mux.HandleFunc("/api/reference", s.handleReference)
	// 前端静态资源（编译进二进制，可脱离源码目录运行）
	sub, err := fs.Sub(liuyao.WebFS, "web")
	if err != nil {
		log.Fatalf("加载内嵌前端资源失败: %v", err)
	}
	s.mux.Handle("/", http.FileServer(http.FS(sub)))
}

// ServeHTTP 实现 http.Handler
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	s.mux.ServeHTTP(w, r)
}

// ---------- 通用响应 ----------

type apiError struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("写响应失败: %v", err)
	}
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, apiError{Error: msg})
}

// ---------- /api/health ----------

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	})
}

// ---------- /api/meta ----------

// MetaResponse 前端初始化所需的静态字典
type MetaResponse struct {
	YaoValues []YaoValueMeta `json:"yaoValues"`
	Methods   []string       `json:"methods"`
}

// YaoValueMeta 爻值说明（供前端渲染摇卦选项）
type YaoValueMeta struct {
	Value   int    `json:"value"`
	Label   string `json:"label"`
	YinYang string `json:"yinYang"`
	Moving  bool   `json:"moving"`
	Backs   int    `json:"backs"`
	Desc    string `json:"desc"`
}

func (s *Server) handleMeta(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, MetaResponse{
		YaoValues: []YaoValueMeta{
			{6, "老阴", "阴", true, 0, "三枚铜钱皆正面（0 背），阴极而变阳，为动爻"},
			{7, "少阳", "阳", false, 1, "一背二字，阳而不变，为静爻"},
			{8, "少阴", "阴", false, 2, "二背一字，阴而不变，为静爻"},
			{9, "老阳", "阳", true, 3, "三枚铜钱皆背面（3 背），阳极而变阴，为动爻"},
		},
		Methods: []string{"coin", "manual"},
	})
}

// ---------- /api/cast ----------

// CastResponse 摇卦结果
type CastResponse struct {
	Results []cast.CoinResult `json:"results"`
	Values  []int             `json:"values"`
	Time    calendar.Info     `json:"time"`
}

// handleCast 摇卦（三枚铜钱六次）。可选 query: at=RFC3339 指定起卦时间。
func (s *Server) handleCast(w http.ResponseWriter, r *http.Request) {
	at, err := parseTime(r)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	results := cast.TossSix()
	vals := make([]int, 0, 6)
	for _, res := range results {
		vals = append(vals, int(res.Value))
	}
	writeJSON(w, http.StatusOK, CastResponse{
		Results: results,
		Values:  vals,
		Time:    calendar.FromSolar(at),
	})
}

// ---------- /api/reference ----------

// handleReference 返回五行生克参考表（静态字典，供使用者查询印证）。
func (s *Server) handleReference(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, divination.BuildReference())
}

// ---------- /api/chart ----------

// ChartRequest 排盘请求
type ChartRequest struct {
	// YaoValues 六爻值，自下而上。可传数字数组或字符串 "7,7,7,7,7,7"
	YaoValues []int  `json:"yaoValues"`
	Question  string `json:"question"`
	Method    string `json:"method"`
	// At 起卦时间，RFC3339 或 "2006-01-02 15:04"。留空取当前时间。
	At string `json:"at"`
	// Timezone 时区名，如 "Asia/Shanghai"，留空取 UTC+8
	Timezone string `json:"timezone"`
	// DayBoundary 换日口径："zi"(23:00 换日，默认) 或 "midnight"(00:00 换日)。
	// 仅在 23:00-23:59 起卦时影响日辰，进而影响暗动/日破判定。
	DayBoundary string `json:"dayBoundary"`
	// YongShen 指定用神六亲：父母/兄弟/子孙/妻财/官鬼。
	// 留空则由 question 自动推断。
	YongShen string `json:"yongShen"`
}

func (s *Server) handleChart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "请使用 POST")
		return
	}
	var req ChartRequest
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	if err := dec.Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "请求体解析失败："+err.Error())
		return
	}

	loc, err := parseLocation(req.Timezone)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	at, err := parseTimeString(req.At, loc)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}

	// 爻值转换
	vals := make([]divination.YaoValue, 0, 6)
	for _, v := range req.YaoValues {
		vals = append(vals, divination.YaoValue(v))
	}

	method := divination.CastMethod(req.Method)
	if method != divination.MethodCoin && method != divination.MethodManual {
		method = divination.MethodManual
	}

	// 所问之事为可选。缺省时不再注入占位文案（否则会在白话里出现
	// 「问的是『未填写…』」这类无意义措辞），改由用神承载主题。
	q := strings.TrimSpace(req.Question)

	info := calendar.FromSolarWithBoundary(at, calendar.ParseDayBoundary(req.DayBoundary))
	ys := strings.TrimSpace(req.YongShen)
	if ys != "" && !divination.IsValidLiuQin(ys) {
		writeErr(w, http.StatusBadRequest, "用神无效，应为 父母/兄弟/子孙/妻财/官鬼 之一")
		return
	}
	// 用神留空时：若提供了所问之事则依问题推断，否则默认取妻财
	if ys == "" && q == "" {
		ys = "妻财"
	}

	chart, err := divination.BuildChart(divination.ChartInput{
		YaoValues: vals,
		Question:  q,
		Method:    method,
		YongShen:  ys,
	}, info.ToGanZhiTime())
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, chart)
}

// ---------- 辅助 ----------

var defaultLocation = time.FixedZone("CST", 8*3600)

func parseLocation(name string) (*time.Location, error) {
	if name == "" {
		return defaultLocation, nil
	}
	loc, err := time.LoadLocation(name)
	if err != nil {
		return nil, errors.New("时区无效：" + name)
	}
	return loc, nil
}

func parseTime(r *http.Request) (time.Time, error) {
	return parseTimeString(r.URL.Query().Get("at"), defaultLocation)
}

func parseTimeString(s string, loc *time.Location) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Now().In(loc), nil
	}
	for _, layout := range []string{time.RFC3339, "2006-01-02 15:04:05", "2006-01-02 15:04", "2006-01-02"} {
		if t, err := time.ParseInLocation(layout, s, loc); err == nil {
			return t, nil
		}
	}
	return time.Time{}, errors.New("时间格式无效，应为 RFC3339 或 2006-01-02 15:04")
}
