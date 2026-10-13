// WAG browser interface. Everything shown comes from the Go engine through
// the JSON API (/api/...): the browser renders the State and posts moves,
// which the engine validates and evaluates. No game logic lives here.
'use strict';

const $ = (sel, root = document) => root.querySelector(sel);
const esc = s => String(s ?? '').replace(/[&<>"']/g, c => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c]));
const short = (s, n = 90) => { s = String(s ?? ''); return s.length > n ? s.slice(0, n - 1) + '…' : s; };
// plain drops the URL of Markdown links, for one-line displays.
const plain = s => String(s ?? '').replace(/\[([^\]]+)\]\((https?:\/\/[^)\s]+)\)/g, '$1');
// rich escapes text and turns Markdown links [text](url), as in Kialo claims, into links.
const rich = s => esc(s).replace(/\[([^\]]+)\]\((https?:\/\/[^)\s]+)\)/g, '<a href="$2" target="_blank" rel="noopener">$1</a>');

let META = null;           // schemes, critical questions, modes
let SCHEME = {}, CQ = {};  // by ID
let G = null;              // the current game's State
const ui = {
  tab: 'play', gameId: null, focus: null, crumbs: [], openCQ: null, moderator: false, autoplay: true,
  open: new Set(), c: null, cKey: '', previewSeq: 0, setup: null, aiTimer: null, busy: false,
};

// ---------------------------------------------------------------------------
// API
// ---------------------------------------------------------------------------
async function api(path, body, raw) {
  const opt = body === undefined ? {} : { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: raw ? body : JSON.stringify(body) };
  const r = await fetch(path, opt);
  const j = await r.json().catch(() => ({ error: r.statusText }));
  if (!r.ok) throw new Error(j.error || r.statusText);
  return j;
}
function toast(msg, ms = 3500) {
  const t = $('#toast'); t.textContent = msg; t.hidden = false;
  clearTimeout(toast.t); toast.t = setTimeout(() => (t.hidden = true), ms);
}

// ---------------------------------------------------------------------------
// Start
// ---------------------------------------------------------------------------
async function init() {
  META = await api('/api/meta');
  for (const s of META.schemes) SCHEME[s.id] = s;
  for (const q of META.cqs) CQ[q.id] = q;
  $('#tabs').addEventListener('click', e => {
    const b = e.target.closest('button[data-tab]'); if (!b) return;
    ui.tab = b.dataset.tab;
    for (const x of $('#tabs').children) x.classList.toggle('active', x === b);
    render();
  });
  const m = location.hash.match(/game=(\w+)/);
  if (m) await loadGame(m[1]);
  render();
  setInterval(poll, 1500);
}

async function loadGame(id) {
  try {
    const st = await api('/api/games/' + id);
    ui.gameId = id; setState(st);
  } catch (e) { toast(e.message); ui.gameId = null; G = null; }
}

function setState(st) {
  for (const k of ['roots', 'questions', 'log', 'players']) st[k] = st[k] || [];
  const firstLoad = !G || G.id !== st.id;
  G = st;
  location.hash = 'game=' + G.id;
  if (firstLoad) { ui.focus = G.roots[0] || Object.keys(G.statements)[0] || null; ui.crumbs = ui.focus ? [ui.focus] : []; ui.c = null; }
  if (ui.focus && !G.statements[ui.focus]) ui.focus = G.roots[0] || null;
  if (!ui.focus && G.roots.length) { ui.focus = G.roots[0]; ui.crumbs = [ui.focus]; }
  const key = [G.id, G.round, G.active, G.log.length, !!G.pending, G.over].join('|');
  if (key !== ui.cKey) { ui.cKey = key; ui.c = null; }
  renderGamebar();
  if (ui.tab === 'play') renderPlay();
  scheduleAI();
}

async function poll() {
  if (!G || ui.busy) return;
  try {
    const st = await api(`/api/games/${G.id}?since=${G.version}`);
    if (!st.unchanged) setState(st);
  } catch (e) { /* the server may be restarting */ }
}

function scheduleAI() {
  clearTimeout(ui.aiTimer);
  if (!G || G.over || G.pending || !G.active_ai || !ui.autoplay) return;
  ui.aiTimer = setTimeout(aiStep, 900);
}
async function aiStep() {
  if (!G || ui.busy) return;
  ui.busy = true;
  try { const r = await api(`/api/games/${G.id}/ai`, {}); setState(r.state); }
  catch (e) { toast(e.message); }
  finally { ui.busy = false; }
}

// ---------------------------------------------------------------------------
// Top-level render
// ---------------------------------------------------------------------------
function render() {
  renderGamebar();
  if (ui.tab === 'nascq') return renderNASCQ();
  if (ui.tab === 'rules') return renderRules();
  if (!G) return renderSetup();
  renderPlay();
}

function renderGamebar() {
  const bar = $('#gamebar');
  if (!G) { bar.innerHTML = ''; return; }
  const players = G.players.map(p => {
    const cls = ['pill', p.id === G.active && !G.over ? 'active' : '', p.side].join(' ');
    const kind = p.kind === 'human' ? '' : ` <span class="muted">(${esc(p.kind)} AI)</span>`;
    return `<span class="${cls}" title="${esc(p.side || '')}">${esc(p.id)}${kind} <b>${p.score.toFixed(1)}</b>${p.resigned ? ' ✗' : ''}</span>`;
  }).join('');
  const verdict = G.config.objective === 'claim'
    ? `<span class="pill">Claim <span class="lbl ${(G.verdict || 'out').toLowerCase()}">${esc(G.verdict || 'Out')}</span></span>` : '';
  const mod = G.config.moderated ? `<label class="inline small"><input type="checkbox" id="modtoggle" ${ui.moderator ? 'checked' : ''}> I am the moderator</label>` : '';
  bar.innerHTML = `
    <span class="pill" title="${esc(G.title)}">${esc(short(G.title, 38))}</span>
    <span class="pill">${esc(G.config.mode)} · round ${G.round}${G.config.max_rounds ? ' / ' + G.config.max_rounds : ''}</span>
    ${players}${verdict}${mod}
    <button class="small" id="newgame">New game</button>
    <a class="small" href="/api/games/${G.id}/export" target="_blank">Export</a>`;
  $('#newgame').onclick = () => { G = null; ui.gameId = null; location.hash = ''; ui.tab = 'play'; render(); };
  const mt = $('#modtoggle'); if (mt) mt.onchange = () => { ui.moderator = mt.checked; renderPlay(); };
}

