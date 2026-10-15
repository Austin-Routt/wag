// WAG dashboard: the Models tab (connect a model service, find a model, add
// it as a player) and the Annotate tab (models annotate claims, a model
// reviews them, you give verdicts). It uses app.js's helpers ($, esc, api,
// toast, rich, short) and the server's /api/llm, /api/annotate and
// /api/store. Keys go to the server once and never come back.
'use strict';

const WB = {
  services: [], players: [], store: '', svc: 'gemini', models: {}, filter: '', chosen: '', test: null, busy: false,
  datasets: [], runs: [], run: null, status: null, report: null, item: null, view: null, reveal: false, timer: null,
  form: { dataset: '', annotators: [], reviewer: '', count: 5 },
};

// The game bar's "Saved" button opens the game's folder.
document.addEventListener('click', async e => {
  const b = e.target.closest('#gamebar [data-open]'); if (!b) return;
  try { await api('/api/store/open', { path: b.dataset.open }); } catch (err) { toast(err.message, 6000); }
});

async function wbLoad() {
  const [s, p] = await Promise.all([api('/api/llm/services'), api('/api/llm/players')]);
  WB.services = s.services || []; WB.store = s.store; WB.players = p.players || [];
}

// Refresh what the Play tab's seat list offers.
async function wbMeta() {
  META = await api('/api/meta');
}

const svcOf = id => WB.services.find(s => s.id === id) || {};
const readyPlayers = () => WB.players.filter(p => p.ready);
// what a player runs on: its model, or its service if it names none (a command)
const runsOn = p => p.model || p.provider;

// The data folder line, with a button that opens it.
function dataFolder(rel, label) {
  if (!WB.store) return '<p class="small warn">This server has no data folder: start it with --store.</p>';
  const sep = WB.store.includes('\\') ? '\\' : '/';
  const path = rel ? WB.store + sep + rel.split('/').join(sep) : WB.store;
  return `<div class="folder"><span class="small muted">${esc(label || 'Saved in')}</span> <span class="mono">${esc(path)}</span>
    <button class="small" data-open="${esc(rel || '')}">Open folder</button></div>`;
}

