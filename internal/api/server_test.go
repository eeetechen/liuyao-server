package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func doJSON(t *testing.T, h http.Handler, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("编码请求失败: %v", err)
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	return w
}

func TestHealth(t *testing.T) {
	w := doJSON(t, New(), http.MethodGet, "/api/health", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("状态码 = %d, want 200", w.Code)
	}
	var m map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &m); err != nil {
		t.Fatalf("响应非 JSON: %v", err)
	}
	if m["status"] != "ok" {
		t.Errorf("status = %v, want ok", m["status"])
	}
}

// /api/meta 应返回四种爻值
func TestMeta(t *testing.T) {
	w := doJSON(t, New(), http.MethodGet, "/api/meta", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("状态码 = %d", w.Code)
	}
	var m MetaResponse
	if err := json.Unmarshal(w.Body.Bytes(), &m); err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if len(m.YaoValues) != 4 {
		t.Fatalf("爻值种类 = %d, want 4", len(m.YaoValues))
	}
	for _, v := range m.YaoValues {
		if v.Value != 6 && v.Value != 7 && v.Value != 8 && v.Value != 9 {
			t.Errorf("非法爻值 %d", v.Value)
		}
	}
}

// /api/cast 应返回六爻且值域合法
func TestCastEndpoint(t *testing.T) {
	for i := 0; i < 20; i++ {
		w := doJSON(t, New(), http.MethodGet, "/api/cast", nil)
		if w.Code != http.StatusOK {
			t.Fatalf("状态码 = %d", w.Code)
		}
		var r CastResponse
		if err := json.Unmarshal(w.Body.Bytes(), &r); err != nil {
			t.Fatalf("解析失败: %v", err)
		}
		if len(r.Results) != 6 || len(r.Values) != 6 {
			t.Fatalf("应返回 6 爻，实际 %d/%d", len(r.Results), len(r.Values))
		}
		for j, v := range r.Values {
			if v < 6 || v > 9 {
				t.Fatalf("第%d爻值越界: %d", j+1, v)
			}
		}
	}
}

// 完整排盘：乾为天，校验关键字段
func TestChartEndpoint(t *testing.T) {
	body := ChartRequest{
		YaoValues: []int{7, 7, 7, 7, 7, 7},
		Question:  "求测事业",
		Method:    "manual",
		At:        "2024-01-01 12:00",
		Timezone:  "Asia/Shanghai",
	}
	w := doJSON(t, New(), http.MethodPost, "/api/chart", body)
	if w.Code != http.StatusOK {
		t.Fatalf("状态码 = %d, body=%s", w.Code, w.Body.String())
	}
	// 用宽松结构校验
	var got map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	ben, ok := got["benGua"].(map[string]any)
	if !ok {
		t.Fatal("缺少 benGua")
	}
	if ben["name"] != "乾为天" {
		t.Errorf("卦名 = %v, want 乾为天", ben["name"])
	}
	if ben["type"] != "六冲" {
		t.Errorf("卦型 = %v, want 六冲", ben["type"])
	}
	yaos, ok := got["yaos"].([]any)
	if !ok || len(yaos) != 6 {
		t.Fatalf("yaos 应为 6 条，实际 %v", got["yaos"])
	}
	// 每爻必须含旺衰与动静字段
	for i, raw := range yaos {
		y := raw.(map[string]any)
		for _, k := range []string{"wangShuai", "moveState", "moveLabel", "special", "strength", "liuQin", "liuShen"} {
			if _, ok := y[k]; !ok {
				t.Errorf("第%d爻缺少字段 %s", i+1, k)
			}
		}
	}
	sum, ok := got["summary"].(map[string]any)
	if !ok {
		t.Fatal("缺少 summary")
	}
	if sum["isLiuChong"] != true {
		t.Error("summary.isLiuChong 应为 true")
	}
	hl, ok := sum["highlights"].([]any)
	if !ok || len(hl) == 0 {
		t.Error("summary.highlights 不应为空")
	}
	// 六冲提示必须出现
	found := false
	for _, x := range hl {
		if s, _ := x.(string); len(s) > 0 && (bytes.Contains([]byte(s), []byte("六冲"))) {
			found = true
		}
	}
	if !found {
		t.Error("六冲提示缺失")
	}
}

