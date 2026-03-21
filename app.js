'use strict';

/* ==========================================================
   BLOCKUDOKU TRAINER  –  app.js
   ========================================================== */

// ── Piece definitions ──────────────────────────────────────
// Each entry is an array of [row, col] offsets.
// All pieces normalised: minRow = 0, minCol = 0.

// Extended set – all 51 canonical pieces (current full set)
const PIECE_DEFS_EXTENDED = [
  // 1-cell
  [[0,0]],
  // Dominoes
  [[0,0],[0,1]],
  [[0,0],[1,0]],
  // Diagonal dominoes (2-block diagonal)
  [[0,0],[1,1]],
  [[0,1],[1,0]],
  // Straight trominoes
  [[0,0],[0,1],[0,2]],
  [[0,0],[1,0],[2,0]],
  // Diagonal trominoes (3-block diagonal)
  [[0,0],[1,1],[2,2]],
  [[0,2],[1,1],[2,0]],
  // Corner trominoes (2 canonical forms)
  [[0,0],[0,1],[1,0]],
  [[0,0],[1,0],[1,1]],
  // 2×2 square
  [[0,0],[0,1],[1,0],[1,1]],
  // Straight tetrominoes
  [[0,0],[0,1],[0,2],[0,3]],
  [[0,0],[1,0],[2,0],[3,0]],
  // L-tetromino
  [[0,0],[1,0],[2,0],[2,1]],
  // Extended corner (wide L in 2×3)
  [[0,0],[0,1],[0,2],[1,0]],
  // T-tetromino
  [[0,0],[0,1],[0,2],[1,1]],
  // S-tetromino (horizontal)
  [[0,1],[0,2],[1,0],[1,1]],
  // Step corner (vertical S)
  [[0,0],[1,0],[1,1],[2,1]],
  // Z-tetromino
  [[0,0],[0,1],[1,1],[1,2]],
  // Straight pentominoes
  [[0,0],[0,1],[0,2],[0,3],[0,4]],
  [[0,0],[1,0],[2,0],[3,0],[4,0]],

  // ── 5-cell shapes ──────────────────────────────────────
  // Plus / X-pentomino (rotationally symmetric)
  [[0,1],[1,0],[1,1],[1,2],[2,1]],
  // T-pentomino (4 orientations)
  [[0,0],[0,1],[0,2],[1,1],[2,1]],
  [[0,2],[1,0],[1,1],[1,2],[2,2]],
  [[0,1],[1,1],[2,0],[2,1],[2,2]],
  [[0,0],[1,0],[1,1],[1,2],[2,0]],
  // V-pentomino / large corner (4 orientations)
  [[0,0],[1,0],[2,0],[2,1],[2,2]],
  [[0,0],[0,1],[0,2],[1,0],[2,0]],
  [[0,0],[0,1],[0,2],[1,2],[2,2]],
  [[0,2],[1,2],[2,0],[2,1],[2,2]],
  // L-pentomino (4 orientations)
  [[0,0],[1,0],[2,0],[3,0],[3,1]],
  [[0,0],[0,1],[0,2],[0,3],[1,0]],
  [[0,0],[0,1],[1,1],[2,1],[3,1]],
  [[0,3],[1,0],[1,1],[1,2],[1,3]],
  // J-pentomino (4 orientations)
  [[0,1],[1,1],[2,1],[3,0],[3,1]],
  [[0,0],[1,0],[1,1],[1,2],[1,3]],
  [[0,0],[0,1],[1,0],[2,0],[3,0]],
  [[0,0],[0,1],[0,2],[0,3],[1,3]],
  // U-pentomino (4 orientations)
  [[0,0],[0,2],[1,0],[1,1],[1,2]],
  [[0,0],[0,1],[1,0],[2,0],[2,1]],
  [[0,0],[0,1],[0,2],[1,0],[1,2]],
  [[0,0],[0,1],[1,1],[2,0],[2,1]],
  // W-pentomino / staircase (4 orientations)
  [[0,0],[1,0],[1,1],[2,1],[2,2]],
  [[0,1],[0,2],[1,0],[1,1],[2,0]],
  [[0,0],[0,1],[1,1],[1,2],[2,2]],
  [[0,2],[1,1],[1,2],[2,0],[2,1]],
  // P-pentomino (2×2 plus one extension)
  [[0,0],[0,1],[1,0],[1,1],[2,0]],
  // F-pentomino (offset cross form)
  [[0,1],[0,2],[1,0],[1,1],[2,1]],
  // Z5-pentomino (5-block zigzag)
  [[0,0],[0,1],[1,1],[2,1],[2,2]],
  // S5-pentomino (5-block mirror of Z)
  [[0,1],[0,2],[1,1],[2,0],[2,1]],
];

// Standard set – matches original Blockudoku shapes (no complex V/W/P/F/Z5/S5 pentominoes)
// Explicitly defined (not filtered by index) for maintainability
const PIECE_DEFS_STANDARD = [
  // 1-cell
  [[0,0]],
  // Dominoes
  [[0,0],[0,1]],
  [[0,0],[1,0]],
  // Diagonal dominoes (2-block diagonal)
  [[0,0],[1,1]],
  [[0,1],[1,0]],
  // Straight trominoes
  [[0,0],[0,1],[0,2]],
  [[0,0],[1,0],[2,0]],
  // Diagonal trominoes (3-block diagonal)
  [[0,0],[1,1],[2,2]],
  [[0,2],[1,1],[2,0]],
  // Corner trominoes (2 canonical forms)
  [[0,0],[0,1],[1,0]],
  [[0,0],[1,0],[1,1]],
  // 2×2 square
  [[0,0],[0,1],[1,0],[1,1]],
  // Straight tetrominoes
  [[0,0],[0,1],[0,2],[0,3]],
  [[0,0],[1,0],[2,0],[3,0]],
  // L-tetromino
  [[0,0],[1,0],[2,0],[2,1]],
  // Extended corner (wide L in 2×3)
  [[0,0],[0,1],[0,2],[1,0]],
  // T-tetromino
  [[0,0],[0,1],[0,2],[1,1]],
  // S-tetromino (horizontal)
  [[0,1],[0,2],[1,0],[1,1]],
  // Step corner (vertical S)
  [[0,0],[1,0],[1,1],[2,1]],
  // Z-tetromino
  [[0,0],[0,1],[1,1],[1,2]],
  // Straight pentominoes
  [[0,0],[0,1],[0,2],[0,3],[0,4]],
  [[0,0],[1,0],[2,0],[3,0],[4,0]],
  // Plus / X-pentomino
  [[0,1],[1,0],[1,1],[1,2],[2,1]],
  // T-pentomino (4 orientations)
  [[0,0],[0,1],[0,2],[1,1],[2,1]],
  [[0,2],[1,0],[1,1],[1,2],[2,2]],
  [[0,1],[1,1],[2,0],[2,1],[2,2]],
  [[0,0],[1,0],[1,1],[1,2],[2,0]],
  // V-pentomino / large corner (4 orientations)
  [[0,0],[1,0],[2,0],[2,1],[2,2]],
  [[0,0],[0,1],[0,2],[1,0],[2,0]],
  [[0,0],[0,1],[0,2],[1,2],[2,2]],
  [[0,2],[1,2],[2,0],[2,1],[2,2]],
  // U-pentomino (4 orientations)
  [[0,0],[0,2],[1,0],[1,1],[1,2]],
  [[0,0],[0,1],[1,0],[2,0],[2,1]],
  [[0,0],[0,1],[0,2],[1,0],[1,2]],
  [[0,0],[0,1],[1,1],[2,0],[2,1]],
];

