package report

const htmlTemplate = `<!doctype html>
<html lang="es">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>{{.Title}}</title>
  <style>
    :root {
      --bg: #f6f8fb;
      --panel: #ffffff;
      --text: #172033;
      --muted: #667085;
      --border: #d9e2ec;
      --primary: #2454d6;
      --primary-soft: #e8efff;
      --keep: #0f766e;
      --discard: #dc6803;
      --danger: #c2410c;
      --purple: #7c3aed;
      --shadow: 0 12px 30px rgba(15, 23, 42, .08);
    }
    * { box-sizing: border-box; }
    body {
      margin: 0;
      font-family: Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
      background: radial-gradient(circle at top left, #eef4ff 0, #f6f8fb 34rem);
      color: var(--text);
    }
    header {
      padding: 34px 36px 22px;
      border-bottom: 1px solid var(--border);
      background: rgba(255,255,255,.78);
      backdrop-filter: blur(10px);
      position: sticky;
      top: 0;
      z-index: 10;
    }
    h1 { margin: 0 0 8px; font-size: 30px; letter-spacing: -.03em; }
    h2 { margin: 0 0 16px; font-size: 20px; letter-spacing: -.02em; }
    h3 { margin: 18px 0 10px; font-size: 16px; }
    .subtitle { color: var(--muted); margin: 0; }
    .toolbar { display: flex; flex-wrap: wrap; gap: 10px; margin-top: 18px; }
    button {
      border: 1px solid var(--border);
      background: var(--panel);
      color: var(--text);
      border-radius: 10px;
      padding: 10px 14px;
      cursor: pointer;
      font-weight: 650;
      box-shadow: 0 2px 8px rgba(15, 23, 42, .04);
    }
    button.primary { background: var(--primary); border-color: var(--primary); color: white; }
    main { padding: 28px 36px 48px; max-width: 1440px; margin: 0 auto; }
    section {
      background: rgba(255,255,255,.92);
      border: 1px solid var(--border);
      border-radius: 18px;
      box-shadow: var(--shadow);
      padding: 22px;
      margin-bottom: 22px;
    }
    .grid { display: grid; gap: 16px; }
    .cards { grid-template-columns: repeat(4, minmax(0, 1fr)); }
    .group-cards { grid-template-columns: repeat(3, minmax(0, 1fr)); }
    .charts { grid-template-columns: repeat(2, minmax(0, 1fr)); }
    .three-col { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 16px; }
    .card {
      background: linear-gradient(180deg, #fff, #f9fbff);
      border: 1px solid var(--border);
      border-radius: 16px;
      padding: 16px;
      min-height: 112px;
    }
    .card .label { color: var(--muted); font-size: 12px; text-transform: uppercase; letter-spacing: .08em; font-weight: 800; }
    .card .value { font-size: 30px; font-weight: 850; margin-top: 8px; letter-spacing: -.03em; }
    .card .hint { color: var(--muted); font-size: 13px; margin-top: 6px; }
    .chart-card {
      border: 1px solid var(--border);
      border-radius: 16px;
      padding: 16px;
      background: #fff;
    }
    .group-card {
      border: 1px solid var(--border);
      border-radius: 18px;
      background: #fff;
      padding: 18px;
      box-shadow: 0 8px 22px rgba(15, 23, 42, .06);
    }
    .group-card h3 { margin-top: 0; }
    .metric-list { display: grid; grid-template-columns: 1fr 1fr; gap: 10px; margin: 14px 0; }
    .metric { background: #f8fbff; border: 1px solid var(--border); border-radius: 12px; padding: 10px; }
    .metric b { display:block; font-size: 18px; margin-top: 4px; }
    .coverage-bar { height: 10px; background: #e5e7eb; border-radius: 999px; overflow:hidden; margin: 10px 0; }
    .coverage-bar span { display:block; height:100%; background: linear-gradient(90deg, var(--primary), var(--keep)); }
    .final-column {
      border: 1px solid var(--border);
      border-radius: 16px;
      padding: 14px;
      background: #fff;
      min-width: 0;
    }
    .final-column pre { max-height: 360px; }
    .coverage-list { padding-left: 18px; }
    .coverage-list li { margin-bottom: 6px; }
    .detected { color: var(--keep); font-weight: 800; }
    .missed { color: var(--danger); font-weight: 800; }
    canvas { width: 100%; height: 280px; display: block; }
    table {
      width: 100%;
      border-collapse: collapse;
      font-size: 13px;
      overflow: hidden;
      border-radius: 12px;
    }
    th, td { border-bottom: 1px solid var(--border); padding: 10px; text-align: left; vertical-align: top; }
    th { background: #f1f5fb; color: #334155; font-size: 12px; text-transform: uppercase; letter-spacing: .06em; }
    tr:hover td { background: #f8fbff; }
    .pill {
      display: inline-flex;
      align-items: center;
      border-radius: 999px;
      padding: 4px 9px;
      font-weight: 750;
      font-size: 12px;
      border: 1px solid transparent;
    }
    .success { color: #047857; background: #ecfdf3; border-color: #abefc6; }
    .failed { color: var(--danger); background: #fff4ed; border-color: #fed7aa; }
    .mono { font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace; }
    .two-col { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; }
    pre {
      white-space: pre-wrap;
      word-break: break-word;
      background: #0f172a;
      color: #dbeafe;
      border-radius: 14px;
      padding: 16px;
      max-height: 520px;
      overflow: auto;
      font-size: 12px;
      line-height: 1.55;
    }
    .findings li { margin-bottom: 8px; }
    .muted { color: var(--muted); }
    .hidden-long .long-content { display: none; }
    .capture header { position: static; }
    .capture section { box-shadow: none; break-inside: avoid; }
    .capture main { max-width: 1280px; }
    @media (max-width: 1000px) {
      .cards, .group-cards, .charts, .two-col, .three-col { grid-template-columns: 1fr; }
      header, main { padding-left: 18px; padding-right: 18px; }
    }
    @media print {
      body { background: #fff; }
      header { position: static; }
      .toolbar { display: none; }
      section { box-shadow: none; page-break-inside: avoid; }
      pre { max-height: none; }
    }
  </style>
</head>
<body>
  <header>
    <h1>Reporte experimental BFMA</h1>
    <p class="subtitle" id="subtitle">Cargando datos…</p>
    <div class="toolbar">
      <button class="primary" onclick="window.print()">Imprimir / Guardar PDF</button>
      <button onclick="copySummary()">Copiar resumen</button>
      <button onclick="document.body.classList.toggle('hidden-long')">Ocultar respuestas largas</button>
      <button onclick="document.body.classList.toggle('capture')">Modo captura para anexos</button>
    </div>
  </header>
  <main>
    <section>
      <h2>Resumen ejecutivo</h2>
      <div class="grid cards" id="cards"></div>
    </section>

    <section>
      <h2>Comparación experimental por grupo</h2>
      <p class="muted">Lectura rápida para asesor: G1 resume, G2 acumula memoria y G3 selecciona contexto bajo presupuesto.</p>
      <div class="grid group-cards" id="groupCards"></div>
    </section>

    <section>
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

    <section>
      <h2>Lectura metodológica</h2>
      <div id="methodology"></div>
    </section>

    <section>
      <h2>Cuadros estadísticos</h2>
      <div class="grid charts">
        <div class="chart-card"><h3>Latencia por turno</h3><canvas id="latencyChart"></canvas></div>
        <div class="chart-card"><h3>Crecimiento de memoria</h3><canvas id="memoryChart"></canvas></div>
        <div class="chart-card"><h3>Decisiones BFMA keep vs discard</h3><canvas id="decisionChart"></canvas></div>
        <div class="chart-card"><h3>Razones de descarte BFMA</h3><canvas id="reasonChart"></canvas></div>
      </div>
    </section>

    <section>
      <h2>Hallazgos observables</h2>
      <ul class="findings" id="findings"></ul>
    </section>

    <section>
      <h2>Tabla por turno</h2>
      <div style="overflow:auto">
        <table id="turnTable"></table>
      </div>
    </section>

    <section class="long-content">
      <h2>Comparación de respuestas finales</h2>
      <div class="three-col" id="finalComparison"></div>
    </section>

    <section class="long-content">
      <h2>Material para anexo</h2>
      <div class="two-col">
        <div>
          <h3>Contexto final seleccionado</h3>
          <pre id="finalContext"></pre>
        </div>
        <div>
          <h3>Respuesta final del agente</h3>
          <pre id="finalAnswer"></pre>
        </div>
      </div>
      <h3>Preguntas finales esperadas</h3>
      <div style="overflow:auto">
        <table id="questionsTable"></table>
      </div>
    </section>
  </main>

  <script>
    const REPORT_PAYLOAD = "{{.Payload}}";
    const bytes = Uint8Array.from(atob(REPORT_PAYLOAD), c => c.charCodeAt(0));
    const DATA = JSON.parse(new TextDecoder().decode(bytes));
    const COLORS = { primary: "#2454d6", keep: "#0f766e", discard: "#dc6803", danger: "#c2410c", purple: "#7c3aed", grid: "#d9e2ec", text: "#172033" };

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
    function renderGroupComparison() {
      const cards = (DATA.group_summaries || []).map(g => {
        const contextLabel = g.group === "g1" ? "líneas resumen" : "memorias contexto";
        const memoryLine = g.group === "g1"
          ? '<div class="metric"><span>Resumen final</span><b>' + g.final_summary_lines + '</b><small>líneas</small></div>'
          : '<div class="metric"><span>Memorias almacenadas</span><b>' + g.final_memory_count + '</b><small>store final</small></div>';
        const bfmaLine = g.group === "g3"
          ? '<div class="metric"><span>BFMA final</span><b>' + g.final_keep + '/' + g.final_discard + '</b><small>keep/discard</small></div>'
          : '<div class="metric"><span>Contexto final</span><b>' + g.final_context_items + '</b><small>' + contextLabel + '</small></div>';
        return '<div class="group-card">'
          + '<h3>' + esc(g.label) + '</h3>'
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
          + '</div>';
      }).join("");
      document.getElementById("groupCards").innerHTML = cards || '<p class="muted">No hay grupos para comparar.</p>';
    }
    function renderMethodology() {
      const hasAll = DATA.groups.includes("g1") && DATA.groups.includes("g2") && DATA.groups.includes("g3");
      const lines = [
        "G1 representa el baseline con resumen incremental: la memoria se comprime y puede sufrir summary drift.",
        "G2 representa memoria persistente: conserva más información, pero puede arrastrar ruido u obsolescencia.",
        "G3 representa BFMA: selecciona el contexto por utilidad y presupuesto, por eso no compite por recordar más sino por recordar lo relevante.",
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
        return '<div class="final-column">'
          + '<h3>' + esc(g.label) + '</h3>'
          + '<p class="muted">Contexto final: ' + g.final_context_items + ' ítems · respuesta: ' + g.final_answer_chars + ' caracteres</p>'
          + '<h4>Criterios esperados detectados</h4><ul class="coverage-list">' + coverage + '</ul>'
          + omissions
          + '<h4>Contexto final inyectado</h4><pre>' + esc(g.final_context || 'Sin contexto') + '</pre>'
          + '<h4>Respuesta final</h4><pre>' + esc(g.final_answer || 'Sin respuesta') + '</pre>'
          + '</div>';
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
      canvas.width = Math.max(640, Math.floor(rect.width * ratio));
      canvas.height = Math.floor(280 * ratio);
      const ctx = canvas.getContext("2d");
      ctx.scale(ratio, ratio);
      return { canvas, ctx, w: canvas.width / ratio, h: canvas.height / ratio };
    }
    function axes(ctx, w, h) {
      ctx.strokeStyle = COLORS.grid; ctx.lineWidth = 1;
      ctx.beginPath(); ctx.moveTo(42, 18); ctx.lineTo(42, h-44); ctx.lineTo(w-12, h-44); ctx.stroke();
    }
    function drawBar(id, points, color, formatter) {
      const {ctx,w,h} = setupCanvas(id); ctx.clearRect(0,0,w,h); axes(ctx,w,h);
      if (!points.length) { empty(ctx,w,h); return; }
      const max = Math.max(...points.map(p => p.value), 1);
      const gap = 6; const bw = Math.max(8, (w-70) / points.length - gap);
      points.forEach((p,i) => {
        const x = 50 + i*(bw+gap);
        const bh = (h-74) * (p.value/max);
        const y = h-44-bh;
        ctx.fillStyle = color; ctx.fillRect(x,y,bw,bh);
        ctx.fillStyle = COLORS.text; ctx.font = "10px sans-serif"; ctx.save(); ctx.translate(x+2,h-30); ctx.rotate(-Math.PI/4); ctx.fillText(p.label,0,0); ctx.restore();
        if (points.length <= 12) ctx.fillText(String(formatter ? formatter(p.value) : p.value), x, y-4);
      });
    }
    function drawLine(id, points, color) {
      const {ctx,w,h} = setupCanvas(id); ctx.clearRect(0,0,w,h); axes(ctx,w,h);
      if (!points.length) { empty(ctx,w,h); return; }
      const max = Math.max(...points.map(p => p.value), 1);
      const step = (w-72) / Math.max(points.length-1, 1);
      ctx.strokeStyle = color; ctx.lineWidth = 3; ctx.beginPath();
      points.forEach((p,i) => {
        const x = 48 + i*step; const y = h-44 - ((h-74)*(p.value/max));
        if (i===0) ctx.moveTo(x,y); else ctx.lineTo(x,y);
      });
      ctx.stroke();
      points.forEach((p,i) => {
        const x = 48 + i*step; const y = h-44 - ((h-74)*(p.value/max));
        ctx.fillStyle = color; ctx.beginPath(); ctx.arc(x,y,4,0,Math.PI*2); ctx.fill();
      });
    }
    function drawStacked(id, points) {
      const {ctx,w,h} = setupCanvas(id); ctx.clearRect(0,0,w,h); axes(ctx,w,h);
      if (!points.length) { empty(ctx,w,h); return; }
      const max = Math.max(...points.map(p => p.keep + p.discard), 1);
      const gap = 6; const bw = Math.max(8, (w-70) / points.length - gap);
      points.forEach((p,i) => {
        const x = 50 + i*(bw+gap);
        const keepH = (h-74) * (p.keep/max);
        const discardH = (h-74) * (p.discard/max);
        ctx.fillStyle = COLORS.keep; ctx.fillRect(x,h-44-keepH,bw,keepH);
        ctx.fillStyle = COLORS.discard; ctx.fillRect(x,h-44-keepH-discardH,bw,discardH);
        ctx.fillStyle = COLORS.text; ctx.font = "10px sans-serif"; ctx.save(); ctx.translate(x+2,h-30); ctx.rotate(-Math.PI/4); ctx.fillText(p.label,0,0); ctx.restore();
      });
    }
    function drawMemoryPressure(id, points) {
      const {ctx,w,h} = setupCanvas(id); ctx.clearRect(0,0,w,h); axes(ctx,w,h);
      if (!points.length) { empty(ctx,w,h); return; }
      const max = Math.max(...points.map(p => Math.max(p.stored || 0, p.selected || 0)), 1);
      const gap = 18; const groupW = Math.max(50, (w-78) / points.length - gap);
      points.forEach((p,i) => {
        const x = 52 + i*(groupW+gap);
        const bw = Math.max(12, groupW/3);
        const storedH = (h-74) * ((p.stored || 0)/max);
        const selectedH = (h-74) * ((p.selected || 0)/max);
        ctx.fillStyle = COLORS.primary; ctx.fillRect(x,h-44-storedH,bw,storedH);
        ctx.fillStyle = COLORS.keep; ctx.fillRect(x+bw+4,h-44-selectedH,bw,selectedH);
        ctx.fillStyle = COLORS.text; ctx.font = "10px sans-serif";
        ctx.fillText(p.label, x, h-28);
        if (points.length <= 6) {
          ctx.fillText(String(p.stored || 0), x, h-48-storedH);
          ctx.fillText(String(p.selected || 0), x+bw+4, h-48-selectedH);
        }
      });
      ctx.fillStyle = COLORS.primary; ctx.fillRect(w-160, 20, 10, 10); ctx.fillStyle = COLORS.text; ctx.fillText("almacenado", w-145, 30);
      ctx.fillStyle = COLORS.keep; ctx.fillRect(w-160, 38, 10, 10); ctx.fillStyle = COLORS.text; ctx.fillText("seleccionado", w-145, 48);
    }
    function empty(ctx,w,h) {
      ctx.fillStyle = "#667085"; ctx.font = "14px sans-serif"; ctx.fillText("Sin datos para este gráfico", 52, h/2);
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