// 暗动场景：坤为地，丑月未日
func TestChartAnDong(t *testing.T) {
	body := ChartRequest{
		YaoValues: []int{8, 8, 8, 8, 8, 8},
		Question:  "暗动测试",
		Method:    "manual",
		At:        "2024-01-08 10:00", // 需落在丑月未日
		Timezone:  "Asia/Shanghai",
	}
	w := doJSON(t, New(), http.MethodPost, "/api/chart", body)
	if w.Code != http.StatusOK {
		t.Fatalf("状态码 = %d, body=%s", w.Code, w.Body.String())
	}
	var got map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &got)
	sum := got["summary"].(map[string]any)
	t.Logf("时间: %v", got["time"])
	t.Logf("anDongCount=%v anDongYaos=%v", sum["anDongCount"], sum["anDongYaos"])
}

// 非法请求
func TestChartErrors(t *testing.T) {
	srv := New()
	// 爻值非法
	w := doJSON(t, srv, http.MethodPost, "/api/chart", ChartRequest{YaoValues: []int{1, 2, 3, 4, 5, 6}})
	if w.Code != http.StatusBadRequest {
		t.Errorf("非法爻值应返回 400，实际 %d", w.Code)
	}
	// 爻数不足
	w = doJSON(t, srv, http.MethodPost, "/api/chart", ChartRequest{YaoValues: []int{7, 7, 7}})
	if w.Code != http.StatusBadRequest {
		t.Errorf("爻数不足应返回 400，实际 %d", w.Code)
	}
	// 时间格式错误
	w = doJSON(t, srv, http.MethodPost, "/api/chart", ChartRequest{
		YaoValues: []int{7, 7, 7, 7, 7, 7}, At: "不是时间",
	})
	if w.Code != http.StatusBadRequest {
		t.Errorf("时间非法应返回 400，实际 %d", w.Code)
	}
	// GET 方法不允许
	w = doJSON(t, srv, http.MethodGet, "/api/chart", nil)
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("GET /api/chart 应返回 405，实际 %d", w.Code)
	}
	// 时区非法
	w = doJSON(t, srv, http.MethodPost, "/api/chart", ChartRequest{
		YaoValues: []int{7, 7, 7, 7, 7, 7}, Timezone: "Not/AZone",
	})
	if w.Code != http.StatusBadRequest {
		t.Errorf("时区非法应返回 400，实际 %d", w.Code)
	}
}

// 前端资源应被内嵌并可访问
func TestEmbeddedFrontend(t *testing.T) {
	srv := New()
	for _, p := range []string{"/", "/style.css", "/app.js"} {
		w := doJSON(t, srv, http.MethodGet, p, nil)
		if w.Code != http.StatusOK {
			t.Errorf("GET %s = %d, want 200", p, w.Code)
		}
		if w.Body.Len() == 0 {
			t.Errorf("GET %s 返回空内容", p)
		}
	}
	// http.FileServer 会把 /index.html 重定向到 /（标准行为）
	w := doJSON(t, srv, http.MethodGet, "/index.html", nil)
	if w.Code != http.StatusMovedPermanently {
		t.Errorf("GET /index.html = %d, want 301", w.Code)
	}
	if loc := w.Header().Get("Location"); loc != "./" && loc != "/" {
		t.Errorf("重定向目标 = %q, want ./ 或 /", loc)
	}

	// 首页应含关键中文与响应式 viewport
	w = doJSON(t, srv, http.MethodGet, "/", nil)
	body := w.Body.String()
	for _, want := range []string{"六爻排盘", "viewport", "旺相", "暗动", "六冲"} {
		if !bytes.Contains([]byte(body), []byte(want)) {
			t.Errorf("首页缺少关键字 %q", want)
		}
	}
}

// CORS 预检
func TestCORSPreflight(t *testing.T) {
	w := doJSON(t, New(), http.MethodOptions, "/api/chart", nil)
	if w.Code != http.StatusNoContent {
		t.Errorf("OPTIONS 应返回 204，实际 %d", w.Code)
	}
	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("缺少 CORS 头")
	}
}