// Keep PIECE_DEFS as an alias so other helpers can reference the active set
let PIECE_DEFS = PIECE_DEFS_STANDARD;

const N = 9;

// ── Animation durations (ms) – keep in sync with styles.css ──
const ANIM_SLOT_SHRINK   = 200;   // matches slotShrink 0.2s
const ANIM_CLEAR         = 380;   // matches clearFlash 0.38s
const ANIM_CLEAR_STAGGER = 120;   // max ripple stagger offset
const ANIM_NO_SPACE_IN   = 700;   // "no more space" fade-in
const ANIM_NO_SPACE_HOLD = 1500;  // "no more space" hold time
const ANIM_NO_SPACE_OUT  = 800;   // "no more space" fade-out

// ── State ─────────────────────────────────────────────────
let board   = [];   // N×N of 0/1
let pieces  = [];   // current rack piece-cell-arrays
let used    = [];   // [bool …] – which slots are placed
let score   = 0;
let bestScore = 0;
let todayScore = 0;
let combo   = 0;
let gameOver = false;
let trainingMode = false;
let extendedPieces = false;
let darkMode     = false;
let colorSetting = 'orange';   // 'orange','blue','green','purple','red','teal','pink','random'
let rackSize     = 3;          // number of pieces shown in the rack (1–3)

const COLOR_NAMES = ['orange','blue','green','purple','red','teal','pink'];

// ── Piece helpers ──────────────────────────────────────────
function bounds(cells) {
  let minR = N, maxR = 0, minC = N, maxC = 0;
  for (const [r, c] of cells) {
    if (r < minR) minR = r; if (r > maxR) maxR = r;
    if (c < minC) minC = c; if (c > maxC) maxC = c;
  }
  return { minR, maxR, minC, maxC, rows: maxR - minR + 1, cols: maxC - minC + 1 };
}

function canPlace(cells, row, col) {
  for (const [dr, dc] of cells) {
    const r = row + dr, c = col + dc;
    if (r < 0 || r >= N || c < 0 || c >= N) return false;
    if (board[r][c]) return false;
  }
  return true;
}

function canPlaceAnywhere(cells) {
  for (let r = 0; r < N; r++)
    for (let c = 0; c < N; c++)
      if (canPlace(cells, r, c)) return true;
  return false;
}

// ── Arbitrary-board placement helpers (for order-checking) ─
function canPlaceOnBoard(cells, row, col, b) {
  for (const [dr, dc] of cells) {
    const r = row + dr, c = col + dc;
    if (r < 0 || r >= N || c < 0 || c >= N) return false;
    if (b[r][c]) return false;
  }
  return true;
}

// Try placing each piece (in the given slot order) at its first available
// position and return whether all can be placed.
function canFitAllInOrder(order) {
  let b = board.map(r => [...r]);
  for (const i of order) {
    let placed = false;
    outer: for (let r = 0; r < N; r++) {
      for (let c = 0; c < N; c++) {
        if (!canPlaceOnBoard(pieces[i], r, c, b)) continue;
        for (const [dr, dc] of pieces[i]) b[r + dr][c + dc] = 1;
        b = applyClears(b, getClearsOnBoard(b));
        placed = true;
        break outer;
      }
    }
    if (!placed) return false;
  }
  return true;
}

// Returns true when only some orderings allow all pieces to be placed –
// meaning the player must choose carefully which piece to play first.
function orderMatters() {
  if (rackSize <= 1) return false;
  const unplaced = Array.from({ length: rackSize }, (_, i) => i).filter(i => !used[i]);
  if (unplaced.length <= 1) return false;

  // Skip the check on nearly-empty boards – not tight enough to matter.
  const fillCount = board.reduce((sum, row) => sum + row.reduce((total, cell) => total + cell, 0), 0);
  if (fillCount < 28) return false;

  const perms = getPermutations(unplaced);
  let worksCount = 0;
  for (const order of perms) {
    if (canFitAllInOrder(order)) worksCount++;
  }
  // True only if at least one ordering works but not all do.
  return worksCount > 0 && worksCount < perms.length;
}

function randomPiece() {
  return PIECE_DEFS[Math.floor(Math.random() * PIECE_DEFS.length)];
}

// ── Difficulty-weighted piece selection ────────────────────
// As score grows, larger/harder pieces become progressively more likely.
function weightedRandomPiece() {
  // fill: 0 at score 0, 1 at score 200, reaches 1.5 at score 300
  const fill = Math.min(1.5, score / 200);
  // Each piece gets a weight = 1 + fill × (cellCount - 1) × 0.5
  // At fill=0 all weights are equal; at fill=1.5 a 5-cell piece is 4× as likely as a 1-cell piece.
  let totalWeight = 0;
  const weights = PIECE_DEFS.map(p => {
    const w = 1 + fill * (p.length - 1) * 0.5;
    totalWeight += w;
    return w;
  });
  let rand = Math.random() * totalWeight;
  for (let i = 0; i < PIECE_DEFS.length; i++) {
    rand -= weights[i];
    if (rand <= 0) return PIECE_DEFS[i];
  }
  return PIECE_DEFS[PIECE_DEFS.length - 1];
}

// ── Smart piece selection ──────────────────────────────────
function canCauseClear(cells) {
  for (let r = 0; r < N; r++) {
    for (let c = 0; c < N; c++) {
      if (!canPlace(cells, r, c)) continue;
      const tmp = board.map(row => [...row]);
      for (const [dr, dc] of cells) tmp[r + dr][c + dc] = 1;
      const clrs = getClearsOnBoard(tmp);
      if (clrs.total > 0) return true;
    }
  }
  return false;
}

