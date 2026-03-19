# Admin UI High-Fi Mock-up Design

**Date:** 2026-03-18

## Goal

Produce a high-fidelity, mobile-first admin UI mock-up for OpenFiltr that is demoable, visually coherent with the existing logo, and useful for validating UX decisions before any production frontend architecture is locked in.

## Why

The repository is backend-first and currently ships no real frontend foundation. Building a production UI now would be premature because:

- issue `#41` explicitly treats the admin UI as a future concern after backend stabilisation
- browser-session hardening and broader frontend architecture are not settled enough to justify implementation-first work
- closed legacy frontend issues describe an intended surface, not a trustworthy current plan

The immediate need is a mock-up that lets the team judge branding, layout, flow, information density, and click count before reopening or replacing frontend issues.

## Decisions

### Product shape

- This phase delivers a mock-up, not a production UI.
- Scope is limited to:
  - first-run setup step 1: create admin account
  - first-run setup step 2: choose `import config` or `continue to dashboard`
  - login
  - main dashboard
- The dashboard should make system state understandable within the first mobile viewport.
- The experience should minimise clicks and avoid deep navigation for core understanding.

### Visual direction

- The UI is clean, flat, simple, and mobile-first.
- The existing OpenFiltr logo carries the visual personality and broad colour range.
- The interface itself stays restrained, with neutral surfaces and one blue-teal accent.
- Avoid glossy effects, glassmorphism, heavy gradients, and decorative chrome.
- Layout clarity should come from spacing, typography, border contrast, and hierarchy rather than visual tricks.

### Branding and colour

Use the existing logo from the shared `.github` asset as the brand anchor.

Source of truth for the mark:

- `https://raw.githubusercontent.com/openfiltr/.github/main/assets/logo.svg`

For the mock-up implementation phase, copy that asset into the prototype tree as a local file rather than hotlinking it.

UI palette direction:

- background: warm off-white
- surface: white
- border: soft grey
- primary text: charcoal
- secondary text: muted slate
- accent: blue-teal, used for primary actions, active states, and key highlights only

The logo's pink, orange, yellow, green, cyan, blue, and purple tones should not be spread across the wider interface. Doing so would make the product look noisy and less trustworthy.

### Interaction principles

- Mobile-first, responsive by default
- Large tap targets and obvious primary actions
- One clear action per screen section
- Minimal form fields in the setup flow
- No explanatory filler copy unless it directly reduces confusion
- Dashboard content should answer:
  - is the service healthy
  - what is it doing
  - what needs attention

## Prototype architecture

The first artefact should be a static high-fidelity prototype with representative sample data and no backend wiring.

Expected characteristics:

- explicitly treated as a mock-up, not an app shell
- implemented as static pages that can be opened locally and shared for review
- uses local sample data to render realistic dashboard states
- uses simple client-side interactions only where needed to demonstrate flow between setup, login, and dashboard
- avoids choosing a long-term frontend stack prematurely

Prototype structure:

- multi-page static prototype
- product-flow pages:
  - `setup.html`
  - `login.html`
  - `dashboard.html`
- optional `index.html` review launcher that links directly to the available pages and review states

State handling in this phase:

- setup and login use lightweight local form behaviour
- dashboard uses local sample data only
- review-only alternate states may be reached with query parameters such as `?state=error` or `?state=low-data`
- those review controls are for critique and demo convenience, not part of the product UI

Happy-path transitions:

- successful submit on setup step 1 advances to setup step 2
- choosing `continue to dashboard` on setup step 2 advances to the dashboard
- choosing `import config` on setup step 2 opens the local stub state and does not leave the setup flow
- successful submit on login advances to the dashboard
- the prototype should be reviewed through a lightweight local static server rather than `file://` so page transitions and review-state URLs behave predictably

Recommended storage location for the prototype:

- `docs/mockups/admin-ui-v1/`

This keeps the work clearly separate from backend code and avoids implying that a production-ready frontend foundation already exists.

## Screen map

### 1. First-run setup, step 1

Purpose:

- create the first admin account with the minimum viable fields

Expected content:

- logo and product name
- short heading explaining that OpenFiltr needs an initial administrator
- username field
- password field
- confirm password field
- submit action

Expected states:

- default form
- inline validation for missing input
- inline validation for weak password input, matching the backend rule of at least 8 characters
- inline validation for password mismatch

Trigger in the prototype:

- submitting invalid setup input must show validation on the same page
- the validation review state must also be directly reachable, for example with `setup.html?state=validation-error`

### 2. First-run setup, step 2

Purpose:

- let the operator choose the next action without forcing unnecessary configuration

Expected content:

- confirmation that setup is complete
- two clear options:
  - import config
  - continue to dashboard

