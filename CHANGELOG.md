# Changelog

## [1.2.0]

### Changed
- **Headerless, "straight to the point" landing page.** Removed the top app-bar entirely; brand wordmark and theme toggle now live in a quiet, borderless utility line that renders on every page.
- **Seamless create form.** The floating elevated card is gone — the form is now set directly into the page as one continuous "sheet" alongside the stats ledger, joined by a single hairline divider. New "Drop a secret." hero, grouped fields under shared eyebrow labels, a compact Text/File switch, a tinted secret well, "Delete after first view" toggle, and an integrated "receipt" success panel.
- File mode now accepts **drag-and-drop** onto the well (previously the "drop it here" label was decorative).

### Fixed
- **Dark mode now paints a dark ground.** Previously the global stylesheet forced `.v-application { background: transparent !important }` over a hard-coded light gradient, so the dark theme's background never actually applied. The page now uses the Vuetify theme background token with a theme-aware radial wash, and the mobile browser chrome color (`theme-color`) tracks the active theme.

## [1.1.0]

### Added
- **Gamified server stats dashboard** on the landing page — a live "server ledger" beside the create form showing lifetime KPIs (Secrets Delivered, Text Secrets, Files Shared, Data Encrypted) and a live "Held right now" count.
  - The hero "Secrets delivered" figure **ticks +1 with a count-up animation** the moment you create a secret.
  - A 30-day stacked activity chart (text vs. file), a text/file split bar, and a milestone progress meter — all hand-rolled inline SVG (no charting-library dependency, keeps the bundle lean).
- New public, aggregate-only **`GET /stats`** endpoint (30s in-memory cache, no per-secret data ever exposed). Toggle with the `STATS_ENABLED` env var (default `true`).
- Lifetime counters (`stat_counters`) and per-day activity (`stat_days`) tables. Counters are **monotonic**: incremented at creation and never decremented by expiry or one-time retrieval, so totals are honest despite secrets being ephemeral. A `since` timestamp keeps the "lifetime" figure honest on pre-existing installs.

### Changed
- Modernized the landing page: new "Share secrets that vanish." hero, refined two-column layout (form + ledger), Space Grotesk display type, and a tasteful, `prefers-reduced-motion`-aware motion budget.
- Extracted the post-create success panel into a reusable `SecretResult` component.
- Send creation now verifies the database write succeeded before returning a link (previously the insert error was ignored).

### Privacy
- **Removed all third-party CDN requests on page load.** Inter and Space Grotesk fonts are now self-hosted via `@fontsource`, and the `animate.css` CDN link was dropped (its handful of animations are re-implemented locally). A privacy tool no longer phones home.

## [10.0.10]

### Added
- Added dedicated curl-friendly creation endpoints:
  - `POST /send/text` for text pastes
  - `POST /send/file` for file pastes
- Added support for raw-body paste creation:
  - Text via raw request body to `/send/text`
  - Files via raw binary body to `/send/file` with `filename` query param (or `X-Filename` header)
- Expanded API documentation in `Readme.md` with practical curl examples for text and file flows.

### Changed
- Kept `POST /send` backward-compatible while improving create handling for CLI workflows.
- Hardened file storage handling to fall back to a safe temp directory when `STORAGE_PATH` is unset.

## [1.0.9]

### Changed
- Improved the "Type" selector on the Create Secret page to use a visually appealing slider toggle with icons for "Text" and "File".
- After creating a secret, all input fields are now cleared for a better user experience.
- Enhanced file upload validation to reliably detect selected files and prevent false "Please select a file" errors.

### Fixed
- Fixed an issue where file uploads would sometimes incorrectly show "Please select a file" even when a file was chosen.

### UI/UX
- The form now resets when navigating to the Create page via the logo or Create+ button.

## [1.0.8]
### UI Enhancements
- **Redesigned Header and Footer**: Modernized header/footer with theme toggle and improved visual consistency.
- **Improved 404 Page**: Updated design for better responsiveness and refined typography.

### Theme and Styling Updates
- **Custom Themes**: Introduced custom light/dark themes and theme switching functionality.
- **Global Styling Updates**: Updated font to "Inter" and added a gradient background.

### Component Enhancements
- **Reusable Password Input**: Created a component for password handling, including visibility and generation.
- **Secret Display Component**: Added a component for secret copying and file downloading with alerts.

### Code Cleanup and Refactoring
- **Removed Unused API Code**: Deleted obsolete `api.js` to simplify codebase.
- **Form Reset Logic**: Refactored `Create.vue` to use a centralized store for form reset.

## [1.0.7]
### Added
- Customizable application title and description through environment variables
- New build arguments in Dockerfile: VITE_APP_TITLE and VITE_APP_DESCRIPTION
- Support for title and description customization in docker-compose configuration

### [1.0.6]
* `cmd/server/main.go`: Added rate limiting to POST requests.
* `ui/src/pages/Create.vue`: Added error handling for 429 responses.

### [1.0.5]
UI Enhancements:

* `ui/index.html`: Added a link to the Animate.css library for animation effects.
* [`ui/src/App.vue`: Integrated animation classes into various elements, including the header logo, buttons, and alerts.

Password Management Enhancements:
* `ui/src/pages/Create.vue`: Added functionality to toggle password visibility and generate random passwords, along with corresponding tooltips and animations.
404 Error Page Improvements:
* `ui/src/pages/Error404.vue`: Redesigned the 404 error page with a more user-friendly card layout, including an icon, message, and a button to navigate back home.

Other Enhancements:
* `ui/src/pages/Create.vue`, `ui/src/pages/View.vue`: Applied animation classes to various elements to provide a more dynamic user experience.

### [1.0.4]
- Reworked CD pipelines to follow semver tagging
- added version.yaml

### [1.0.1]

#### Added
- Initial unified Dockerfile combining backend and frontend.
- Support for Traefik reverse proxy configuration.
- Environment variables for configuring API URLs dynamically.
- Ability to copy the generated shareable link with improved formatting.
- New CORS configuration to support wildcard origins in development mode.
- Automatic HTTPS redirection using Traefik middlewares.

#### Fixed
- Corrected Vite's `VITE_API_URL` handling to avoid hardcoded URLs.
- Resolved 404 errors for static assets when accessed via Traefik.
- Fixed MIME type issues for serving CSS files behind Traefik.

---

### [1.0.0] - 2024-12-18

#### Added
- Secure one-time secret sharing with encrypted text and file support.
- Password protection for shared secrets.
- One-time retrieval mechanism to ensure secrets are accessed only once.
- Expiration settings for shared secrets.
- Responsive UI built with Vue.js and Vuetify.
- Dockerized deployment with `docker-compose.yml` for easy setup.
- Nginx reverse proxy configuration for serving the frontend and API.
- Health checks for PostgreSQL in `docker-compose.yml`.

#### Changed
- Updated Docker images to use multi-stage builds for backend and frontend.
- Improved project documentation and added sections for installation, configuration, and deployment.

---