// Board-agnostic version of canCauseClear (used for look-ahead on simulated boards)
function canCauseClearOnBoard(cells, b) {
  for (let r = 0; r < N; r++) {
    for (let c = 0; c < N; c++) {
      if (!canPlaceOnBoard(cells, r, c, b)) continue;
      const tmp = b.map(row => [...row]);
      for (const [dr, dc] of cells) tmp[r + dr][c + dc] = 1;
      if (getClearsOnBoard(tmp).total > 0) return true;
    }
  }
  return false;
}

// Piece well-suited to an early game (sparse board, score < 200).
// Prefers 3–4 cell pieces for meaningful scoring on an open board.
function earlyPiece() {
  const pool = PIECE_DEFS.filter(p => p.length >= 3 && p.length <= 4);
  if (pool.length === 0) return weightedRandomPiece();
  return pool[Math.floor(Math.random() * pool.length)];
}

// Verify the next `rounds` future rounds will still have clearing opportunities.
// Modifies `p` in-place if needed to maintain strategic play.
function ensureLookahead(p, rounds) {
  // Look-ahead requires at least 2 pieces to simulate meaningful future states.
  if (rackSize < 2) return;

  // Simulate placing all current rack pieces at their first available position
  let b = board.map(r => [...r]);
  for (const pc of p) {
    let placed = false;
    outer: for (let r = 0; r < N; r++) {
      for (let c = 0; c < N; c++) {
        if (!canPlaceOnBoard(pc, r, c, b)) continue;
        for (const [dr, dc] of pc) b[r + dr][c + dc] = 1;
        b = applyClears(b, getClearsOnBoard(b));
        placed = true;
        break outer;
      }
    }
    if (!placed) return; // simulation failed, skip look-ahead
  }

  // Check that future rounds still offer clearing opportunities
  for (let round = 0; round < rounds; round++) {
    const futureHasClear = PIECE_DEFS.some(pc => canCauseClearOnBoard(pc, b));
    if (!futureHasClear) {
      // Future board is too dense, inject a clear-enabling piece into current rack
      // canCauseClear(pc) implies canPlaceAnywhere(pc), so no separate placement check needed
      const clearNow = PIECE_DEFS.find(pc => canCauseClear(pc));
      if (clearNow) {
        const swapIdx = p.findIndex(pc => !canCauseClear(pc));
        if (swapIdx >= 0) p[swapIdx] = clearNow;
      }
      return;
    }

    // Advance simulation by one round of typical future pieces
    const futurePieces = Array.from({ length: rackSize }, () => weightedRandomPiece());
    for (const pc of futurePieces) {
      let placed = false;
      outer: for (let r = 0; r < N; r++) {
        for (let c = 0; c < N; c++) {
          if (!canPlaceOnBoard(pc, r, c, b)) continue;
          for (const [dr, dc] of pc) b[r + dr][c + dc] = 1;
          b = applyClears(b, getClearsOnBoard(b));
          placed = true;
          break outer;
        }
      }
      if (!placed) return; // board too full, stop simulation
    }
  }
}

function smartPieces() {
  const filled = board.reduce((s, r) => s + r.reduce((t, c) => t + c, 0), 0);
  const density = filled / (N * N);
  const earlyGame = density < 0.10 && score < 200;

  // Generate candidate rack based on difficulty
  const p = Array.from({ length: rackSize }, () =>
    earlyGame ? earlyPiece() : weightedRandomPiece()
  );

  // Guarantee at least one clearing opportunity in this rack
  if (!p.some(pc => canCauseClear(pc))) {
    const candidates = [];
    for (const pc of PIECE_DEFS) {
      if (canCauseClear(pc)) {
        candidates.push(pc);
        if (candidates.length >= 8) break; // 8 candidates give good random variety
      }
    }
    if (candidates.length > 0) {
      const slot = Math.floor(Math.random() * rackSize);
      p[slot] = candidates[Math.floor(Math.random() * candidates.length)];
    }
  }

  // Light 2-round look-ahead: keep the game strategic and clearable
  ensureLookahead(p, 2);

  // Final safety: ensure at least one piece can be placed
  if (!p.some(pc => canPlaceAnywhere(pc))) {
    outerLoop: for (let i = 0; i < rackSize; i++) {
      for (const pc of PIECE_DEFS) {
        if (canPlaceAnywhere(pc)) { p[i] = pc; break outerLoop; }
      }
    }
  }

  return p;
}

// ── Colour / theme helpers ─────────────────────────────────
function applyColor(name) {
  if (name === 'random') {
    const pick = COLOR_NAMES[Math.floor(Math.random() * COLOR_NAMES.length)];
    document.documentElement.dataset.color = pick;
  } else {
    document.documentElement.dataset.color = name;
  }
}

function applyDarkMode(on) {
  document.documentElement.dataset.theme = on ? 'dark' : '';
  document.querySelector('meta[name="theme-color"]')
    .setAttribute('content', on ? '#2f2722' : '#f6f1e8');
}

function applyExtendedPieces(on) {
  PIECE_DEFS = on ? PIECE_DEFS_EXTENDED : PIECE_DEFS_STANDARD;
}

function saveSettings() {
  localStorage.setItem('bst-settings', JSON.stringify({
    training:  trainingMode,
    extended:  extendedPieces,
    dark:      darkMode,
    color:     colorSetting,
    rackSize:  rackSize,
  }));
}

function loadSettings() {
  try {
    const s = JSON.parse(localStorage.getItem('bst-settings') || '{}');
    if (typeof s.training === 'boolean')  trainingMode   = s.training;
    if (typeof s.extended === 'boolean')  extendedPieces = s.extended;
    if (typeof s.color === 'string')      colorSetting   = s.color;
    if (typeof s.rackSize === 'number' && s.rackSize >= 1 && s.rackSize <= 3)
      rackSize = s.rackSize;
    // Respect saved dark preference; fall back to OS preference on first launch
    if (typeof s.dark === 'boolean') {
      darkMode = s.dark;
    } else {
      darkMode = window.matchMedia('(prefers-color-scheme: dark)').matches;
    }
  } catch (_) { /* ignore corrupt data */ }
}

// ── Board helpers ──────────────────────────────────────────
function emptyBoard() {
  return Array.from({ length: N }, () => new Array(N).fill(0));
}

function cellEl(r, c) {
  return document.querySelector(`#board .cell[data-r="${r}"][data-c="${c}"]`);
}

function cellSize() {
  const el = document.getElementById('board');
  return el ? el.getBoundingClientRect().width / N : 40;
}