// ---------------------------------------------------------------------------
// Models
// ---------------------------------------------------------------------------
async function renderModels() {
  const main = $('#main');
  if (!WB.services.length) {
    main.innerHTML = '<div class="setup"><p class="muted">Loading…</p></div>';
    try { await wbLoad(); } catch (e) { main.innerHTML = `<div class="setup"><p class="warn">${esc(e.message)}</p></div>`; return; }
  }
  const sv = svcOf(WB.svc);
  const keyState = s => s.key_source ? `<span class="pill on">key ✓</span>` : s.needs_key ? '<span class="pill">needs a key</span>' : '<span class="pill">no key needed</span>';
  const list = WB.services.map(s => `<button data-svc="${esc(s.id)}" class="${s.id === WB.svc ? 'on' : ''}">${esc(s.name)} ${keyState(s)}</button>`).join('');
  const models = WB.models[WB.svc];
  const f = WB.filter.toLowerCase();
  const shown = (models || []).filter(m => !f || (m.id + ' ' + (m.display_name || '')).toLowerCase().includes(f));
  const effortHelp = { gemini: 'Thinking level: minimal, low, medium, high (models that think)', anthropic: 'Effort: low, medium, high, xhigh, max', openai: 'Reasoning effort: low, medium, high (reasoning models)' }[WB.svc] || 'Reasoning effort, if the model takes one';
  main.innerHTML = `<div class="setup workbench">
    <div class="panel"><h2>Models</h2>
      <p class="small muted">Connect a model service, find a model, and add it as a <b>player</b>. A player can take a seat in a game (Play → New game: against you, a rule bot or another model) or annotate claims (Annotate).
      Keys stay on this computer, with this server: in memory until it stops, or, if you tick <i>Remember</i>, in <span class="mono">keys.json</span> in your data folder (not encrypted). A key is never shown again, and never written into a game, a trace or a run.</p>
      ${dataFolder('', 'Your data folder:')}
    </div>
    <div class="grid2">
      <div class="panel"><h3>1. Service</h3><div class="list">${list}</div></div>
      <div class="panel">
        <h3>2. Connect ${esc(sv.name || '')}</h3>
        <p class="small muted">${esc(sv.about || '')}</p>
        ${sv.note ? `<p class="small note">${esc(sv.note)}</p>` : ''}
        ${sv.key_source ? `<p class="small">Key: <b>${esc(sv.key_source)}</b> ${esc(sv.key_hint || '')} ${sv.key_source.startsWith('environment') ? '' : '<button class="small" id="forget">Forget it</button>'}</p>` : ''}
        ${sv.needs_key || !sv.local ? `<div class="row keyrow"><input type="password" id="key" placeholder="${sv.key_source ? 'Replace the key' : 'Paste your API key'}" autocomplete="off">
          <label class="inline small"><input type="checkbox" id="remember"> Remember on this computer</label>
          <button class="primary" id="savekey">Save key</button></div>
          ${sv.get_key ? `<p class="small muted">Get a key: <a href="${esc(sv.get_key)}" target="_blank" rel="noopener">${esc(sv.get_key)}</a></p>` : ''}` : ''}
        ${sv.local || sv.id === 'openai-compatible' ? `<label class="small">Server address<input type="text" id="baseurl" value="${esc(WB.baseURL ?? sv.base_url ?? '')}" placeholder="http://localhost:1234/v1"></label>` : ''}
        <h3>3. Model</h3>
        <div class="row"><button id="findmodels">Find models</button>
          <input type="text" id="filter" placeholder="Search" value="${esc(WB.filter)}" ${models ? '' : 'disabled'}></div>
        ${models ? `<div class="modellist">${shown.map(m => `<label class="modelrow ${m.id === WB.chosen ? 'on' : ''}"><input type="radio" name="model" value="${esc(m.id)}" ${m.id === WB.chosen ? 'checked' : ''}>
            <b>${esc(m.id)}</b> <span class="small muted">${esc(m.display_name && m.display_name !== m.id ? m.display_name : '')} ${esc(m.details || '')}</span></label>`).join('') || '<p class="small muted">No model matches.</p>'}</div>
          <p class="small muted">${shown.length} of ${models.length} models.</p>` : '<p class="small muted">Find the models this key can use, or type a model name below.</p>'}
        <h3>4. Player</h3>
        <div class="grid2">
          <label class="small">Model<input type="text" id="pmodel" value="${esc(WB.chosen)}" placeholder="e.g. gemini-2.5-flash"></label>
          <label class="small">Player name<input type="text" id="pname" value="${esc(WB.pname || '')}" placeholder="(the model's name)"></label>
          <label class="small">Temperature<input type="number" id="ptemp" step="0.1" min="0" max="2" placeholder="the model's default"></label>
          <label class="small">Seconds between calls<input type="number" id="pinterval" min="0" step="1" value="${WB.svc === 'gemini' ? 7 : 0}" title="Free tiers allow few requests a minute: 7 seconds stays under about 8 a minute"></label>
          <label class="small" title="${esc(effortHelp)}">Thinking / effort<input type="text" id="peffort" placeholder="${esc(effortHelp)}"></label>
          <label class="small">Answer format<select id="poutput"><option value="json">JSON (recommended)</option><option value="">Free text with JSON at the end</option></select></label>
        </div>
        <div class="row"><button id="testmodel">Test it</button> <button class="primary" id="addplayer">Add player</button></div>
        ${WB.test ? `<p class="small ${WB.test.error ? 'warn' : 'note good'}">${WB.test.error ? esc(WB.test.error) : `It answered “${esc(WB.test.text)}” (${esc(WB.test.model || '')}, ${WB.test.latency_ms} ms${WB.test.usage && WB.test.usage.input_tokens ? `, ${WB.test.usage.input_tokens} tokens in` : ''}).`}</p>` : ''}
      </div>
    </div>
    <div class="panel"><h3>Players</h3>
      ${WB.players.length ? `<table><tr><th>Player</th><th>Service</th><th>Model</th><th>Settings</th><th>Status</th><th></th></tr>
        ${WB.players.map(p => `<tr><td><b>${esc(p.name)}</b></td><td>${esc(svcOf(p.provider).name || p.provider)}</td><td class="mono">${esc(p.model)}</td>
          <td class="small">${[p.temperature != null ? 'temperature ' + p.temperature : '', p.output === 'json' ? 'JSON' : '', p.min_interval && p.min_interval !== '0s' ? p.min_interval + ' between calls' : '', p.effort ? 'effort ' + p.effort : ''].filter(Boolean).map(esc).join(', ')}</td>
          <td class="small">${p.ready ? '<span class="lbl in">ready</span>' : `<span class="warn">${esc(p.problem || 'not connected')}</span>`}${p.from === 'command line' ? ' <span class="muted">(command line)</span>' : ''}</td>
          <td>${p.from === 'dashboard' ? `<button class="small" data-rm="${esc(p.name)}">Remove</button>` : ''}</td></tr>`).join('')}</table>
        <p class="small muted">Use them in <b>Play → New game</b> (a seat) or in <b>Annotate</b>. Games with a model seat are saved in your data folder, under <span class="mono">games</span>.</p>`
      : '<p class="small muted">No players yet.</p>'}
    </div></div>`;
  wireModels();
}