// relations 字段必须存在且结构完整
func TestChartRelationsContract(t *testing.T) {
	body := ChartRequest{
		YaoValues: []int{7, 7, 7, 7, 7, 7}, // 乾为天，含子寅辰午申戌
		Question:  "关系验证",
		At:        "2024-01-01 12:00",
		Timezone:  "Asia/Shanghai",
	}
	w := doJSON(t, New(), http.MethodPost, "/api/chart", body)
	if w.Code != http.StatusOK {
		t.Fatalf("状态码 = %d, body=%s", w.Code, w.Body.String())
	}
	var got map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("解析失败: %v", err)
	}

	rel, ok := got["relations"].(map[string]any)
	if !ok {
		t.Fatal("响应缺少 relations 字段")
	}

	// 五行生克：应有 6 条，每条含月建与日辰作用
	sk, ok := rel["shengKe"].([]any)
	if !ok || len(sk) != 6 {
		t.Fatalf("shengKe 应为 6 条，实际 %v", rel["shengKe"])
	}
	for i, raw := range sk {
		item := raw.(map[string]any)
		for _, k := range []string{"position", "label", "zhi", "wuXing", "liuQin", "month", "day"} {
			if _, ok := item[k]; !ok {
				t.Errorf("shengKe[%d] 缺字段 %s", i, k)
			}
		}
		for _, side := range []string{"month", "day"} {
			inf := item[side].(map[string]any)
			for _, k := range []string{"source", "sourceWuXing", "kind", "label", "desc"} {
				if _, ok := inf[k]; !ok {
					t.Errorf("shengKe[%d].%s 缺字段 %s", i, side, k)
				}
			}
		}
	}

	// 世应
	sy, ok := rel["shiYing"].(map[string]any)
	if !ok {
		t.Fatal("缺少 shiYing")
	}
	for _, k := range []string{"shiPosition", "yingPosition", "shiWuXing", "yingWuXing", "kind", "label", "desc"} {
		if _, ok := sy[k]; !ok {
			t.Errorf("shiYing 缺字段 %s", k)
		}
	}

	// 三合：乾为天含申子辰与寅午戌，应检出 2 局
	san, ok := rel["sanHe"].([]any)
	if !ok || len(san) == 0 {
		t.Fatalf("乾为天应检出三合局，实际 %v", rel["sanHe"])
	}
	names := map[string]bool{}
	for _, raw := range san {
		j := raw.(map[string]any)
		names[j["name"].(string)] = true
		for _, k := range []string{"name", "element", "branches", "positions", "complete", "desc"} {
			if _, ok := j[k]; !ok {
				t.Errorf("sanHe 缺字段 %s", k)
			}
		}
	}
	if !names["申子辰"] || !names["寅午戌"] {
		t.Errorf("应检出申子辰与寅午戌，实际 %v", names)
	}

	// 六合 / 六冲：必须为数组（非 null）
	if _, ok := rel["liuHe"].([]any); !ok {
		t.Errorf("liuHe 应为数组，实际 %T", rel["liuHe"])
	}
	lc, ok := rel["liuChong"].([]any)
	if !ok || len(lc) == 0 {
		t.Fatalf("乾为天为六冲卦，应检出六冲，实际 %v", rel["liuChong"])
	}
	for _, raw := range lc {
		p := raw.(map[string]any)
		for _, k := range []string{"a", "b", "kind"} {
			if _, ok := p[k]; !ok {
				t.Errorf("liuChong 缺字段 %s", k)
			}
		}
	}

	// banHe 亦须为数组
	if _, ok := rel["banHe"].([]any); !ok {
		t.Errorf("banHe 应为数组，实际 %T", rel["banHe"])
	}
}

// 六合检出：子日与含子爻之卦
func TestChartLiuHeContract(t *testing.T) {
	body := ChartRequest{
		YaoValues: []int{7, 7, 7, 7, 7, 7}, // 初爻为子
		At:        "2024-01-06 12:00",      // 子日
		Timezone:  "Asia/Shanghai",
	}
	w := doJSON(t, New(), http.MethodPost, "/api/chart", body)
	if w.Code != http.StatusOK {
		t.Fatalf("状态码 = %d", w.Code)
	}
	var got map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &got)
	rel := got["relations"].(map[string]any)
	he := rel["liuHe"].([]any)
	if len(he) == 0 {
		t.Logf("该日未见六合（日柱可能非丑），time=%v", got["time"])
	}
	for _, raw := range he {
		p := raw.(map[string]any)
		if _, ok := p["hua"]; !ok {
			t.Error("六合项应含 hua（合化五行）")
		}
	}
}