// ── DOM – board ────────────────────────────────────────────
function initBoardDOM() {
  const el = document.getElementById('board');
  el.innerHTML = '';
  for (let r = 0; r < N; r++) {
    for (let c = 0; c < N; c++) {
      const d = document.createElement('div');
      d.className = 'cell';
      d.dataset.r = r;
      d.dataset.c = c;
      const br = Math.floor(r / 3), bc = Math.floor(c / 3);
      if ((br + bc) % 2 === 0) d.classList.add('cell-alt');
      el.appendChild(d);
    }
  }
}

function renderBoard() {
  for (let r = 0; r < N; r++) {
    for (let c = 0; c < N; c++) {
      const el = cellEl(r, c);
      if (el) {
        el.classList.remove('clearing');
        el.classList.toggle('filled', !!board[r][c]);
      }
    }
  }
}

// ── DOM – rack ─────────────────────────────────────────────
const RACK_CELL = 18; // px per cell in rack

function initRackDOM() {
  const rack = document.getElementById('rack');
  rack.innerHTML = '';
  for (let i = 0; i < rackSize; i++) {
    const slot = document.createElement('div');
    slot.className = 'slot';
    slot.id = `slot-${i}`;
    rack.appendChild(slot);
  }
}

function renderRack() {
  for (let i = 0; i < rackSize; i++) renderSlot(i);
}

function renderSlot(i) {
  const slot = document.getElementById(`slot-${i}`);
  slot.innerHTML = '';
  slot.classList.remove('used', 'dragging', 'hint-slot', 'hint-slot-2', 'hint-slot-3', 'unplayable');

  if (used[i]) { slot.classList.add('used'); return; }

  const cells = pieces[i];
  const b = bounds(cells);

  const inner = document.createElement('div');
  inner.className = 'piece-inner entering';
  inner.style.width  = (b.cols * RACK_CELL) + 'px';
  inner.style.height = (b.rows * RACK_CELL) + 'px';

  for (const [r, c] of cells) {
    const blk = document.createElement('div');
    blk.className = 'piece-block';
    blk.style.cssText = `
      width:${RACK_CELL - 2}px; height:${RACK_CELL - 2}px;
      left:${(c - b.minC) * RACK_CELL + 1}px;
      top:${(r - b.minR) * RACK_CELL + 1}px;
    `;
    inner.appendChild(blk);
  }

  // Stagger entrance by slot index
  inner.style.animationDelay = (i * 80) + 'ms';
  inner.addEventListener('animationend', () => {
    inner.classList.remove('entering');
    inner.style.animationDelay = '';
  }, { once: true });

  slot.appendChild(inner);

  // Slot number label, helps players match hint text ("play slot 2 first") to the rack
  const label = document.createElement('span');
  label.className = 'slot-label';
  label.textContent = String(i + 1);
  slot.appendChild(label);

  attachDragListeners(slot, i);
}

// Grey out any piece that cannot be placed anywhere on the current board.
function updateRackPlayability() {
  for (let i = 0; i < rackSize; i++) {
    if (used[i]) continue;
    const slot = document.getElementById(`slot-${i}`);
    if (slot) slot.classList.toggle('unplayable', !canPlaceAnywhere(pieces[i]));
  }
}

// ── Drag & drop ────────────────────────────────────────────
let drag = null;

function attachDragListeners(slot, i) {
  slot.addEventListener('touchstart', e => {
    e.preventDefault();
    const t = e.changedTouches[0];
    startDrag(t.clientX, t.clientY, i);
  }, { passive: false });

  slot.addEventListener('mousedown', e => {
    e.preventDefault();
    startDrag(e.clientX, e.clientY, i);
  });
}

document.addEventListener('touchmove', e => {
  if (!drag) return;
  e.preventDefault();
  const t = e.changedTouches[0];
  moveDrag(t.clientX, t.clientY);
}, { passive: false });

document.addEventListener('touchend', e => {
  if (!drag) return;
  e.preventDefault();
  const t = e.changedTouches[0];
  endDrag(t.clientX, t.clientY);
}, { passive: false });

document.addEventListener('touchcancel', () => { if (drag) cancelDrag(); });

document.addEventListener('mousemove', e => { if (drag) moveDrag(e.clientX, e.clientY); });
document.addEventListener('mouseup',   e => { if (drag) endDrag(e.clientX, e.clientY); });

function startDrag(cx, cy, slotIdx) {
  if (drag || used[slotIdx] || gameOver) return;
  const slotEl = document.getElementById(`slot-${slotIdx}`);
  if (slotEl && slotEl.classList.contains('unplayable')) return;
  clearHint();

  const cells = pieces[slotIdx];
  const cs    = cellSize();
  const b     = bounds(cells);

  const ghost = document.createElement('div');
  ghost.className = 'ghost';
  ghost.style.width  = (b.cols * cs) + 'px';
  ghost.style.height = (b.rows * cs) + 'px';

  for (const [r, c] of cells) {
    const blk = document.createElement('div');
    blk.className = 'ghost-block';
    blk.style.cssText = `
      width:${cs - 2}px; height:${cs - 2}px;
      left:${(c - b.minC) * cs + 1}px;
      top:${(r - b.minR) * cs + 1}px;
    `;
    ghost.appendChild(blk);
  }
  document.body.appendChild(ghost);

  document.getElementById(`slot-${slotIdx}`).classList.add('dragging');

  drag = { slotIdx, cells, ghost, b, cs, snapR: -99, snapC: -99 };
  updateGhost(cx, cy);
  updatePreview(cx, cy);
}

function moveDrag(cx, cy) {
  updateGhost(cx, cy);
  updatePreview(cx, cy);
}

function updateGhost(cx, cy) {
  if (!drag) return;
  const { ghost, b, cs } = drag;
  const x = cx - (b.cols * cs) / 2;
  const y = cy - b.rows * cs - cs * 0.9;
  ghost.style.left = x + 'px';
  ghost.style.top  = y + 'px';
}

function getSnap(cx, cy) {
  const boardRect = document.getElementById('board').getBoundingClientRect();
  const { b, cs } = drag;
  const ghostX = cx - (b.cols * cs) / 2;
  const ghostY = cy - b.rows * cs - cs * 0.9;
  const col = Math.round((ghostX - boardRect.left) / cs);
  const row = Math.round((ghostY - boardRect.top)  / cs);
  return { row, col };
}