function playerForm() {
  const num = id => { const v = $('#' + id).value; return v === '' ? null : Number(v); };
  const eff = $('#peffort').value.trim();
  return {
    service: WB.svc, model: $('#pmodel').value.trim(), name: $('#pname').value.trim(),
    base_url: $('#baseurl') ? $('#baseurl').value.trim() : '', temperature: num('ptemp'), min_interval_seconds: num('pinterval') || 0,
    effort: eff, output: $('#poutput').value,
  };
}

function wireModels() {
  const main = $('#main');
  main.onclick = async e => {
    const b = e.target.closest('button'); if (!b) return;
    try {
      if (b.dataset.svc) { WB.svc = b.dataset.svc; WB.chosen = ''; WB.filter = ''; WB.test = null; WB.baseURL = undefined; return renderModels(); }
      if (b.dataset.open !== undefined) return void await api('/api/store/open', { path: b.dataset.open });
      if (b.id === 'savekey') {
        const key = $('#key').value.trim(); if (!key) return toast('Paste the key first.');
        const r = await api('/api/llm/key', { service: WB.svc, key, remember: $('#remember').checked });
        WB.services = r.services; WB.players = r.players; await wbMeta(); toast('Key saved.'); return renderModels();
      }
      if (b.id === 'forget') {
        const r = await api('/api/llm/forget', { service: WB.svc });
        WB.services = r.services; WB.players = r.players; await wbMeta(); toast('Key forgotten.'); return renderModels();
      }
      if (b.id === 'findmodels') {
        b.disabled = true; b.textContent = 'Asking…';
        if ($('#baseurl')) WB.baseURL = $('#baseurl').value.trim();
        const r = await api('/api/llm/models', { service: WB.svc, base_url: WB.baseURL || '' });
        WB.models[WB.svc] = r.models; return renderModels();
      }
      if (b.id === 'testmodel') {
        const f = playerForm(); if (!f.model) return toast('Choose or type a model first.');
        b.disabled = true; b.textContent = 'Testing…';
        WB.test = await api('/api/llm/test', f); WB.chosen = f.model; WB.pname = f.name; return renderModels();
      }
      if (b.id === 'addplayer') {
        const f = playerForm(); if (!f.model) return toast('Choose or type a model first.');
        const r = await api('/api/llm/players', f);
        WB.players = r.players; await wbMeta(); toast(r.warning || 'Player added: it can take a seat in Play, or annotate.', r.warning ? 8000 : 3500);
        WB.pname = ''; return renderModels();
      }
      if (b.dataset.rm) {
        const r = await api('/api/llm/players/remove', { name: b.dataset.rm }); WB.players = r.players; await wbMeta(); return renderModels();
      }
    } catch (err) { toast(err.message, 8000); renderModels(); }
  };
  main.onchange = e => {
    if (e.target.name === 'model') { WB.chosen = e.target.value; $('#pmodel').value = WB.chosen; for (const x of main.querySelectorAll('.modelrow')) x.classList.toggle('on', x.querySelector('input').checked); }
  };
  const fl = $('#filter');
  if (fl) fl.oninput = () => { WB.filter = fl.value; const pos = fl.selectionStart; renderModels(); const f2 = $('#filter'); f2.focus(); f2.setSelectionRange(pos, pos); };
}