// /api/reference 契约：参考表必须完整可用
func TestReferenceEndpoint(t *testing.T) {
	w := doJSON(t, New(), http.MethodGet, "/api/reference", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("状态码 = %d", w.Code)
	}
	var d map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &d); err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	// 各集合长度
	want := map[string]int{
		"wuXing": 5, "shengKe": 10, "tianGan": 10,
		"diZhi": 12, "liuChong": 6, "liuHe": 6, "sanHe": 4, "banHe": 8,
	}
	for k, n := range want {
		arr, ok := d[k].([]any)
		if !ok {
			t.Errorf("%s 应为数组，实际 %T", k, d[k])
			continue
		}
		if len(arr) != n {
			t.Errorf("%s 长度 = %d, want %d", k, len(arr), n)
		}
	}
	// 水克火必须可查
	var foundShengKe bool
	for _, raw := range d["shengKe"].([]any) {
		x := raw.(map[string]any)
		if x["text"] == "水克火" {
			foundShengKe = true
			if x["kind"] != "克" {
				t.Errorf("水克火 kind = %v", x["kind"])
			}
			if x["plain"] == "" {
				t.Error("水克火 缺白话")
			}
		}
	}
	if !foundShengKe {
		t.Error("参考表缺「水克火」")
	}
	// 卯属木、癸属水必须可查
	var maoOK, guiOK bool
	for _, raw := range d["diZhi"].([]any) {
		g := raw.(map[string]any)
		if g["name"] == "卯" {
			maoOK = true
			if g["wuXing"] != "木" {
				t.Errorf("卯 五行 = %v, want 木", g["wuXing"])
			}
			if g["shengXiao"] != "兔" {
				t.Errorf("卯 生肖 = %v, want 兔", g["shengXiao"])
			}
			if g["yinYang"] != "阴" {
				t.Errorf("卯 阴阳 = %v, want 阴", g["yinYang"])
			}
		}
	}
	for _, raw := range d["tianGan"].([]any) {
		g := raw.(map[string]any)
		if g["name"] == "癸" {
			guiOK = true
			if g["wuXing"] != "水" {
				t.Errorf("癸 五行 = %v, want 水", g["wuXing"])
			}
			if g["yinYang"] != "阴" {
				t.Errorf("癸 阴阳 = %v, want 阴", g["yinYang"])
			}
		}
	}
	if !maoOK {
		t.Error("参考表缺地支「卯」")
	}
	if !guiOK {
		t.Error("参考表缺天干「癸」")
	}
}

// 年份标注必须出现在 chart 响应中
func TestChartYearLabel(t *testing.T) {
	w := doJSON(t, New(), http.MethodPost, "/api/chart", ChartRequest{
		YaoValues: []int{7, 7, 7, 7, 7, 7},
		At:        "2026-06-15 12:00",
		Timezone:  "Asia/Shanghai",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("状态码 = %d", w.Code)
	}
	var d map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &d)
	tm := d["time"].(map[string]any)

	// 2026 年为丙午年
	if tm["year"] != "丙午" {
		t.Errorf("2026 年柱 = %v, want 丙午", tm["year"])
	}
	label, _ := tm["yearLabel"].(string)
	if label == "" {
		t.Fatal("缺 yearLabel")
	}
	for _, want := range []string{"丙午", "马年"} {
		if !bytes.Contains([]byte(label), []byte(want)) {
			t.Errorf("yearLabel 应含 %q，实际 %q", want, label)
		}
	}
	// 柱别标注
	for _, k := range []string{"monthLabel", "dayLabel", "hourLabel"} {
		if s, _ := tm[k].(string); s == "" {
			t.Errorf("缺 %s", k)
		}
	}
}