Behaviour in this phase:

- `import config` is a mock-up stub only
- it may open a lightweight local panel or message that makes clear the import flow is not designed in this slice
- it must not imply that real import behaviour or file handling already exists in the prototype

Expected states:

- default action selection
- optional short note that configuration can also be imported later

### 3. Login

Purpose:

- provide the simplest possible re-entry point

Expected content:

- compact logo lock-up
- username field
- password field
- sign-in action

Constraints:

- no marketing copy
- no multi-column layout on mobile
- no unnecessary secondary actions in the primary viewport

Expected states:

- default
- invalid credentials error banner or inline message

Trigger in the prototype:

- submitting the local invalid state must show the error on the same page
- the same state must also be directly reachable for review, for example with `login.html?state=error`

### 4. Dashboard

Purpose:

- give the operator immediate confidence about system health and recent behaviour

Above-the-fold mobile priority:

- service health status
- key stats
- recent activity preview

Secondary sections below:

- top blocked domains
- small set of quick actions

Dashboard content blocks:

- service health card
  - running state
  - HTTP/API health
  - DNS service state
- key stats row or stack
  - total requests
  - blocked requests
  - allowed requests
  - block rate
- recent activity list
  - newest events first
  - concise row design for domain, action, client, and time
- top blocked domains list
  - domain
  - block count
- low-emphasis quick actions
  - import config
  - view activity
  - manage rules

## Information architecture

For this mock-up, navigation should remain intentionally shallow.

- Only `setup.html`, `login.html`, and `dashboard.html` are interactive in this slice.
- Within the wider product navigation, the dashboard is the only fully active product destination.
- Other product areas can appear as low-emphasis navigation placeholders if needed for realism, but they should not distract from the demo flow and must not behave like finished pages.
- Mobile navigation should favour a compact bottom bar or similarly simple pattern once more destinations exist, but the prototype should not pretend that deeper IA has been solved.

## Content and tone

- Write copy in plain language.
- Prefer short headings and explicit labels over clever phrasing.
- Avoid internal jargon where a simpler term exists.
- Keep the tone competent and friendly, not playful and not enterprise-theatrical.

## Error handling and edge cases

The mock-up should show enough non-happy-path detail to validate the UX without ballooning scope.

Include:

- setup validation state
- login failure state
- dashboard low-data state for recent activity and top blocked domains

Required presentation:

- login failure state must show a clear inline or banner message without breaking layout
- dashboard low-data state must keep service health and key stats visible while replacing list-heavy sections with calm empty-state copy

Required triggers in the prototype:

- setup validation state reachable by invalid submission and by direct review URL
- login failure state reachable by local invalid submission and by direct review URL
- dashboard low-data state reachable by direct review URL such as `dashboard.html?state=low-data`

Exclude for now:

- password reset
- session expiry flows
- CSRF or browser security messaging
- import workflow details
- full settings or account management

## Issue strategy

- Keep issue `#41` as the high-level UI placeholder.
- Treat the closed frontend issues as historical backlog input only.
- After the mock-up is approved and the backend surface is stable enough, reopen or replace the relevant issues with narrower tickets tied to the approved design.
- Do not revive the old issue list wholesale; it predates the current backend-first reality and the absence of a frontend foundation.

## File map

Likely new files for the next phase:

- `docs/superpowers/plans/2026-03-18-admin-ui-high-fi-mockup.md`
- `docs/mockups/admin-ui-v1/login.html`
- `docs/mockups/admin-ui-v1/setup.html`
- `docs/mockups/admin-ui-v1/dashboard.html`
- `docs/mockups/admin-ui-v1/styles.css`
- `docs/mockups/admin-ui-v1/mock-data.js`

Optional review helper:

- `docs/mockups/admin-ui-v1/index.html` as a non-product launcher for jumping to pages and review states

Potential supporting assets:

- `docs/mockups/admin-ui-v1/assets/openfiltr-logo.svg`

## Risks

- The logo is visually rich; careless colour reuse will make the UI look unfocused.
- A desktop-first layout would fail the stated goal and produce the wrong feedback.
- Building too much navigation around one dashboard mock-up would create fake certainty about the wider IA.
- Turning the mock-up into a framework bootstrap would waste time and bias future architecture decisions.

## Verification strategy

- review screens at narrow mobile widths first, then tablet and desktop
- verify the first mobile viewport shows service health, key stats, and recent activity without excessive scrolling
- confirm the setup and login flows use the minimum viable number of inputs and actions
- check that the palette stays mostly neutral and that brand colour remains concentrated in the logo plus a restrained accent
- validate that the prototype remains obviously separate from any production frontend commitment