// ---------------------------------------------------------------------------
// Annotate
// ---------------------------------------------------------------------------
async function annLoad() {
  const [d, r, p] = await Promise.all([api('/api/annotate/datasets'), api('/api/annotate/runs'), api('/api/llm/players')]);
  if (META.kialo && !WB.kialo) WB.kialo = await api('/api/kialo').catch(() => []);
  WB.datasets = d.datasets || []; WB.datasetFolder = d.folder; WB.datasetErrors = d.errors || []; WB.runs = r.runs || []; WB.players = p.players || [];
  if (!WB.form.dataset && WB.datasets.length) WB.form.dataset = WB.datasets[0].id;
  if (WB.run) await runLoad();
}

async function runLoad() {
  const r = await api('/api/annotate/runs/' + WB.run);
  WB.status = r.status; WB.notes = r.notes || [];
  for (const x of WB.runs) if (x.status && x.status.id === WB.run) x.status = r.status; // the list on the left
  if (!WB.status.busy) WB.report = await api(`/api/annotate/runs/${WB.run}/report`);
  if (WB.item) WB.view = await api(`/api/annotate/runs/${WB.run}/item/${WB.item}`);
}

function annPoll() {
  clearTimeout(WB.timer);
  if (ui.tab !== 'annotate' || !WB.status || !WB.status.busy) return;
  WB.timer = setTimeout(async () => { try { await runLoad(); } catch (e) { /* keep polling */ } renderAnnotate(true); }, 2000);
}