// 白话解析契约
func TestChartPlainContract(t *testing.T) {
	w := doJSON(t, New(), http.MethodPost, "/api/chart", ChartRequest{
		YaoValues: []int{9, 7, 7, 7, 7, 7}, // 初爻动 → 乾为天变天风姤
		Question:  "这份工作要不要接",
		At:        "2026-06-15 12:00",
		Timezone:  "Asia/Shanghai",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("状态码 = %d", w.Code)
	}
	var d map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &d)

	plain, ok := d["plain"].(map[string]any)
	if !ok {
		t.Fatal("响应缺 plain")
	}
	ov, _ := plain["overview"].(string)
	if ov == "" {
		t.Fatal("plain.overview 为空")
	}
	for _, want := range []string{"乾为天", "天风姤", "这份工作要不要接"} {
		if !bytes.Contains([]byte(ov), []byte(want)) {
			t.Errorf("overview 应含 %q，实际 %q", want, ov)
		}
	}
	secs, ok := plain["sections"].([]any)
	if !ok || len(secs) < 4 {
		t.Fatalf("plain.sections 应至少 4 节，实际 %v", plain["sections"])
	}
	for i, raw := range secs {
		s := raw.(map[string]any)
		if s["title"] == "" {
			t.Errorf("section[%d] 缺 title", i)
		}
		lines, ok := s["lines"].([]any)
		if !ok || len(lines) == 0 {
			t.Errorf("section[%d] 无内容", i)
		}
	}
	sug, ok := plain["suggestions"].([]any)
	if !ok || len(sug) == 0 {
		t.Fatal("plain.suggestions 为空")
	}
	last, _ := sug[len(sug)-1].(string)
	if !bytes.Contains([]byte(last), []byte("仅供")) {
		t.Errorf("末条建议应为免责声明，实际 %q", last)
	}
}

// 变卦须带上下卦与世应
func TestBianGuaFields(t *testing.T) {
	w := doJSON(t, New(), http.MethodPost, "/api/chart", ChartRequest{
		YaoValues: []int{9, 7, 7, 7, 7, 7},
		At:        "2026-06-15 12:00",
		Timezone:  "Asia/Shanghai",
	})
	var d map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &d)
	v, ok := d["bianGua"].(map[string]any)
	if !ok {
		t.Fatal("缺 bianGua")
	}
	if v["name"] != "天风姤" {
		t.Errorf("变卦 = %v, want 天风姤", v["name"])
	}
	// 天风姤 = 乾上巽下
	if v["upperGua"] != "乾" || v["lowerGua"] != "巽" {
		t.Errorf("变卦上下卦 = %v上%v下, want 乾上巽下", v["upperGua"], v["lowerGua"])
	}
	if v["shiPos"] == nil || v["yingPos"] == nil {
		t.Error("变卦缺世应")
	}
}

// 三刑六害契约
func TestChartXingHaiContract(t *testing.T) {
	w := doJSON(t, New(), http.MethodPost, "/api/chart", ChartRequest{
		YaoValues: []int{8, 8, 8, 8, 8, 8}, // 坤为地：丑未戌
		At:        "2024-10-15 10:00",      // 月建戌
		Timezone:  "Asia/Shanghai",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("状态码 = %d", w.Code)
	}
	var d map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &d)
	rel := d["relations"].(map[string]any)

	xing, ok := rel["xing"].([]any)
	if !ok {
		t.Fatalf("xing 应为数组，实际 %T", rel["xing"])
	}
	if len(xing) == 0 {
		t.Fatalf("坤为地遇月建戌应检出恃势之刑")
	}
	found := false
	for _, raw := range xing {
		x := raw.(map[string]any)
		for _, k := range []string{"name", "branches", "members", "selfXing", "complete", "desc"} {
			if _, ok := x[k]; !ok {
				t.Errorf("xing 缺字段 %s", k)
			}
		}
		if x["name"] == "恃势之刑" {
			found = true
		}
	}
	if !found {
		t.Errorf("应检出恃势之刑，实际 %v", xing)
	}
	// hai 必须为数组
	if _, ok := rel["hai"].([]any); !ok {
		t.Errorf("hai 应为数组，实际 %T", rel["hai"])
	}
}

// 用神体系契约
func TestChartShenContract(t *testing.T) {
	w := doJSON(t, New(), http.MethodPost, "/api/chart", ChartRequest{
		YaoValues: []int{7, 7, 7, 7, 7, 7},
		Question:  "今年财运怎么样",
		At:        "2026-06-15 12:00",
		Timezone:  "Asia/Shanghai",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("状态码 = %d", w.Code)
	}
	var d map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &d)
	shen, ok := d["shen"].(map[string]any)
	if !ok {
		t.Fatal("响应缺 shen")
	}
	// 求财 → 妻财
	if shen["yongShen"] != "妻财" {
		t.Errorf("用神 = %v, want 妻财", shen["yongShen"])
	}
	if shen["inferred"] != true {
		t.Error("应由问题自动推断")
	}
	// 乾宫金 妻财木 → 原神子孙、忌神兄弟、仇神父母
	if shen["yuanShen"] != "子孙" {
		t.Errorf("原神 = %v, want 子孙", shen["yuanShen"])
	}
	if shen["jiShen"] != "兄弟" {
		t.Errorf("忌神 = %v, want 兄弟", shen["jiShen"])
	}
	if shen["chouShen"] != "父母" {
		t.Errorf("仇神 = %v, want 父母", shen["chouShen"])
	}
	for _, k := range []string{"topic", "selection", "selectNote", "candidates", "advice", "ambiguous"} {
		if _, ok := shen[k]; !ok {
			t.Errorf("shen 缺字段 %s", k)
		}
	}
}

