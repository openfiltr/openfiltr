'use strict';

/* ==========================================================
   BLOCKUDOKU TRAINER  –  app.js
   ========================================================== */

// ── Piece definitions ──────────────────────────────────────
// Each entry is an array of [row, col] offsets.
// All pieces normalised: minRow = 0, minCol = 0.
const PIECE_DEFS = [
  // 1-cell
  [[0,0]],
  // Dominoes
  [[0,0],[0,1]],
  [[0,0],[1,0]],
  // Straight trominoes
  [[0,0],[0,1],[0,2]],
  [[0,0],[1,0],[2,0]],
  // Corner trominoes (4 orientations)
  [[0,0],[0,1],[1,0]],
  [[0,0],[0,1],[1,1]],
  [[0,0],[1,0],[1,1]],
  [[0,1],[1,0],[1,1]],
  // 2×2 square
  [[0,0],[0,1],[1,0],[1,1]],
  // Straight tetrominoes
  [[0,0],[0,1],[0,2],[0,3]],
  [[0,0],[1,0],[2,0],[3,0]],
  // Straight pentominoes
  [[0,0],[0,1],[0,2],[0,3],[0,4]],
  [[0,0],[1,0],[2,0],[3,0],[4,0]],
  // L-tetrominoes (4 orientations)
  [[0,0],[1,0],[2,0],[2,1]],
  [[0,0],[0,1],[0,2],[1,0]],
  [[0,0],[0,1],[1,1],[2,1]],
  [[0,2],[1,0],[1,1],[1,2]],
  // J-tetrominoes (4 orientations)
  [[0,0],[0,1],[1,0],[2,0]],
  [[0,0],[1,0],[1,1],[1,2]],
  [[0,1],[1,1],[2,0],[2,1]],
  [[0,0],[0,1],[0,2],[1,2]],
  // T-tetrominoes (4 orientations)
  [[0,0],[0,1],[0,2],[1,1]],
  [[0,1],[1,0],[1,1],[2,1]],
  [[0,1],[1,0],[1,1],[1,2]],
  [[0,0],[1,0],[1,1],[2,0]],
  // S-tetrominoes (2 orientations)
  [[0,1],[0,2],[1,0],[1,1]],
  [[0,0],[1,0],[1,1],[2,1]],
  // Z-tetrominoes (2 orientations)
  [[0,0],[0,1],[1,1],[1,2]],
  [[0,1],[1,0],[1,1],[2,0]],
  // 3×3 full
  [[0,0],[0,1],[0,2],[1,0],[1,1],[1,2],[2,0],[2,1],[2,2]],
  // Rectangles
  [[0,0],[0,1],[1,0],[1,1],[2,0],[2,1]],
  [[0,0],[0,1],[0,2],[1,0],[1,1],[1,2]],
];

const N = 9;

// ── State ─────────────────────────────────────────────────
let board   = [];   // N×N of 0/1
let pieces  = [];   // current 3 piece-cell-arrays
let used    = [];   // [bool, bool, bool] – which slots are placed
let score   = 0;
let bestScore = 0;
let todayScore = 0;
let combo   = 0;
let gameOver = false;
let trainingMode = false;

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