async function renderAnnotate(noLoad) {
  const main = $('#main');
  if (!noLoad) {
    try { await annLoad(); } catch (e) { main.innerHTML = `<div class="setup"><p class="warn">${esc(e.message)}</p></div>`; return; }
  }
  const ready = readyPlayers();
  const ds = WB.datasets.find(d => d.id === WB.form.dataset) || {};
  const runs = WB.runs.map(x => {
    const st = x.status; if (!st) return `<button disabled>${esc(x.id)}: ${esc(x.error)}</button>`;
    const done = (st.progress || []).filter(p => p.role === 'annotator').map(p => `${p.done}/${p.total}`).join(' · ');
    return `<button data-run="${esc(st.id)}" class="${st.id === WB.run ? 'on' : ''}"><b>${esc(st.id)}</b><br><span class="small muted">${esc(st.settings.annotators.map(a => a.name).join(', '))}${st.settings.reviewer ? ' · reviewer ' + esc(st.settings.reviewer.name) : ''} · ${done}${st.busy ? ' · <b>' + esc(st.busy) + '…</b>' : ''}</span></button>`;
  }).join('');
  main.innerHTML = `<div class="annotate">
    <div class="side">
      <div class="panel"><h2>Annotate</h2>
        <p class="small muted">Models annotate claims as instances of Walton's schemes, under the guidelines, each claim in its own call; a model reviews where they disagree; you give your verdict, claim by claim. Every run is a folder in your data folder, under <span class="mono">annotation</span>.</p>
        <h3>New run</h3>
        <label class="small">Dataset<select id="fdataset">${WB.datasets.map(d => `<option value="${esc(d.id)}" ${d.id === WB.form.dataset ? 'selected' : ''}>${esc(d.name)}</option>`).join('')}</select></label>
        <p class="small muted">${esc(ds.description || '')} ${ds.items || 0} claims${ds.has_gold ? ', with a gold annotation to score against' : ''}. ${ds.source && ds.source !== 'built in' ? 'Folder: <span class="mono">' + esc(ds.source) + '</span>' : ''}</p>
        <details class="small"><summary>New dataset from a Kialo discussion</summary>
          ${(WB.kialo || []).length ? `<label class="small">A captured discussion<select id="kid">${WB.kialo.map(k => `<option value="${esc(k.id)}">${esc(k.id)}: ${esc(short(k.title, 60))}</option>`).join('')}</select></label>` : ''}
          <label class="small">${(WB.kialo || []).length ? 'or a' : 'A'} <span class="mono">graph.json</span> of the Kialo research visualizer<input type="file" id="kfile" accept=".json,application/json"></label>
          <div class="grid2"><label class="small">Levels below the thesis<input type="number" id="kdepth" min="0" value="3" title="0: all"></label>
            <label class="small">At most, claims<input type="number" id="klimit" min="0" value="60" title="0: all"></label></div>
          <button id="mkdataset">Make the dataset</button>
          <p class="small muted">It becomes a folder in <span class="mono">${esc(WB.datasetFolder || '')}</span>, with the claims top-down and their Kialo context (no gold).</p>
        </details>
        <p class="small muted">Your own datasets: a folder with <span class="mono">items.json</span> (and optionally <span class="mono">guidelines.md</span>, <span class="mono">gold.json</span>, <span class="mono">lexicon.json</span>) in <span class="mono">${esc(WB.datasetFolder || '')}</span>.</p>
        ${(WB.datasetErrors || []).map(e => `<p class="small warn">${esc(e)}</p>`).join('')}
        <div class="small"><b>Annotators</b> (one or more players)</div>
        ${ready.length ? ready.map(p => `<label class="inline small"><input type="checkbox" data-ann="${esc(p.name)}" ${WB.form.annotators.includes(p.name) ? 'checked' : ''}> ${esc(p.name)} <span class="muted">(${esc(runsOn(p))})</span></label>`).join('<br>')
          : '<p class="small warn">No player is ready: add one in the Models tab.</p>'}
        <label class="small">Reviewer (optional)<select id="freviewer"><option value="">— none —</option>${ready.map(p => `<option value="${esc(p.name)}" ${WB.form.reviewer === p.name ? 'selected' : ''}>${esc(p.name)} (${esc(runsOn(p))})</option>`).join('')}</select></label>
        <label class="small">Claims: the first<input type="number" id="fcount" min="1" max="${ds.items || 1}" value="${WB.form.count}"> of ${ds.items || 0}</label>
        <p class="small muted">Each call sends the guidelines and the scheme collection: about 35,000–40,000 tokens. On a free tier, start with a few claims.</p>
        <button class="primary" id="newrun" ${ready.length ? '' : 'disabled'}>Create the run</button>
      </div>
      <div class="panel"><h3>Runs</h3><div class="list">${runs || '<p class="small muted">None yet.</p>'}</div></div>
    </div>
    <div class="mainpane">${WB.status ? runPane() : '<div class="panel"><p class="muted">Create a run, or open one.</p></div>'}</div>
  </div>`;
  wireAnnotate();
  annPoll();
}