function updatePreview(cx, cy) {
  clearPreview();
  if (!drag) return;

  const { row, col } = getSnap(cx, cy);
  drag.snapR = row;
  drag.snapC = col;

  const { cells } = drag;
  const onBoard = cells.some(([dr, dc]) => {
    const r = row + dr, c = col + dc;
    return r >= 0 && r < N && c >= 0 && c < N;
  });
  if (!onBoard) return;

  const valid = canPlace(cells, row, col);
  for (const [dr, dc] of cells) {
    const r = row + dr, c = col + dc;
    if (r < 0 || r >= N || c < 0 || c >= N) continue;
    const el = cellEl(r, c);
    if (el) el.classList.add(valid ? 'preview-ok' : 'preview-bad');
  }

  if (valid) {
    for (const [r, c] of simClears(cells, row, col)) {
      const el = cellEl(r, c);
      if (el) el.classList.add('preview-clr');
    }
  }
}

function clearPreview() {
  document.querySelectorAll('#board .cell.preview-ok, #board .cell.preview-bad, #board .cell.preview-clr')
    .forEach(el => el.classList.remove('preview-ok', 'preview-bad', 'preview-clr'));
}

function endDrag(cx, cy) {
  if (!drag) return;
  clearPreview();

  const { slotIdx, cells, ghost, snapR, snapC } = drag;
  ghost.remove();
  document.getElementById(`slot-${slotIdx}`).classList.remove('dragging');
  drag = null;

  if (canPlace(cells, snapR, snapC)) {
    doPlace(slotIdx, snapR, snapC);
  }
}

function cancelDrag() {
  if (!drag) return;
  clearPreview();
  drag.ghost.remove();
  document.getElementById(`slot-${drag.slotIdx}`).classList.remove('dragging');
  drag = null;
}

// ── Game actions ───────────────────────────────────────────
function doPlace(slotIdx, row, col) {
  const cells = pieces[slotIdx];

  // Place blocks on board
  for (const [dr, dc] of cells) board[row + dr][col + dc] = 1;
  score += cells.length;
  updateScoreUI();

  used[slotIdx] = true;

  // Animate slot shrinking out, then render board with placement pop
  const slotEl = document.getElementById(`slot-${slotIdx}`);
  slotEl.classList.add('shrinking');
  setTimeout(() => {
    slotEl.classList.remove('shrinking');
    renderSlot(slotIdx);
  }, ANIM_SLOT_SHRINK);

  renderBoard();

  // Add placement pop animation to newly placed cells
  for (const [dr, dc] of cells) {
    const el = cellEl(row + dr, col + dc);
    if (el) {
      el.classList.add('just-placed');
      el.addEventListener('animationend', () => el.classList.remove('just-placed'), { once: true });
    }
  }

  // Check clears
  const cleared = doClears();

  if (cleared.size) {
    showPointsPopup(cleared.size);
    if (combo > 0) showComboPopup(combo);
    animateClears(cleared, () => {
      renderBoard();
      afterPlace();
    });
  } else {
    afterPlace();
  }
}

function afterPlace() {
  updateRackPlayability();
  updateTrainingPanel();
  if (used.every(Boolean)) {
    // All pieces placed → new round
    setTimeout(newRound, 80);
  } else {
    if (isGameOver()) setTimeout(triggerGameOver, 150);
  }
}

function doClears() {
  const rowFull = [], colFull = [], boxFull = [];

  for (let r = 0; r < N; r++)
    if (board[r].every(Boolean)) rowFull.push(r);

  for (let c = 0; c < N; c++)
    if (board.every(row => row[c])) colFull.push(c);

  for (let br = 0; br < 3; br++) {
    for (let bc = 0; bc < 3; bc++) {
      let full = true;
      outer: for (let r = br * 3; r < br * 3 + 3; r++) {
        for (let c = bc * 3; c < bc * 3 + 3; c++) {
          if (!board[r][c]) { full = false; break outer; }
        }
      }
      if (full) boxFull.push([br, bc]);
    }
  }

  const total = rowFull.length + colFull.length + boxFull.length;
  if (!total) { combo = 0; return new Set(); }

  const cleared = new Set();
  for (const r of rowFull) {
    for (let c = 0; c < N; c++) cleared.add(`${r},${c}`);
  }
  for (const c of colFull) {
    for (let r = 0; r < N; r++) cleared.add(`${r},${c}`);
  }
  for (const [br, bc] of boxFull) {
    for (let r = br * 3; r < br * 3 + 3; r++) {
      for (let c = bc * 3; c < bc * 3 + 3; c++) cleared.add(`${r},${c}`);
    }
  }

  // Scoring: cells cleared + multi-clear bonus + combo
  let pts = cleared.size;
  if (total > 1) pts += (total - 1) * 10;
  combo += total;   // combo grows by every region cleared in this move
  pts += combo * 5;
  score += pts;

  for (const key of cleared) {
    const [r, c] = key.split(',').map(Number);
    board[r][c] = 0;
  }

  updateScoreUI();
  return cleared;
}

function animateClears(cleared, cb) {
  // Stagger clear animation for a ripple effect
  const cells = [...cleared].map(key => key.split(',').map(Number));
  // Sort by distance from centroid for ripple effect
  const cx = cells.reduce((s, [r]) => s + r, 0) / cells.length;
  const cy = cells.reduce((s, [, c]) => s + c, 0) / cells.length;
  cells.sort((a, b) => {
    const da = Math.abs(a[0] - cx) + Math.abs(a[1] - cy);
    const db = Math.abs(b[0] - cx) + Math.abs(b[1] - cy);
    return da - db;
  });

  const step = cells.length > 1 ? ANIM_CLEAR_STAGGER / (cells.length - 1) : 0;

  cells.forEach(([r, c], i) => {
    const el = cellEl(r, c);
    if (el) {
      el.style.animationDelay = (i * step) + 'ms';
      el.classList.add('clearing');
    }
  });

  setTimeout(() => {
    // Clean up animation delays
    cells.forEach(([r, c]) => {
      const el = cellEl(r, c);
      if (el) el.style.animationDelay = '';
    });
    cb();
  }, ANIM_CLEAR + ANIM_CLEAR_STAGGER);
}