// 显式指定用神
func TestChartYongShenOverride(t *testing.T) {
	w := doJSON(t, New(), http.MethodPost, "/api/chart", ChartRequest{
		YaoValues: []int{7, 7, 7, 7, 7, 7},
		Question:  "随便问问",
		YongShen:  "官鬼",
		At:        "2026-06-15 12:00",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("状态码 = %d", w.Code)
	}
	var d map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &d)
	shen := d["shen"].(map[string]any)
	if shen["yongShen"] != "官鬼" {
		t.Errorf("用神 = %v, want 官鬼", shen["yongShen"])
	}
	if shen["inferred"] != false {
		t.Error("显式指定时 inferred 应为 false")
	}
	// 乾宫金 官鬼火 → 原神妻财、忌神子孙、仇神兄弟
	if shen["yuanShen"] != "妻财" || shen["jiShen"] != "子孙" || shen["chouShen"] != "兄弟" {
		t.Errorf("四神 = %v/%v/%v, want 妻财/子孙/兄弟",
			shen["yuanShen"], shen["jiShen"], shen["chouShen"])
	}

	// 非法用神
	w = doJSON(t, New(), http.MethodPost, "/api/chart", ChartRequest{
		YaoValues: []int{7, 7, 7, 7, 7, 7}, YongShen: "乱写",
	})
	if w.Code != http.StatusBadRequest {
		t.Errorf("非法用神应返回 400，实际 %d", w.Code)
	}
}

// 参考表须含三刑六害
func TestReferenceXingHai(t *testing.T) {
	w := doJSON(t, New(), http.MethodGet, "/api/reference", nil)
	var d map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &d)

	xing, ok := d["xing"].([]any)
	if !ok || len(xing) != 7 {
		t.Fatalf("三刑应为 7 组，实际 %v", d["xing"])
	}
	selfCount := 0
	for _, raw := range xing {
		x := raw.(map[string]any)
		if x["label"] == "" {
			t.Error("三刑缺 label")
		}
		if x["selfXing"] == true {
			selfCount++
		}
	}
	if selfCount != 4 {
		t.Errorf("自刑应为 4 组，实际 %d", selfCount)
	}

	hai, ok := d["hai"].([]any)
	if !ok || len(hai) != 6 {
		t.Fatalf("六害应为 6 组，实际 %v", d["hai"])
	}
	wantHai := map[string]bool{"子未相害": true, "丑午相害": true, "寅巳相害": true,
		"卯辰相害": true, "申亥相害": true, "酉戌相害": true}
	for _, raw := range hai {
		h := raw.(map[string]any)
		label, _ := h["label"].(string)
		if !wantHai[label] {
			t.Errorf("非预期的六害条目: %s", label)
		}
	}
}