function runPane() {
  const st = WB.status, rep = WB.report;
  const ann = st.settings.annotators.map(a => a.name);
  const prog = (st.progress || []).map(p => `<tr><td><b>${esc(p.name)}</b> <span class="muted small">${esc(p.role)}</span></td><td>${p.done}${p.total ? ' / ' + p.total : ''}</td><td>${p.failed || ''}</td>
    <td class="small">${p.calls} calls, ${p.input_tokens.toLocaleString()} tokens in, ${p.output_tokens.toLocaleString()} out</td>
    <td class="small">${p.running ? `working on ${esc(p.current || '')}` : ''} ${p.last_error ? `<span class="warn">${esc(short(p.last_error, 160))}</span>` : ''}</td></tr>`).join('');
  const fmt = f => f ? `${f.agreement.toFixed(2)}${f.kappa != null ? ` <span class="muted">κ ${f.kappa.toFixed(2)}</span>` : ''}` : '';
  const fields = ['scheme', 'fit', 'role', 'polarity', 'target_claim', 'target_part', 're_parented', 'evidence'];
  const agreeRows = [];
  if (rep) {
    for (const p of rep.pairs || []) agreeRows.push([`${p.a} ↔ ${p.b}`, p.items, p.fields]);
    for (const a of rep.annotators || []) if (a.vs_gold) agreeRows.push([`${a.name} ↔ gold`, a.vs_gold.items, a.vs_gold.fields]);
    if (rep.reviewer && rep.reviewer.adjudicated_vs_gold) agreeRows.push([`reviewer ↔ gold`, rep.reviewer.adjudicated_vs_gold.items, rep.reviewer.adjudicated_vs_gold.fields]);
  }
  const items = st.settings.items;
  const claimList = items.map(id => {
    const dis = rep && rep.disagreements && rep.disagreements[id];
    return `<button data-item="${esc(id)}" class="${id === WB.item ? 'on' : ''}">${esc(id)}${dis ? ` <span class="pill">differs: ${dis.length}</span>` : ''}</button>`;
  }).join('');
  return `<div class="panel"><h2>${esc(st.id)}</h2>
      <p class="small muted">${esc(st.dataset.name)} · ${items.length} claims · annotators ${esc(ann.join(', '))}${st.settings.reviewer ? ' · reviewer ' + esc(st.settings.reviewer.name) : ''}</p>
      ${dataFolder('annotation/' + st.id, 'Its files are in')}
      <p class="small">${(st.files || []).map(f => `<a href="/api/store/file?path=${encodeURIComponent('annotation/' + st.id + '/' + f)}">${esc(f)}</a>`).join(' · ')}
        <span class="muted">(README.txt says what each is)</span></p>
      ${(WB.notes || []).map(n => `<p class="small warn">${esc(n)}</p>`).join('')}
      <div class="row">
        <button id="next" ${st.busy || !st.next_item ? 'disabled' : ''} title="Every annotator annotates the next claim; then look at it">Annotate the next claim${st.next_item ? ' (' + esc(st.next_item) + ')' : ''}</button>
        <button class="primary" id="startall" ${st.busy || !st.next_item ? 'disabled' : ''} title="Runs in the server: you can close this page">Annotate all the rest</button>
        <button id="review" ${st.busy || !st.settings.reviewer || st.annotated_by_two === 0 ? 'disabled' : ''} title="The reviewer reviews the claims whose annotations differ">Review the disagreements</button>
        <button id="reviewall" ${st.busy || !st.settings.reviewer || st.annotated_by_two === 0 ? 'disabled' : ''}>Review all</button>
        <button id="stop" ${st.busy ? '' : 'disabled'}>Stop</button>
      </div>
      ${st.busy ? `<p class="small"><b>${esc(st.busy)}…</b> This goes on in the server, in the background: you can close this page and come back.</p>` : ''}
      <table><tr><th>Who</th><th>Done</th><th>Failed</th><th>Cost</th><th></th></tr>${prog}</table>
    </div>
    ${agreeRows.length ? `<div class="panel"><h3>Agreement</h3><div class="netwrap"><table><tr><th></th><th>claims</th>${fields.map(f => `<th>${esc(f.replace('_', ' '))}</th>`).join('')}</tr>
      ${agreeRows.map(([who, n, f]) => `<tr><td>${esc(who)}</td><td>${n}</td>${fields.map(k => `<td class="small">${fmt(f[k])}</td>`).join('')}</tr>`).join('')}</table></div>
      ${rep ? `<p class="small muted">${(rep.annotators || []).map(a => `${esc(a.name)}: ${a.annotated} annotated, ${a.valid} with no problem found`).join(' · ')}. Agreement is the share of claims labelled alike; κ corrects it for chance.</p>` : ''}</div>` : ''}
    <div class="grid-claims">
      <div class="panel"><h3>Claims</h3><div class="list claims">${claimList}</div></div>
      <div class="panel">${WB.view ? itemPane() : '<p class="small muted">Open a claim to compare its annotations and give your verdict.</p>'}</div>
    </div>`;
}

const primaryOf = a => { const ins = (a && a.instances) || []; return ins.find(i => i.primary) || ins[0] || {}; };

