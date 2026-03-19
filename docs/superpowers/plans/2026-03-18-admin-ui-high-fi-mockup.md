# Admin UI High-Fi Mock-up Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a static, mobile-first high-fidelity prototype for OpenFiltr covering setup, login, and dashboard flows without committing to a production frontend stack.

**Architecture:** Keep all prototype assets under `docs/mockups/admin-ui-v1/` so the work remains clearly separate from the backend product surface. Use plain HTML, CSS, and lightweight JavaScript for local state changes, sample data rendering, and review-state URLs, plus one shell smoke-check script for basic verification.

**Tech Stack:** Static HTML, CSS, vanilla JavaScript, shell, Python static server for local review

---

## Chunk 1: Static Prototype Foundation

### Task 1: Add Prototype Scaffold, Assets, And Smoke Checks

**Files:**
- Create: `docs/mockups/admin-ui-v1/assets/openfiltr-logo.svg`
- Create: `docs/mockups/admin-ui-v1/index.html`
- Create: `docs/mockups/admin-ui-v1/styles.css`
- Create: `docs/mockups/admin-ui-v1/mock-data.js`
- Create: `docs/mockups/admin-ui-v1/prototype.js`
- Create: `scripts/check_admin_ui_mockup.sh`

- [ ] **Step 1: Write the failing smoke check**

Create `scripts/check_admin_ui_mockup.sh` first. It should fail until the prototype files exist and expose the required landmarks:
- `index.html`
- the logo asset
- a stylesheet and shared JavaScript files
- a basic review-launcher shell for the prototype

- [ ] **Step 2: Run the smoke check to verify failure**

Run:
```bash
bash scripts/check_admin_ui_mockup.sh
```
Expected: FAIL because the prototype files do not exist yet.

- [ ] **Step 3: Create the shared prototype foundation**

Implement:
- local logo copy at `docs/mockups/admin-ui-v1/assets/openfiltr-logo.svg`
- `styles.css` with the flat neutral palette, blue-teal accent, form controls, cards, layout utilities, and mobile-first breakpoints
- `mock-data.js` with representative dashboard data for normal and low-data states
- `prototype.js` with shared helpers only for query-param review states, class toggles, and page bootstrapping hooks
- `index.html` as a review launcher linking to each prototype page and its review states

- [ ] **Step 4: Re-run the smoke check**

Run:
```bash
bash scripts/check_admin_ui_mockup.sh
```
Expected: PASS for the shared foundation checks in this task.

- [ ] **Step 5: Commit**

Run:
```bash
git add docs/mockups/admin-ui-v1/assets/openfiltr-logo.svg docs/mockups/admin-ui-v1/index.html docs/mockups/admin-ui-v1/styles.css docs/mockups/admin-ui-v1/mock-data.js docs/mockups/admin-ui-v1/prototype.js scripts/check_admin_ui_mockup.sh
git commit -s -m "feat(mockup): add prototype foundation"
```

### Task 2: Build Setup And Login Prototype Pages

**Files:**
- Create: `docs/mockups/admin-ui-v1/setup.html`
- Create: `docs/mockups/admin-ui-v1/login.html`
- Modify: `docs/mockups/admin-ui-v1/styles.css`
- Modify: `docs/mockups/admin-ui-v1/prototype.js`
- Modify: `scripts/check_admin_ui_mockup.sh`

- [ ] **Step 1: Extend the failing smoke check for auth flow details**

Update the smoke check so it asserts:
- `setup.html` contains the admin-account step and the setup-complete step
- `setup.html?state=validation-error` is supported by shared JavaScript markers
- `login.html` contains a same-page error state reachable as `login.html?state=error`
- `login.html` exposes a clear success target or redirect hook pointing to `dashboard.html`
- both pages use the shared logo, shared stylesheet, and shared script

- [ ] **Step 2: Run the smoke check to verify failure**

Run:
```bash
bash scripts/check_admin_ui_mockup.sh
```
Expected: FAIL because setup and login files or required state markers are still missing.

- [ ] **Step 3: Implement the setup flow**

Build `setup.html` as a two-step mobile-first flow:
- step 1 creates the admin account with username, password, and confirm password
- step 2 offers `import config` and `continue to dashboard`
- validation state shows missing input, weak password, and mismatch errors inline
- `import config` opens a local stub panel or message and keeps the user in the setup flow
- successful submit on step 1 advances to step 2 locally

- [ ] **Step 4: Implement the login flow**

Build `login.html` with:
- compact logo lock-up
- username and password inputs
- primary sign-in action
- same-page failure state with a clear message
- successful local submit targeting `dashboard.html` through a clear link or redirect hook, even before the dashboard page is implemented fully

- [ ] **Step 5: Re-run the smoke check**

Run:
```bash
bash scripts/check_admin_ui_mockup.sh
```
Expected: PASS for setup and login expectations, but dashboard-specific checks may still fail if already added.

- [ ] **Step 6: Commit**

Run:
```bash
git add docs/mockups/admin-ui-v1/setup.html docs/mockups/admin-ui-v1/login.html docs/mockups/admin-ui-v1/styles.css docs/mockups/admin-ui-v1/prototype.js scripts/check_admin_ui_mockup.sh
git commit -s -m "feat(mockup): add setup and login flows"
```