// ---------------------------------------------------------------------------
// Setup
// ---------------------------------------------------------------------------
function renderSetup() {
  const s = ui.setup || (ui.setup = {
    mode: 'debate', seats: [{ name: 'Alice', kind: 'human' }, { name: 'Bob', kind: 'greedy' }],
    allow_other: true, moderated: false, require_evidence: false, title: '', claim: '', standard: 'PE',
    scoring: '', max_rounds: '', answer_within: '', target: '',
  });
  const modes = META.modes.map(m => `<div class="mode ${m.Name === s.mode ? 'on' : ''}" data-mode="${m.Name}"><b>${esc(m.Name)}</b><span class="small muted">${esc(m.Description)}</span></div>`).join('');
  const seats = s.seats.map((x, i) => `<div class="seat">
      <input type="text" data-seat="${i}" data-f="name" value="${esc(x.name)}" placeholder="Name">
      <select data-seat="${i}" data-f="kind">${['human', 'random', 'greedy'].map(k => `<option value="${k}" ${x.kind === k ? 'selected' : ''}>${k === 'human' ? 'Human (this browser)' : k + ' AI'}</option>`).join('')}</select>
      <button class="small" data-del="${i}" title="Remove">✕</button></div>`).join('');
  $('#main').innerHTML = `<div class="setup">
    <div class="panel"><h2>New game</h2>
      <h3>Mode</h3><div class="modes">${modes}</div>
      <div class="grid2">
        <div><h3>Players</h3><div class="seats">${seats}</div>
          <button class="small" id="addseat" ${s.mode === 'solo' ? 'disabled' : ''}>Add a player</button>
          <p class="small muted">Debates: even seats argue for the claim (pro), odd seats against it (con). Hot-seat: all human seats play in this browser; other tabs can watch or moderate.</p></div>
        <div><h3>Options</h3>
          <label class="inline"><input type="checkbox" id="o_other" ${s.allow_other ? 'checked' : ''}> Allow the “Other” scheme (your own premises and conclusion)</label>
          <label class="inline"><input type="checkbox" id="o_mod" ${s.moderated ? 'checked' : ''}> Moderator: every move waits for approval</label>
          <label class="inline"><input type="checkbox" id="o_ev" ${s.require_evidence ? 'checked' : ''}> Require sources for schemes that need evidence</label>
          <div class="grid2">
            <div><label>Claim (debates; empty: pro's first argument states it)</label><input type="text" id="o_claim" value="${esc(s.claim)}" placeholder="e.g. Governments should tax meat."></div>
            <div><label>Proof standard of the claim</label><select id="o_std">${['PE', 'CCE', 'BRD'].map(x => `<option ${s.standard === x ? 'selected' : ''}>${x}</option>`).join('')}</select></div>
          </div>
          <div class="grid3">
            <div><label>Max rounds (blank: mode's)</label><input type="number" id="o_rounds" min="0" value="${esc(s.max_rounds)}"></div>
            <div><label>Answer questions within (turns)</label><input type="number" id="o_within" min="0" value="${esc(s.answer_within)}"></div>
            <div><label>Scoring</label><select id="o_scoring"><option value="">mode's</option><option value="move" ${s.scoring === 'move' ? 'selected' : ''}>once per move</option><option value="live" ${s.scoring === 'live' ? 'selected' : ''}>live</option></select></div>
          </div>
          <label>Title</label><input type="text" id="o_title" value="${esc(s.title)}" placeholder="WAG — ${esc(s.mode)}">
        </div>
      </div>
      <p><button class="primary" id="start">Start the game</button></p>
    </div>
    <div class="grid2" style="margin-top:12px">
      <div class="panel"><h2>Import a Kialo discussion</h2>
        <p class="small muted">A captured discussion (<span class="mono">data/&lt;id&gt;/graph.json</span> of the Kialo research visualizer) becomes a debate on its thesis: every pro and con is an argument of the “Other” scheme, weighted by its impact votes, ready to be questioned, answered and re-classified.</p>
        <div id="kialolist" class="small muted">${META.kialo ? 'Loading…' : 'Start the server with <span class="mono">--kialo-dir</span> to list captured discussions here, or choose a graph.json:'}</div>
        <label>graph.json</label><input type="file" id="kfile" accept=".json,application/json">
        <label>Depth limit (0: everything)</label><input type="number" id="kdepth" value="0" min="0">
        <p><button id="kimport">Import the file</button></p>
      </div>
      <div class="panel"><h2>Games on this server</h2><ul class="games" id="gamelist"><li class="muted">Loading…</li></ul></div>
    </div></div>`;
  $('#main').onclick = e => {
    const m = e.target.closest('[data-mode]');
    if (m) { s.mode = m.dataset.mode; if (s.mode === 'solo') s.seats = s.seats.slice(0, 1); renderSetup(); }
    const d = e.target.closest('[data-del]');
    if (d && s.seats.length > 1) { s.seats.splice(+d.dataset.del, 1); renderSetup(); }
  };
  $('#main').oninput = e => {
    const t = e.target;
    if (t.dataset.seat !== undefined) s.seats[+t.dataset.seat][t.dataset.f] = t.value;
  };
  $('#main').onchange = $('#main').oninput;
  $('#addseat').onclick = () => { s.seats.push({ name: `Player${s.seats.length + 1}`, kind: 'human' }); renderSetup(); };
  $('#start').onclick = async () => {
    const num = id => { const v = $(id).value.trim(); return v === '' ? undefined : +v; };
    s.allow_other = $('#o_other').checked; s.moderated = $('#o_mod').checked; s.require_evidence = $('#o_ev').checked;
    s.claim = $('#o_claim').value; s.standard = $('#o_std').value; s.title = $('#o_title').value;
    s.max_rounds = $('#o_rounds').value; s.answer_within = $('#o_within').value; s.scoring = $('#o_scoring').value;
    const req = {
      title: s.title, mode: s.mode, seats: s.seats, allow_other: s.allow_other, moderated: s.moderated,
      require_evidence: s.require_evidence, claim: s.claim, standard: s.standard, scoring: s.scoring || undefined,
      max_rounds: num('#o_rounds'), answer_within: num('#o_within'),
    };
    try { const st = await api('/api/games', req); ui.gameId = st.id; G = null; setState(st); ui.moderator = s.moderated; render(); }
    catch (e) { toast(e.message); }
  };
  $('#kimport').onclick = async () => {
    const f = $('#kfile').files[0]; if (!f) return toast('Choose a graph.json first.');
    toast('Importing…', 10000);
    try { const r = await api('/api/kialo/import?depth=' + (+$('#kdepth').value || 0), await f.text(), true); afterImport(r); }
    catch (e) { toast(e.message); }
  };
  api('/api/games').then(list => {
    if (!$('#gamelist')) return;
    $('#gamelist').innerHTML = (list || []).map(g => `<li><a href="#game=${g.ID}" data-load="${g.ID}">${esc(g.Title)}</a> <span class="muted small">${esc(g.Mode)} · round ${g.Round}${g.Over ? ' · over' : ''}</span></li>`).join('') || '<li class="muted">None yet.</li>';
    $('#gamelist').onclick = async e => { const a = e.target.closest('[data-load]'); if (!a) return; e.preventDefault(); await loadGame(a.dataset.load); render(); };
  });
  if (META.kialo) api('/api/kialo').then(list => {
    if (!$('#kialolist')) return;
    $('#kialolist').innerHTML = list.length ? '<ul>' + list.map(d => `<li><button class="link" data-k="${esc(d.id)}">${esc(d.title)}</button> <span class="muted">(${(d.size / 1e6).toFixed(1)} MB)</span></li>`).join('') + '</ul>' : 'No graph.json found.';
    $('#kialolist').onclick = async e => {
      const b = e.target.closest('[data-k]'); if (!b) return;
      toast('Importing…', 10000);
      try { afterImport(await api(`/api/kialo/import?id=${encodeURIComponent(b.dataset.k)}&depth=${+$('#kdepth').value || 0}`, {})); }
      catch (err) { toast(err.message); }
    };
  });
}
function afterImport(r) {
  const p = r.report;
  toast(`Imported “${p.title}”: ${p.arguments} arguments (${p.pros} pros, ${p.cons} cons; ${p.voted} with votes), ${p.cycles} left out as cycles.`, 8000);
  G = null; setState(r.state); render();
}

// ---------------------------------------------------------------------------
// Play
// ---------------------------------------------------------------------------
function renderPlay() {
  if (ui.tab !== 'play') return;
  if (!G) return renderSetup();
  if (!$('#play')) {
    $('#main').innerHTML = `<div class="play" id="play">
      <aside><div class="panel sticky" id="mappanel"></div></aside>
      <section id="board"></section>
      <aside><div class="sticky" id="side"></div></aside></div>`;
  }
  renderMap(); renderBoard(); renderSide();
}

function lbl(l) { return `<span class="lbl ${l}">${{ in: 'IN', out: 'OUT', undecided: 'UND' }[l] || l}</span>`; }
function stmtText(id) { const s = G.statements[id]; return s ? plain(s.text) : id; }
function argTitle(id) { const a = G.arguments[id]; if (!a) return id; return `${id} · ${a.scheme_name}${a.owner ? ' · ' + a.owner : ''}`; }