// ── Clears simulation (for preview) ───────────────────────
function simClears(cells, row, col) {
  const tmp = board.map(r => [...r]);
  for (const [dr, dc] of cells) tmp[row + dr][col + dc] = 1;

  const result = [];
  for (let r = 0; r < N; r++) {
    if (tmp[r].every(Boolean)) {
      for (let c = 0; c < N; c++) result.push([r, c]);
    }
  }
  for (let c = 0; c < N; c++) {
    if (tmp.every(row => row[c])) {
      for (let r = 0; r < N; r++) result.push([r, c]);
    }
  }
  for (let br = 0; br < 3; br++) {
    for (let bc = 0; bc < 3; bc++) {
      let full = true;
      outer: for (let r = br * 3; r < br * 3 + 3; r++) {
        for (let c = bc * 3; c < bc * 3 + 3; c++) {
          if (!tmp[r][c]) { full = false; break outer; }
        }
      }
      if (full) {
        for (let r = br * 3; r < br * 3 + 3; r++) {
          for (let c = bc * 3; c < bc * 3 + 3; c++) result.push([r, c]);
        }
      }
    }
  }
  return result;
}

// ── Game over ──────────────────────────────────────────────
// Game over only when every remaining unplaced piece is blocked.
// (Individual pieces that can't fit are just greyed out; the game continues
//  as long as at least one piece can still be placed.)
function isGameOver() {
  for (let i = 0; i < rackSize; i++) {
    if (used[i]) continue;
    if (canPlaceAnywhere(pieces[i])) return false;
  }
  return true;
}

function triggerGameOver() {
  gameOver = true;

  if (score > bestScore) {
    bestScore = score;
    localStorage.setItem('bst-best', bestScore);
  }

  const todayKey = new Date().toISOString().slice(0, 10);
  const td = JSON.parse(localStorage.getItem('bst-today') || '{"d":"","s":0}');
  todayScore = (td.d === todayKey) ? Math.max(td.s, score) : score;
  localStorage.setItem('bst-today', JSON.stringify({ d: todayKey, s: todayScore }));
  updateScoreUI();

  // Fade in "No more space!", hold, then fade out before showing the game-over card.
  showNoMoreSpaceMsg(() => {
    document.getElementById('go-score').textContent = `Score: ${score}`;
    document.getElementById('go-best').textContent  = `Best: ${bestScore}`;
    showOverlay('ov-gameover');
  });
}

// ── End-of-game messages ───────────────────────────────────
function showNoMoreSpaceMsg(cb) {
  const overlay = document.createElement('div');
  overlay.className = 'no-space-overlay';
  const span = document.createElement('span');
  span.className = 'no-space-text';
  span.textContent = 'No more space!';
  overlay.appendChild(span);
  document.body.appendChild(overlay);

  // After fade-in + hold, fade out then invoke callback.
  setTimeout(() => {
    overlay.classList.add('fading-out');
    setTimeout(() => {
      overlay.remove();
      if (cb) cb();
    }, ANIM_NO_SPACE_OUT);
  }, ANIM_NO_SPACE_IN + ANIM_NO_SPACE_HOLD);
}

function showChooseCarefullyMsg() {
  const boardRect = document.getElementById('board-wrap').getBoundingClientRect();
  const msg = document.createElement('div');
  msg.className = 'choose-carefully-msg';
  msg.textContent = 'Choose carefully…';
  // Centre the pill vertically in the board
  msg.style.top = (boardRect.top + boardRect.height / 2) + 'px';
  document.body.appendChild(msg);
  msg.addEventListener('animationend', () => msg.remove(), { once: true });
}

// ── New round / restart ────────────────────────────────────
function newRound() {
  used    = Array(rackSize).fill(false);
  pieces  = smartPieces();
  if (colorSetting === 'random') applyColor('random');
  renderRack();
  updateRackPlayability();
  if (isGameOver()) {
    setTimeout(triggerGameOver, 150);
  } else if (rackSize > 1 && orderMatters()) {
    showChooseCarefullyMsg();
  }
}

function startNewGame() {
  board    = emptyBoard();
  score    = 0;
  combo    = 0;
  gameOver = false;
  used     = Array(rackSize).fill(false);
  pieces   = smartPieces();

  applyColor(colorSetting);
  updateScoreUI();
  renderBoard();
  renderRack();
  updateRackPlayability();
  clearHint();
  updateTrainingPanel();

  hideOverlay('ov-gameover');
  document.getElementById('move-eval').textContent = '';
  document.getElementById('strategy-note').textContent = '';
}

// ── Score UI ───────────────────────────────────────────────
function updateScoreUI() {
  const el = document.getElementById('score-main');
  const prev = el.textContent;
  el.textContent = score;
  document.getElementById('today-val').textContent  = Math.max(todayScore, score);
  document.getElementById('best-val').textContent   = Math.max(bestScore, score);

  // Bump animation when score changes
  if (String(score) !== prev) {
    el.classList.remove('bump');
    void el.offsetWidth; // reflow to restart animation
    el.classList.add('bump');
  }
}

// ── Training: metrics ──────────────────────────────────────
function countHoles(b) {
  b = b || board;
  let n = 0;
  for (let r = 0; r < N; r++) {
    for (let c = 0; c < N; c++) {
      if (b[r][c]) continue;
      const nbrs = [[r-1,c],[r+1,c],[r,c-1],[r,c+1]];
      if (nbrs.every(([nr,nc]) => nr < 0 || nr >= N || nc < 0 || nc >= N || b[nr][nc]))
        n++;
    }
  }
  return n;
}

function countOpenLanes(b) {
  b = b || board;
  let n = 0;
  for (let r = 0; r < N; r++) if (b[r].every(v => !v)) n++;
  for (let c = 0; c < N; c++) if (b.every(row => !row[c])) n++;
  return n;
}

function centreCongestion(b) {
  b = b || board;
  let n = 0;
  for (let r = 3; r < 6; r++) for (let c = 3; c < 6; c++) if (b[r][c]) n++;
  return n;
}

function fragmentation(b) {
  b = b || board;
  // Count number of empty connected components (rough metric)
  const visited = Array.from({ length: N }, () => new Array(N).fill(false));
  let components = 0;
  for (let r = 0; r < N; r++) {
    for (let c = 0; c < N; c++) {
      if (b[r][c] || visited[r][c]) continue;
      components++;
      // BFS flood fill
      const q = [[r, c]];
      visited[r][c] = true;
      while (q.length) {
        const [cr, cc] = q.shift();
        for (const [nr, nc] of [[cr-1,cc],[cr+1,cc],[cr,cc-1],[cr,cc+1]]) {
          if (nr >= 0 && nr < N && nc >= 0 && nc < N && !b[nr][nc] && !visited[nr][nc]) {
            visited[nr][nc] = true;
            q.push([nr, nc]);
          }
        }
      }
    }
  }
  return components;
}