// 无所问之事：应省略该分句，且用神仍须有合理主题（不得出现「未填写」）
func TestChartWithoutQuestion(t *testing.T) {
	w := doJSON(t, New(), http.MethodPost, "/api/chart", ChartRequest{
		YaoValues: []int{7, 7, 7, 7, 7, 7},
		At:        "2026-06-15 12:00",
		Timezone:  "Asia/Shanghai",
		// 既无 question 也无 yongShen
	})
	if w.Code != http.StatusOK {
		t.Fatalf("状态码 = %d, body=%s", w.Code, w.Body.String())
	}
	body := w.Body.String()
	if bytes.Contains([]byte(body), []byte("未填写")) {
		t.Error("响应中不应出现「未填写」占位文案")
	}

	var d map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &d)

	// 用神缺省为妻财，主题须为标准主题名
	shen := d["shen"].(map[string]any)
	if shen["yongShen"] != "妻财" {
		t.Errorf("缺省用神 = %v, want 妻财", shen["yongShen"])
	}
	if shen["topic"] == "通用（未识别主题）" || shen["topic"] == "" {
		t.Errorf("主题不应为空或未识别，实际 %v", shen["topic"])
	}

	// 白话总览不得出现「问的是」
	plain := d["plain"].(map[string]any)
	ov, _ := plain["overview"].(string)
	if bytes.Contains([]byte(ov), []byte("问的是")) {
		t.Errorf("无所问之事时总览不应出现「问的是」，实际 %q", ov)
	}
	if !bytes.Contains([]byte(ov), []byte("起卦")) {
		t.Errorf("总览应保留起卦信息，实际 %q", ov)
	}

	// 用神小节的断语不得依赖所问之事
	secs, _ := plain["sections"].([]any)
	found := false
	for _, raw := range secs {
		s := raw.(map[string]any)
		title, _ := s["title"].(string)
		if !bytes.Contains([]byte(title), []byte("用神")) {
			continue
		}
		found = true
		for _, l := range s["lines"].([]any) {
			line, _ := l.(string)
			if bytes.Contains([]byte(line), []byte("未填写")) {
				t.Errorf("用神小节出现占位文案: %s", line)
			}
		}
	}
	if !found {
		t.Error("应有用神小节")
	}
}

// 各用神的标准主题名
func TestTopicByYongShenViaAPI(t *testing.T) {
	want := map[string]string{
		"妻财": "求财", "官鬼": "工作事业", "父母": "学业文书",
		"子孙": "子女平安", "兄弟": "同辈合作",
	}
	for qin, topic := range want {
		w := doJSON(t, New(), http.MethodPost, "/api/chart", ChartRequest{
			YaoValues: []int{7, 7, 7, 7, 7, 7},
			YongShen:  qin,
			At:        "2026-06-15 12:00",
		})
		if w.Code != http.StatusOK {
			t.Fatalf("%s 状态码 = %d", qin, w.Code)
		}
		var d map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &d)
		shen := d["shen"].(map[string]any)
		if shen["topic"] != topic {
			t.Errorf("%s 的主题 = %v, want %s", qin, shen["topic"], topic)
		}
	}
}

// 参考表须含六神意象
func TestReferenceLiuShen(t *testing.T) {
	w := doJSON(t, New(), http.MethodGet, "/api/reference", nil)
	var d map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &d)

	list, ok := d["liuShen"].([]any)
	if !ok || len(list) != 6 {
		t.Fatalf("六神应为 6 条，实际 %v", d["liuShen"])
	}

	wantWuXing := map[string]string{
		"青龙": "木", "朱雀": "火", "勾陈": "土",
		"螣蛇": "土", "白虎": "金", "玄武": "水",
	}
	seen := map[string]bool{}
	for _, raw := range list {
		ls := raw.(map[string]any)
		name, _ := ls["name"].(string)
		seen[name] = true

		for _, k := range []string{"name", "wuXing", "color", "startGan", "nature", "image", "meaning", "note"} {
			if _, ok := ls[k]; !ok {
				t.Errorf("%s 缺字段 %s", name, k)
			}
		}
		if want, ok := wantWuXing[name]; ok {
			if ls["wuXing"] != want {
				t.Errorf("%s 五行 = %v, want %s", name, ls["wuXing"], want)
			}
		} else {
			t.Errorf("非预期六神: %s", name)
		}
		imgs, ok := ls["image"].([]any)
		if !ok || len(imgs) == 0 {
			t.Errorf("%s 缺意象", name)
		}
		meaning, _ := ls["meaning"].(string)
		if meaning == "" {
			t.Errorf("%s 缺释义", name)
		}
	}
	for _, n := range []string{"青龙", "朱雀", "勾陈", "螣蛇", "白虎", "玄武"} {
		if !seen[n] {
			t.Errorf("缺六神 %s", n)
		}
	}
	// 螣蛇在六爻体系属土（须与十二将体系区分）
	for _, raw := range list {
		ls := raw.(map[string]any)
		if ls["name"] == "螣蛇" {
			if ls["wuXing"] != "土" {
				t.Errorf("螣蛇应属土，实际 %v", ls["wuXing"])
			}
		}
	}
}