function annCard(name, a, problems, diff) {
  if (!a) return '';
  const p = primaryOf(a), s = p.scheme || {}, r = a.role || {}, t = (a.target || {}).actual || {};
  const mark = f => diff.includes(f) ? ' class="differs"' : '';
  const prem = (p.premises || []).map(x => `<li><span class="muted">${esc(x.role)}</span> <span class="mono">${esc(x.formula)}</span>${x.status ? ` <span class="muted small">(${esc(x.status)})</span>` : ''}</li>`).join('');
  return `<div class="anncard"><h4>${esc(name)}</h4>
    <table class="wtable">
      <tr${mark('scheme')}><td>Scheme</td><td><b>${esc(s.walton_id)}</b> ${s.compendium ? '<span class="muted">book ' + esc(s.compendium) + '</span>' : ''}</td></tr>
      <tr${mark('fit')}><td>Fit</td><td>${esc(s.fit)}${s.confidence != null ? ` <span class="muted">(confidence ${esc(s.confidence)})</span>` : ''}</td></tr>
      <tr${mark('role')}><td>Role</td><td>${esc(r.kind)}${r.cq ? ' — ' + esc(r.cq) : ''}</td></tr>
      <tr${mark('polarity')}><td>Polarity</td><td>${esc(r.polarity)}</td></tr>
      <tr${mark('target_claim') || mark('target_part')}><td>Target</td><td>${esc(t.claim)} · ${esc(t.part)} ${t.formula ? '<span class="mono">' + esc(t.formula) + '</span>' : ''}</td></tr>
      <tr${mark('re_parented')}><td>Re-parented</td><td>${esc((a.target || {}).re_parented)}</td></tr>
      <tr${mark('evidence')}><td>Evidence</td><td>${esc((a.evidence || {}).status)}</td></tr>
    </table>
    ${prem ? `<div class="small">Premises<ul class="small">${prem}</ul></div>` : ''}
    ${p.conclusion ? `<div class="small">Conclusion <span class="mono">${esc(p.conclusion.formula)}</span></div>` : ''}
    ${(problems || []).length ? `<div class="small warn">${problems.map(esc).join('<br>')}</div>` : '<div class="small muted">The check finds nothing wrong.</div>'}
    ${a.notes && a.notes.text ? `<div class="small muted">Note: ${esc(a.notes.text)}</div>` : ''}
  </div>`;
}

function itemPane() {
  const v = WB.view, it = v.item || {};
  const names = Object.keys(v.annotations);
  const parent = (it.kialo_parents || [])[0];
  const parentText = parent && ((it.ancestors || []).find(x => x.claim_id === parent.claim_id) || {}).text;
  const cards = names.map(n => annCard(n, v.annotations[n], v.problems[n], v.disagree || [])).join('');
  const fails = Object.entries(v.failures || {}).map(([n, e]) => `<p class="small warn">${esc(n)} failed: ${esc(e)}</p>`).join('');
  const rv = v.review;
  const review = rv && WB.reveal ? `<div class="panel review"><h4>The reviewer (${esc(rv.reviewer)})</h4>
      ${rv.error ? `<p class="small warn">${esc(rv.error)}</p>` : `<table class="wtable">${Object.entries(rv.fields || {}).map(([f, c]) => `<tr><td>${esc(f)}</td><td><b>${esc(c.annotator || c.choice)}</b> <span class="small muted">${esc(c.why || '')}</span></td></tr>`).join('')}</table>
      <p class="small">${rv.needs_human ? '<b>Asks for a person.</b> ' : ''}${rv.confidence ? 'Confidence ' + esc(rv.confidence) + '. ' : ''}${esc(rv.notes || '')}</p>`}</div>` : '';
  const choice = (v.verdict || {}).choice || '';
  const opts = [...names.map(n => [n, n + ' is right']), ...(rv && !rv.error ? [['reviewer', 'The reviewer is right']] : []), ...(v.gold ? [['gold', 'The gold is right']] : []), ['none', 'None is right'], ['unsure', 'Unsure']];
  return `<h3>${esc(v.id)}</h3>
    <p class="claimtext">${rich(it.text || '')}</p>
    ${parent ? `<p class="small muted">${esc(parent.relation)} of ${esc(parent.claim_id)}${parentText ? ': ' + esc(short(parentText, 160)) : ''}</p>` : '<p class="small muted">The thesis.</p>'}
    ${v.disagree && v.disagree.length ? `<p class="small">They differ on: <b>${esc(v.disagree.join(', ').replace(/_/g, ' '))}</b></p>` : names.length > 1 ? '<p class="small">They agree on every field compared.</p>' : ''}
    <div class="anncards">${cards || '<p class="small muted">Not annotated yet.</p>'}</div>${fails}
    <label class="inline small"><input type="checkbox" id="reveal" ${WB.reveal ? 'checked' : ''}> Show the reviewer${v.gold ? ' and the gold' : ''} (leave it off to judge first)</label>
    ${review}
    ${v.gold && WB.reveal ? `<div class="anncards">${annCard('Gold (adjudicated)', v.gold, [], v.disagree || [])}</div>` : ''}
    <div class="verdict"><h4>Your verdict</h4>
      <div class="seg">${opts.map(([k, l]) => `<button data-choice="${esc(k)}" class="${choice === k ? 'on' : ''}">${esc(l)}</button>`).join('')}</div>
      <input type="text" id="vnote" placeholder="Why (optional)" value="${esc((v.verdict || {}).note || '')}">
      <p class="small muted">Saved in the run's <span class="mono">review/human.json</span>.</p>
    </div>`;
}