function randomPiece() {
  return PIECE_DEFS[Math.floor(Math.random() * PIECE_DEFS.length)];
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

function renderRack() {
  for (let i = 0; i < 3; i++) renderSlot(i);
}

function renderSlot(i) {
  const slot = document.getElementById(`slot-${i}`);
  slot.innerHTML = '';
  slot.classList.remove('used', 'dragging', 'hint-slot');

  if (used[i]) { slot.classList.add('used'); return; }

  const cells = pieces[i];
  const b = bounds(cells);

  const inner = document.createElement('div');
  inner.className = 'piece-inner';
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

  slot.appendChild(inner);
  attachDragListeners(slot, i);
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
  renderBoard();
  renderSlot(slotIdx);

  // Check clears
  const cleared = doClears();

  if (cleared.size) {
    animateClears(cleared, () => {
      renderBoard();
      afterPlace();
    });
  } else {
    afterPlace();
  }
}

function afterPlace() {
  updateTrainingPanel();
  if (used.every(Boolean)) {
    // All 3 placed → new round
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
  combo++;
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
  for (const key of cleared) {
    const [r, c] = key.split(',').map(Number);
    const el = cellEl(r, c);
    if (el) el.classList.add('clearing');
  }
  setTimeout(cb, 340);
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
function isGameOver() {
  for (let i = 0; i < 3; i++) {
    if (used[i]) continue;
    if (!canPlaceAnywhere(pieces[i])) return true;
  }
  return false;
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

  document.getElementById('go-score').textContent = `Score: ${score}`;
  document.getElementById('go-best').textContent  = `Best: ${bestScore}`;
  document.getElementById('ov-gameover').hidden = false;
}

// ── New round / restart ────────────────────────────────────
function newRound() {
  used    = [false, false, false];
  pieces  = [randomPiece(), randomPiece(), randomPiece()];
  renderRack();
  if (isGameOver()) triggerGameOver();
}

function startNewGame() {
  board    = emptyBoard();
  score    = 0;
  combo    = 0;
  gameOver = false;
  used     = [false, false, false];
  pieces   = [randomPiece(), randomPiece(), randomPiece()];

  updateScoreUI();
  renderBoard();
  renderRack();
  clearHint();
  updateTrainingPanel();

  document.getElementById('ov-gameover').hidden = true;
  document.getElementById('move-eval').textContent = '';
  document.getElementById('strategy-note').textContent = '';
}

// ── Score UI ───────────────────────────────────────────────
function updateScoreUI() {
  document.getElementById('score-main').textContent = score;
  document.getElementById('today-val').textContent  = Math.max(todayScore, score);
  document.getElementById('best-val').textContent   = Math.max(bestScore, score);
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
  if (holes > 4)   return '⚠️ Many isolated holes — avoid blocking empty cells.';
  if (centre > 6)  return '⚠️ Centre is congested — try to clear those boxes soon.';
  if (lanes < 4)   return '⚠️ Few open lanes — prioritise clearing rows/cols.';
  if (combo > 2)   return `🔥 ${combo}× combo! Keep clearing to maximise score.`;
  if (holes === 0 && lanes >= 12) return '✅ Clean board — build towards a multi-clear.';
  return '💡 Look for placements that complete a full row, column or 3×3 box.';
}

// ── Training: hint ─────────────────────────────────────────
let hintActive = false;

function showHint() {
  clearHint();

  let bestVal  = -Infinity;
  let bestMove = null;

  for (let i = 0; i < 3; i++) {
    if (used[i]) continue;
    const cells = pieces[i];
    for (let r = 0; r < N; r++) {
      for (let c = 0; c < N; c++) {
        if (!canPlace(cells, r, c)) continue;
        const val = evalMove(cells, r, c);
        if (val > bestVal) { bestVal = val; bestMove = { i, r, c, cells }; }
      }
    }
  }

  if (!bestMove) return;

  for (const [dr, dc] of bestMove.cells) {
    const el = cellEl(bestMove.r + dr, bestMove.c + dc);
    if (el) el.classList.add('hint-cell');
  }
  document.getElementById(`slot-${bestMove.i}`).classList.add('hint-slot');
  document.getElementById('move-eval').textContent = explainMove(bestMove.cells, bestMove.r, bestMove.c);
  hintActive = true;
}

function clearHint() {
  document.querySelectorAll('.hint-cell').forEach(el => el.classList.remove('hint-cell'));
  document.querySelectorAll('.hint-slot').forEach(el => el.classList.remove('hint-slot'));
  if (hintActive) {
    document.getElementById('move-eval').textContent = '';
    hintActive = false;
  }
}

// ── Move evaluation heuristics ─────────────────────────────
function evalMove(cells, row, col) {
  const tmp = board.map(r => [...r]);
  for (const [dr, dc] of cells) tmp[row + dr][col + dc] = 1;

  const clrs = getClearsOnBoard(tmp);
  const afterBoard = applyClears(tmp, clrs);

  let val = cells.length;                        // blocks placed
  val += clrs.total * 18;                        // region clears
  if (clrs.total > 1) val += (clrs.total - 1) * 12; // multi-clear bonus
  val -= countHoles(afterBoard) * 7;             // penalise holes
  val += countOpenLanes(afterBoard) * 2;         // reward open lanes
  val -= centreCongestion(afterBoard) * 2;       // penalise centre congestion
  val -= fragmentation(afterBoard) * 3;          // penalise fragmentation
  return val;
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

  if (clrs.total >= 3) return `✅ Best move — clears ${clrs.total} regions at once!`;
  if (clrs.total === 2) return `✅ Great — clears ${clrs.total} regions simultaneously.`;
  if (clrs.total === 1) {
    if (newHoles > 1) return `⚠️ Clears a region but creates ${newHoles} holes.`;
    return '✅ Clears a region — good for score and space.';
  }
  if (newHoles > 2)   return `⚠️ Risky — creates ${newHoles} isolated holes.`;
  if (newHoles > 0)   return `⚠️ Creates ${newHoles} hole(s). Consider alternatives.`;
  if (countOpenLanes(after) >= countOpenLanes(board))
    return '✅ Safe — preserves open lanes for future pieces.';
  return '💡 Neutral placement — no immediate clears or major penalties.';
}

// ── Settings / overlays ────────────────────────────────────
document.getElementById('btn-settings').addEventListener('click', () => {
  document.getElementById('chk-training').checked = trainingMode;
  document.getElementById('ov-settings').hidden = false;
});

document.getElementById('btn-done').addEventListener('click', () => {
  const prev = trainingMode;
  trainingMode = document.getElementById('chk-training').checked;
  document.getElementById('ov-settings').hidden = true;
  document.getElementById('training-panel').hidden = !trainingMode;
  if (trainingMode && !prev) updateTrainingPanel();
  if (!trainingMode) {
    clearHint();
    document.getElementById('move-eval').textContent = '';
    document.getElementById('strategy-note').textContent = '';
  }
});

document.getElementById('btn-back').addEventListener('click', () => {
  // No-op on single-page app; could reset or show menu in future
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

  initBoardDOM();
  startNewGame();
}

init();
