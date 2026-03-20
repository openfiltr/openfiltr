# Admin UI Operations Expansion Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Expand the static admin UI mock-up into a more convincing self-hosted control plane with dedicated activity, DNS records, allow-list, and block-list screens, smarter quick actions, and concise inline DNS help.

**Architecture:** Keep the prototype as static multi-page HTML in `docs/mockups/admin-ui-v1/`, but move it to a shared operational shell used by overview and the new management pages. Extend `mock-data.js` and `prototype.js` so seeded review states, deep links, quick actions, and info panels all work through one small prototype runtime rather than per-page hacks.

**Tech Stack:** Static HTML, Tailwind CSS CDN, shared CSS, vanilla JavaScript, shell, Node-backed smoke checks, local browser review

---

## Chunk 1: Prototype Foundation

### Task 1: Extend Smoke Checks And Runtime Contracts

**Files:**
- Modify: `docs/mockups/admin-ui-v1/mock-data.js`
- Modify: `docs/mockups/admin-ui-v1/prototype.js`
- Modify: `scripts/check_admin_ui_mockup.sh`

- [ ] **Step 1: Tighten the smoke check first**

Update `scripts/check_admin_ui_mockup.sh` so it fails until the expansion adds:
- `activity.html`
- `dns-records.html`
- `allow-list.html`
- `block-list.html`
- quick-action routes that replace the old dashboard dead ends
- seeded data for activity, DNS records, allow-list, and block-list views
- small info-panel support for DNS and policy terms

- [ ] **Step 2: Run the smoke check to verify failure**

Run:
```bash
bash scripts/check_admin_ui_mockup.sh
```
Expected: FAIL because the new pages, route markers, and seeded data do not exist yet.

- [ ] **Step 3: Add shared runtime scaffolding**

Extend `mock-data.js` and `prototype.js` with:
- named page data for overview, activity, DNS records, allow list, and block list
- query-state resolution helpers for:
  - `add-record`
  - `add-rule`
  - `validation-error`
  - `record-added`
  - `rule-added`
  - `import-preview`
  - `filter=blocked`
- concise info-panel copy for DNS and policy terms
- shared helpers for opening or closing inline help and seeded side-panel states

- [ ] **Step 4: Re-run the smoke check**

Run:
```bash
bash scripts/check_admin_ui_mockup.sh
```
Expected: still FAIL on missing page structure, but PASS for the new runtime and data checks.

### Task 2: Expand Shared Shell Primitives For Operations Screens

**Files:**
- Modify: `docs/mockups/admin-ui-v1/styles.css`
- Modify: `docs/mockups/admin-ui-v1/theme.js`

- [ ] **Step 1: Add failing structure checks**

Update the smoke check so it expects shared shell primitives for:
- app navigation
- page toolbars
- records or policy tables
- inline help triggers and panels
- side-panel or inline editor states

- [ ] **Step 2: Run the smoke check to verify failure**

Run:
```bash
bash scripts/check_admin_ui_mockup.sh
```
Expected: FAIL because the shared operations shell and help-panel primitives are not defined yet.

- [ ] **Step 3: Add the shared visual primitives**

Extend the shared styling with reusable patterns for:
- left or compact nav framing
- page header or toolbar rows
- filter bars
- tabular or list management rows
- split content and side-panel layouts
- info trigger buttons and anchored help panels
- stronger block-policy emphasis without changing the overall tone

Keep the current design direction: calm first, sharper only where enforcement matters.

- [ ] **Step 4: Re-run the smoke check**

Run:
```bash
bash scripts/check_admin_ui_mockup.sh
```
Expected: FAIL only because the new page HTML is still missing or incomplete.

## Chunk 2: New Operational Pages

### Task 3: Build The Shared Application Shell On New Pages

