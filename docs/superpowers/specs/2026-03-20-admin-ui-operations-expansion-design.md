# Admin UI Operations Expansion Design

**Date:** 2026-03-20

## Goal

Expand the admin UI mock-up so it no longer stops at generic health and placeholder actions.

The revised prototype should show how a self-hosting admin would:

- inspect resolver activity
- review suspicious or blocked domains
- add and edit DNS records
- manage allow-list and block-list rules

This remains a static mock-up. The point is to make the operational story believable, not to imply a production frontend stack or real persistence.

## User and product posture

### Primary user

A self-hosting admin managing one OpenFiltr deployment.

### Primary job

Inspect what the resolver is doing and act on suspicious domains quickly.

### Product tone

The interface should feel:

- calm
- trustworthy
- operational
- lightly sharper when enforcement or policy actions are involved

It should stay aligned with the current mock-up direction: credible control plane, restrained branding, no theatrical security styling, no fake enterprise theatre.

## Why this change is needed

The current mock-up proves the shell and the basic auth flow, but it still avoids the actual admin work.

Current gaps:

- quick actions suggest capability without showing it
- `Import config` and `Manage rules` still behave like placeholders
- there is no concrete surface for DNS records
- there is no concrete surface for allow or block policy management
- the dashboard still does too much hinting and not enough demonstrating

The prototype needs to stop hand-waving. It should show the main operator tasks in enough depth for design review without pretending the whole product is already built.

## Design direction

Keep the existing control-plane direction as the baseline, then push it further towards a real self-hosted operations surface.

### Visual character

- clean, pale control-plane shell
- restrained blue-led accent system
- compact, authored layout rhythm
- slightly firmer contrast and emphasis on policy or blocked states
- sparse but intentional network-operations cues

### Do not introduce

- decorative charts
- neon or cyber-security styling
- oversized hero sections
- modal-heavy CRUD
- long explanatory marketing copy

## Product structure

The prototype should expand from four screens to a small but coherent control plane.

### Screen set

- `index.html`: internal review launcher only
- `setup.html`: first-run account creation
- `login.html`: sign-in
- `dashboard.html`: overview and triage landing page
- `activity.html`: recent resolver activity and filtering
- `dns-records.html`: zone and DNS record management
- `allow-list.html`: exception management
- `block-list.html`: enforcement management

## Shell and navigation

### Shared shell

Overview, activity, DNS records, allow list, and block list should all use one shared application shell.

Shell requirements:

- steady top bar with product identity and environment context
- lightweight navigation framing
- same spacing, borders, radii, badges, and section headers across pages
- content area that feels authored rather than stretched

### Navigation model

Primary navigation should include:

- Overview
- Activity
- DNS records
- Allow list
- Block list

The nav should feel like a compact control-plane frame, not a bloated application menu.

### Quick actions

The overview page should include a quick-actions panel for immediate operational actions:

- Add DNS record
- Allow domain
- Block domain
- Review blocked traffic

This panel should replace the current dead-end quick actions rather than sit beside them.

These actions should route to real mock-up pages, not placeholders.

## Screen design

### Launcher

Treat `index.html` as an internal preview page only.

It should:

- link to the full auth and control-plane flow
- expose seeded review states for the new pages
- stay quieter than the product screens

### Setup

Keep setup focused and short.

Changes:

- keep the existing two-step account creation structure
- keep the hand-off into the product
- replace the current `Import config` dead-end with a meaningful next-step route into DNS records in an import-ready review state

The setup screen should still feel like product entry, not like a full admin workspace.

### Login

Keep login minimal.

Changes:

- preserve the quiet sign-in surface
- keep the seeded error state
- make the hand-off clearly lead into the control-plane overview

### Overview

`dashboard.html` should remain the landing page, but it should behave more like an operational triage surface.

Required sections:

- current service health
- request totals and policy effect
- recent activity preview
- top blocked domains preview
- quick actions
- lightweight nav frame

The page should answer:

- is the system healthy?
- what is being blocked right now?
- what should I do next?

### Activity

`activity.html` should be a dedicated inspection surface.

Required content:

- filter bar
- recent request list
- per-row domain, client, category, action, matched rule, time
- compact detail view for the selected or focused event
- obvious actions to allow or block from the inspected domain

The point of this page is review and triage, not analytics theatre.

### DNS records

`dns-records.html` should show believable day-to-day DNS administration.

Required content:

- zone context and zone selector if needed
- search and basic filtering
- records table
- add-record action
- add-record form or side panel
- visible common record types first: `A`, `AAAA`, `CNAME`, `MX`, `TXT`
- concise validation and recent-success state

The surface should feel like structured CRUD, not like a toy form.

### Allow list

`allow-list.html` should show explicit exceptions to blocking behaviour.

Required content:

- search and filtering
- existing allow rules
- scope or target detail
- short reason or note
- recent hit or usage context
- add or edit rule panel

Tone:

- quiet
- precise
- less visually aggressive than block list