function updateTrainingPanel() {
  if (!trainingMode) return;

  const holes  = countHoles();
  const lanes  = countOpenLanes();
  const centre = centreCongestion();

  document.querySelector('#m-holes b').textContent  = holes;
  document.querySelector('#m-lanes b').textContent  = lanes;
  document.querySelector('#m-centre b').textContent = centre <= 3 ? 'ok' : centre <= 6 ? 'busy' : 'full';
  document.querySelector('#m-combo b').textContent  = combo;

  document.getElementById('strategy-note').textContent = strategyNote(holes, lanes, centre);
}

function strategyNote(holes, lanes, centre) {
  if (holes > 4)   return '⚠️ Many isolated holes, avoid blocking empty cells.';
  if (centre > 6)  return '⚠️ Centre is congested, try to clear those boxes soon.';
  if (lanes < 4)   return '⚠️ Few open lanes, prioritise clearing rows and columns.';
  if (combo > 2)   return `🔥 ${combo}× combo! Keep clearing to maximise score.`;
  if (holes === 0 && lanes >= 12) return '✅ Clean board, build towards a multi-clear.';
  return '💡 Look for placements that complete a full row, column or 3×3 box.';
}

// ── Training: hint ─────────────────────────────────────────
let hintActive = false;

// Return all permutations of an array
function getPermutations(arr) {
  if (arr.length <= 1) return [arr.slice()];
  const result = [];
  for (let i = 0; i < arr.length; i++) {
    const rest = arr.filter((_, j) => j !== i);
    for (const perm of getPermutations(rest)) result.push([arr[i], ...perm]);
  }
  return result;
}

// Evaluate a move on an arbitrary board snapshot (not the live board)
function evalMoveOnBoard(cells, row, col, b) {
  const tmp = b.map(r => [...r]);
  for (const [dr, dc] of cells) tmp[row + dr][col + dc] = 1;

  const clrs = getClearsOnBoard(tmp);
  const afterBoard = applyClears(tmp, clrs);

  let val = cells.length;
  val += clrs.total * 18;
  if (clrs.total > 1) val += (clrs.total - 1) * 12;
  val -= countHoles(afterBoard) * 7;
  val += countOpenLanes(afterBoard) * 2;
  val -= centreCongestion(afterBoard) * 2;
  val -= fragmentation(afterBoard) * 3;
  return val;
}

// Greedy best-placement sequence for a given ordering of slot indices, starting from board b
function greedySequence(order, startBoard) {
  let b = startBoard.map(r => [...r]);
  const moves = [];
  let totalScore = 0;

  for (const slotIdx of order) {
    const cells = pieces[slotIdx];
    let bestVal = -Infinity;
    let bestMove = null;

    for (let r = 0; r < N; r++) {
      for (let c = 0; c < N; c++) {
        let ok = true;
        for (const [dr, dc] of cells) {
          const nr = r + dr, nc = c + dc;
          if (nr < 0 || nr >= N || nc < 0 || nc >= N || b[nr][nc]) { ok = false; break; }
        }
        if (!ok) continue;
        const val = evalMoveOnBoard(cells, r, c, b);
        if (val > bestVal) { bestVal = val; bestMove = { slotIdx, r, c, cells }; }
      }
    }

    if (!bestMove) return null; // can't place this piece
    moves.push(bestMove);
    totalScore += bestVal;

    // Apply placement + clears to b
    for (const [dr, dc] of cells) b[bestMove.r + dr][bestMove.c + dc] = 1;
    const clrs = getClearsOnBoard(b);
    b = applyClears(b, clrs);
  }

  return { moves, score: totalScore };
}

// Find the best sequence of placements for all unplaced slots
function findBestSequence() {
  const unplaced = Array.from({ length: rackSize }, (_, i) => i).filter(i => !used[i]);
  if (unplaced.length === 0) return null;

  let bestScore = -Infinity;
  let bestMoves = null;

  for (const order of getPermutations(unplaced)) {
    const result = greedySequence(order, board);
    if (result && result.score > bestScore) {
      bestScore = result.score;
      bestMoves = result.moves;
    }
  }

  return bestMoves;
}

function showHint() {
  clearHint();

  const sequence = findBestSequence();
  if (!sequence || sequence.length === 0) return;

  const hintClasses = ['hint-cell', 'hint-cell-2', 'hint-cell-3'];
  const hintSlotClasses = ['hint-slot', 'hint-slot-2', 'hint-slot-3'];
  sequence.forEach((move, idx) => {
    const cls = hintClasses[idx] || 'hint-cell';
    for (const [dr, dc] of move.cells) {
      const el = cellEl(move.r + dr, move.c + dc);
      if (el) el.classList.add(cls);
    }
    const slotCls = hintSlotClasses[idx] || 'hint-slot';
    document.getElementById(`slot-${move.slotIdx}`).classList.add(slotCls);
  });

  const first = sequence[0];
  const suffix = sequence.length > 1 ? ` · Play slot ${first.slotIdx + 1} first.` : '';
  document.getElementById('move-eval').textContent = explainMove(first.cells, first.r, first.c) + suffix;
  hintActive = true;
}

function clearHint() {
  document.querySelectorAll('.hint-cell, .hint-cell-2, .hint-cell-3')
    .forEach(el => el.classList.remove('hint-cell', 'hint-cell-2', 'hint-cell-3'));
  document.querySelectorAll('.hint-slot, .hint-slot-2, .hint-slot-3')
    .forEach(el => el.classList.remove('hint-slot', 'hint-slot-2', 'hint-slot-3'));
  if (hintActive) {
    document.getElementById('move-eval').textContent = '';
    hintActive = false;
  }
}

// ── Move evaluation heuristics ─────────────────────────────
function evalMove(cells, row, col) {
  return evalMoveOnBoard(cells, row, col, board);
}

function getClearsOnBoard(b) {
  const rows = [], cols = [], boxes = [];
  for (let r = 0; r < N; r++) { if (b[r].every(Boolean)) rows.push(r); }
  for (let c = 0; c < N; c++) { if (b.every(row => row[c])) cols.push(c); }
  for (let br = 0; br < 3; br++) {
    for (let bc = 0; bc < 3; bc++) {
      let full = true;
      outer: for (let r = br * 3; r < br * 3 + 3; r++) {
        for (let c = bc * 3; c < bc * 3 + 3; c++) {
          if (!b[r][c]) { full = false; break outer; }
        }
      }
      if (full) boxes.push([br, bc]);
    }
  }
  return { rows, cols, boxes, total: rows.length + cols.length + boxes.length };
}

function applyClears(b, clrs) {
  const out = b.map(r => [...r]);
  for (const r of clrs.rows) {
    for (let c = 0; c < N; c++) out[r][c] = 0;
  }
  for (const c of clrs.cols) {
    for (let r = 0; r < N; r++) out[r][c] = 0;
  }
  for (const [br, bc] of clrs.boxes) {
    for (let r = br * 3; r < br * 3 + 3; r++) {
      for (let c = bc * 3; c < bc * 3 + 3; c++) out[r][c] = 0;
    }
  }
  return out;
}

