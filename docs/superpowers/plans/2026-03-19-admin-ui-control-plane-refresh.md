# Admin UI Control-Plane Refresh Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Rework the static admin UI mock-up into a cleaner, more product-grade control-plane prototype with Tailwind CSS, shared component primitives, and stronger branding discipline.

**Architecture:** Keep the prototype as static multi-page HTML under `docs/mockups/admin-ui-v1/`, but replace the current page-specific visual styling with a shared Tailwind-driven theme and a small set of reusable component-like classes. Preserve the existing JavaScript flow model for setup, login, and dashboard review states while refreshing the page structure and presentation.

**Tech Stack:** Static HTML, Tailwind CSS CDN, plain CSS, vanilla JavaScript, shell, Python static server for local review

---

## Chunk 1: Shared Design System Foundation

### Task 1: Add Tailwind Theme Scaffolding And Shared Primitives

**Files:**
- Create: `docs/mockups/admin-ui-v1/theme.js`
- Modify: `docs/mockups/admin-ui-v1/styles.css`
- Modify: `docs/mockups/admin-ui-v1/index.html`
- Modify: `docs/mockups/admin-ui-v1/setup.html`
- Modify: `docs/mockups/admin-ui-v1/login.html`
- Modify: `docs/mockups/admin-ui-v1/dashboard.html`
- Modify: `scripts/check_admin_ui_mockup.sh`

- [ ] **Step 1: Extend the smoke check first**

Update `scripts/check_admin_ui_mockup.sh` so it fails until the refresh adds:
- local `theme.js`
- Tailwind CDN script usage on all prototype pages
- a shared component-system vocabulary across the pages, such as common shell, panel, button, input, badge, and list-row primitives
- continued presence of the existing flow and review-state markers

- [ ] **Step 2: Run the smoke check to verify failure**

Run:
```bash
bash scripts/check_admin_ui_mockup.sh
```
Expected: FAIL because the Tailwind theme file and new shared component markers do not exist yet.

- [ ] **Step 3: Add the shared theme and reusable primitives**

Implement:
- `theme.js` with the Tailwind config for the control-plane palette, spacing, shadow, radius, and font stacks
- a rewritten `styles.css` with a small set of reusable component-like classes for:
  - shell
  - top bar
  - panel
  - section heading
  - button variants
  - input shell
  - badge or status chip
  - metric tile
  - operational list row
- shared page head updates so each prototype page loads:
  - `theme.js`
  - Tailwind CDN
  - `styles.css`

- [ ] **Step 4: Re-run the smoke check**

Run:
```bash
bash scripts/check_admin_ui_mockup.sh
```
Expected: PASS for the shared theme and component-foundation checks, even if some page-level structure checks still fail.

- [ ] **Step 5: Commit**

Run:
```bash
git add docs/mockups/admin-ui-v1/theme.js docs/mockups/admin-ui-v1/styles.css docs/mockups/admin-ui-v1/index.html docs/mockups/admin-ui-v1/setup.html docs/mockups/admin-ui-v1/login.html docs/mockups/admin-ui-v1/dashboard.html scripts/check_admin_ui_mockup.sh
git commit -s -m "feat(mockup): add control-plane design system foundation"
```

### Task 2: Refresh Setup And Login With The Shared System

**Files:**
- Modify: `docs/mockups/admin-ui-v1/setup.html`
- Modify: `docs/mockups/admin-ui-v1/login.html`
- Modify: `docs/mockups/admin-ui-v1/prototype.js`
- Modify: `docs/mockups/admin-ui-v1/styles.css`
- Modify: `scripts/check_admin_ui_mockup.sh`

- [ ] **Step 1: Tighten the smoke check for the auth refresh**

Update the smoke check so it asserts that:
- setup and login use the Tailwind theme and shared primitives
- setup keeps both steps but presents them in a stronger product shell
- login keeps the same-page error state and success target
- auth forms include recognisable shared input and button treatments instead of one-off page styling

- [ ] **Step 2: Run the smoke check to verify failure**

Run:
```bash
bash scripts/check_admin_ui_mockup.sh
```
Expected: FAIL because the auth pages still reflect the older layout and primitive structure.

- [ ] **Step 3: Rebuild the setup page**

Refresh `setup.html` so it feels product-grade:
- restrained control-plane header
- two-step setup shell with stronger hierarchy
- shared panel, input, button, and stepper primitives
- cleaner completion state that feels like a hand-off into the product
- preserved validation behaviour and review state support

Adjust `prototype.js` only where needed to support updated hooks or shared state helpers.

- [ ] **Step 4: Rebuild the login page**

Refresh `login.html` so it uses:
- the same shared shell and brand treatment as setup
- a more disciplined single-panel sign-in form
- stronger error presentation
- shared primitives instead of page-specific auth styles

- [ ] **Step 5: Re-run the smoke check**

Run:
```bash
bash scripts/check_admin_ui_mockup.sh
```
Expected: PASS for the auth refresh checks.

- [ ] **Step 6: Commit**

