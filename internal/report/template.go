package report

const htmlTemplate = `<!doctype html>
<html lang="es">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>{{.Title}}</title>
  <style>
    :root {
      --notion-bg: #f7f6f3;
      --notion-page: #ffffff;
      --notion-text: #37352f;
      --notion-muted: #787774;
      --notion-border: #e9e7e3;
      --notion-hover: #f1f1ef;
      --notion-blue: #2f80ed;
      --notion-green: #448361;
      --notion-orange: #d9730d;
      --notion-red: #d44c47;
      --notion-purple: #6940a5;
      --notion-code: #f7f6f3;
      --notion-shadow: 0 1px 2px rgba(15, 15, 15, .06);
    }
    * { box-sizing: border-box; }
    html { scroll-behavior: smooth; }
    body.notion-page {
      margin: 0;
      background: var(--notion-bg);
      color: var(--notion-text);
      font-family: ui-sans-serif, -apple-system, BlinkMacSystemFont, "Segoe UI", Helvetica, Arial, sans-serif;
      line-height: 1.55;
    }
    a { color: inherit; text-decoration: none; }
    .notion-shell {
      display: grid;
      grid-template-columns: 248px minmax(0, 1fr);
      gap: 0;
      max-width: 1520px;
      margin: 0 auto;
    }
    .notion-toc {
      position: sticky;
      top: 0;
      height: 100vh;
      padding: 24px 16px;
      border-right: 1px solid var(--notion-border);
      background: rgba(247, 246, 243, .92);
      overflow: auto;
    }
    .toc-title {
      font-size: 12px;
      color: var(--notion-muted);
      text-transform: uppercase;
      letter-spacing: .08em;
      margin: 0 0 12px;
      font-weight: 700;
    }
    .toc-link {
      display: block;
      padding: 7px 9px;
      border-radius: 6px;
      color: var(--notion-muted);
      font-size: 14px;
    }
    .toc-link:hover { background: var(--notion-hover); color: var(--notion-text); }
    .notion-document {
      min-width: 0;
      background: var(--notion-page);
      padding: 42px clamp(18px, 5vw, 72px) 72px;
    }
    .doc-header {
      max-width: 1120px;
      margin: 0 auto 28px;
    }
    .doc-icon { font-size: 58px; line-height: 1; margin-bottom: 14px; }
    h1, h2, h3, h4 { color: var(--notion-text); letter-spacing: -.02em; }
    h1 { font-size: clamp(34px, 6vw, 54px); line-height: 1.05; margin: 0 0 12px; font-weight: 750; }
    h2 { font-size: 25px; line-height: 1.2; margin: 0 0 14px; font-weight: 720; }
    h3 { font-size: 17px; margin: 0 0 10px; font-weight: 680; }
    h4 { font-size: 14px; margin: 18px 0 8px; font-weight: 680; }
    .subtitle { color: var(--notion-muted); margin: 0 0 14px; font-size: 15px; }
    .notion-meta {
      display: flex;
      flex-wrap: wrap;
      gap: 8px;
      margin: 14px 0 18px;
    }
    .notion-badge, .pill {
      display: inline-flex;
      align-items: center;
      gap: 6px;
      border-radius: 999px;
      padding: 4px 9px;
      background: var(--notion-hover);
      color: var(--notion-muted);
      border: 1px solid transparent;
      font-size: 12px;
      font-weight: 650;
      white-space: nowrap;
    }
    .badge-blue { color: #1d4f8f; background: #eaf3ff; }
    .badge-green { color: #2f6f4e; background: #edf7f1; }
    .badge-orange { color: #9a4f05; background: #fbf0e3; }
    .badge-red, .failed { color: #9f2f2c; background: #fdebec; border-color: #f4c7c3; }
    .success { color: #2f6f4e; background: #edf7f1; border-color: #cfe7d8; }
    .toolbar {
      display: flex;
      flex-wrap: wrap;
      gap: 8px;
      margin-top: 14px;
      overflow-x: auto;
      padding-bottom: 2px;
    }
    button {
      border: 1px solid var(--notion-border);
      background: #fff;
      color: var(--notion-text);
      border-radius: 6px;
      padding: 7px 10px;
      cursor: pointer;
      font-weight: 620;
      font-size: 13px;
      min-height: 34px;
    }
    button:hover { background: var(--notion-hover); }
    button.primary { background: var(--notion-text); color: white; border-color: var(--notion-text); }
    .notion-block {
      max-width: 1120px;
      margin: 0 auto 18px;
      padding: 20px 0;
      border-top: 1px solid var(--notion-border);
      scroll-margin-top: 20px;
    }
    .notion-block:first-of-type { border-top: 0; }
    .notion-callout {
      display: flex;
      gap: 12px;
      align-items: flex-start;
      padding: 14px 16px;
      margin: 16px 0;
      border-radius: 8px;
      background: var(--notion-bg);
      border: 1px solid var(--notion-border);
    }
    .notion-callout .emoji { font-size: 22px; line-height: 1.2; }
    .notion-callout p { margin: 0; }
    .grid { display: grid; gap: 12px; }
    .cards { grid-template-columns: repeat(4, minmax(0, 1fr)); }
    .group-cards { grid-template-columns: repeat(3, minmax(0, 1fr)); }
    .charts { grid-template-columns: repeat(2, minmax(0, 1fr)); }
    .three-col { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 12px; }
    .card, .group-card, .chart-card, .final-column {
      background: #fff;
      border: 1px solid var(--notion-border);
      border-radius: 8px;
      box-shadow: var(--notion-shadow);
      min-width: 0;
    }
    .card { padding: 14px; }
    .card .label {
      color: var(--notion-muted);
      font-size: 11px;
      text-transform: uppercase;
      letter-spacing: .08em;
      font-weight: 720;
    }
    .card .value { font-size: 25px; font-weight: 740; margin-top: 5px; letter-spacing: -.03em; }
    .card .hint { color: var(--notion-muted); font-size: 12px; margin-top: 4px; }
    .group-card { padding: 16px; }
    .group-card h3 { display: flex; align-items: center; gap: 8px; }
    .metric-list { display: grid; grid-template-columns: 1fr 1fr; gap: 8px; margin: 12px 0; }
    .metric { background: var(--notion-bg); border: 1px solid var(--notion-border); border-radius: 7px; padding: 9px; font-size: 12px; color: var(--notion-muted); }
    .metric b { display:block; color: var(--notion-text); font-size: 17px; margin-top: 2px; }
    .metric small { display:block; margin-top:2px; }
    .coverage-bar { height: 8px; background: #ebebea; border-radius: 999px; overflow:hidden; margin: 10px 0; }
    .coverage-bar span { display:block; height:100%; background: linear-gradient(90deg, var(--notion-blue), var(--notion-green)); }
    .chart-card { padding: 14px; overflow: hidden; }
    .chart-card h3 { margin-bottom: 8px; }
    canvas { width: 100%; max-width: 100%; height: 260px; display: block; }
    .table-wrap { overflow-x: auto; border: 1px solid var(--notion-border); border-radius: 8px; }
    table { width: 100%; min-width: 760px; border-collapse: collapse; font-size: 13px; }
    th, td { border-bottom: 1px solid var(--notion-border); padding: 9px 10px; text-align: left; vertical-align: top; }
    th { background: var(--notion-bg); color: var(--notion-muted); font-size: 11px; text-transform: uppercase; letter-spacing: .05em; }
    tr:hover td { background: #fbfbfa; }
    .mono { font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace; }
    pre {
      white-space: pre-wrap;
      word-break: break-word;
      background: var(--notion-code);
      color: var(--notion-text);
      border: 1px solid var(--notion-border);
      border-radius: 8px;
      padding: 13px;
      max-height: 420px;
      overflow: auto;
      font-size: 12px;
      line-height: 1.55;
    }
    details.notion-details {
      border: 1px solid var(--notion-border);
      border-radius: 8px;
      background: #fff;
      margin: 10px 0;
      overflow: hidden;
    }
    details.notion-details > summary {
      cursor: pointer;
      padding: 12px 14px;
      font-weight: 680;
      background: #fff;
    }
    details.notion-details[open] > summary { border-bottom: 1px solid var(--notion-border); background: var(--notion-bg); }
    .details-body { padding: 14px; }
    .final-column { padding: 0; }
    .final-column > summary { list-style-position: inside; }
    .final-column .details-body { padding-top: 10px; }
    .coverage-list { padding-left: 18px; }
    .coverage-list li, .findings li { margin-bottom: 6px; }
    .detected { color: var(--notion-green); font-weight: 800; }
    .missed { color: var(--notion-red); font-weight: 800; }
    .muted { color: var(--notion-muted); }
    .hidden-long .long-content { display: none; }
    .capture .notion-toc, .capture .toolbar { display: none; }
    .capture .notion-shell { display: block; max-width: 1280px; }
    .capture .notion-document { padding: 24px; }
    .capture .notion-block { break-inside: avoid; }
    @media (max-width: 1200px) {
      .notion-shell { display: block; }
      .notion-toc { position: static; height: auto; border-right: 0; border-bottom: 1px solid var(--notion-border); display: flex; gap: 8px; overflow-x: auto; padding: 10px 14px; }
      .toc-title { display: none; }
      .toc-link { white-space: nowrap; }
      .notion-document { padding-top: 28px; }
    }
    @media (max-width: 900px) {
      .cards { grid-template-columns: repeat(2, minmax(0, 1fr)); }
      .group-cards, .charts, .two-col, .three-col { grid-template-columns: 1fr; }
    }
    @media (max-width: 640px) {
      .notion-document { padding: 20px 14px 42px; }
      .doc-icon { font-size: 42px; }
      h1 { font-size: 32px; }
      h2 { font-size: 22px; }
      .cards, .metric-list { grid-template-columns: 1fr; }
      .toolbar { flex-wrap: nowrap; }
      button { flex: 0 0 auto; }
      canvas { height: 230px; }
      pre { max-height: 320px; }
    }
    @media print {
      body.notion-page { background: #fff; }
      .notion-shell { display:block; }
      .notion-toc, .toolbar { display: none; }
      .notion-document { padding: 20px; }
      .notion-block, .card, .group-card, .chart-card, details.notion-details { box-shadow: none; break-inside: avoid; }
      pre { max-height: none; }
    }
  </style>
</head>
<body class="notion-page">
  <div class="notion-shell">
    <aside class="notion-toc" aria-label="Tabla de contenido">
      <p class="toc-title">Contenido</p>
      <a class="toc-link" href="#resumen">📌 Resumen</a>
      <a class="toc-link" href="#comparacion">🧪 Comparación G1/G2/G3</a>
      <a class="toc-link" href="#charts">📊 Cuadros comparativos</a>
      <a class="toc-link" href="#tesis-metrics">📐 Métricas de tesis</a>
      <a class="toc-link" href="#metodologia">💡 Lectura metodológica</a>
      <a class="toc-link" href="#respuestas">📝 Respuestas finales</a>
      <a class="toc-link" href="#anexos">📎 Anexos</a>
    </aside>
    <main class="notion-document">
      <header class="doc-header">
        <div class="doc-icon">🧠</div>
        <h1>Reporte experimental BFMA</h1>
        <p class="subtitle" id="subtitle">Cargando datos…</p>
        <div class="notion-meta" id="metadataBadges">
          <span class="notion-badge">⏳ Cargando metadata</span>
        </div>
        <div class="toolbar">
          <button class="primary" onclick="window.print()">Imprimir / Guardar PDF</button>
          <button onclick="copySummary()">Copiar resumen</button>
          <button onclick="document.body.classList.toggle('hidden-long')">Ocultar respuestas largas</button>
          <button onclick="document.body.classList.toggle('capture')">Modo captura para anexos</button>
        </div>
        <div class="notion-callout">
          <span class="emoji">📌</span>
          <p>Este reporte compara G1, G2 y G3 para observar cómo BFMA selecciona memoria relevante bajo presupuesto. Está diseñado para explicación al asesor y capturas para anexos.</p>
        </div>
        <div class="notion-callout">
          <span class="emoji">⚠️</span>
          <p><b>Reporte demo técnico:</b> la cobertura visual es exploratoria. Las métricas F1/SubEM/falsas memorias/contradicciones se calculan en una sección separada y sirven para validar instrumentación, no para contrastar todavía la hipótesis.</p>
        </div>
      </header>

      <section class="notion-block" id="resumen">
        <h2>Resumen ejecutivo</h2>
        <div class="grid cards" id="cards"></div>
      </section>

      <section class="notion-block" id="comparacion">
        <h2>Comparación experimental por grupo</h2>
        <p class="muted">Lectura rápida para asesor: G1 resume, G2 acumula memoria y G3 selecciona contexto bajo presupuesto.</p>
        <div class="grid group-cards" id="groupCards"></div>
      </section>

      <section class="notion-block" id="tesis-metrics">
        <h2>Métricas exploratorias alineadas a la matriz de consistencia</h2>
        <div class="notion-callout">
          <span class="emoji">📐</span>
          <div id="thesisConclusion"></div>
        </div>
        <div class="details-body table-wrap">
          <table id="thesisMetricsTable"></table>
        </div>
      </section>

      <section class="notion-block" id="charts">
        <h2>Cuadros comparativos G1/G2/G3</h2>
        <div class="grid charts">
          <div class="chart-card"><h3>Latencia promedio por grupo</h3><canvas id="groupLatencyChart"></canvas></div>
          <div class="chart-card"><h3>Tamaño de respuesta final</h3><canvas id="groupAnswerChart"></canvas></div>
          <div class="chart-card"><h3>Cobertura automática estimada</h3><canvas id="groupCoverageChart"></canvas></div>
          <div class="chart-card"><h3>Memoria/contexto final por grupo</h3><canvas id="groupContextChart"></canvas></div>
          <div class="chart-card"><h3>Presión de memoria: almacenado vs seleccionado</h3><canvas id="memoryPressureChart"></canvas></div>
          <div class="chart-card"><h3>Decisiones BFMA acumuladas</h3><canvas id="groupDecisionChart"></canvas></div>
        </div>
      </section>

      <section class="notion-block" id="metodologia">
        <h2>Lectura metodológica</h2>
        <div class="notion-callout">
          <span class="emoji">💡</span>
          <div id="methodology"></div>
        </div>
      </section>

      <section class="notion-block">
        <h2>Cuadros estadísticos por turno</h2>
        <div class="grid charts">
          <div class="chart-card"><h3>Latencia por turno</h3><canvas id="latencyChart"></canvas></div>
          <div class="chart-card"><h3>Crecimiento de memoria</h3><canvas id="memoryChart"></canvas></div>
          <div class="chart-card"><h3>Decisiones BFMA keep vs discard</h3><canvas id="decisionChart"></canvas></div>
          <div class="chart-card"><h3>Razones de descarte BFMA</h3><canvas id="reasonChart"></canvas></div>
        </div>
      </section>

      <section class="notion-block">
        <h2>Hallazgos observables</h2>
        <div class="notion-callout">
          <span class="emoji">📌</span>
          <ul class="findings" id="findings"></ul>
        </div>
      </section>

      <section class="notion-block">
        <details class="notion-details" open>
          <summary>Tabla por turno</summary>
          <div class="details-body table-wrap">
            <table id="turnTable"></table>
          </div>
        </details>
      </section>

      <section class="notion-block long-content" id="respuestas">
        <h2>Comparación de respuestas finales</h2>
        <div class="three-col" id="finalComparison"></div>
      </section>

      <section class="notion-block long-content" id="anexos">
        <h2>Material para anexo</h2>
        <details class="notion-details">
          <summary>Contexto final seleccionado</summary>
          <div class="details-body"><pre id="finalContext"></pre></div>
        </details>
        <details class="notion-details">
          <summary>Respuesta final del agente</summary>
          <div class="details-body"><pre id="finalAnswer"></pre></div>
        </details>
        <details class="notion-details" open>
          <summary>Preguntas finales esperadas</summary>
          <div class="details-body table-wrap"><table id="questionsTable"></table></div>
        </details>
      </section>
    </main>
  </div>

  <script>
    const REPORT_PAYLOAD = "{{.Payload}}";
    const bytes = Uint8Array.from(atob(REPORT_PAYLOAD), c => c.charCodeAt(0));
    const DATA = JSON.parse(new TextDecoder().decode(bytes));
    const COLORS = { primary: "#2f80ed", keep: "#448361", discard: "#d9730d", danger: "#d44c47", purple: "#6940a5", grid: "#e9e7e3", text: "#37352f" };

    function fmtMs(ms) {
      if (!ms) return "0 ms";
      if (ms >= 1000) return (ms / 1000).toFixed(1) + " s";
      return ms + " ms";
    }
    function esc(s) {
      return String(s ?? "").replace(/[&<>"']/g, ch => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#039;" }[ch]));
    }
    function card(label, value, hint) {
      return '<div class="card"><div class="label">'+esc(label)+'</div><div class="value">'+esc(value)+'</div><div class="hint">'+esc(hint || "")+'</div></div>';
    }
    function render() {
      document.getElementById("subtitle").textContent = DATA.run_id + " · " + DATA.scenario_id + " · grupos: " + DATA.groups.join(", ");
      document.getElementById("metadataBadges").innerHTML = [
        '<span class="notion-badge badge-blue">Run ' + esc(DATA.run_id) + '</span>',
        '<span class="notion-badge">Escenario ' + esc(DATA.scenario_id) + '</span>',
        '<span class="notion-badge badge-green">Grupos ' + esc(DATA.groups.join(", ")) + '</span>',
        '<span class="notion-badge">Generado ' + esc(DATA.generated_at) + '</span>'
      ].join("");
      document.getElementById("cards").innerHTML = [
        card("Run ID", DATA.run_id, DATA.generated_at),
        card("Turnos", DATA.total_turns, DATA.success_turns + " exitosos · " + DATA.failed_turns + " fallidos"),
        card("Latencia promedio", fmtMs(DATA.avg_latency_ms), "Promedio de turnos exitosos"),
        card("Memoria final", DATA.final_memory_count, "Memorias experimentales registradas"),
        card("Keep BFMA", DATA.total_keep, "Decisiones de conservación"),
        card("Discard BFMA", DATA.total_discard, "Decisiones de olvido/descarte"),
        card("Reintentos", DATA.retry_count, DATA.has_failures ? "Hubo fallos registrados" : "Sin fallos registrados"),
        card("Grupos", DATA.groups.join(", "), DATA.groups.length === 1 ? "Corrida parcial" : "Corrida comparativa")
      ].join("");
      document.getElementById("findings").innerHTML = DATA.findings.map(x => "<li>"+esc(x)+"</li>").join("");
      renderGroupComparison();
      renderThesisMetrics();
      renderMethodology();
      renderFinalComparison();
      renderTurnTable();
      renderQuestions();
      document.getElementById("finalContext").textContent = DATA.final_context || "Sin contexto final.";
      document.getElementById("finalAnswer").textContent = DATA.final_answer || "Sin respuesta final.";
      drawBar("groupLatencyChart", DATA.group_latency.map(x => ({ label: x.label, value: x.value })), COLORS.purple, v => fmtMs(v));
      drawBar("groupAnswerChart", DATA.group_answer_size.map(x => ({ label: x.label, value: x.value })), COLORS.primary, v => v + " palabras");
      drawBar("groupCoverageChart", DATA.group_coverage.map(x => ({ label: x.label, value: x.value })), COLORS.keep, v => v + "%");
      drawBar("groupContextChart", DATA.group_context_size.map(x => ({ label: x.label, value: x.value })), COLORS.primary, v => v);
      drawMemoryPressure("memoryPressureChart", DATA.memory_pressure || []);
      drawStacked("groupDecisionChart", (DATA.group_summaries || []).map(x => ({ label: x.group.toUpperCase(), keep: x.total_keep, discard: x.total_discard, group: x.group })));
      drawBar("latencyChart", DATA.latency_by_turn.map(x => ({ label: x.label, value: x.value })), COLORS.purple, v => fmtMs(v));
      drawLine("memoryChart", DATA.memory_by_turn.map(x => ({ label: x.label, value: x.value })), COLORS.primary);
      drawStacked("decisionChart", DATA.keep_discard_by_turn);
      drawBar("reasonChart", DATA.reason_distribution.map(x => ({ label: x.label, value: x.value })), COLORS.discard, v => v);
    }
    function groupBadge(g) {
      if (g === "g1") return '<span class="notion-badge badge-orange">G1 baseline</span>';
      if (g === "g2") return '<span class="notion-badge badge-blue">G2 memoria persistente</span>';
      if (g === "g3") return '<span class="notion-badge badge-green">G3 BFMA</span>';
      return '<span class="notion-badge">' + esc(g.toUpperCase()) + '</span>';
    }
    function renderGroupComparison() {
      const cards = (DATA.group_summaries || []).map(g => {
        const contextLabel = g.group === "g1" ? "líneas resumen" : "memorias contexto";
        const memoryLine = g.group === "g1"
          ? '<div class="metric"><span>Resumen final</span><b>' + g.final_summary_lines + '</b><small>líneas</small></div>'
          : '<div class="metric"><span>Memorias almacenadas</span><b>' + g.final_memory_count + '</b><small>store final</small></div>';
        const bfmaLine = g.group === "g3"
          ? '<div class="metric"><span>BFMA final</span><b>' + g.final_keep + '/' + g.final_discard + '</b><small>keep/discard</small></div>'
          : '<div class="metric"><span>Contexto final</span><b>' + g.final_context_items + '</b><small>' + contextLabel + '</small></div>';
        return '<article class="group-card">'
          + '<h3>' + esc(g.label) + '</h3>'
          + groupBadge(g.group)
          + '<p class="muted">' + esc(g.strategy) + '</p>'
          + '<div class="coverage-bar"><span style="width:' + g.coverage_percent + '%"></span></div>'
          + '<p><b>Cobertura automática estimada:</b> ' + g.coverage_detected + '/' + g.coverage_total + ' (' + g.coverage_percent + '%)</p>'
          + '<div class="metric-list">'
          + '<div class="metric"><span>Turnos</span><b>' + g.turns + '</b><small>' + g.success_turns + ' exitosos</small></div>'
          + '<div class="metric"><span>Latencia promedio</span><b>' + fmtMs(g.avg_latency_ms) + '</b><small>OpenCode real</small></div>'
          + memoryLine
          + bfmaLine
          + '<div class="metric"><span>Respuesta final</span><b>' + g.final_answer_words + '</b><small>palabras</small></div>'
          + '<div class="metric"><span>Budget</span><b>' + (g.token_budget ? (g.token_used + '/' + g.token_budget) : '—') + '</b><small>tokens contexto</small></div>'
          + '</div>'
          + '<p>' + esc(g.methodological_reading) + '</p>'
          + '</article>';
      }).join("");
      document.getElementById("groupCards").innerHTML = cards || '<p class="muted">No hay grupos para comparar.</p>';
    }
    function renderThesisMetrics() {
      const conclusion = DATA.evaluation_conclusion || [];
      const warnings = DATA.evaluation_warnings || [];
      const groups = DATA.evaluation_groups || [];
      const intro = DATA.evaluation_available
        ? conclusion.map(x => '<li>' + esc(x) + '</li>').join('')
        : '<li>No se pudo calcular la evaluación formal desde los logs y el escenario.</li>';
      const warningList = warnings.length
        ? '<h4>Advertencias</h4><ul class="findings">' + warnings.map(x => '<li>' + esc(x) + '</li>').join('') + '</ul>'
        : '';
      document.getElementById("thesisConclusion").innerHTML = '<ul class="findings">' + intro + '</ul>' + warningList;
      const rows = groups.map(g => '<tr>'
        + '<td class="mono">' + esc(g.architecture) + '</td>'
        + '<td>' + Number(g.avg_f1 || 0).toFixed(4) + '</td>'
        + '<td>' + Number(g.subem_hits || 0) + '/' + Number(g.subem_total || 0) + ' (' + Number(g.subem_rate || 0).toFixed(4) + ')</td>'
        + '<td>' + Number(g.false_memory_hits || 0) + '/' + Number(g.false_memory_total || 0) + ' (' + Number(g.false_memory_rate || 0).toFixed(4) + ')</td>'
        + '<td>' + Number(g.contradiction_hits || 0) + '/' + Number(g.contradiction_total || 0) + ' (' + Number(g.contradiction_rate || 0).toFixed(4) + ')</td>'
        + '<td>' + Number(g.avg_memory_tokens || 0).toFixed(1) + '</td>'
        + '<td>' + Number(g.avg_storage_items || 0).toFixed(1) + '</td>'
        + '<td>' + fmtMs(Math.round(g.avg_latency_ms || 0)) + '</td>'
        + '<td>' + esc(g.interpretation || '') + '</td>'
        + '</tr>').join('');
      document.getElementById("thesisMetricsTable").innerHTML = '<thead><tr><th>Arquitectura</th><th>F1</th><th>SubEM entidades</th><th>Falsas memorias detectadas</th><th>Contradicciones detectadas</th><th>Tokens memoria</th><th>Almacenamiento</th><th>Latencia</th><th>Lectura</th></tr></thead><tbody>'
        + (rows || '<tr><td colspan="9">Sin métricas exploratorias disponibles.</td></tr>')
        + '</tbody>';
    }
    function renderMethodology() {
      const hasAll = DATA.groups.includes("g1") && DATA.groups.includes("g2") && DATA.groups.includes("g3");
      const lines = [
        "G1 representa el baseline con resumen incremental: la memoria se comprime y puede sufrir summary drift.",
        "G2 representa memoria persistente: conserva más información, pero puede arrastrar ruido u obsolescencia.",
        "G3 representa BFMA: selecciona el contexto por utilidad y presupuesto, por eso no compite por recordar más sino por recordar lo relevante.",
        "El olvido presupuestado observado en este demo significa exclusión del contexto activo, no borrado permanente del almacén de memoria.",
        hasAll ? "Este run contiene G1, G2 y G3; por lo tanto sirve como evidencia comparativa visual." : "Este run no contiene todos los grupos; usarlo como evidencia parcial."
      ];
      document.getElementById("methodology").innerHTML = '<ul class="findings">' + lines.map(x => '<li>' + esc(x) + '</li>').join("") + '</ul>';
    }
    function renderFinalComparison() {
      const columns = (DATA.final_by_group || []).map(g => {
        const coverage = (g.coverage || []).map(c => '<li><span class="' + (c.detected ? 'detected' : 'missed') + '">' + (c.detected ? '✓' : '✗') + '</span> ' + esc(c.id) + ': ' + esc(c.expected_answer) + '</li>').join("");
        const omissions = (g.omissions || []).length
          ? '<p><b>Posibles omisiones automáticas:</b> ' + g.omissions.map(x => esc(x.id)).join(", ") + '</p>'
          : '<p><b>Posibles omisiones automáticas:</b> ninguna detectada por coincidencia simple.</p>';
        return '<details class="notion-details final-column" open>'
          + '<summary>' + esc(g.label) + '</summary>'
          + '<div class="details-body">'
          + '<p class="muted">Contexto final: ' + g.final_context_items + ' ítems · respuesta: ' + g.final_answer_chars + ' caracteres</p>'
          + '<h4>Criterios esperados detectados</h4><ul class="coverage-list">' + coverage + '</ul>'
          + omissions
          + '<h4>Contexto final inyectado</h4><pre>' + esc(g.final_context || 'Sin contexto') + '</pre>'
          + '<h4>Respuesta final</h4><pre>' + esc(g.final_answer || 'Sin respuesta') + '</pre>'
          + '</div></details>';
      }).join("");
      document.getElementById("finalComparison").innerHTML = columns || '<p class="muted">No hay respuestas finales por grupo.</p>';
    }
    function renderTurnTable() {
      const rows = DATA.turns.map(t => {
        const budget = t.token_budget ? (t.token_used + "/" + t.token_budget) : "—";
        const statusClass = t.status === "failed" ? "failed" : "success";
        return '<tr>'
          + '<td class="mono">' + esc(t.group) + ' T' + t.turn + '</td>'
          + '<td><span class="pill ' + statusClass + '">' + esc(t.status) + '</span></td>'
          + '<td>' + (t.attempt || 0) + '/' + (t.max_attempts || 0) + '</td>'
          + '<td>' + fmtMs(t.latency_ms) + '</td>'
          + '<td>' + t.memory_count + '</td>'
          + '<td style="color:' + COLORS.keep + ';font-weight:800">' + t.keep + '</td>'
          + '<td style="color:' + COLORS.discard + ';font-weight:800">' + t.discard + '</td>'
          + '<td>' + budget + '</td>'
          + '<td>' + esc(t.prompt_short) + '</td>'
          + '<td>' + esc(t.error || t.answer_short) + '</td>'
          + '</tr>';
      }).join("");
      document.getElementById("turnTable").innerHTML =
        '<thead><tr><th>Turno</th><th>Estado</th><th>Intento</th><th>Latencia</th><th>Memoria</th><th>Keep</th><th>Discard</th><th>Budget</th><th>Prompt</th><th>Respuesta/Error</th></tr></thead>'
        + '<tbody>' + rows + '</tbody>';
    }
    function renderQuestions() {
      const rows = (DATA.final_questions || []).map(q =>
        '<tr><td class="mono">' + esc(q.id) + '</td><td>' + esc(q.question) + '</td><td>' + esc(q.expected_answer) + '</td></tr>'
      ).join("");
      document.getElementById("questionsTable").innerHTML =
        '<thead><tr><th>ID</th><th>Pregunta</th><th>Respuesta esperada</th></tr></thead><tbody>'
        + (rows || '<tr><td colspan="3">Sin preguntas finales.</td></tr>')
        + '</tbody>';
    }
    function setupCanvas(id) {
      const canvas = document.getElementById(id);
      const ratio = window.devicePixelRatio || 1;
      const rect = canvas.getBoundingClientRect();
      const cssWidth = Math.max(280, rect.width || 320);
      canvas.width = Math.floor(cssWidth * ratio);
      canvas.height = Math.floor(260 * ratio);
      const ctx = canvas.getContext("2d");
      ctx.setTransform(ratio, 0, 0, ratio, 0, 0);
      return { canvas, ctx, w: cssWidth, h: 260 };
    }
    function axes(ctx, w, h) {
      ctx.strokeStyle = COLORS.grid; ctx.lineWidth = 1;
      ctx.beginPath(); ctx.moveTo(38, 16); ctx.lineTo(38, h-40); ctx.lineTo(w-10, h-40); ctx.stroke();
    }
    function drawBar(id, points, color, formatter) {
      const {ctx,w,h} = setupCanvas(id); ctx.clearRect(0,0,w,h); axes(ctx,w,h);
      if (!points.length) { empty(ctx,w,h); return; }
      const max = Math.max(...points.map(p => p.value), 1);
      const gap = Math.max(4, Math.min(8, 80 / points.length));
      const bw = Math.max(6, (w-58) / points.length - gap);
      points.forEach((p,i) => {
        const x = 46 + i*(bw+gap);
        const bh = (h-68) * (p.value/max);
        const y = h-40-bh;
        ctx.fillStyle = color; ctx.fillRect(x,y,bw,bh);
        ctx.fillStyle = COLORS.text; ctx.font = "10px sans-serif";
        ctx.save(); ctx.translate(x+2,h-26); ctx.rotate(-Math.PI/4); ctx.fillText(p.label,0,0); ctx.restore();
        if (points.length <= 8) ctx.fillText(String(formatter ? formatter(p.value) : p.value), x, y-4);
      });
    }
    function drawLine(id, points, color) {
      const {ctx,w,h} = setupCanvas(id); ctx.clearRect(0,0,w,h); axes(ctx,w,h);
      if (!points.length) { empty(ctx,w,h); return; }
      const max = Math.max(...points.map(p => p.value), 1);
      const step = (w-58) / Math.max(points.length-1, 1);
      ctx.strokeStyle = color; ctx.lineWidth = 2.5; ctx.beginPath();
      points.forEach((p,i) => {
        const x = 44 + i*step; const y = h-40 - ((h-68)*(p.value/max));
        if (i===0) ctx.moveTo(x,y); else ctx.lineTo(x,y);
      });
      ctx.stroke();
      points.forEach((p,i) => {
        const x = 44 + i*step; const y = h-40 - ((h-68)*(p.value/max));
        ctx.fillStyle = color; ctx.beginPath(); ctx.arc(x,y,3.5,0,Math.PI*2); ctx.fill();
      });
    }
    function drawStacked(id, points) {
      const {ctx,w,h} = setupCanvas(id); ctx.clearRect(0,0,w,h); axes(ctx,w,h);
      if (!points.length) { empty(ctx,w,h); return; }
      const max = Math.max(...points.map(p => p.keep + p.discard), 1);
      const gap = Math.max(4, Math.min(8, 80 / points.length));
      const bw = Math.max(6, (w-58) / points.length - gap);
      points.forEach((p,i) => {
        const x = 46 + i*(bw+gap);
        const keepH = (h-68) * (p.keep/max);
        const discardH = (h-68) * (p.discard/max);
        ctx.fillStyle = COLORS.keep; ctx.fillRect(x,h-40-keepH,bw,keepH);
        ctx.fillStyle = COLORS.discard; ctx.fillRect(x,h-40-keepH-discardH,bw,discardH);
        ctx.fillStyle = COLORS.text; ctx.font = "10px sans-serif"; ctx.save(); ctx.translate(x+2,h-26); ctx.rotate(-Math.PI/4); ctx.fillText(p.label,0,0); ctx.restore();
      });
    }
    function drawMemoryPressure(id, points) {
      const {ctx,w,h} = setupCanvas(id); ctx.clearRect(0,0,w,h); axes(ctx,w,h);
      if (!points.length) { empty(ctx,w,h); return; }
      const max = Math.max(...points.map(p => Math.max(p.stored || 0, p.selected || 0)), 1);
      const gap = 14; const groupW = Math.max(42, (w-64) / points.length - gap);
      points.forEach((p,i) => {
        const x = 46 + i*(groupW+gap);
        const bw = Math.max(10, groupW/3);
        const storedH = (h-68) * ((p.stored || 0)/max);
        const selectedH = (h-68) * ((p.selected || 0)/max);
        ctx.fillStyle = COLORS.primary; ctx.fillRect(x,h-40-storedH,bw,storedH);
        ctx.fillStyle = COLORS.keep; ctx.fillRect(x+bw+4,h-40-selectedH,bw,selectedH);
        ctx.fillStyle = COLORS.text; ctx.font = "10px sans-serif";
        ctx.fillText(p.label, x, h-24);
        if (points.length <= 6) {
          ctx.fillText(String(p.stored || 0), x, h-44-storedH);
          ctx.fillText(String(p.selected || 0), x+bw+4, h-44-selectedH);
        }
      });
      ctx.fillStyle = COLORS.primary; ctx.fillRect(w-142, 18, 9, 9); ctx.fillStyle = COLORS.text; ctx.fillText("almacenado", w-128, 27);
      ctx.fillStyle = COLORS.keep; ctx.fillRect(w-142, 34, 9, 9); ctx.fillStyle = COLORS.text; ctx.fillText("seleccionado", w-128, 43);
    }
    function empty(ctx,w,h) {
      ctx.fillStyle = "#787774"; ctx.font = "14px sans-serif"; ctx.fillText("Sin datos para este gráfico", 44, h/2);
    }
    function copySummary() {
      const text = [
        "Reporte BFMA " + DATA.run_id,
        "Escenario: " + DATA.scenario_id,
        "Turnos: " + DATA.total_turns,
        "Keep: " + DATA.total_keep + " · Discard: " + DATA.total_discard,
        "Memoria final: " + DATA.final_memory_count,
        ...DATA.findings
      ].join("\n");
      navigator.clipboard.writeText(text).then(() => alert("Resumen copiado"));
    }
    window.addEventListener("resize", () => {
      drawBar("groupLatencyChart", DATA.group_latency.map(x => ({ label: x.label, value: x.value })), COLORS.purple, v => fmtMs(v));
      drawBar("groupAnswerChart", DATA.group_answer_size.map(x => ({ label: x.label, value: x.value })), COLORS.primary, v => v + " palabras");
      drawBar("groupCoverageChart", DATA.group_coverage.map(x => ({ label: x.label, value: x.value })), COLORS.keep, v => v + "%");
      drawBar("groupContextChart", DATA.group_context_size.map(x => ({ label: x.label, value: x.value })), COLORS.primary, v => v);
      drawMemoryPressure("memoryPressureChart", DATA.memory_pressure || []);
      drawStacked("groupDecisionChart", (DATA.group_summaries || []).map(x => ({ label: x.group.toUpperCase(), keep: x.total_keep, discard: x.total_discard, group: x.group })));
      drawBar("latencyChart", DATA.latency_by_turn.map(x => ({ label: x.label, value: x.value })), COLORS.purple, v => fmtMs(v));
      drawLine("memoryChart", DATA.memory_by_turn.map(x => ({ label: x.label, value: x.value })), COLORS.primary);
      drawStacked("decisionChart", DATA.keep_discard_by_turn);
      drawBar("reasonChart", DATA.reason_distribution.map(x => ({ label: x.label, value: x.value })), COLORS.discard, v => v);
    });
    render();
  </script>
</body>
</html>`