### Block list

`block-list.html` should share the same management pattern as allow list, but it should feel slightly firmer because it is the enforcement surface.

Required content:

- search and filtering
- populated rule list
- reason, scope, and recent-hit context
- add or edit rule panel
- stronger visual emphasis around enforced behaviour

This page should look more decisive, not louder.

## Interaction model

### General rule

The mock-up should stop at the point where user intent is clear.

It should show:

- open and closed panels
- seeded validation states
- seeded success states
- seeded empty states
- prefilled fast-action states

It should not fake:

- backend persistence
- live syncing
- real auth depth
- complex settings that do not exist elsewhere in the product

### Deep-link behaviour

Quick actions from overview should open meaningful target states.

Examples:

- `Add DNS record` -> `dns-records.html?state=add-record`
- `Allow domain` -> `allow-list.html?state=add-rule&domain=...`
- `Block domain` -> `block-list.html?state=add-rule&domain=...`
- `Review blocked traffic` -> `activity.html?filter=blocked`

The review launcher should expose similar seeded entry points for critique.

### Replacing current stubs

Existing stubs should become real mock-up flows:

- setup `Import config` should open the DNS records page in an import-ready or import-preview state
- dashboard `Review blocked traffic` should open `activity.html?filter=blocked`
- dashboard `Block domain` should open `block-list.html?state=add-rule&domain=...`
- dashboard `Allow domain` should open `allow-list.html?state=add-rule&domain=...`
- dashboard `Add DNS record` should open `dns-records.html?state=add-record`

The old dashboard actions should be replaced, not retained as a second action set.

## Micro-help pattern

The prototype should add small info triggers next to terms that reviewers may not know.

### Terms in scope

- `A`
- `AAAA`
- `CNAME`
- `MX`
- `TXT`
- `TTL`
- `allow list`
- `block list`
- `matched rule`

### Behaviour

- use a compact info icon beside the relevant term or label
- open a small anchored panel, not a modal
- only one help panel open at a time
- close on outside click, `Esc`, or opening another panel
- on mobile, place the panel below the trigger rather than letting it drift off-screen

### Copy style

Each panel should contain:

- a short heading
- one short definition
- one tiny example where useful

Example:

- `A record`: points a hostname to an IPv4 address. Example: `app.example.com -> 203.0.113.10`

These help panels should feel like compact operator guidance, not onboarding bubbles.

## Data and seeded states

`mock-data.js` should expand to include realistic data for all new pages.

### Overview states

- normal
- low data

### Activity states

- normal feed
- blocked-only filtered view
- quiet or low-activity state

### DNS records states

- populated zone
- empty zone
- add-record open
- validation error
- record added
- import-ready or import-preview state

### Allow-list states

- populated list
- empty list
- add-rule open
- rule added

### Block-list states

- populated list
- empty list
- add-rule open
- rule added

## Technical shape

Keep the same static prototype model.

### Prototype unit boundaries

Implementation should stay split into clear responsibilities:

- page HTML files own page-specific layout and review-state markup
- `styles.css` owns shared tokens, primitives, shell styling, and small reusable interaction surfaces
- `mock-data.js` owns all seeded fixtures and named review states
- `prototype.js` owns query-param state resolution, seeded page boot logic, quick-action prefill behaviour, and micro-help toggling

No single page should contain bespoke interaction patterns that bypass the shared shell or the shared prototype helpers without a strong reason.

### Files to add

- `docs/mockups/admin-ui-v1/activity.html`
- `docs/mockups/admin-ui-v1/dns-records.html`
- `docs/mockups/admin-ui-v1/allow-list.html`
- `docs/mockups/admin-ui-v1/block-list.html`

### Files to update

- `docs/mockups/admin-ui-v1/index.html`
- `docs/mockups/admin-ui-v1/setup.html`
- `docs/mockups/admin-ui-v1/login.html`
- `docs/mockups/admin-ui-v1/dashboard.html`
- `docs/mockups/admin-ui-v1/mock-data.js`
- `docs/mockups/admin-ui-v1/prototype.js`
- `docs/mockups/admin-ui-v1/styles.css`
- `scripts/check_admin_ui_mockup.sh`

### Implementation boundary

- keep Tailwind via CDN for the mock-up
- keep local JavaScript for stateful prototype behaviour
- prefer one shared shell and reusable primitives over one-off page styling
- keep interaction logic shallow, clear, and reviewable

## Review expectations

The expanded prototype should be judged on:

- whether it now shows believable admin work rather than vague placeholder actions
- whether overview still feels like a triage page rather than an everything page
- whether DNS records and policy management feel concrete enough to review
- whether the allow and block surfaces feel related but distinct
- whether the info panels help without turning into documentation clutter
- whether the whole control-plane set still feels like one product

## Non-goals

- no production frontend scaffold
- no real persistence or backend calls
- no fake advanced analytics
- no attempt to mock the entire product surface
- no speculative settings or features outside DNS records, activity, and allow/block policy management