function explainMove(cells, row, col) {
  const tmp = board.map(r => [...r]);
  for (const [dr, dc] of cells) tmp[row + dr][col + dc] = 1;

  const clrs  = getClearsOnBoard(tmp);
  const after = applyClears(tmp, clrs);
  const hBefore = countHoles(board);
  const hAfter  = countHoles(after);
  const newHoles = hAfter - hBefore;

  if (clrs.total >= 3) return `✅ Best move, clears ${clrs.total} regions at once!`;
  if (clrs.total === 2) return `✅ Great, clears ${clrs.total} regions simultaneously.`;
  if (clrs.total === 1) {
    if (newHoles > 1) return `⚠️ Clears a region but creates ${newHoles} holes.`;
    return '✅ Clears a region, good for score and space.';
  }
  if (newHoles > 2)   return `⚠️ Risky, creates ${newHoles} isolated holes.`;
  if (newHoles > 0)   return `⚠️ Creates ${newHoles} hole(s). Consider alternatives.`;
  if (countOpenLanes(after) >= countOpenLanes(board))
    return '✅ Safe, preserves open lanes for future pieces.';
  return '💡 Neutral placement, no immediate clears or major penalties.';
}

// ── Animation helpers ──────────────────────────────────────
function showComboPopup(c) {
  const label = c >= 5 ? '🔥🔥🔥' : c >= 3 ? '🔥🔥' : '🔥';
  const popup = document.createElement('div');
  popup.className = 'combo-popup';
  popup.textContent = `${label} ${c}× Combo!`;
  // Position above the board
  const boardRect = document.getElementById('board-wrap').getBoundingClientRect();
  popup.style.top = (boardRect.top + boardRect.height * 0.3) + 'px';
  document.body.appendChild(popup);
  popup.addEventListener('animationend', () => popup.remove());
}

function showPointsPopup(pts) {
  const popup = document.createElement('div');
  popup.className = 'points-popup';
  popup.textContent = `+${pts}`;
  const boardRect = document.getElementById('board-wrap').getBoundingClientRect();
  popup.style.top = (boardRect.top + boardRect.height * 0.45) + 'px';
  document.body.appendChild(popup);
  popup.addEventListener('animationend', () => popup.remove());
}

function showOverlay(id) {
  const ov = document.getElementById(id);
  ov.hidden = false;
  ov.classList.remove('show');
  void ov.offsetWidth; // reflow
  ov.classList.add('show');
}

function hideOverlay(id) {
  const ov = document.getElementById(id);
  ov.classList.remove('show');
  ov.hidden = true;
}

// ── Settings / overlays ────────────────────────────────────
document.getElementById('btn-settings').addEventListener('click', () => {
  document.getElementById('chk-coach').checked = trainingMode;
  document.getElementById('chk-extended').checked = extendedPieces;
  document.getElementById('chk-dark').checked = darkMode;
  document.getElementById('sel-color').value = colorSetting;
  document.getElementById('sel-rack').value = String(rackSize);
  showOverlay('ov-settings');
});

document.getElementById('btn-done').addEventListener('click', () => {
  const prev = trainingMode;
  const prevRackSize = rackSize;
  trainingMode   = document.getElementById('chk-coach').checked;
  extendedPieces = document.getElementById('chk-extended').checked;
  darkMode       = document.getElementById('chk-dark').checked;
  colorSetting   = document.getElementById('sel-color').value;
  rackSize       = parseInt(document.getElementById('sel-rack').value, 10);

  applyDarkMode(darkMode);
  applyColor(colorSetting);
  applyExtendedPieces(extendedPieces);
  saveSettings();

  hideOverlay('ov-settings');
  document.getElementById('coach-panel').hidden = !trainingMode;
  if (trainingMode && !prev) updateTrainingPanel();
  if (!trainingMode) {
    clearHint();
    document.getElementById('move-eval').textContent = '';
    document.getElementById('strategy-note').textContent = '';
  }

  // Rebuild rack and restart game if rack size changed
  if (rackSize !== prevRackSize) {
    initRackDOM();
    startNewGame();
  }
});

document.getElementById('btn-clear-data').addEventListener('click', async () => {
  if (!confirm('Clear all game progress and cached data?\nThis cannot be undone.')) return;

  // Remove game progress from localStorage
  localStorage.removeItem('bst-best');
  localStorage.removeItem('bst-today');
  localStorage.removeItem('bst-settings');

  // Unregister service workers so new assets are fetched on next load
  if ('serviceWorker' in navigator) {
    const regs = await navigator.serviceWorker.getRegistrations();
    await Promise.all(regs.map(r => r.unregister()));
  }

  // Delete all cached responses (style, script, etc.)
  if ('caches' in window) {
    const keys = await caches.keys();
    await Promise.all(keys.map(k => caches.delete(k)));
  }

  location.reload();
});



document.getElementById('btn-hint').addEventListener('click', showHint);

document.getElementById('btn-restart').addEventListener('click', startNewGame);

document.getElementById('btn-new').addEventListener('click', startNewGame);

// Prevent body scroll while dragging on iOS
document.addEventListener('touchstart', e => {
  if (e.target.closest('.slot')) e.preventDefault();
}, { passive: false });

// ── Init ───────────────────────────────────────────────────
function init() {
  bestScore  = parseInt(localStorage.getItem('bst-best') || '0', 10);
  const todayKey = new Date().toISOString().slice(0, 10);
  const td   = JSON.parse(localStorage.getItem('bst-today') || '{"d":"","s":0}');
  todayScore = (td.d === todayKey) ? td.s : 0;

  loadSettings();
  applyDarkMode(darkMode);
  applyColor(colorSetting);
  applyExtendedPieces(extendedPieces);
  document.getElementById('coach-panel').hidden = !trainingMode;

  // Follow OS dark-mode changes dynamically when the user hasn't set
  // an explicit preference (i.e. no saved 'dark' key in settings yet).
  const darkMQ = window.matchMedia('(prefers-color-scheme: dark)');
  darkMQ.addEventListener('change', e => {
    const s = JSON.parse(localStorage.getItem('bst-settings') || '{}');
    if (typeof s.dark !== 'boolean') {
      darkMode = e.matches;
      applyDarkMode(darkMode);
    }
  });

  initBoardDOM();
  initRackDOM();
  startNewGame();
}

init();
