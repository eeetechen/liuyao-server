/* 六爻排盘前端逻辑：纯原生 JS，无依赖。与后端通过 /api/* 通信。 */
(function () {
  'use strict';

  var API = '';
  var POS = ['初爻', '二爻', '三爻', '四爻', '五爻', '上爻'];

  // 手动起卦：6 爻，索引 0 = 初爻
  var state = {
    manual: POS.map(function () { return { yin: false, moving: false }; })
  };

  /* ---------- 工具 ---------- */
  function $(id) { return document.getElementById(id); }

  function esc(s) {
    return String(s == null ? '' : s).replace(/[&<>"']/g, function (c) {
      return { '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c];
    });
  }

  function wxCn(wx) {
    return { '木': 'Mu', '火': 'Huo', '土': 'Tu', '金': 'Jin', '水': 'Shui' }[wx] || '';
  }

  function localNow() {
    var d = new Date(), p = function (n) { return (n < 10 ? '0' : '') + n; };
    return d.getFullYear() + '-' + p(d.getMonth() + 1) + '-' + p(d.getDate()) +
      'T' + p(d.getHours()) + ':' + p(d.getMinutes());
  }

  /* 用 CSS 绘制爻符号，避免 Unicode 方块字符在不同字体下粗细/断口不一致 */
  function yaoGlyph(y, mini) {
    var yang = y.yinYang === 'yang';
    var cls = 'yao ' + (mini ? 'mini ' : '') + (yang ? 'yang' : 'yin') + (y.isChanging ? ' moving' : '');
    var inner = yang ? '<i></i>' : '<i></i><i></i>';
    var dot = y.isChanging ? '<span class="dot">' + (yang ? '○' : '×') + '</span>' : '';
    return '<span class="' + cls + '">' + inner + dot + '</span>';
  }

  /* ---------- 手动起卦 UI ---------- */
  function renderManual() {
    var box = $('manualList');
    box.innerHTML = state.manual.map(function (y, i) {
      return '<div class="manual-row">' +
        '<span class="pos-label">' + POS[i] + '</span>' +
        '<div class="seg" data-i="' + i + '">' +
        '<button data-v="yang" class="' + (y.yin ? '' : 'on') + '">阳</button>' +
        '<button data-v="yin" class="' + (y.yin ? 'on' : '') + '">阴</button>' +
        '</div>' +
        '<label class="chk"><input type="checkbox" data-m="' + i + '"' + (y.moving ? ' checked' : '') + '>动爻</label>' +
        '</div>';
    }).join('');

    box.querySelectorAll('.seg').forEach(function (seg) {
      var i = +seg.dataset.i;
      seg.querySelectorAll('button').forEach(function (b) {
        b.addEventListener('click', function () {
          state.manual[i].yin = (b.dataset.v === 'yin');
          renderManual();
        });
      });
    });
    box.querySelectorAll('input[type=checkbox]').forEach(function (c) {
      c.addEventListener('change', function () {
        state.manual[+c.dataset.m].moving = c.checked;
      });
    });
  }

  function currentValues() {
    return state.manual.map(function (y) {
      if (y.yin) return y.moving ? 6 : 8;   // 老阴 / 少阴
      return y.moving ? 9 : 7;              // 老阳 / 少阳
    });
  }

  /* ---------- 排盘 ---------- */
  $('btnChart').addEventListener('click', function () {
    $('errBox').textContent = '';
    var btn = this;
    btn.disabled = true; btn.textContent = '排盘中…';

    fetch(API + '/api/chart', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        yaoValues: currentValues(),
        method: 'manual',
        at: ($('atTime').value || '').replace('T', ' '),
        timezone: $('tz').value,
        dayBoundary: $('boundary').value,
        yongShen: $('yongShen').value
      })
    })
      .then(function (r) { return r.json().then(function (d) { return { ok: r.ok, d: d }; }); })
      .then(function (res) {
        if (!res.ok) throw new Error(res.d.error || '排盘失败');
        render(res.d);
      })
      .catch(function (e) { $('errBox').textContent = e.message; })
      .finally(function () { btn.disabled = false; btn.textContent = '开 始 排 盘'; });
  });

  /* ---------- 渲染结果 ---------- */
  function render(c) {
    $('placeholder').classList.add('hidden');
    var box = $('result');
    box.classList.remove('hidden');

    box.innerHTML =
      renderHead(c) +
      renderHighlights(c) +
      renderGuaCompare(c) +
      renderChart(c) +
      renderFuShen(c) +
      renderShen(c) +
      renderRelations(c) +
      renderPlain(c) +
      renderReference();
    if (c.plain && c.plain.overview) box.setAttribute('data-overview', c.plain.overview);
  }

  function renderHead(c) {
    var t = c.time;
    return '<div class="res-head">' +
      '<div class="res-question">用神：' + esc((c.shen && c.shen.yongShen) || '—') +
      (c.shen && c.shen.topic ? '<span class="res-topic">' + esc(c.shen.topic) + '</span>' : '') +
      '</div>' +
      '<div class="res-time">' +
      '<div class="res-solar">' + esc(t.solarText) + '　<span class="lunar">' + esc(t.lunarText || '') + '</span></div>' +
      '<div class="year-line">' +
      '<span class="year-badge">' + esc(t.yearLabel || (t.year + '年')) + '</span>' +
      '<span class="year-hint">年份以立春为界</span>' +
      '</div>' +
      '<div class="pillar-grid">' +
      '<span class="pillar"><i>年柱</i><b>' + esc(t.year) + '</b></span>' +
      '<span class="pillar"><i>月柱</i><b>' + esc(t.month) + '</b><em>月建' + esc(t.monthZhi) + '</em></span>' +
      '<span class="pillar"><i>日柱</i><b>' + esc(t.day) + '</b><em>日辰' + esc(t.dayZhi) + '</em></span>' +
      '<span class="pillar"><i>时柱</i><b>' + esc(t.hour) + '</b></span>' +
      '<span class="pillar kong"><i>旬空</i><b>' + esc(t.xunKong) + '</b></span>' +
      '</div>' +
      '</div></div>';
  }

  function typeBadge(name) {
    if (!name) return '';
    var cls = name === '六冲' ? 'chong' : (name === '六合' ? 'he' : 'hun');
    return '<span class="badge ' + cls + '">' + esc(name) + '</span>';
  }

  // 四柱：本卦与变卦左右并列，便于对照；窄屏可横向滑动查看
  function renderGuaCompare(c) {
    var b = c.benGua, v = c.bianGua;

    function card(g, o) {
      if (!g) {
        return '<div class="gua-card empty">' +
          '<div class="gc-title">变卦</div>' +
          '<div class="gc-empty">六爻皆静<br>无变卦</div>' +
          '</div>';
      }
      return '<div class="gua-card' + (o.bian ? ' bian' : '') + '">' +
        '<div class="gc-title">' + (o.bian ? '变卦' : '本卦') + '</div>' +
        '<div class="gc-name">' + esc(g.name) + typeBadge(g.type) + '</div>' +
        '<div class="gc-rows">' +
        '<div class="gc-row"><span>卦宫</span><b>' + esc(g.gong) + '宫</b></div>' +
        '<div class="gc-row"><span>五行</span><b class="wx-' + wxCn(g.element) + '">' + esc(g.element) + '</b></div>' +
        '<div class="gc-row"><span>上下卦</span><b>' + esc(g.upperGua) + '上 ' + esc(g.lowerGua) + '下</b></div>' +
        '<div class="gc-row"><span>世应</span><b>世' + g.shiPos + ' 应' + g.yingPos + '</b></div>' +
        (o.bian && o.fromText
          ? '<div class="gc-row"><span>由来</span><b>' + esc(o.fromText) + '</b></div>'
          : '') +
        '</div>' +
        (o.yaoHtml ? '<div class="gc-yaos">' + o.yaoHtml + '</div>' : '') +
        '</div>';
    }

    // 本卦六爻（自上而下显示）
    var benYaos = c.yaos.slice().reverse().map(function (y) {
      return '<div class="gc-yao">' +
        '<span class="gcy-pos">' + esc(y.name) + '</span>' +
        yaoGlyph(y, true) +
        '<span class="gcy-text">' + esc(y.liuQin) + '</span>' +
        '</div>';
    }).join('');

    // 变卦六爻（仅动爻标出变化）
    var bianYaos = '';
    if (v) {
      bianYaos = c.yaos.slice().reverse().map(function (y, ri) {
        var i = 5 - ri;                     // 还原为自下而上的索引
        var bnj = (v.naJia && v.naJia[i]) || {};
        var qin = (v.liuQin && v.liuQin[i]) || '';
        var glyph;
        if (y.isChanging) {
          glyph = yaoGlyph({ yinYang: y.changedYinYang, isChanging: false }, true);
          return '<div class="gc-yao changed">' +
            '<span class="gcy-pos">' + esc(y.name) + '</span>' +
            glyph +
            '<span class="gcy-text">' + esc(bnj.ganZhi || '') + ' ' + esc(qin) + '</span>' +
            '</div>';
        }
        glyph = yaoGlyph({ yinYang: y.yinYang, isChanging: false }, true);
        return '<div class="gc-yao dim">' +
          '<span class="gcy-pos">' + esc(y.name) + '</span>' +
          glyph +
          '<span class="gcy-text">' + esc(bnj.ganZhi || '') + ' ' + esc(qin) + '</span>' +
          '</div>';
      }).join('');
    }

    var fromText = '';
    var ps = [];
    c.yaos.forEach(function (y) { if (y.isChanging) ps.push(y.name); });
    if (ps.length) fromText = ps.join('、') + ' 发动';

    return '<div class="gua-compare">' +
      '<div class="gc-scroll">' +
      card(b, { bian: false, yaoHtml: benYaos }) +
      (v ? '<div class="gc-arrow">→</div>' : '') +
      card(v, { bian: true, yaoHtml: bianYaos, fromText: fromText }) +
      '</div>' +
      (v ? '' : '<p class="gc-note">六爻皆静，事情暂无明显变化，以本卦断之。</p>') +
      '</div>';
  }

  function renderHighlights(c) {
    var hs = c.summary.highlights || [];
    if (!hs.length) return '';
    return '<div class="highlights">' + hs.map(function (x) {
      var cls = 'hl', ic = '※';
      if (/三合|半合/.test(x)) { cls += ' sanhe'; ic = '⬡'; }
      else if (/六冲/.test(x)) { cls += ' chong'; ic = '⚡'; }
      else if (/六合/.test(x)) { cls += ' he'; ic = '◉'; }
      else if (/暗动/.test(x)) { cls += ' andong'; ic = '⚑'; }
      else if (/月破|日破/.test(x)) { cls += ' break'; ic = '✕'; }
      else if (/旬空/.test(x)) { cls += ' kong'; ic = '○'; }
      else if (/世应|生|克/.test(x)) { cls += ' shengke'; ic = '⇄'; }
      return '<div class="' + cls + '"><span class="ic">' + ic + '</span><span>' + esc(x) + '</span></div>';
    }).join('') + '</div>';
  }

  function renderChart(c) {
    var rows = c.yaos.slice().reverse().map(function (y) {   // 上爻在上
      var trCls = '';
      if (y.isChanging) trCls = 'moving';
      else if (y.moveState === 'hidden_moving') trCls = 'andong';
      if (y.isShi) trCls += ' shi';
      if (y.isYing) trCls += ' ying';

      var sym = '<span class="sym-cell">' + yaoGlyph(y) + '</span>';

      var badges = '';
      badges += '<span class="badge-2 b-' + esc(y.wangShuai) + '">' + esc(y.wangShuai) + '</span> ';
      if (y.special === 'anDong') badges += '<span class="badge-2 b-andong">暗动</span> ';
      if (y.special === 'riPo') badges += '<span class="badge-2 b-ripo">日破</span> ';
      if (y.special === 'yuePo') badges += '<span class="badge-2 b-yuepo">月破</span> ';
      if (y.isChanging) badges += '<span class="badge-2 b-move">明动</span> ';
      if (y.isKong) badges += '<span class="badge-2 b-kong">旬空</span> ';
      if (y.isShi) badges += '<span class="badge-2 b-shi">世</span> ';
      if (y.isYing) badges += '<span class="badge-2 b-ying">应</span> ';

      var strength = '<span class="badge-2 b-' + (y.strength === 'strong' ? 'strong' : y.strength === 'moderate' ? 'mid' : 'weak') + '">' +
        (y.strength === 'strong' ? '有力' : y.strength === 'moderate' ? '平平' : '无力') + '</span>';

      var ev = (y.evidence && y.evidence.length)
        ? '<div class="evidence">' + y.evidence.map(function (e) { return '<span>' + esc(e) + '</span>'; }).join('') + '</div>'
        : '';

      var changed = '';
      if (y.isChanging && y.changedGanZhi) {
        changed = '<div class="gua-sub">变 ' + esc(y.changedGanZhi) + ' ' + esc(y.changedLiuQin) +
          (y.huaType ? ' · ' + esc(y.huaType) : '') + '</div>';
      }

      return '<tr class="' + trCls.trim() + '">' +
        '<td class="pos-cell">' + esc(y.name) + '</td>' +
        '<td class="sym-cell">' + sym + '</td>' +
        '<td class="col-gz"><span class="wx-' + wxCn(y.wuXing) + '">' + esc(y.ganZhi) + esc(y.wuXing) + '</span>' + changed + '</td>' +
        '<td class="col-qin qin">' + esc(y.liuQin) + '</td>' +
        '<td class="col-shen hide-sm">' + esc(y.liuShen) + '</td>' +
        '<td class="col-status">' + badges + '</td>' +
        '<td class="col-strength hide-sm">' + strength + '</td>' +
        '<td class="col-ev">' + ev + '</td>' +
        '</tr>';
    }).join('');

    return '<div class="chart-wrap"><table class="gua">' +
      '<thead><tr>' +
      '<th>爻位</th><th>卦象</th><th>纳甲</th><th>六亲</th><th>六神</th><th>旺衰·动静</th><th>力量</th><th>判定依据</th>' +
      '</tr></thead><tbody>' + rows + '</tbody></table></div>';
  }

  function renderFuShen(c) {
    if (!c.fuShen || !c.fuShen.length) return '';
    return '<div class="fushen"><h3>伏神</h3>' + c.fuShen.map(function (f) {
      return '<div class="fs-item">' + POS[f.position - 1] + ' 伏 <b>' + esc(f.liuQin) + '</b> ' +
        '<span class="wx-' + wxCn(f.wuXing) + '">' + esc(f.ganZhi) + esc(f.wuXing) + '</span></div>';
    }).join('') + '</div>';
  }

  /* ---------- 用神体系 ---------- */

  function renderShen(c) {
    var sh = c.shen;
    if (!sh) return '';
    if (!sh.yongShen) return '';

    var h = '<section class="shen-block">' +
      '<h3 class="rel-title"><span class="rel-ic shen">◎</span>用神体系' +
      '<span class="rel-count">' + (sh.inferred ? '自动取用' : '指定用神') + '</span></h3>';

    // 四神链
    // 依「助我者—主事者—伤我者」之序排列，便于对读
    h += '<div class="shen-chain">' +
      shenCard('原神', sh.yuanShen, '生用神', 'yuan') +
      shenCard('用神', sh.yongShen, sh.topic || '主事之爻', 'yong') +
      shenCard('忌神', sh.jiShen, '克用神', 'ji') +
      shenCard('仇神', sh.chouShen, '生忌神', 'chou') +
      '</div>';

    // 取用说明
    if (sh.selectNote) {
      h += '<div class="shen-note">' + esc(sh.selectNote) + '</div>';
    }

    // 主用神详表
    if (sh.yongYao) {
      var y = sh.yongYao;
      h += '<div class="shen-detail">' +
        '<span class="sd-tag">主用神</span>' +
        '<span class="sd-main">' + esc(y.label) + ' ' + esc(y.ganZhi) +
        '<b class="wx-' + wxCn(y.wuXing) + '">' + esc(y.wuXing) + '</b></span>' +
        '<span class="badge-2 b-' + esc(y.wangShuai) + '">' + esc(y.wangShuai) + '</span> ' +
        (y.moving ? '<span class="badge-2 b-move">明动</span> ' : '') +
        (y.special === 'anDong' ? '<span class="badge-2 b-andong">暗动</span> ' : '') +
        (y.special === 'riPo' ? '<span class="badge-2 b-ripo">日破</span> ' : '') +
        (y.special === 'yuePo' ? '<span class="badge-2 b-yuepo">月破</span> ' : '') +
        (y.isKong ? '<span class="badge-2 b-kong">旬空</span> ' : '') +
        (y.isShi ? '<span class="badge-2 b-shi">世</span> ' : '') +
        (y.isYing ? '<span class="badge-2 b-ying">应</span> ' : '') +
        (y.isFuShen ? '<span class="badge-2 b-fushen">伏神</span> ' : '') +
        '<span class="badge-2 b-' + (y.strength === 'strong' ? 'strong' : y.strength === 'moderate' ? 'mid' : 'weak') + '">' +
        (y.strength === 'strong' ? '有力' : y.strength === 'moderate' ? '平平' : '无力') + '</span>' +
        '</div>';
    }

    // 候选（多现时）
    if (sh.candidates && sh.candidates.length > 1) {
      h += '<div class="shen-cands"><span class="sc-label">并见</span>' +
        sh.candidates.map(function (y) {
          return '<span class="sc-chip">' + esc(y.label) + ' ' + esc(y.ganZhi) +
            '<b class="wx-' + wxCn(y.wuXing) + '">' + esc(y.wuXing) + '</b>' +
            '<i>' + esc(y.wangShuai) + '</i></span>';
        }).join('') + '</div>';
    }

    // 断语
    if (sh.advice && sh.advice.length) {
      h += '<ul class="shen-advice">' + sh.advice.map(function (a) {
        return '<li>' + esc(a) + '</li>';
      }).join('') + '</ul>';
    }

    return h + '</section>';
  }

  function shenCard(role, qin, note, cls) {
    return '<div class="shen-card ' + cls + '">' +
      '<div class="sc-role">' + esc(role) + '</div>' +
      '<div class="sc-qin">' + esc(qin || '—') + '</div>' +
      (note ? '<div class="sc-note">' + esc(note) + '</div>' : '') +
      '</div>';
  }

  /* ---------- 生合冲克 ---------- */

  function infBadge(influence) {
    if (!influence || !influence.kind) return '<span class="inf inf-none">—</span>';
    return '<span class="inf inf-' + esc(influence.kind) + '">' + esc(influence.label) + '</span>';
  }

  function renderShengKe(c) {
    var r = c.relations;
    if (!r || !r.shengKe || !r.shengKe.length) return '';

    var rows = r.shengKe.map(function (sk) {
      var yao = c.yaos[sk.position - 1] || {};
      var trCls = yao.isShi ? 'shi' : (yao.isYing ? 'ying' : '');
      // data-label 供手机端（表头隐藏时）标注列义
      return '<tr class="' + trCls + '">' +
        '<td class="pos-cell" data-label="爻位">' + esc(sk.label) + '</td>' +
        '<td data-label="纳甲"><span class="wx-' + wxCn(sk.wuXing) + '">' + esc(sk.zhi + sk.wuXing) + '</span></td>' +
        '<td class="qin" data-label="六亲">' + esc(sk.liuQin) + '</td>' +
        '<td data-label="月建作用">' + infBadge(sk.month) + '</td>' +
        '<td data-label="日辰作用">' + infBadge(sk.day) + '</td>' +
        '</tr>';
    }).join('');

    var h = '<section class="rel-block">' +
      '<h3 class="rel-title"><span class="rel-ic">⇄</span>五行生克</h3>' +
      '<div class="chart-wrap"><table class="rel-table">' +
      '<thead><tr><th>爻位</th><th>纳甲</th><th>六亲</th><th>月建作用</th><th>日辰作用</th></tr></thead>' +
      '<tbody>' + rows + '</tbody></table></div>';

    if (r.shiYing) {
      var sr = r.shiYing;
      h += '<div class="shiying">' +
        '<span class="sy-label">世应关系</span>' +
        '<span class="sy-main">' + esc(sr.label) + '</span>' +
        '<span class="sy-detail">世' + sr.shiPosition + ' <span class="wx-' + wxCn(sr.shiWuXing) + '">' + esc(sr.shiWuXing) + '</span>' +
        ' · 应' + sr.yingPosition + ' <span class="wx-' + wxCn(sr.yingWuXing) + '">' + esc(sr.yingWuXing) + '</span>' +
        ' — ' + esc(sr.desc) + '</span>' +
        '</div>';
    }
    return h + '</section>';
  }

  function pairChip(ref) {
    return '<span class="bc-chip' + (ref.active ? ' active' : '') + '">' +
      esc(ref.label) + '<b class="wx-' + wxCn(ref.wuXing) + '">' + esc(ref.zhi) + '</b>' +
      (ref.moving ? '<em>动</em>' : '') + '</span>';
  }

  function kindTag(kind) {
    var m = { 'yao-yao': '爻间', 'yao-day': '日辰', 'yao-month': '月建' };
    return '<span class="bc-kind">' + (m[kind] || '') + '</span>';
  }

  function renderPairs(list, type, title, ic, emptyText) {
    var h = '<section class="rel-block">' +
      '<h3 class="rel-title"><span class="rel-ic ' + type + '">' + ic + '</span>' + title +
      '<span class="rel-count">' + list.length + '</span></h3>';
    if (!list.length) {
      return h + '<p class="rel-empty">' + emptyText + '</p></section>';
    }
    h += '<div class="bc-list">' + list.map(function (p) {
      var hua = p.hua ? '<span class="bc-hua">合化' + esc(p.hua) + '</span>' : '';
      var note = p.note ? '<span class="bc-note">' + esc(p.note) + '</span>' : '';
      return '<div class="bc-item ' + type + '">' +
        pairChip(p.a) + '<span class="bc-op">' + (type === 'he' ? '合' : '冲') + '</span>' + pairChip(p.b) +
        kindTag(p.kind) + hua + note +
        '</div>';
    }).join('') + '</div></section>';
    return h;
  }

  function renderSanHe(c) {
    var r = c.relations;
    var full = (r && r.sanHe) || [];
    var ban = (r && r.banHe) || [];
    if (!full.length && !ban.length) {
      return '<section class="rel-block">' +
        '<h3 class="rel-title"><span class="rel-ic sanhe">⬡</span>三合<span class="rel-count">0</span></h3>' +
        '<p class="rel-empty">六爻与日月未见三合局或半合。</p></section>';
    }
    var h = '<section class="rel-block">' +
      '<h3 class="rel-title"><span class="rel-ic sanhe">⬡</span>三合' +
      '<span class="rel-count">' + (full.length + ban.length) + '</span></h3>';

    function ju(j, isBan) {
      var flags = '';
      if (j.withMonth) flags += '<span class="ju-flag">含月建</span>';
      if (j.withDay) flags += '<span class="ju-flag">含日辰</span>';
      if (j.hasActive) flags += '<span class="ju-flag on">有动爻</span>';
      if (j.complete) flags += '<span class="ju-flag on">三支全</span>';
      return '<div class="ju-item' + (isBan ? ' ban' : '') + '">' +
        '<div class="ju-head">' +
        '<span class="ju-name">' + esc(j.name) + '</span>' +
        '<span class="ju-el wx-' + wxCn(j.element) + '">' + esc(j.element) + '局</span>' +
        '<span class="ju-type">' + (isBan ? '半合' : '三合') + '</span>' +
        '<span class="ju-pos">' + (j.positions && j.positions.length ? '爻位 ' + j.positions.join('/') : '仅日月') + '</span>' +
        '</div>' +
        '<div class="ju-desc">' + esc(j.desc) + '</div>' +
        '<div class="ju-flags">' + flags + '</div>' +
        '</div>';
    }

    if (full.length) h += '<div class="ju-list">' + full.map(function (j) { return ju(j, false); }).join('') + '</div>';
    if (ban.length) h += '<div class="ju-list">' + ban.map(function (j) { return ju(j, true); }).join('') + '</div>';
    return h + '</section>';
  }

  function renderRelations(c) {
    var r = c.relations;
    if (!r) return '';
    return '<div class="relations">' +
      renderShengKe(c) +
      renderSanHe(c) +
      renderPairs(r.liuHe || [], 'he', '六合', '◉', '未见六合。') +
      renderPairs(r.liuChong || [], 'chong', '六冲', '⚡', '未见六冲。') +
      renderXing(r.xing || []) +
      renderHai(r.hai || []) +
      '</div>';
  }

  /* ---------- 白话解析 ---------- */

  function renderPlain(c) {
    var p = c.plain;
    if (!p) return '';
    var h = '<section class="plain-block">' +
      '<h3 class="rel-title"><span class="rel-ic plain">◷</span>白话解析' +
      '<span class="rel-count">通俗版</span></h3>';

    if (p.overview) {
      h += '<div class="plain-overview">' + esc(p.overview) + '</div>';
    }

    h += (p.sections || []).map(function (s) {
      return '<div class="plain-sec">' +
        '<div class="ps-title">' + esc(s.title) + '</div>' +
        '<ul class="ps-lines">' + (s.lines || []).map(function (l) {
          return '<li>' + esc(l) + '</li>';
        }).join('') + '</ul>' +
        '</div>';
    }).join('');

    if (p.suggestions && p.suggestions.length) {
      // 序号由实际小节数推导，避免与后端小节编号重复
      var CN = ['一', '二', '三', '四', '五', '六', '七', '八', '九', '十'];
      var nextNo = CN[(p.sections || []).length] || String((p.sections || []).length + 1);
      h += '<div class="plain-sec">' +
        '<div class="ps-title">' + nextNo + '、参考建议</div>' +
        '<ul class="ps-lines suggest">' + p.suggestions.map(function (l, i) {
          var last = i === p.suggestions.length - 1;
          return '<li' + (last ? ' class="disclaimer"' : '') + '>' + esc(l) + '</li>';
        }).join('') + '</ul>' +
        '</div>';
    }

    return h + '</section>';
  }

  /* ---------- 五行生克参考表 ---------- */

  var REF = null;          // 参考表缓存
  var refLoading = false;

  function loadReference(cb) {
    if (REF) { cb(REF); return; }
    if (refLoading) return;
    refLoading = true;
    fetch(API + '/api/reference')
      .then(function (r) { return r.json(); })
      .then(function (d) { REF = d; refLoading = false; cb(d); })
      .catch(function () { refLoading = false; });
  }

  function renderXing(list) {
    var h = '<section class="rel-block">' +
      '<h3 class="rel-title"><span class="rel-ic xing">✕</span>三刑' +
      '<span class="rel-count">' + list.length + '</span></h3>';
    if (!list.length) {
      return h + '<p class="rel-empty">未见三刑。</p></section>';
    }
    h += '<div class="bc-list">' + list.map(function (x) {
      var chips = (x.members || []).map(function (m) {
        return '<span class="bc-chip' + (m.active ? ' active' : '') + '">' +
          esc(m.label) + '<b class="wx-' + wxCn(m.wuXing) + '">' + esc(m.zhi) + '</b></span>';
      }).join('<span class="bc-op">·</span>');
      return '<div class="bc-item xing">' +
        '<span class="bc-name">' + esc(x.name) + (x.selfXing ? '<em>自刑</em>' : '') + '</span>' +
        chips +
        '</div>';
    }).join('') + '</div>' +
      '<p class="rel-note">刑主刑伤、纠纷、官讼、身体损伤，多指内部摩擦或人际失和。</p></section>';
    return h;
  }

  function renderHai(list) {
    var h = '<section class="rel-block">' +
      '<h3 class="rel-title"><span class="rel-ic hai">○</span>六害' +
      '<span class="rel-count">' + list.length + '</span></h3>';
    if (!list.length) {
      return h + '<p class="rel-empty">未见六害。</p></section>';
    }
    h += '<div class="bc-list">' + list.map(function (p) {
      return '<div class="bc-item hai">' +
        pairChip(p.a) + '<span class="bc-op">害</span>' + pairChip(p.b) +
        kindTag(p.kind) +
        '</div>';
    }).join('') + '</div>' +
      '<p class="rel-note">害主暗损、彼此猜忌、小事牵绊，其力弱于刑冲。</p></section>';
    return h;
  }

  function renderReference() {
    // 先返回占位骨架，随后异步填充（参考表为静态数据，不随卦变）
    setTimeout(function () {
      loadReference(function (d) {
        var el = $('refBody');
        if (el) el.innerHTML = referenceHTML(d);
      });
    }, 0);

    return '<section class="ref-block">' +
      '<h3 class="rel-title"><span class="rel-ic ref">▦</span>五行生克参考表' +
      '<span class="rel-count">查询印证</span></h3>' +
      '<p class="ref-hint">查不到的对照关系可以在这里核对。例如「水克火」「卯属木」「六冲有哪几组」。</p>' +
      '<div id="refBody" class="ref-body"><p class="rel-empty">载入中…</p></div>' +
      '</section>';
  }

  function referenceHTML(d) {
    function wxTable() {
      return '<div class="ref-sub">' +
        '<h4>五行属性与生克</h4>' +
        '<div class="ref-grid">' +
        d.wuXing.map(function (w) {
          return '<div class="wx-card">' +
            '<div class="wx-head"><b class="wx-' + wxCn(w.wuXing) + '">' + esc(w.wuXing) + '</b>' +
            '<span>' + esc(w.color) + '色 · ' + esc(w.direction) + ' · ' + esc(w.season) + '</span></div>' +
            '<div class="wx-rel">' +
            '<span class="wx-line">' + esc(w.wuXing) + ' 生 <b class="wx-' + wxCn(w.sheng) + '">' + esc(w.sheng) + '</b></span>' +
            '<span class="wx-line">' + esc(w.wuXing) + ' 克 <b class="wx-' + wxCn(w.ke) + '">' + esc(w.ke) + '</b></span>' +
            '<span class="wx-line dim">' + esc(w.beiSheng) + ' 生我</span>' +
            '<span class="wx-line dim">' + esc(w.beiKe) + ' 克我</span>' +
            '</div></div>';
        }).join('') +
        '</div>' +
        '<div class="sk-pairs">' +
        d.shengKe.map(function (x) {
          return '<span class="sk-chip ' + (x.kind === '生' ? 'sheng' : 'ke') + '">' + esc(x.text) + '</span>';
        }).join('') +
        '</div>' +
        '</div>';
    }

    function ganzhiTable() {
      function rows(list, title) {
        return '<div class="ref-sub"><h4>' + title + '</h4>' +
          '<div class="gz-grid">' +
          list.map(function (g) {
            var extra = g.shengXiao ? '<em>' + esc(g.shengXiao) + '</em>' : '';
            return '<span class="gz-chip">' +
              '<b class="wx-' + wxCn(g.wuXing) + '">' + esc(g.name) + '</b>' +
              '<i class="wx-' + wxCn(g.wuXing) + '">' + esc(g.wuXing) + '</i>' +
              '<s>' + esc(g.yinYang) + '</s>' + extra +
              '</span>';
          }).join('') +
          '</div></div>';
      }
      return rows(d.tianGan, '十天干所属五行') + rows(d.diZhi, '十二地支所属五行与生肖');
    }

    function pairTable(list, title, kind) {
      return '<div class="ref-sub"><h4>' + title + '</h4>' +
        '<div class="pair-grid ' + kind + '">' +
        list.map(function (p) {
          var mid = kind === 'chong' ? '冲' : (kind === 'hai' ? '害' : '合');
          var hua = p.hua ? '<em>化' + esc(p.hua) + '</em>' : '';
          return '<span class="pair-chip ' + kind + '">' +
            '<b class="wx-' + wxCn(wuXingOfZhi(p.a)) + '">' + esc(p.a) + '</b>' +
            '<u>' + mid + '</u>' +
            '<b class="wx-' + wxCn(wuXingOfZhi(p.b)) + '">' + esc(p.b) + '</b>' +
            hua +
            '</span>';
        }).join('') +
        '</div></div>';
    }

    function sanheTable() {
      return '<div class="ref-sub"><h4>三合局</h4>' +
        '<div class="pair-grid he">' +
        d.sanHe.map(function (j) {
          return '<span class="pair-chip he sanhe">' +
            '<b>' + esc(j.name) + '</b><u>合</u>' +
            '<em class="wx-' + wxCn(j.element) + '">' + esc(j.element) + '局</em>' +
            '</span>';
        }).join('') +
        '</div></div>';
    }

    return wxTable() + ganzhiTable() +
      pairTable(d.liuChong, '六冲（六组）', 'chong') +
      pairTable(d.liuHe, '六合（六组）', 'he') +
      sanheTable() +
      pairTable(d.banHe, '半合（八组，须缺第三支）', 'he') +
      xingTable(d.xing) +
      pairTable(d.hai, '六害（六组）', 'hai') +
      liuShenTable(d.liuShen) +
      '<div class="ref-sub"><h4>说明</h4><ul class="ref-notes">' +
      d.notes.map(function (n) { return '<li>' + esc(n) + '</li>'; }).join('') +
      '</ul></div>';
  }

  function liuShenTable(list) {
    if (!list || !list.length) return '';
    return '<div class="ref-sub"><h4>六神意象（六兽）</h4>' +
      '<p class="ls-intro">六神按日干起初爻、顺次上排：甲乙起青龙、丙丁起朱雀、戊起勾陈、' +
      '己起螣蛇、庚辛起白虎、壬癸起玄武。六神只论吉凶之象，本身不论旺衰。</p>' +
      '<div class="ls-grid">' +
      list.map(function (ls) {
        var natureCls = ls.nature === '吉' ? 'ji' : (ls.nature === '凶' ? 'xiong' : 'zhong');
        return '<div class="ls-card ' + natureCls + '">' +
          '<div class="ls-head">' +
          '<b class="wx-' + wxCn(ls.wuXing) + '">' + esc(ls.name) + '</b>' +
          '<span class="ls-wx wx-' + wxCn(ls.wuXing) + '">' + esc(ls.wuXing) + '</span>' +
          '<span class="ls-nature ' + natureCls + '">' + esc(ls.nature) + '</span>' +
          '<span class="ls-gan">' + esc(ls.startGan.join('')) + '日起</span>' +
          '</div>' +
          '<div class="ls-meaning">' + esc(ls.meaning) + '</div>' +
          '<ul class="ls-image">' +
          ls.image.map(function (im) {
            var i = im.indexOf('：');
            if (i > 0) {
              return '<li><i>' + esc(im.slice(0, i)) + '</i>' + esc(im.slice(i + 1)) + '</li>';
            }
            return '<li>' + esc(im) + '</li>';
          }).join('') +
          '</ul>' +
          (ls.note ? '<div class="ls-note">' + esc(ls.note) + '</div>' : '') +
          '</div>';
      }).join('') +
      '</div></div>';
  }

  function xingTable(list) {
    if (!list || !list.length) return '';
    return '<div class="ref-sub"><h4>三刑</h4>' +
      '<div class="pair-grid xing">' +
      list.map(function (x) {
        if (x.selfXing) {
          return '<span class="pair-chip xing self">' +
            '<b class="wx-' + wxCn(wuXingOfZhi(x.branches[0])) + '">' + esc(x.branches[0]) + '</b>' +
            '<u>自刑</u></span>';
        }
        var inner = x.branches.map(function (b, i) {
          var sep = i > 0 ? '<u>刑</u>' : '';
          return sep + '<b class="wx-' + wxCn(wuXingOfZhi(b)) + '">' + esc(b) + '</b>';
        }).join('');
        return '<span class="pair-chip xing"><span class="pc-name">' + esc(x.name) + '</span>' + inner + '</span>';
      }).join('') +
      '</div></div>';
  }

  // 由参考表反查地支五行（前端本地表，避免额外请求）
  var ZHI_WX = {
    '子': '水', '丑': '土', '寅': '木', '卯': '木', '辰': '土', '巳': '火',
    '午': '火', '未': '土', '申': '金', '酉': '金', '戌': '土', '亥': '水'
  };
  function wuXingOfZhi(z) { return ZHI_WX[z] || ''; }

  /* ---------- 用神参考文本 ---------- */
  // 每个用神对应的取用参考，随下拉选择同步更新并常驻显示
  var YONG_REF = {
    '妻财': '求财、买卖、投资、收益、工资、货款；男占婚姻亦取妻财。主财物与可控之利。',
    '官鬼': '工作、事业、职位、升迁、功名；亦主官非、诉讼、疾病、阻碍。凡有约束力者多属之。',
    '父母': '学业、考试、文书、合同、证件、房产、车船；亦主长辈、上级、公司单位。主庇护与文书。',
    '子孙': '子女、晚辈、生育；亦主平安、解忧、医药、僧道、宠物。为福神，能克制官鬼。',
    '兄弟': '同辈、朋友、同事、合作、竞争；亦主破财、借贷、花费。为劫财之神，与用神争利。'
  };

  function updateYongHint() {
    var v = $('yongShen').value;
    var el = $('yongHint');
    if (!el) return;
    el.textContent = YONG_REF[v] || '';
    el.setAttribute('data-yong', v);
  }

  $('yongShen').addEventListener('change', updateYongHint);

  /* ---------- 初始化 ---------- */
  $('atTime').value = localNow();
  renderManual();
  updateYongHint();
})();