function wireAnnotate() {
  const main = $('#main');
  main.onclick = async e => {
    const b = e.target.closest('button'); if (!b) return;
    try {
      if (b.dataset.open !== undefined) return void await api('/api/store/open', { path: b.dataset.open });
      if (b.dataset.run) { WB.run = b.dataset.run; WB.item = null; WB.view = null; await runLoad(); return renderAnnotate(true); }
      if (b.dataset.item) { WB.item = b.dataset.item; WB.view = await api(`/api/annotate/runs/${WB.run}/item/${WB.item}`); return renderAnnotate(true); }
      if (b.id === 'mkdataset') {
        const body = { depth: Number($('#kdepth').value) || 0, limit: Number($('#klimit').value) || 0 };
        const file = $('#kfile').files[0];
        if (file) body.graph = JSON.parse(await file.text());
        else if ($('#kid')) body.kialo_id = $('#kid').value;
        else return toast('Choose a graph.json first.');
        b.disabled = true; b.textContent = 'Making it…';
        const r = await api('/api/annotate/datasets/kialo', body);
        WB.form.dataset = r.dataset; toast(`Dataset made: ${r.claims} claims, in ${r.dir}`, 8000); return renderAnnotate();
      }
      if (b.id === 'newrun') {
        const f = WB.form;
        if (!f.annotators.length) return toast('Tick at least one annotator.');
        const r = await api('/api/annotate/runs', { dataset: f.dataset, annotators: f.annotators, reviewer: f.reviewer, count: Number(f.count) || 0 });
        WB.run = r.status.id; WB.item = null; WB.view = null; toast('Run created: annotate the next claim, or all of them.'); return renderAnnotate();
      }
      const act = { next: 'next', startall: 'start', review: 'review', reviewall: 'review', stop: 'stop' }[b.id];
      if (act) {
        b.disabled = true;
        const body = b.id === 'reviewall' ? { all: true } : {};
        const r = await api(`/api/annotate/runs/${WB.run}/${act}`, body);
        WB.status = r.status;
        if (b.id === 'next' && r.status.next_item) WB.item = r.status.next_item;
        return renderAnnotate(true);
      }
      if (b.dataset.choice !== undefined) {
        await api(`/api/annotate/runs/${WB.run}/verdict`, { item: WB.item, choice: b.dataset.choice, note: $('#vnote').value });
        await runLoad(); toast('Verdict saved.'); return renderAnnotate(true);
      }
    } catch (err) { toast(err.message, 8000); }
  };
  main.onchange = e => {
    const t = e.target;
    if (t.id === 'fdataset') { WB.form.dataset = t.value; renderAnnotate(true); }
    if (t.id === 'freviewer') WB.form.reviewer = t.value;
    if (t.id === 'fcount') WB.form.count = t.value;
    if (t.dataset.ann) { const n = t.dataset.ann; WB.form.annotators = t.checked ? [...WB.form.annotators, n] : WB.form.annotators.filter(x => x !== n); }
    if (t.id === 'reveal') { WB.reveal = t.checked; renderAnnotate(true); }
  };
}