**Files:**
- Create: `docs/mockups/admin-ui-v1/activity.html`
- Create: `docs/mockups/admin-ui-v1/dns-records.html`
- Create: `docs/mockups/admin-ui-v1/allow-list.html`
- Create: `docs/mockups/admin-ui-v1/block-list.html`
- Modify: `scripts/check_admin_ui_mockup.sh`

- [ ] **Step 1: Extend the smoke check for the new pages**

Require each new page to include:
- the shared theme and styles
- the shared navigation shell
- page-specific headings
- obvious entry points from quick actions or review-state links
- `prototype.js`

- [ ] **Step 2: Run the smoke check to verify failure**

Run:
```bash
bash scripts/check_admin_ui_mockup.sh
```
Expected: FAIL because the page files do not exist yet.

- [ ] **Step 3: Create the new page shells**

Add the four HTML files with:
- shared top-level app shell
- page headers and nav
- placeholder-free page structure for each screen’s main job
- page data hooks needed by `prototype.js`

Do not fake behaviour yet beyond seeded layout hooks.

- [ ] **Step 4: Re-run the smoke check**

Run:
```bash
bash scripts/check_admin_ui_mockup.sh
```
Expected: PASS for file existence and shell markers, but FAIL until page-specific content is complete.

### Task 4: Implement Activity And DNS Records Screens

**Files:**
- Modify: `docs/mockups/admin-ui-v1/activity.html`
- Modify: `docs/mockups/admin-ui-v1/dns-records.html`
- Modify: `docs/mockups/admin-ui-v1/mock-data.js`
- Modify: `docs/mockups/admin-ui-v1/prototype.js`
- Modify: `docs/mockups/admin-ui-v1/styles.css`
- Modify: `scripts/check_admin_ui_mockup.sh`

- [ ] **Step 1: Add failing checks for activity and records**

Require:
- activity filters and blocked-only review state
- recent request rows with action context
- DNS records table with common record types
- add-record panel state
- concise DNS info help for record types and TTL

- [ ] **Step 2: Run the smoke check to verify failure**

Run:
```bash
bash scripts/check_admin_ui_mockup.sh
```
Expected: FAIL because the pages do not yet show the required activity or record-management structures.

- [ ] **Step 3: Build `activity.html`**

Implement:
- filter bar
- operational event rows
- selected or focused event detail area
- direct allow or block actions that deep-link to the policy pages

- [ ] **Step 4: Build `dns-records.html`**

Implement:
- zone context
- records table
- add-record panel
- validation and recently-added seeded states
- compact info help for `A`, `AAAA`, `CNAME`, `MX`, `TXT`, and `TTL`

- [ ] **Step 5: Re-run the smoke check**

Run:
```bash
bash scripts/check_admin_ui_mockup.sh
```
Expected: PASS for activity and DNS-records requirements.

### Task 5: Implement Allow-List And Block-List Screens

**Files:**
- Modify: `docs/mockups/admin-ui-v1/allow-list.html`
- Modify: `docs/mockups/admin-ui-v1/block-list.html`
- Modify: `docs/mockups/admin-ui-v1/mock-data.js`
- Modify: `docs/mockups/admin-ui-v1/prototype.js`
- Modify: `docs/mockups/admin-ui-v1/styles.css`
- Modify: `scripts/check_admin_ui_mockup.sh`

- [ ] **Step 1: Add failing checks for policy pages**

Require:
- populated and empty policy states
- add-rule panel state
- shared layout pattern between allow and block
- different tone between quieter allow and firmer block
- concise help for `allow list`, `block list`, and `matched rule`

- [ ] **Step 2: Run the smoke check to verify failure**

Run:
```bash
bash scripts/check_admin_ui_mockup.sh
```
Expected: FAIL because policy-page structure and state markers are not complete.

- [ ] **Step 3: Build `allow-list.html`**

Implement:
- searchable rule list
- exception-focused copy
- add/edit rule panel
- seeded empty and recently-added states

- [ ] **Step 4: Build `block-list.html`**