## Chunk 2: Dashboard And Review Verification

### Task 3: Build The Dashboard Prototype

**Files:**
- Create: `docs/mockups/admin-ui-v1/dashboard.html`
- Modify: `docs/mockups/admin-ui-v1/styles.css`
- Modify: `docs/mockups/admin-ui-v1/mock-data.js`
- Modify: `docs/mockups/admin-ui-v1/prototype.js`
- Modify: `docs/mockups/admin-ui-v1/index.html`
- Modify: `scripts/check_admin_ui_mockup.sh`

- [ ] **Step 1: Extend the failing smoke check for dashboard expectations**

Update the smoke check so it asserts that `dashboard.html` includes:
- service health card
- key stats for total requests, blocked requests, allowed requests, and block rate
- recent activity list
- top blocked domains list
- low-emphasis quick actions
- low-data review state reachable as `dashboard.html?state=low-data`

- [ ] **Step 2: Run the smoke check to verify failure**

Run:
```bash
bash scripts/check_admin_ui_mockup.sh
```
Expected: FAIL because the dashboard page and its state hooks are not complete yet.

- [ ] **Step 3: Implement the dashboard page**

Build `dashboard.html` so the first narrow mobile viewport shows:
- service health at the top
- the four key stats immediately below
- recent activity preview without excessive scrolling

Below that, add:
- top blocked domains
- low-emphasis quick actions for:
  - `import config`
  - `view activity`
  - `manage rules`
- shallow placeholder navigation that does not pretend other pages are finished
- review-launcher updates in `index.html` for `dashboard.html` and `dashboard.html?state=low-data`

Service health content must show:
- running state
- HTTP/API health
- DNS service state

Key stats must show:
- total requests
- blocked requests
- allowed requests
- block rate

Recent activity rows must show, newest first:
- domain
- action
- client
- time

Top blocked domain rows must show:
- domain
- block count

Use `mock-data.js` for normal and low-data states. In the low-data state:
- keep service health and key stats visible
- replace recent-activity and blocked-domain lists with calm empty-state copy
- keep the page visually stable rather than looking broken

- [ ] **Step 4: Re-run the smoke check**

Run:
```bash
bash scripts/check_admin_ui_mockup.sh
```
Expected: PASS.

- [ ] **Step 5: Commit**

Run:
```bash
git add docs/mockups/admin-ui-v1/dashboard.html docs/mockups/admin-ui-v1/styles.css docs/mockups/admin-ui-v1/mock-data.js docs/mockups/admin-ui-v1/prototype.js docs/mockups/admin-ui-v1/index.html scripts/check_admin_ui_mockup.sh
git commit -s -m "feat(mockup): add dashboard prototype"
```

### Task 4: Verify Review Flows And Demo Readiness

**Files:**
- Modify: `docs/mockups/admin-ui-v1/index.html`
- Modify: `docs/mockups/admin-ui-v1/dashboard.html`
- Modify: `docs/mockups/admin-ui-v1/styles.css`
- Modify: `docs/mockups/admin-ui-v1/mock-data.js`
- Modify: `docs/mockups/admin-ui-v1/prototype.js`
- Modify: `scripts/check_admin_ui_mockup.sh`

- [ ] **Step 1: Run formatting and diff sanity checks**

Run:
```bash
git diff --check
bash scripts/check_admin_ui_mockup.sh
```
Expected:
- no whitespace or merge-marker problems
- smoke check PASS

- [ ] **Step 2: Run the local review server**

Run:
```bash
python3 -m http.server 4173 -d docs/mockups/admin-ui-v1
```
Expected: local server starts on `http://127.0.0.1:4173/`.

- [ ] **Step 3: Verify the key review URLs**

Check in a browser at minimum:
- `/index.html`
- `/setup.html`
- `/setup.html?state=validation-error`
- `/login.html`
- `/login.html?state=error`
- `/dashboard.html`
- `/dashboard.html?state=low-data`

Confirm:
- mobile-first layout remains readable at narrow widths
- first dashboard viewport shows health, stats, and recent activity
- `dashboard.html?state=low-data` keeps health and stats visible while list-heavy sections switch to calm empty-state copy
- state pages remain visually stable

- [ ] **Step 4: Run final verification**

Run:
```bash
git diff --check
bash scripts/check_admin_ui_mockup.sh
```
Expected:
- no diff formatting issues
- smoke check PASS

- [ ] **Step 5: Commit any final polish**

Run:
```bash
git add docs/mockups/admin-ui-v1/index.html docs/mockups/admin-ui-v1/dashboard.html docs/mockups/admin-ui-v1/styles.css docs/mockups/admin-ui-v1/mock-data.js docs/mockups/admin-ui-v1/prototype.js scripts/check_admin_ui_mockup.sh
git commit -s -m "chore(mockup): polish review flow"
```

- [ ] **Step 6: Confirm the branch is clean after the final commit**

Run:
```bash
git status --short
```
Expected:
- no output