Run:
```bash
git add docs/mockups/admin-ui-v1/setup.html docs/mockups/admin-ui-v1/login.html docs/mockups/admin-ui-v1/prototype.js docs/mockups/admin-ui-v1/styles.css scripts/check_admin_ui_mockup.sh
git commit -s -m "feat(mockup): refresh setup and login screens"
```

## Chunk 2: Dashboard And Demo Review

### Task 3: Rebuild The Dashboard And Review Launcher

**Files:**
- Modify: `docs/mockups/admin-ui-v1/index.html`
- Modify: `docs/mockups/admin-ui-v1/dashboard.html`
- Modify: `docs/mockups/admin-ui-v1/mock-data.js`
- Modify: `docs/mockups/admin-ui-v1/prototype.js`
- Modify: `docs/mockups/admin-ui-v1/styles.css`
- Modify: `scripts/check_admin_ui_mockup.sh`

- [ ] **Step 1: Extend the smoke check for the refreshed dashboard**

Update the smoke check so it asserts that:
- `dashboard.html` uses the shared control-plane shell and primitives
- the dashboard keeps health, key totals, recent activity, top blocked domains, and quick actions
- the layout is less card-repetitive and more operational
- the review launcher remains present but clearly secondary
- `dashboard.html?state=low-data` still works

- [ ] **Step 2: Run the smoke check to verify failure**

Run:
```bash
bash scripts/check_admin_ui_mockup.sh
```
Expected: FAIL because the dashboard and launcher still use the old structure.

- [ ] **Step 3: Rebuild the review launcher**

Refresh `index.html` into a disciplined internal preview page:
- quiet product header
- clearer grouping of core pages versus review states
- shared panel and action primitives
- no fake product-home-page treatment

- [ ] **Step 4: Rebuild the dashboard**

Refresh `dashboard.html` so it feels like a DNS control plane:
- stable top bar with brand and environment context
- immediate health and request summary
- operational recent-activity list with GitHub-like row discipline
- more deliberate blocked-domain panel
- restrained quick actions
- subtle DNS-inspired visual cues without gimmicks

Update `mock-data.js` and `prototype.js` only where needed to support the revised presentation and data labels.

- [ ] **Step 5: Re-run the smoke check**

Run:
```bash
bash scripts/check_admin_ui_mockup.sh
```
Expected: PASS.

- [ ] **Step 6: Commit**

Run:
```bash
git add docs/mockups/admin-ui-v1/index.html docs/mockups/admin-ui-v1/dashboard.html docs/mockups/admin-ui-v1/mock-data.js docs/mockups/admin-ui-v1/prototype.js docs/mockups/admin-ui-v1/styles.css scripts/check_admin_ui_mockup.sh
git commit -s -m "feat(mockup): refresh dashboard control plane"
```

### Task 4: Verify Demo Readiness And Control-Plane Fit

**Files:**
- Modify: `docs/mockups/admin-ui-v1/index.html`
- Modify: `docs/mockups/admin-ui-v1/setup.html`
- Modify: `docs/mockups/admin-ui-v1/login.html`
- Modify: `docs/mockups/admin-ui-v1/dashboard.html`
- Modify: `docs/mockups/admin-ui-v1/styles.css`
- Modify: `docs/mockups/admin-ui-v1/mock-data.js`
- Modify: `docs/mockups/admin-ui-v1/prototype.js`
- Modify: `scripts/check_admin_ui_mockup.sh`

- [ ] **Step 1: Run file sanity checks**

Run:
```bash
git diff --check
bash scripts/check_admin_ui_mockup.sh
```
Expected:
- no whitespace or merge-marker issues
- smoke check PASS

- [ ] **Step 2: Run the local review server**

Run:
```bash
python3 -m http.server 4173 -d docs/mockups/admin-ui-v1
```
Expected: local server starts on `http://127.0.0.1:4173/`.

- [ ] **Step 3: Verify the key review URLs in a browser**

Check at minimum:
- `/index.html`
- `/setup.html`
- `/setup.html?state=validation-error`
- `/login.html`
- `/login.html?state=error`
- `/dashboard.html`
- `/dashboard.html?state=low-data`

Confirm:
- the shared component-system feel is obvious across all screens
- setup and login look like part of the same product shell
- dashboard feels calmer, cleaner, and more product-grade than the original mock-up
- mobile keeps the operational priority order intact
- low-data state remains stable and readable

- [ ] **Step 4: Run final verification**

Run:
```bash
git diff --check
bash scripts/check_admin_ui_mockup.sh
go test ./...
```
Expected:
- all checks PASS

- [ ] **Step 5: Commit final polish if needed**

Run only if verification required final clean-up edits:
```bash
git add docs/mockups/admin-ui-v1/index.html docs/mockups/admin-ui-v1/setup.html docs/mockups/admin-ui-v1/login.html docs/mockups/admin-ui-v1/dashboard.html docs/mockups/admin-ui-v1/mock-data.js docs/mockups/admin-ui-v1/prototype.js docs/mockups/admin-ui-v1/styles.css scripts/check_admin_ui_mockup.sh
git commit -s -m "chore(mockup): polish control-plane refresh"
```