function focusOn(id, push = true) {
  if (!G.statements[id]) return;
  ui.focus = id; ui.openCQ = null;
  if (push) { const i = ui.crumbs.indexOf(id); ui.crumbs = i >= 0 ? ui.crumbs.slice(0, i + 1) : [...ui.crumbs, id].slice(-8); }
  renderMap(); renderBoard();
  $('#board').scrollIntoView({ block: 'nearest' });
}

// The map: an outline from the claim and the other roots.
function renderMap() {
  const p = $('#mappanel'); if (!p) return;
  const disputed = sid => { const s = G.statements[sid]; return s && (s.pro.length || s.questions.length || ((G.statements[s.opposite] || {}).pro || []).length); };
  const node = (sid, path, depth, tag) => {
    const s = G.statements[sid]; if (!s) return '';
    const key = path + '/' + sid;
    const kids = [...s.pro.map(a => ['pro', a]), ...((G.statements[s.opposite] || {}).pro || []).map(a => ['con', a]), ...s.questions.map(a => ['q', a])];
    const open = ui.open.has(key) || (depth < 1 && !ui.open.has('!' + key));
    const toggle = kids.length ? `<button class="toggle" data-tog="${esc(key)}" data-open="${open ? 1 : 0}">${open ? '▾' : '▸'}</button>` : '<span class="toggle"></span>';
    let html = `<li><div class="node ${sid === ui.focus ? 'focus' : ''}" data-focus="${esc(sid)}">${toggle}<span class="dot ${s.label}"></span><span class="t">${tag ? `<span class="tag">${tag}</span> ` : ''}${esc(plain(s.text))}</span></div>`;
    if (open && kids.length && depth < 12) {
      html += '<ul>' + kids.map(([side, aid]) => {
        const a = G.arguments[aid]; if (!a) return '';
        const sign = side === 'pro' ? '＋' : side === 'con' ? '−' : '?';
        const sub = (x, tag) => path.includes('/' + x + '/') ? '' : node(x, key + '/' + aid, depth + 1, tag);
        // Premises, then the exceptions and the undercutter once someone argues about them.
        const prem = a.premises.map(x => sub(x)).join('') +
          (a.exceptions || []).filter(disputed).map(x => sub(x, 'exception')).join('') +
          (a.undercutter && disputed(a.undercutter) ? sub(a.undercutter, 'inference') : '');
        return `<li><div class="arg"><span class="sign ${side}">${sign}</span><span>${esc(argTitle(aid))}${a.in_force ? '' : ' (out of force)'}</span></div>${prem ? '<ul>' + prem + '</ul>' : ''}</li>`;
      }).join('') + '</ul>';
    }
    return html + '</li>';
  };
  const roots = G.roots.length ? G.roots : Object.keys(G.statements).slice(0, 1);
  p.innerHTML = `<h2>Map</h2><div class="map"><ul>${roots.map(r => node(r, '', 0)).join('') || '<li class="muted">No arguments yet.</li>'}</ul></div>`;
  p.onclick = e => {
    const t = e.target.closest('[data-tog]');
    if (t) { const k = t.dataset.tog; if (t.dataset.open === '1') { ui.open.delete(k); ui.open.add('!' + k); } else { ui.open.add(k); ui.open.delete('!' + k); } renderMap(); return; }
    const f = e.target.closest('[data-focus]'); if (f) focusOn(f.dataset.focus);
  };
}

// The board: the focused statement, with the arguments for and against it.
function renderBoard() {
  const b = $('#board'); if (!b) return;
  if (!ui.focus || !G.statements[ui.focus]) {
    b.innerHTML = `<div class="panel"><h2>The board is empty</h2><p>Put forward the first argument with the composer on the right: write it in your own words, choose the scheme it fits, and fill in the scheme's parts.</p>${G.config.objective === 'claim' ? '<p class="muted">In a debate, pro\'s first argument states the claim.</p>' : ''}</div>`;
    return;
  }
  const s = G.statements[ui.focus];
  const crumbs = ui.crumbs.map(id => `<button data-focus="${esc(id)}" title="${esc(stmtText(id))}">${esc(short(stmtText(id), 40))}</button>`).join(' › ');
  const badges = [];
  if (ui.focus === G.claim) badges.push('<span class="pill pro">the claim</span>');
  if (s.issue) badges.push(`<span class="pill">at issue · ${esc(s.standard)}</span>`);
  if (s.decision) badges.push(`<span class="pill">${esc(s.decision)} by the moderator</span>`);
  if (s.presumed_by.length) badges.push(`<span class="pill" title="Premises hold until attacked or questioned">presumed by ${esc(s.presumed_by.join(', '))}</span>`);
  if (s.exception_of.length) badges.push(`<span class="pill">exception of ${esc(s.exception_of.join(', '))}</span>`);
  const modbtns = G.config.moderated && ui.moderator && !G.over ? `<div class="badges"><span class="small muted">Moderator:</span>
      <button class="small" data-decide="accepted">Accept</button><button class="small" data-decide="rejected">Reject</button><button class="small" data-decide="undecided">Clear</button></div>` : '';
  const opp = G.statements[s.opposite];
  const forIds = s.pro, againstIds = opp ? opp.pro : [];
  const qs = s.questions.map(qid => { const q = G.arguments[qid]; return `<div class="argcard con"><div class="head"><span class="scheme">Question</span> by ${esc(q.owner)} ${q.in_force ? '' : '· answered'}</div><div class="small">${esc(questionTextFor(qid))}</div></div>`; }).join('');
  const parents = s.premise_of.map(a => `<li>premise of <button class="link" data-focus="${esc(G.arguments[a]?.conclusion)}">${esc(argTitle(a))}</button>: ${esc(short(stmtText(G.arguments[a]?.conclusion), 80))}</li>`).join('');
  const exOf = s.exception_of.map(a => `<li>exception of <button class="link" data-focus="${esc(G.arguments[a]?.conclusion)}">${esc(argTitle(a))}</button></li>`).join('');
  b.innerHTML = `<div class="crumbs">${crumbs}</div>
    <div class="panel focus-card ${s.label}">
      <div>${lbl(s.label)} <span class="small muted mono">${esc(s.id)}</span></div>
      <div class="focus-text">${rich(s.text)}</div>
      <div class="explain">${esc(s.explain)}</div>
      <div class="badges">${badges.join('')}</div>${modbtns}
      ${parents || exOf ? `<ul class="parents">${parents}${exOf}</ul>` : ''}
    </div>
    <div class="columns">
      <div><div class="col-head pro">For it (${forIds.length})</div>${forIds.map(a => argCard(a, 'pro')).join('') || '<div class="empty">No argument for it yet.</div>'}</div>
      <div><div class="col-head con">Against it (${againstIds.length + s.questions.length})</div>${againstIds.map(a => argCard(a, 'con')).join('')}${qs}${againstIds.length + s.questions.length ? '' : '<div class="empty">Nothing against it yet.</div>'}</div>
    </div>`;
  b.onclick = boardClick;
}

function questionTextFor(qid) {
  for (const a of Object.values(G.arguments)) for (const c of a.cqs) if (c.asked_by && G.log.some(l => l.args && l.args.includes(qid))) {
    const entry = G.log.find(l => l.args && l.args.includes(qid)); return entry ? entry.text : '';
  }
  const entry = G.log.find(l => l.args && l.args.includes(qid)); return entry ? entry.text : '';
}

