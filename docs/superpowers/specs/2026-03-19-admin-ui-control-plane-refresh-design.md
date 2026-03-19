# Admin UI Control-Plane Refresh Design

**Date:** 2026-03-19

## Goal

Redesign the existing admin UI mock-up so it looks like a credible modern product demo rather than a polished wireframe.

This refresh keeps the same prototype scope:

- first-run setup
- login
- main dashboard
- review launcher

The change is visual, structural, and presentational. It is not a move to a production frontend stack.

## Why

The first high-fidelity mock-up validated the basic flow, but its presentation still feels cheap. The layout is too card-heavy, the shell is too generic, and the overall result does not yet look like a product someone would trust to run network infrastructure.

The revised mock-up should feel:

- modern
- simple
- professional
- clean
- product-grade
- mobile-first

It should also stay visibly aligned with the OpenFiltr logo without turning the interface into a rainbow brand exercise.

## Direction

### Primary reference style

The visual baseline should lean closer to DigitalOcean than Cloudflare:

- calmer blue-led control-plane feel
- more whitespace
- softer but still disciplined surfaces
- less overt enterprise severity

### Secondary reference influence

Add a small amount of Cloudflare and GitHub product discipline:

- firmer hierarchy
- cleaner operational lists
- sharper panel rhythm
- restrained shadow use
- compact, utility-like status treatments

The result should not look like a clone of any of those products. It should look like OpenFiltr adopting the same level of seriousness.

## Product identity

OpenFiltr needs a DNS and network operations flavour without falling into cliché infrastructure art.

Use subtle cues only:

- route-like line work in quiet background treatments
- resolver and network language in metadata and headings
- operational spacing and alignment that feels like a control surface
- occasional monospace treatment for timestamps, server identifiers, and other machine-like data

Do not use:

- loud topology illustrations
- fake terminal skins
- neon security styling
- colourful charts or decorative graphs

## Visual system

### Brand fit

- Keep the logo as the most colourful element in the experience.
- Surround it with enough whitespace to make it feel premium.
- Do not spread the logo palette throughout the wider UI.

### Palette

Use a restrained, product-style palette:

- background: near-white with a cool slate cast
- elevated surface: white
- muted surface: pale blue-grey
- borders: soft slate-grey
- primary text: deep navy-charcoal
- secondary text: steel blue-grey
- primary accent: one strong product blue
- status colours: muted green, amber, and red, used sparingly

The accent should feel more like product chrome than marketing colour.

### Typography

- Use a clean sans-serif stack with modern product tone.
- Headings should be firm and compact, not oversized.
- Body text should be plain and readable.
- Operational metadata can use a monospace stack in small doses.

### Surfaces and depth

- Flat first, with restrained depth only where needed.
- Prefer borders and spacing over shadow.
- Radii should be moderate and consistent.
- Avoid soft consumer-style pillowness.

### Component system feel

The mock-up should feel like it comes from a disciplined internal component library, not from page-by-page custom styling.

That means:

- repeated UI patterns should look intentional and reusable
- spacing, radius, borders, focus states, and typography scales should follow a clear token set
- panels, buttons, inputs, badges, step indicators, section headers, and list rows should feel like shared primitives
- pages should feel assembled from the same system rather than redesigned from scratch each time

The goal is not to literally build a reusable component library in this prototype. The goal is to design with that library-like discipline so later implementation can stay visually consistent across the whole product.

## Layout principles

### Overall shell

- Mobile-first by default.
- Desktop should feel like a real control plane, not stretched mobile cards.
- Use a steady top bar with brand, environment context, and light navigation framing.
- Keep the content width disciplined so the interface feels authored rather than sprawling.

### Density

- Generous whitespace, but not empty luxury spacing.
- Prioritise rhythm and grouping over decorative separation.
- Let content breathe without pushing critical information below the fold.

### Component style

- Fewer repeated cards.
- More deliberate sections with clear internal hierarchy.
- Lists and tables should feel closer to GitHub than to a dashboard template.
- Status chips should feel operational.
- Primary buttons should be obvious without dominating the page.
- Shared primitives should be recognisable across screens, especially buttons, input shells, section frames, badges, and list treatments.

## Screen design

### Review launcher

Treat `index.html` as an internal review surface only.

- keep it simple
- make it look like a disciplined internal preview page
- avoid letting it become a second product home page

### Setup

Setup should feel like a proper first-run product screen:

- restrained branded header
- short reassurance copy
- a focused form panel
- cleaner two-step progression
- stronger hierarchy between the account creation step and the completion step

Step two should feel like a controlled hand-off into the product, not like another form.

### Login

Login should be even more restrained than setup:

- one strong panel
- minimal helper copy
- clear validation and error state
- enough top-level branding to feel like a product entry point, not a generic auth box

### Dashboard

The dashboard should feel operational within seconds.

Top of page:

- compact top bar
- environment or host context
- immediate system summary

Primary body:

- service health
- key request totals
- recent activity in a structured operational list

Secondary body:

- top blocked domains
- a small set of obvious actions
- placeholder navigation framing that hints at future depth without pretending it exists already

The dashboard should rely more on layout and information hierarchy than on stacked card repetition.

## Interaction principles

- keep clicks low
- keep primary actions obvious
- use short validation copy
- use strong focus states
- keep motion minimal
- avoid decorative transitions

The interface should feel stable and dependable, not animated.

## Technical approach for the mock-up

Use Tailwind CSS through the CDN build for this prototype refresh only.

Constraints:

- no frontend build pipeline
- no framework choice implied
- no production stack commitment
- keep the existing static multi-page prototype model

Implementation shape:

- add Tailwind via CDN to each prototype page
- define a small inline Tailwind config for colours, typography, spacing, shadows, and radius
- define a small set of reusable utility patterns and component-like class groupings so the pages read as one coherent system
- keep existing local JavaScript for flow and review states
- either remove or sharply reduce the custom stylesheet so Tailwind owns the visual system

## Files in scope

The refresh should stay inside the prototype surface:

- `docs/mockups/admin-ui-v1/index.html`
- `docs/mockups/admin-ui-v1/setup.html`
- `docs/mockups/admin-ui-v1/login.html`
- `docs/mockups/admin-ui-v1/dashboard.html`
- `docs/mockups/admin-ui-v1/mock-data.js`
- `docs/mockups/admin-ui-v1/prototype.js`
- `docs/mockups/admin-ui-v1/styles.css`
- `scripts/check_admin_ui_mockup.sh`

## Review expectations

The revised prototype should be judged on:

- whether it feels product-grade
- whether it looks modern without looking generic
- whether the dashboard feels trustworthy for DNS operations
- whether mobile still prioritises health, totals, and recent activity
- whether branding feels aligned with the logo without becoming noisy
- whether the screens feel like they belong to one reusable design system rather than one-off page mock-ups

## Non-goals

- no real backend integration
- no React or production frontend scaffold
- no full navigation architecture
- no analytics-heavy visualisation layer
- no implementation of config import
- no attempt to settle long-term frontend technology choices