Implement:
- searchable rule list
- enforcement-focused copy
- add/edit rule panel
- seeded empty and recently-added states
- slightly firmer visual treatment than allow list

- [ ] **Step 5: Re-run the smoke check**

Run:
```bash
bash scripts/check_admin_ui_mockup.sh
```
Expected: PASS for the policy-management requirements.

## Chunk 3: Existing Page Integration And Review

### Task 6: Wire Existing Pages Into The Expanded Control Plane

**Files:**
- Modify: `docs/mockups/admin-ui-v1/index.html`
- Modify: `docs/mockups/admin-ui-v1/setup.html`
- Modify: `docs/mockups/admin-ui-v1/login.html`
- Modify: `docs/mockups/admin-ui-v1/dashboard.html`
- Modify: `docs/mockups/admin-ui-v1/mock-data.js`
- Modify: `docs/mockups/admin-ui-v1/prototype.js`
- Modify: `docs/mockups/admin-ui-v1/styles.css`
- Modify: `scripts/check_admin_ui_mockup.sh`

- [ ] **Step 1: Add failing integration checks**

Require:
- launcher links to the new pages and their seeded review states
- setup `Import config` routes into DNS records import-preview state
- dashboard quick actions use:
  - `Add DNS record`
  - `Allow domain`
  - `Block domain`
  - `Review blocked traffic`
- old dead-end quick actions are removed

- [ ] **Step 2: Run the smoke check to verify failure**

Run:
```bash
bash scripts/check_admin_ui_mockup.sh
```
Expected: FAIL because the old launcher or dashboard wiring is still present.

- [ ] **Step 3: Refresh the existing pages**

Update:
- `index.html` to expose the new pages and seeded states cleanly
- `setup.html` so import flows into DNS records review state
- `login.html` only where needed for shell continuity
- `dashboard.html` so it acts as overview and triage rather than a stubbed endpoint

- [ ] **Step 4: Re-run the smoke check**

Run:
```bash
bash scripts/check_admin_ui_mockup.sh
```
Expected: PASS.

### Task 7: Verify Browser Behaviour And Final Readiness

**Files:**
- Modify: `docs/mockups/admin-ui-v1/index.html`
- Modify: `docs/mockups/admin-ui-v1/setup.html`
- Modify: `docs/mockups/admin-ui-v1/login.html`
- Modify: `docs/mockups/admin-ui-v1/dashboard.html`
- Modify: `docs/mockups/admin-ui-v1/activity.html`
- Modify: `docs/mockups/admin-ui-v1/dns-records.html`
- Modify: `docs/mockups/admin-ui-v1/allow-list.html`
- Modify: `docs/mockups/admin-ui-v1/block-list.html`
- Modify: `docs/mockups/admin-ui-v1/mock-data.js`
- Modify: `docs/mockups/admin-ui-v1/prototype.js`
- Modify: `docs/mockups/admin-ui-v1/styles.css`
- Modify: `scripts/check_admin_ui_mockup.sh`

- [ ] **Step 1: Run file and smoke checks**

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

- [ ] **Step 3: Review the key URLs in a browser**

Check at minimum:
- `/index.html`
- `/setup.html`
- `/login.html`
- `/dashboard.html`
- `/activity.html`
- `/dns-records.html`
- `/allow-list.html`
- `/block-list.html`
- `/dashboard.html?state=low-data`
- `/activity.html?filter=blocked`
- `/dns-records.html?state=add-record`
- `/dns-records.html?state=import-preview`
- `/allow-list.html?state=add-rule`
- `/block-list.html?state=add-rule`

Confirm:
- the shared shell feels coherent
- overview still reads as triage
- the new management pages feel concrete enough to review
- info panels are concise and not annoying
- mobile layout still works

- [ ] **Step 4: Run final verification**

Run:
```bash
git diff --check
bash scripts/check_admin_ui_mockup.sh
```
Expected: PASS.