function argCard(id, side) {
  const a = G.arguments[id]; if (!a) return '';
  const status = a.in_force ? '<span class="lbl in">in force</span>' : `<span class="lbl out" title="${esc(a.why_out)}">out of force</span>`;
  const answers = a.answers ? `<div class="answers">${a.answers.positive ? 'Defends' : 'Attacks'} <button class="link" data-focus="${esc(G.arguments[a.answers.target]?.conclusion)}">${esc(a.answers.target)}</button> on the ${esc(a.answers.cq_name)} (${esc(a.answers.locus)})</div>` : '';
  const premises = a.premises.map(p => { const ps = G.statements[p]; return `<div class="premise" data-focus="${esc(p)}"><span class="dot ${ps ? ps.label : ''}"></span><span class="ptxt">${rich(ps ? ps.text : p)}</span></div>`; }).join('');
  const exIn = (a.exceptions || []).filter(e => G.statements[e]?.label === 'in');
  const warn = !a.in_force && a.why_out ? `<div class="warn">✗ ${esc(a.why_out)}</div>` : '';
  const exc = exIn.map(e => {
    const by = (G.statements[e].pro || []).map(x => `${x} (${G.arguments[x] ? G.arguments[x].owner : ''})`).join(', ');
    return `<div class="warn" data-focus="${esc(e)}" style="cursor:pointer">⚡ exception holds: ${esc(stmtText(e))}${by ? ` — argued by ${esc(by)}` : ''}</div>`;
  }).join('');
  const ev = (a.evidence || []).map(x => /^https?:\/\//.test(x) ? `<a href="${esc(x)}" target="_blank" rel="noopener">${esc(x)}</a>` : esc(x)).join('<br>');
  const chips = a.cqs.map(c => {
    const st = c.open ? 'asked' : c.attacks ? 'attacked' : c.defences ? 'defended' : '';
    const icons = (c.open ? ' ?' : '') + (c.attacks ? ` ⚔${c.attacks.length}` : '') + (c.defences ? ` 🛡${c.defences.length}` : '');
    return `<button class="cq ${st}" data-cq="${esc(c.cq)}" data-arg="${esc(id)}" title="${esc(c.question)}">${esc(c.name.replace(/ Question$/, ''))}${icons}</button>`;
  }).join('');
  const panel = ui.openCQ && ui.openCQ.arg === id ? cqPanel(a, ui.openCQ.cq) : '';
  return `<div class="argcard ${side} ${a.in_force ? '' : 'outforce'}">
    <div class="head"><span class="scheme ${a.free ? 'free' : ''}">${esc(a.scheme_name)}</span><span class="muted">by ${esc(a.owner)} · w ${a.base_weight.toFixed(2)}</span>${status}<span class="muted mono">${esc(id)}</span>${a.moderation ? `<span class="pill small">${esc(a.moderation)}</span>` : ''}${a.kialo ? `<span class="pill small" title="Kialo location ${esc(a.kialo.location)}">Kialo</span>` : ''}</div>
    ${answers}
    ${a.text && !a.kialo ? `<div class="nl">“${rich(a.text)}”</div>` : ''}
    ${premises}
    <div class="therefore">∴ ${esc(short(stmtText(a.conclusion), 120))}</div>
    ${warn}${exc}
    ${ev ? `<div class="ev">Sources: ${ev}</div>` : ''}
    ${chips ? `<div class="cqs">${chips}</div>` : ''}${panel}
  </div>`;
}

function cqPanel(a, cqid) {
  const c = a.cqs.find(x => x.cq === cqid); if (!c) return '';
  const stmts = (c.stmts || []).map(s => `<li><button class="link" data-focus="${esc(s)}">${esc(stmtText(s) || s)}</button></li>`).join('');
  const hist = [c.asked_by ? `asked by ${esc(c.asked_by)}${c.open ? ' (open)' : ''}` : '',
    c.attacks ? 'attacked by ' + c.attacks.map(x => `<button class="link" data-goarg="${esc(x)}">${esc(x)}</button>`).join(', ') : '',
    c.defences ? 'defended by ' + c.defences.map(x => `<button class="link" data-goarg="${esc(x)}">${esc(x)}</button>`).join(', ') : ''].filter(Boolean).join('; ');
  const can = k => k !== 'question' || c.locus === 'premise' || c.locus === 'argument';
  const btn = (k, label) => `<button class="small" data-act="${k}" data-arg="${esc(a.id)}" data-cqid="${esc(cqid)}" ${G.can_move && can(k) ? '' : 'disabled'} ${can(k) ? '' : 'title="Raise an exception or a counterclaim with an argument: whoever raises it must prove it"'}>${label}</button>`;
  return `<div class="cqpanel"><b>${esc(c.name)}</b> <span class="muted">— about the ${esc(c.locus)}</span><div>${esc(c.question)}</div>
    ${stmts ? `<ul class="small">${stmts}</ul>` : ''}${hist ? `<div class="small muted">${hist}</div>` : ''}
    <div class="row">${btn('attack', 'Attack: answer against')}${btn('defend', 'Defend: answer for')}${btn('question', 'Ask (no answer)')}</div>
    ${G.can_move ? '' : '<div class="small muted">Wait for your turn to act on it.</div>'}</div>`;
}

function boardClick(e) {
  const act = e.target.closest('[data-act]');
  if (act) { startTargeted(act.dataset.act, act.dataset.arg, act.dataset.cqid); return; }
  const cq = e.target.closest('[data-cq]');
  if (cq) { const k = { arg: cq.dataset.arg, cq: cq.dataset.cq }; ui.openCQ = ui.openCQ && ui.openCQ.arg === k.arg && ui.openCQ.cq === k.cq ? null : k; renderBoard(); return; }
  const ga = e.target.closest('[data-goarg]');
  if (ga) { const a = G.arguments[ga.dataset.goarg]; if (a) focusOn(a.conclusion); return; }
  const d = e.target.closest('[data-decide]');
  if (d) { decide(d.dataset.decide); return; }
  const f = e.target.closest('[data-focus]');
  if (f && f.dataset.focus) focusOn(f.dataset.focus);
}

async function decide(decision) {
  const note = decision === 'undecided' ? '' : prompt(`Why is it ${decision}? (optional note)`) || '';
  try { const r = await api(`/api/games/${G.id}/decide`, { statement: ui.focus, decision, note }); setState(r.state); }
  catch (e) { toast(e.message); }
}

// ---------------------------------------------------------------------------
// Side: outcome, moderator, composer, questions, log
// ---------------------------------------------------------------------------
function renderSide() {
  const side = $('#side'); if (!side) return;
  const parts = [];
  if (G.over) parts.push(outcomePanel());
  if (G.pending) parts.push(pendingPanel());
  if (!G.over) parts.push('<div class="panel" id="composer"></div>');
  if (G.questions && G.questions.length) parts.push(`<div class="panel"><h2>Open questions</h2><ul class="small">${G.questions.map(q => `<li><b>${esc(q.asker)}</b> asks <b>${esc(q.owner)}</b> about <button class="link" data-goarg="${esc(q.target)}">${esc(q.target)}</button>: ${esc(q.question)}${q.turns_left ? ` <span class="req">(${q.turns_left} turn(s) left)</span>` : ''}</li>`).join('')}</ul></div>`);
  parts.push(`<div class="panel"><h2>History</h2><div class="log"><ol>${G.log.map(l => `<li class="${l.accepted ? '' : 'rej'}"><span class="r">${l.round}.</span> ${esc(l.text)}${l.deltas && Object.keys(l.deltas).length ? ' <span class="muted">(' + Object.entries(l.deltas).map(([k, v]) => `${esc(k)} ${v > 0 ? '+' : ''}${v.toFixed(1)}`).join(', ') + ')</span>' : ''}${l.nl ? `<div class="muted small">“${esc(short(l.nl, 160))}”</div>` : ''}</li>`).join('') || '<li class="muted">No moves yet.</li>'}</ol></div></div>`);
  side.innerHTML = parts.join('');
  side.onclick = e => {
    const ga = e.target.closest('[data-goarg]'); if (ga) { const a = G.arguments[ga.dataset.goarg]; if (a) focusOn(a.conclusion); }
  };
  if (G.pending) wirePending();
  if (!G.over) renderComposer();
  const log = $('.log', side); if (log) log.scrollTop = log.scrollHeight;
}

function outcomePanel() {
  const o = G.outcome || {};
  const win = o.winner ? `<b>${esc(o.winner)}</b> wins` : (o.draw ? 'A draw' : (G.players.length === 1 ? 'Solo game' : 'No winner'));
  return `<div class="panel outcome"><h2>Game over</h2><p>${win}. ${esc(o.reason || '')}</p>${o.claim ? `<p>Claim: ${esc(stmtText(o.claim))} — ${lbl((o.verdict || 'out').toLowerCase())}</p>` : ''}</div>`;
}

function pendingPanel() {
  const p = G.pending, a = p.action;
  const sc = SCHEME[a.scheme_id];
  const can = ui.moderator;
  return `<div class="panel moderator"><h2>Awaiting the moderator</h2>
    <p><b>${esc(p.player)}</b>: ${esc(p.description || a.type)}</p>
    ${a.text ? `<p class="muted">“${esc(a.text)}”</p>` : ''}
    ${a.evidence && a.evidence.length ? `<p class="small">Sources: ${a.evidence.map(esc).join('; ')}</p>` : '<p class="small muted">No sources cited.</p>'}
    ${can ? `<label>Weight (the scheme's is ${p.weight.toFixed(2)})</label><input type="number" id="mweight" min="0.05" max="1" step="0.05" value="${p.weight || 0.5}">
    <label>Note</label><input type="text" id="mnote" placeholder="Reason (shown in the history)">
    <p><button class="primary" id="mapprove">Approve</button> <button id="mreject">Reject</button></p>`
      : `<p class="small muted">Tick “I am the moderator” at the top to decide${sc ? '' : ''}. Another browser tab on this game can act as the moderator.</p>`}</div>`;
}
function wirePending() {
  const ap = $('#mapprove'); if (!ap) return;
  const go = async approve => {
    try { const r = await api(`/api/games/${G.id}/moderate`, { approve, weight: +$('#mweight').value || 0, note: $('#mnote').value }); setState(r.state); }
    catch (e) { toast(e.message); }
  };
  ap.onclick = () => go(true); $('#mreject').onclick = () => go(false);
}

// ---------------------------------------------------------------------------
// Composer
// ---------------------------------------------------------------------------
function newComposer(kind) {
  return { kind, scheme: '', text: '', vars: {}, evidence: '', target: '', opt: '', nP: 2, suggestions: null, preview: null };
}

function startTargeted(kind, arg, cqid) {
  ui.c = newComposer(kind === 'defend' ? 'defend' : kind);
  ui.c.target = arg; ui.c.wantCQ = cqid;
  renderComposer();
  $('#composer').scrollIntoView({ behavior: 'smooth', block: 'start' });
}
const optKey = o => o.cq + '|' + (o.stmt || '');

function renderComposer() {
  const box = $('#composer'); if (!box) return;
  const me = G.players.find(p => p.id === G.active);
  if (G.pending) { box.innerHTML = `<h2>${esc(G.active)}</h2><p class="muted">The move waits for the moderator.</p>`; return; }
  if (G.active_ai) {
    box.innerHTML = `<div class="who">${esc(G.active)} <span class="muted">(${esc(G.active_ai)} AI)</span> is to move</div>
      <p class="small muted">The AI plays through the same engine: it tries moves and keeps the one that leaves it best placed.</p>
      <label class="inline"><input type="checkbox" id="autoplay" ${ui.autoplay ? 'checked' : ''}> Play automatically</label>
      <p><button id="aistep">Let it move now</button></p>`;
    $('#autoplay').onchange = e => { ui.autoplay = e.target.checked; scheduleAI(); };
    $('#aistep').onclick = aiStep;
    return;
  }
  const c = ui.c || (ui.c = newComposer('argue'));
  const kinds = [['argue', 'Argue'], ['attack', 'Attack'], ['defend', 'Defend'], ['question', 'Question'], ['pass', 'Pass'], ['resign', 'Resign']];
  const owes = (G.questions || []).filter(q => q.owner === G.active);
  box.innerHTML = `<div class="who">${esc(G.active)} to move${me && me.side ? ` <span class="pill ${me.side}">${me.side}</span>` : ''}</div>
    ${G.last_rejection ? `<div class="warn">Last move rejected: ${esc(G.last_rejection)}</div>` : ''}
    ${owes.length ? `<div class="warn">You owe an answer: ${owes.map(q => `${esc(q.question)} (about ${esc(q.target)})`).join('; ')}</div>` : ''}
    <div class="seg">${kinds.map(([k, l]) => `<button data-kind="${k}" class="${c.kind === k ? 'on' : ''}">${l}</button>`).join('')}</div>
    <div id="cbody"></div>`;
  box.querySelector('.seg').onclick = e => { const b = e.target.closest('[data-kind]'); if (!b) return; ui.c = newComposer(b.dataset.kind); renderComposer(); };
  const body = $('#cbody');
  if (c.kind === 'argue') return composeArgue(body, c);
  if (c.kind === 'pass') return composeSimple(body, 'Pass this turn. In a debate, passing while the verdict is against your side concedes.', { type: 'abstain' }, 'Pass');
  if (c.kind === 'resign') return composeSimple(body, 'Resign: you lose the game (with three or more players on points, the others play on).', { type: 'resign' }, 'Resign');
  return composeTargeted(body, c);
}

function composeSimple(body, text, action, label) {
  body.innerHTML = `<p>${esc(text)}</p><button class="primary" id="play1">${label}</button>`;
  $('#play1').onclick = () => submit(action);
}

function schemeOptions(ids, selected) {
  const official = ids.filter(i => !SCHEME[i].free), other = ids.filter(i => SCHEME[i].free);
  const opt = i => `<option value="${esc(i)}" ${i === selected ? 'selected' : ''}>${esc(SCHEME[i].name)} (${SCHEME[i].base_weight.toFixed(2)})</option>`;
  return `<option value="">— choose a scheme —</option><optgroup label="Walton schemes (NASCQ)">${official.map(opt).join('')}</optgroup>${other.length ? `<optgroup label="Other">${other.map(opt).join('')}</optgroup>` : ''}`;
}

function templateHTML(t) { return esc(t).replace(/\{(\w+)\}/g, '<span class="var">$1</span>'); }

function schemeCard(sc, cq) {
  const tmpl = `<div class="tmpl">${sc.premise_templates_en.filter((p, i) => !sc.free || i < 2).map(p => '<div>' + templateHTML(p) + '</div>').join('')}<div>∴ ${templateHTML(sc.conclusion_template_en)}</div></div>`;
  let why = '';
  if (cq) {
    const linked = (CQ[cq].backed_by || []).includes(sc.id);
    why = `<div class="small">${linked ? '🔗 linked in the NASCQ: this scheme backs answers to this question' : sc.thesis_independent ? '◎ thesis-independent: it can back an answer to any question' : ''}
      · semantic types: question <b>${esc(CQ[cq].semantic_type)}</b>, scheme <b>${esc(sc.semantic_type)}</b>${CQ[cq].semantic_type === sc.semantic_type ? ' ✓' : ''}</div>`;
  }
  const ev = sc.evidence ? `<div class="small">${sc.evidence === 'required' ? '<span class="req">Evidence required</span>' : 'Evidence recommended'}: ${esc(sc.evidence_note_en)}</div>` : '';
  return `<div class="schemecard"><b>${esc(sc.name)}</b> <span class="muted small">${esc(sc.id)} · ${esc(sc.semantic_type)}${sc.thesis_independent ? ' · thesis-independent' : ''}</span>
    <div class="small">${esc(sc.description_en || '')}</div>${tmpl}${why}${ev}</div>`;
}

function varInputs(sc, c, answerVar) {
  if (!sc) return '';
  if (sc.free) {
    let h = '';
    for (let i = 1; i <= c.nP; i++) h += `<label>Premise ${i}</label><input type="text" data-var="P${i}" value="${esc(c.vars['P' + i] || '')}" placeholder="In your own words" list="stmtlist">`;
    if (c.nP < 6) h += '<button class="small" id="addp">Add a premise</button>';
    if (!answerVar) h += `<label>Conclusion</label><input type="text" data-var="C" value="${esc(c.vars.C || '')}" placeholder="What follows" list="stmtlist">`;
    return h + stmtDatalist();
  }
  return sc.variable_order.filter(v => v !== answerVar).map(v => {
    const spec = sc.variables[v];
    return `<label><span class="var">${esc(v)}</span> ${esc(spec.description)} <span class="muted">(${esc(spec.type)})</span></label><input type="text" data-var="${esc(v)}" value="${esc(c.vars[v] || '')}">`;
  }).join('');
}
function stmtDatalist() {
  return `<datalist id="stmtlist">${Object.values(G.statements).filter(s => !s.id.startsWith('undercut(')).slice(0, 300).map(s => `<option value="${esc(s.id.startsWith('text_') ? s.text : s.id)}">${esc(short(s.text, 80))}</option>`).join('')}</datalist>`;
}

function composeArgue(body, c) {
  const ids = META.schemes.filter(s => !s.free || G.config.allow_other).map(s => s.id);
  const sc = SCHEME[c.scheme];
  body.innerHTML = `
    <label>1. Your argument, in your own words</label>
    <textarea id="nl" placeholder="e.g. According to the IPCC, the climate experts, human activity is warming the planet.">${esc(c.text)}</textarea>
    <button class="small" id="suggest">Which scheme fits?</button>
    <div class="suggest" id="sugg">${(c.suggestions || []).map(s => `<button class="small" data-pick="${esc(s.scheme_id)}" title="matched: ${esc(s.matched.join(', '))}">${esc(s.name)}</button>`).join('') || (c.suggestions ? '<span class="small muted">No cue words matched: choose a scheme yourself' + (G.config.allow_other ? ', or “Other”' : '') + '.</span>' : '')}</div>
    <label>2. The scheme it fits (official Walton schemes)</label>
    <select id="scheme">${schemeOptions(ids, c.scheme)}</select>
    ${sc ? schemeCard(sc) : ''}
    ${sc ? '<label>3. Break it into the scheme\'s parts</label>' + varInputs(sc, c, '') : ''}
    ${sc ? evidenceInput(sc, c) : ''}
    <div id="preview"></div>
    <p><button class="primary" id="play1" disabled>Put forward this argument</button></p>`;
  wireComposer(body, c, () => buildArgue(c));
  $('#suggest').onclick = async () => {
    c.text = $('#nl').value;
    try { c.suggestions = (await api('/api/suggest', { text: c.text })).suggestions || []; renderComposer(); }
    catch (e) { toast(e.message); }
  };
  body.querySelector('#sugg').onclick = e => { const b = e.target.closest('[data-pick]'); if (b) { c.scheme = b.dataset.pick; renderComposer(); } };
}

function evidenceInput(sc, c) {
  const req = sc.evidence === 'required' && G.config.require_evidence;
  return `<label>Sources ${req ? '<span class="req">(required)</span>' : '(optional)'}: one per line</label><textarea id="ev" placeholder="${esc(sc.evidence_note_en || 'A citation or a link')}">${esc(c.evidence)}</textarea>`;
}

function buildArgue(c) {
  if (!c.scheme) return null;
  return { type: 'support', scheme_id: c.scheme, variable_bindings: { ...c.vars }, text: c.text, evidence: c.evidence.split('\n').map(x => x.trim()).filter(Boolean) };
}

// moveOptions fetches (once per game version) the critical questions the
// player to move can raise or answer about an argument.
async function moveOptions(target) {
  const key = G.id + '|' + G.version + '|' + target;
  ui.opts = ui.opts || {};
  if (!ui.opts[key]) ui.opts[key] = await api(`/api/games/${G.id}/moves?target=${encodeURIComponent(target)}`);
  return ui.opts[key];
}

function composeTargeted(body, c) {
  const kind = c.kind; // attack, defend, question
  const fits = o => kind === 'question' ? o.question : o.schemes.length > 0;
  const tlist = Object.values(G.arguments).filter(a => a.role === 'main' && a.cqs.length).map(a => a.id)
    .sort((x, y) => (+x.slice(1)) - (+y.slice(1)));
  if (!tlist.length) { body.innerHTML = `<p class="muted">Nothing to ${kind} yet: put forward an argument first.</p>`; return; }
  if (!tlist.includes(c.target)) { c.target = tlist[tlist.length - 1]; c.opt = ''; }
  const key = G.id + '|' + G.version + '|' + c.target;
  if (!(ui.opts && ui.opts[key])) {
    body.innerHTML = '<p class="muted">Loading the critical questions…</p>';
    moveOptions(c.target).then(() => { if (ui.c === c) renderComposer(); }).catch(e => toast(e.message));
    return;
  }
  const opts = ui.opts[key].filter(fits);
  if (!opts.length) {
    body.innerHTML = `<label>1. The argument</label><select id="target">${tlist.map(t => `<option value="${esc(t)}" ${t === c.target ? 'selected' : ''}>${esc(argTitle(t))}: ${esc(short(stmtText(G.arguments[t].conclusion), 60))}</option>`).join('')}</select>
      <p class="muted">No critical question about ${esc(c.target)} can be ${kind === 'question' ? 'asked alone (exceptions and counterclaims need an argument)' : 'answered'} now.</p>`;
    $('#target').onchange = e => { c.target = e.target.value; c.opt = ''; renderComposer(); };
    return;
  }
  if (c.wantCQ) { const w = opts.find(x => x.cq === c.wantCQ); if (w) c.opt = optKey(w); c.wantCQ = ''; }
  if (!opts.some(o => optKey(o) === c.opt)) c.opt = optKey(opts[0]);
  const o = opts.find(x => optKey(x) === c.opt);
  const q = CQ[o.cq];
  const a = G.arguments[c.target];
  const answer = kind === 'question' ? '' : kind === 'attack' ? o.attack_text : o.defence_text;
  const schemes = o.schemes;
  if (kind !== 'question' && !schemes.includes(c.scheme)) c.scheme = '';
  const sc = SCHEME[c.scheme];
  const what = { attack: 'an answer against it (a negative answer)', defend: 'an answer in its favour (a positive answer)', question: 'the question alone: the proponent must answer it' }[kind];
  body.innerHTML = `
    <label>1. The argument</label>
    <select id="target">${tlist.map(t => `<option value="${esc(t)}" ${t === c.target ? 'selected' : ''}>${esc(argTitle(t))}: ${esc(short(stmtText(G.arguments[t].conclusion), 60))}</option>`).join('')}</select>
    <div class="small muted">∴ ${esc(stmtText(a.conclusion))}</div>
    <label>2. The critical question</label>
    <select id="opt">${opts.map(x => `<option value="${esc(optKey(x))}" ${optKey(x) === c.opt ? 'selected' : ''}>${esc(CQ[x.cq].name)}${x.stmt ? ' — ' + esc(short(stmtText(x.stmt), 50)) : ''} (${esc(x.locus)})</option>`).join('')}</select>
    <div class="answerbox"><b>${esc(questionOf(a, o.cq))}</b><div class="small">About the ${esc(o.locus)}${o.stmt ? ': ' + esc(stmtText(o.stmt).replace(/\.$/, '')) : ''}. You give ${what}.</div>
      ${answer ? `<div>Your answer: <b>${esc(plain(answer))}</b></div>` : ''}</div>
    ${kind === 'question' ? '' : `
      <label>3. The scheme that backs your answer (from the NASCQ links)</label>
      <select id="scheme">${schemeOptions(schemes, c.scheme)}</select>
      ${sc ? schemeCard(sc, o.cq) : ''}
      ${sc ? `<label>4. In your own words (optional)</label><textarea id="nl">${esc(c.text)}</textarea>` : ''}
      ${sc ? '<label>5. The scheme\'s parts' + (sc.answer_variable ? ` (<span class="var">${esc(sc.answer_variable)}</span> is your answer)` : '') + '</label>' + varInputs(sc, c, sc.answer_variable) : ''}
      ${sc ? evidenceInput(sc, c) : ''}`}
    <div id="preview"></div>
    <p><button class="primary" id="play1" disabled>${{ attack: 'Attack', defend: 'Defend', question: 'Ask' }[kind]}</button></p>`;
  $('#target').onchange = e => { c.target = e.target.value; c.opt = ''; renderComposer(); };
  $('#opt').onchange = e => { c.opt = e.target.value; renderComposer(); };
  wireComposer(body, c, () => buildTargeted(c, o));
}

function questionOf(a, cqid) { const x = a.cqs.find(y => y.cq === cqid); return x ? x.question : CQ[cqid].question_template_en; }

function buildTargeted(c, o) {
  const base = { answers_cq_id: o.cq, target_arg_id: c.target, target_premise_id: o.locus === 'premise' ? o.stmt : '' };
  if (c.kind === 'question') return { type: 'question', ...base };
  if (!c.scheme) return null;
  return { type: c.kind === 'defend' ? 'support' : 'attack', scheme_id: c.scheme, ...base, variable_bindings: { ...c.vars }, text: c.text, evidence: c.evidence.split('\n').map(x => x.trim()).filter(Boolean) };
}

function wireComposer(body, c, build) {
  const sel = $('#scheme', body);
  if (sel) sel.onchange = e => { c.scheme = e.target.value; c.vars = {}; renderComposer(); };
  const nl = $('#nl', body); if (nl) nl.oninput = () => { c.text = nl.value; };
  const ev = $('#ev', body); if (ev) ev.oninput = () => { c.evidence = ev.value; schedulePreview(c, build); };
  body.querySelectorAll('[data-var]').forEach(inp => inp.oninput = () => { c.vars[inp.dataset.var] = inp.value; schedulePreview(c, build); });
  const addp = $('#addp', body); if (addp) addp.onclick = () => { c.nP++; renderComposer(); };
  $('#play1', body).onclick = () => { const a = build(); if (a) submit(a); };
  schedulePreview(c, build, 0);
}

function schedulePreview(c, build, delay = 350) {
  clearTimeout(schedulePreview.t);
  schedulePreview.t = setTimeout(async () => {
    const a = build(); const box = $('#preview'); const btn = $('#play1');
    if (!box) return;
    if (!a) { box.innerHTML = ''; btn.disabled = true; return; }
    const seq = ++ui.previewSeq;
    let p;
    try { p = await api(`/api/games/${G.id}/preview`, { action: a }); } catch (e) { p = { ok: false, error: e.message }; }
    if (seq !== ui.previewSeq || !$('#preview')) return;
    const changes = (p.changes || []).slice(0, 8).map(x => `<div class="chg">${lbl(x.before === 'new' ? 'undecided' : x.before)} → ${lbl(x.after)} ${esc(short(x.text, 70))}</div>`).join('');
    const deltas = p.deltas && Object.keys(p.deltas).length ? `<div class="small">Scores: ${Object.entries(p.deltas).map(([k, v]) => `${esc(k)} ${v > 0 ? '+' : ''}${v.toFixed(1)}`).join(', ')}</div>` : '';
    box.innerHTML = `<div class="preview ${p.ok ? 'ok' : 'bad'}">
      ${p.premises && p.premises.length ? '<b>You argue</b><ol>' + p.premises.map(x => `<li>${rich(x)}</li>`).join('') + '</ol>' : ''}
      ${p.effect ? `<div>∴ <b>${rich(p.effect)}</b></div>` : ''}
      ${p.ok ? `<div class="small" style="margin-top:6px">${changes ? 'What changes:' : 'Legal move.'}</div>${changes}${deltas}` : `<div class="warn">${esc(p.error)}</div>`}</div>`;
    btn.disabled = !p.ok;
  }, delay);
}

async function submit(action) {
  if (ui.busy) return;
  ui.busy = true;
  try {
    const r = await api(`/api/games/${G.id}/move`, { action });
    const res = r.result || {};
    if (res.pending) toast('Sent to the moderator.');
    else if (!res.accepted) toast('Rejected: ' + (res.error || ''), 6000);
    else if (res.arg_ids && res.arg_ids.length) {
      const a = r.state.arguments[res.arg_ids[0]];
      if (a) { ui.focus = a.conclusion; ui.crumbs = [...ui.crumbs, a.conclusion].slice(-8); }
    }
    if (res.accepted) ui.c = null;
    setState(r.state);
  } catch (e) { toast(e.message, 6000); }
  finally { ui.busy = false; }
}

// ---------------------------------------------------------------------------
// The NASCQ: schemes, critical questions and their links
// ---------------------------------------------------------------------------
function renderNASCQ() {
  const sel = ui.nsel || (ui.nsel = { kind: 'scheme', id: META.schemes[0].id });
  const list = (items, kind) => items.map(x => `<button class="${sel.kind === kind && sel.id === x.id ? 'on' : ''}" data-n="${kind}" data-id="${esc(x.id)}">${esc(x.name)}${x.optional ? ' <span class="muted small">(optional)</span>' : ''}</button>`).join('');
  $('#main').innerHTML = `<div class="nascq">
    <div><div class="panel"><h2>Schemes</h2><div class="list">${list(META.schemes, 'scheme')}</div>
      <h2 style="margin-top:14px">Critical questions</h2><div class="list">${list(META.cqs, 'cq')}</div></div></div>
    <div><div class="panel"><h2>The network</h2>${networkSVG(sel)}
      <div class="legend small"><span><i class="locus-premise" style="background:#3157d5"></i>asks about a premise</span><span><i style="background:#c4302b"></i>an exception</span><span><i style="background:#9a6700"></i>the conclusion</span><span><i style="background:#8250df"></i>the inference</span><span><i style="border-top:2px dashed #1a7f37;height:0"></i>backs answers (link)</span></div>
      <p class="small muted">Thesis-independent schemes back every question; their links are not drawn. Click a node to see it.</p></div>
      <div class="panel">${sel.kind === 'scheme' ? schemeDetail(SCHEME[sel.id]) : cqDetail(CQ[sel.id])}</div></div></div>`;
  $('#main').onclick = e => {
    const b = e.target.closest('[data-n]');
    if (b) { ui.nsel = { kind: b.dataset.n, id: b.dataset.id }; renderNASCQ(); }
  };
}

function networkSVG(sel) {
  const S = META.schemes, Q = META.cqs;
  const rowH = 26, h = Math.max(S.length, Q.length) * rowH + 20, w = 760, xs = 190, xq = 560;
  const ys = {}, yq = {};
  S.forEach((s, i) => (ys[s.id] = 20 + i * (h - 40) / Math.max(1, S.length - 1)));
  Q.forEach((q, i) => (yq[q.id] = 14 + i * rowH));
  const hot = (s, q) => (sel.kind === 'scheme' && sel.id === s) || (sel.kind === 'cq' && sel.id === q);
  let edges = '';
  for (const q of Q) for (const a of q.asked_of) {
    const cls = `edge locus-${a.locus} ${hot(a.scheme, q.id) ? 'hot' : 'dim'}`;
    edges += `<path class="${cls}" d="M${xs + 6},${ys[a.scheme]} C${xs + 160},${ys[a.scheme]} ${xq - 160},${yq[q.id]} ${xq - 6},${yq[q.id]}"><title>${esc(q.name)} asked of ${esc(SCHEME[a.scheme].name)} (${esc(a.locus)}${a.pattern ? ': ' + esc(a.pattern) : ''})</title></path>`;
  }
  for (const q of Q) for (const sid of q.backed_by) {
    const cls = `edge backup ${hot(sid, q.id) ? 'hot' : 'dim'}`;
    edges += `<path class="${cls}" d="M${xs + 6},${ys[sid] + 4} C${xs + 200},${ys[sid] + 40} ${xq - 200},${yq[q.id] + 30} ${xq - 6},${yq[q.id] + 3}"><title>${esc(SCHEME[sid].name)} backs answers to ${esc(q.name)}</title></path>`;
  }
  const nodes = S.map(s => `<g data-n="scheme" data-id="${esc(s.id)}" style="cursor:pointer"><circle cx="${xs}" cy="${ys[s.id]}" r="5" fill="${s.thesis_independent ? '#1a7f37' : '#3157d5'}"></circle><text x="${xs - 10}" y="${ys[s.id] + 4}" text-anchor="end" font-weight="${sel.id === s.id ? 700 : 400}">${esc(short(s.name.replace(/^Argument from /, ''), 26))}</text></g>`).join('')
    + Q.map(q => `<g data-n="cq" data-id="${esc(q.id)}" style="cursor:pointer"><rect x="${xq - 4}" y="${yq[q.id] - 4}" width="8" height="8" fill="#8250df"></rect><text x="${xq + 10}" y="${yq[q.id] + 4}" font-weight="${sel.id === q.id ? 700 : 400}">${esc(short(q.name.replace(/ Question$/, ''), 28))}</text></g>`).join('');
  return `<div class="netwrap"><svg width="${w}" height="${h}" viewBox="0 0 ${w} ${h}">${edges}${nodes}</svg></div>`;
}

function schemeDetail(s) {
  if (!s) return '';
  const vars = s.variable_order.map(v => `<tr><td><span class="var">${esc(v)}</span></td><td>${esc(s.variables[v].type)}</td><td>${esc(s.variables[v].description)}</td></tr>`).join('');
  const asked = s.asked.map(a => `<tr><td><button class="link" data-n="cq" data-id="${esc(a.cq)}">${esc(CQ[a.cq].name)}</button></td><td>${esc(a.locus)}</td><td class="mono">${esc(a.pattern || '')}</td><td>${esc(CQ[a.cq].question_template_en)}</td></tr>`).join('');
  const cn = s.carneades;
  return `<h2>${esc(s.name)} <span class="muted small mono">${esc(s.id)}</span></h2>
    <p>${esc(s.description_en || '')}</p>
    <p class="small">Semantic type <b>${esc(s.semantic_type)}</b> · weight ${s.base_weight.toFixed(2)}${s.thesis_independent ? ' · thesis-independent (backs every question)' : ''}${s.answer_variable ? ` · answer variable <span class="var">${esc(s.answer_variable)}</span>` : ''}${s.evidence ? ` · evidence ${esc(s.evidence)}: ${esc(s.evidence_note_en)}` : ''}</p>
    <div class="schemecard tmpl">${s.premise_templates_en.map(p => '<div>' + templateHTML(p) + '</div>').join('')}<div>∴ ${templateHTML(s.conclusion_template_en)}</div></div>
    <h3>Variables</h3><table>${vars}</table>
    <h3>Critical questions asked of it</h3><table><tr><th>Question</th><th>About</th><th>Formula</th><th>Text</th></tr>${asked || '<tr><td colspan=4 class="muted">None</td></tr>'}</table>
    <h3>Backs answers to</h3><p>${s.thesis_independent ? 'every question (thesis-independent)' : s.backs.map(q => `<button class="link" data-n="cq" data-id="${esc(q)}">${esc(CQ[q].name)}</button>`).join(', ') || 'no question yet'}</p>
    ${cn ? `<h3>Carneades scheme ${esc(cn.id)}</h3><table>
      <tr><th>Premises</th><td class="mono">${(cn.premises || []).map(esc).join('<br>')}</td></tr>
      <tr><th>Assumptions</th><td class="mono">${(cn.assumptions || []).map(esc).join('<br>') || '—'}</td></tr>
      <tr><th>Exceptions</th><td class="mono">${(cn.exceptions || []).map(esc).join('<br>') || '—'}</td></tr>
      <tr><th>Conclusion</th><td class="mono">${(cn.conclusions || []).map(esc).join('<br>')}</td></tr></table>` : ''}
    ${s.cues_en && s.cues_en.length ? `<h3>Cue words (scheme suggestions)</h3><p class="small">${s.cues_en.map(esc).join(', ')}</p>` : ''}`;
}

function cqDetail(q) {
  if (!q) return '';
  const asked = q.asked_of.map(a => `<tr><td><button class="link" data-n="scheme" data-id="${esc(a.scheme)}">${esc(SCHEME[a.scheme].name)}</button></td><td>${esc(a.locus)}</td><td class="mono">${esc(a.pattern || '')}</td></tr>`).join('');
  const ti = META.schemes.filter(s => s.thesis_independent).map(s => s.name);
  return `<h2>${esc(q.name)} <span class="muted small mono">${esc(q.id)}</span></h2>
    <p><b>${esc(q.question_template_en)}</b></p>
    <p class="small">Semantic type <b>${esc(q.semantic_type)}</b>. A negative answer ${esc(q.caes_mapping.negative_answer_type)}s; a positive answer supports.</p>
    <h3>Asked of</h3><table><tr><th>Scheme</th><th>About</th><th>Formula</th></tr>${asked}</table>
    <h3>Answers backed by</h3><p>${q.backed_by.map(s => `<button class="link" data-n="scheme" data-id="${esc(s)}">${esc(SCHEME[s].name)}</button> (linked)`).join(', ')}${q.backed_by.length ? ', and ' : ''}the thesis-independent schemes: ${ti.map(esc).join(', ')}.</p>`;
}

// ---------------------------------------------------------------------------
// How it works
// ---------------------------------------------------------------------------
function renderRules() {
  $('#main').innerHTML = `<div class="rules panel">
  <h2>How WAG works</h2>
  <p><b>Arguments are built from Walton's argumentation schemes.</b> You may write your argument in your own words first; then you choose the scheme it fits from the official list (the NASCQ) and break your words into the scheme's parts, its variables. The engine instantiates the scheme in Carneades: its premises, its assumptions and its exceptions become statements on the board, and every statement is labelled IN, OUT or UNDECIDED after every move.</p>
  <p><b>Every scheme has critical questions, and each question is linked to the schemes that can back an answer to it.</b> The NASCQ (Pyon &amp; Ri 2022) records both links: <i>ask</i> edges (which questions apply to an argument made with a scheme, and which part of it each is about) and <i>backup</i> edges (which schemes can support an answer). Thesis-independent schemes (expert opinion, position to know, analogy) can back any answer. The schemes and questions tab shows the whole network.</p>
  <p><b>A critical question is about one part of the argument</b> (Gordon, Prakken &amp; Walton 2007):</p>
  <ul>
    <li>a <b>premise</b> (or assumption): it holds by presumption; asking the question makes it OUT until the proponent answers; arguing the negative answer undermines it;</li>
    <li>an <b>exception</b>: it does not hold unless someone argues it, so whoever raises it must argue for it; if it holds, the argument is undercut;</li>
    <li>the <b>conclusion</b>: the negative answer is a rebuttal, decided by weight under the proof standard;</li>
    <li>the <b>inference</b>: the negative answer undercuts the argument.</li>
  </ul>
  <p><b>Answers are arguments too.</b> A negative answer attacks, a positive answer defends, and either one is a new instance of a scheme with its own critical questions. So a dispute deepens: each answer protects the part above it until it is knocked out.</p>
  <p><b>Moves:</b> Argue (a new argument) · Attack (answer a question against an argument) · Defend (answer it in the argument's favour) · Question (ask without answering: the proponent must answer) · Pass · Resign.</p>
  <p><b>Options:</b> the <i>Other</i> scheme lets players write premises and a conclusion freely when no listed scheme fits (it has generic questions, and its arguments are exported for mining new schemes); a <i>moderator</i> approves or rejects each move, may set its weight, and may accept or reject statements (Gordon &amp; Walton 2006); <i>evidence</i> can be required for schemes such as expert opinion.</p>
  <p><b>Winning</b> depends on the mode: points (once per move, or kept live), a debate on a claim (pro carries the burden of proof; the verdict decides), a critical-question duel (an unanswered question loses), solo belief revision, and more. See <span class="mono">docs/game-rules.md</span>.</p>
  </div>`;
}

init().catch(e => { document.body.innerHTML = '<p style="padding:20px">Cannot reach the WAG server: ' + esc(e.message) + '</p>'; });
