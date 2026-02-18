# Railway UI/UX Analysis — Reference Document for Lucity Dashboard

## Context

This document captures a detailed analysis of Railway's dashboard UI/UX patterns, aimed at frontend engineers and UI/UX designers implementing the Lucity dashboard. Railway is the closest competitive reference for Lucity's developer experience. This analysis was performed by navigating through a live Railway account with a project containing a GitHub-deployed web service and a Postgres database.

---

## 1. Global Chrome & Navigation

### Top Navigation Bar
- **Dark theme throughout**
- **Left side**: Railway logo (small icon) → avatar → workspace name ("Christian Blattler's Projects") with plan badge ("Trial") → chevron dropdown for workspace switching
- **Right side**: "Help" text link, notification bell icon, user avatar (circular)
- **Announcement banner** (top, full-width): colored text (green) with a dismissable (X) CTA — e.g. funding announcements. Sits above the main nav.

### Breadcrumb Navigation (Project Context)
When inside a project, the top bar transforms into a breadcrumb:
- `Railway logo / avatar / project-name ▾ / environment-name ▾`
- Both project name and environment name are **dropdown selectors** (chevrons)
- **Right side tabs**: Architecture | Observability | Logs | Settings
- **Far right**: billing badge ("30 days or $4.97 left"), share icon, changelog icon, notification bell, user avatar

### Environment Switcher
- Dropdown from the environment name in breadcrumb
- Lists all environments with a **checkmark** next to the active one
- Shows: environment name, last updated timestamp, user avatar
- Footer: "+ New Environment" action
- No search — simple flat list (environments are typically few)

---

## 2. Projects Dashboard (`/dashboard`)

### Layout
- **Header**: "Projects" title (left), "+ New" button (right, purple with white text, rounded pill shape)
- **Upgrade banner**: green-bordered bar with trial info and "Choose a Plan" CTA
- **Project count & sort**: "1 Project" label, "Sort By: Recent Activity" dropdown
- **View toggle**: "Views" label → "Architecture" | "List" segmented control (right-aligned)

### Architecture View (Default)
Each project is a **card** (~300px wide):
- **Header**: project name in bold white text
- **Body**: miniature dotted canvas background showing service icons (GitHub logo, Postgres elephant, etc.) as small square tiles with rounded corners
- **Footer**: green status dot + environment name + "2/2 services online" text

### List View
Compact single-row per project:
- Project name (left, bold white)
- Status info (right): green dot + "production · 2/2 services online"
- Full-width clickable row with subtle border/background

---

## 3. Project Architecture Canvas (`/project/:id`)

### Canvas Area
- **Full-viewport interactive canvas** with a dark background and subtle dot grid pattern
- Services appear as **draggable cards** arranged vertically (default auto-layout)
- Canvas supports **zoom** (+ / - buttons) and **pan** (drag empty space)
- Canvas is the background; service detail panels slide in from the right as overlays
- **Implementation**: use [vue-flow](https://vueflow.dev/) for the canvas — it provides nodes, edges, zoom, pan, minimap, and drag-and-drop out of the box

### Left Toolbar (Vertical, floating)
A vertical pill-shaped toolbar on the left edge with icon-only buttons:
1. **Grid/minimap** icon (top)
2. **Zoom in** (+)
3. **Zoom out** (-)
4. **Fit to screen** (expand arrows)
5. **Undo** (curved arrow left)
6. **Redo** (curved arrow right)
7. **Layers/stack** icon (bottom)

### Top-Right Actions
- "Sync" button (ghost style, with refresh icon)
- "+ Create" button (outlined, white text)

### Service Cards (on canvas)
Two distinct card types observed:

#### Web Service Card
- **Icon**: GitHub/source provider logo (circular, ~24px)
- **Title**: service name in bold ("zeitlos-website")
- **Subtitle**: public domain ("zeitlos.up.railway.app") in muted text
- **Status**: green dot + "Online" text
- **Border**: subtle `1px` border, rounded corners (~12px), semi-transparent background

#### Database Card
- **Icon**: database logo (Postgres elephant) — branded/colored
- **Title**: "Postgres" in bold
- **Status**: green dot + "Online"
- **Additional info**: volume icon + "postgres-volume" (shows attached storage)
- **Border**: same style as web service, slightly taller due to volume info

---

## 4. Service Detail Panel (Right Slide-In)

Clicking a service card opens a **right-side panel** (~55% width) that slides over the canvas. The canvas remains visible on the left (~45%) with the clicked card highlighted (purple border).

### Panel Header
- Source provider icon (GitHub, Postgres logo, etc.)
- **Service name** in large bold text ("zeitlos-website") with an edit (pencil) icon
- **Close button** (X) in top-right corner
- Escape key also closes the panel

### In-Panel Navigation (Stacking)
The panel supports **in-panel drill-down navigation**. When you click "View logs" on a deployment, the panel doesn't open a new panel — it **navigates within the same panel** to a deployment detail view:

1. **Service level**: `zeitlos-website` → tabs: Deployments | Variables | Metrics | Settings
2. **Deployment level** (drill-down): `zeitlos-website / 2295d8ef` with "Active" badge → tabs: Details | Build Logs | Deploy Logs | HTTP Logs | Network Flow Logs

The panel header becomes a **breadcrumb**: clicking the service name ("zeitlos-website") navigates back to the service level. This is URL-driven — each level has its own URL with hash fragments (`#deploy`, `#details`).

The **Deployment Details** sub-view shows:
- Status banner ("Deployment successful") with expandable details
- Variable count
- Source info: commit message, repo name, branch (with icons)
- Configuration section with "Pretty" / "Code" toggle showing: Build (builder type), Deploy (region, replicas, restart policy)

### Tab Bar
Tabs differ by service type:

**Web Service**: Deployments | Variables | Metrics | Settings
**Database Service**: Deployments | Database | Backups | Variables | Metrics | Settings

---

### 4a. Deployments Tab

#### Active Deployment Card
- **Badge**: "ACTIVE" in green pill/chip
- **Commit info**: user avatar (circular, ~32px) with small source icon overlay → commit message text → "14 hours ago via GitHub" timestamp
- **Actions**: "View logs" button (ghost), three-dot menu (⋮)
- **Expandable section**: green checkmark + "Deployment successful" with a chevron to expand details

#### Metadata Bar (above deployments)
- Globe icon + public URL (clickable)
- Server icon + node info ("node@22.22.0")
- Location pin + region ("europe-west4-drams3a")
- Replica icon + "1 Replica"

#### History Section
- Collapsible "HISTORY" header with chevron
- Previous deployments listed as cards:
  - **Badge**: "REMOVED" in muted/gray pill
  - Same layout as active: avatar, commit message, timestamp, "View logs" button, three-dot menu

---

### 4b. Variables Tab

#### Header Row
- "Service Variables" title (left)
- Actions (right): "Shared Variable" (link icon), "Raw Editor" (code braces icon), "+ New Variable" (purple)

#### Smart Prompt
- Contextual hint: rocket icon + "Trying to connect a database? Add Variable" — links to variable reference helper

#### Empty State
- Card/box with centered text: "No Environment Variables"
- Subtitle: "Import all your variables using the Raw Editor" (link)

#### Platform Variables
- Collapsible: "> 8 variables added by Railway" — expandable to show auto-injected vars

---

### 4c. Metrics Tab

#### Time Range Selector
- Segmented control: 1h | 6h | 1d | 7d | 30d (default: 1h, green highlight on active)

#### Layout Toggle
- Two icons (right-aligned): list view / grid view

#### Metric Cards (2x2 grid)
Each card is a bordered panel with:
- **Title**: "CPU", "Memory", "Public Network Traffic", "Requests"
- **Legend**: colored dots — "Sum" (pink/purple), "Replicas" (outlined circle)
- **Chart**: line/area chart with time on X-axis, metric value on Y-axis
- **Y-axis labels**: "0.8 vCPU", "400 MB", "0 B" etc.
- **X-axis labels**: timestamps (e.g., "9:45 PM", "10:00 PM")
- Empty states: flat line at zero, or "No request metrics available" text

---

### 4d. Settings Tab

#### Search/Filter
- Full-width search input: "Filter Settings..." with keyboard shortcut hint (`/`)

#### Right-Side Section Nav
A sticky right-aligned vertical nav listing all sections:
- Source
- Networking
- Scale
- Build
- Deploy
- Config-as-code
- Danger

Clicking a nav item scrolls to that section. Each section has an icon + heading.

#### Source Section
- **Source Repo**: card with GitHub icon + repo name ("cblaettl/zeitlos-website"), edit (pencil) icon, "Disconnect" button
- **Root Directory**: "+ Add Root Directory" link
- **Branch**: branch icon + "main" dropdown + "Disconnect" button
- **Wait for CI**: toggle switch + description text

#### Scale Section
- **Regions & Replicas**: region dropdown (with flag emoji: EU West), replica count input
- **Replica Limits**: CPU slider (0-2 vCPU), Memory slider (0-1 GB), each showing plan limits
- Upgrade upsell link

#### Deploy Section
- **Cron Schedule**: "+ Add Schedule" button
- **Healthcheck Path**: "+ Healthcheck Path" button
- **Serverless**: toggle + description text + docs link
- **Restart Policy**: dropdown ("On Failure") with description
- **Retry count**: number input (default: 10) with plan limit notice

#### Config-as-code Section
- **Railway Config File**: "+ Add File Path" button with docs link

#### Danger Section
- Red-themed area
- "Delete Service" heading in red
- Warning text in red (bold "permanently delete", italic "this environment")
- Red "Delete service" button

---

## 5. Database Service Detail

Same panel pattern but with additional tabs:

### Database Tab (unique to DB services)
#### Sub-tabs: Data | Stats | Config
- **Connect button** (top-right, purple text with plug icon): for connection string access
- **Data sub-tab**: table browser — empty state: "You have no tables" + purple "Create table" button + "Read the docs" link
- Built-in database management (query/browse tables) directly in the UI

### Backups Tab
- Backup management (not explored in detail but present as a top-level tab)

---

## 6. Observability Page (`/project/:id/observability`)

### Full-page layout (no canvas split)
- **Time range selector** (top-left): "Last 1 hour" dropdown with clock icon
- **"+ Add block"** button (top-right)

### Empty State
- Illustration (4 mini chart placeholders in 2x2 grid)
- "Observe this environment" heading
- Description: "Monitor project usage, resource metrics and custom log dashboards"
- Two CTAs: "Add new item" (ghost) and "Start with a simple dashboard" (purple, primary)

### Design Pattern
- **Customizable dashboard**: users build their own observability views by adding blocks
- Canvas/grid-based layout for dashboard panels

---

## 7. Log Explorer (`/project/:id/logs`)

### Full-page layout
- **Search bar** (top, full-width): "Filter and search logs" with keyboard shortcut hint (`/`)
- **Time range** (top-right): "Last 15 min" dropdown with clock icon
- **Actions** (top-right): pause/play toggle icon, download icon

### Timeline Bar
- Horizontal bar showing log density/distribution over time
- Left timestamp ("10:19 PM") <-- bar --> right timestamp ("10:34 PM")
- Color-coded histogram bars

### Empty State
- Illustration (faded terminal/log lines icon)
- "No logs in this time range"
- "Logs will show up here as they are found"

---

## 8. Project Settings (`/project/:id/settings`)

### Full-page modal overlay with left sidebar navigation
- **Header**: "Project Settings" (large bold) + close (X) button

### Left Sidebar Nav (icon + label, vertical list)
1. General (gear icon)
2. Usage (bar chart icon)
3. Environments (layers icon)
4. Shared Variables (globe icon)
5. Webhooks (webhook icon)
6. Members (people icon)
7. Tokens (key icon)
8. Integrations (puzzle icon)
9. Danger (warning triangle icon)

### General Section
- **Project Info**: Name input, Description input (placeholder text), Project ID (read-only with copy button)
- **Update button** (purple)
- **Visibility**: "This project is currently PRIVATE" text + "Change visibility" button

### Environments Section
- List of environments as rows: name, "Updated X hours ago", user avatar, three-dot menu
- Active environment has a checkmark
- "+ New Environment" button (purple)
- **PR Environments** section: description text + "Enable PR Environments" button (ghost)

---

## 9. Create / New Service Flow

### Command Palette Pattern
Both "New Project" (from dashboard) and "Create" (from within project) use the same **command palette / spotlight search** pattern:

- **Centered modal overlay** with dark backdrop (blurs background)
- **Search input** at top: "What would you like to create?" placeholder
- **Options list** below — each option is a row with:
  - Left: branded icon (GitHub logo, DB icon, template icon, Docker whale, etc.)
  - Center: label text
  - Right: chevron (>) for drill-down sub-menus

#### New Project Options (from dashboard):
1. GitHub Repository >
2. Database >
3. Template >
4. Docker Image >
5. Function
6. Bucket
7. Empty Project

#### New Service Options (from within project):
1. GitHub Repository >
2. Database >
3. Template >
4. Docker Image >
5. Function
6. Bucket
7. Volume
8. Empty Service

(Note: "Volume" only appears in-project; "Empty Project" vs "Empty Service" naming differs)

### Drill-Down Sub-Menu (e.g., Database)
- Breadcrumb chip appears above the search ("Database")
- Same search input remains
- Options change to sub-items:
  - PostgreSQL (branded elephant icon)
  - Redis (branded icon)
  - MongoDB (branded icon)
  - MySQL (branded icon)

---

## 10. Design System Observations

### Interactive Patterns
| Pattern | Usage |
|---------|-------|
| **Command palette** | Create new project/service — centered modal with search + drill-down options |
| **Split-panel** | Service detail — canvas left, detail right, Escape to close |
| **In-panel navigation** | Drill-down within panel (service > deployment) with breadcrumb back |
| **Tabbed interface** | Service tabs (Deployments/Variables/Metrics/Settings), Metric sub-tabs |
| **Segmented control** | View toggle (Architecture/List), time range (1h/6h/1d/7d/30d) |
| **Dropdown** | Environment switcher, sort order, region selector |
| **Toggle switch** | Feature flags (Wait for CI, Serverless) |
| **Slider** | CPU/Memory limits |
| **Section nav (sticky)** | Settings right-rail navigation scrolling to sections |
| **Expandable/collapsible** | Deployment history, platform variables |
| **Dotted canvas** | Architecture view background, new project background |
| **Status indicators** | Green dot = online, badges for deployment state |

### Empty States
Railway uses **illustrated empty states** consistently:
- Small illustration/icon (stylized, not photo-realistic)
- Bold heading describing the state
- Muted description text
- One or two action buttons (primary + ghost/link secondary)

### Keyboard Shortcuts
- `/` focuses search/filter inputs (shown as hint badge in the input)
- `Escape` closes panels and modals

---

## 11. Key UX Patterns to Adopt for Lucity

### Must-Have Patterns
1. **Architecture canvas** (via vue-flow) as the primary project view — services as draggable cards with status, environment-scoped
2. **Split-panel detail** — click a service card > right panel slides in, canvas shrinks left
3. **In-panel navigation** — drill-down within the panel (service > deployment detail) with breadcrumb-based back navigation
4. **Command palette for creation** — searchable, drill-down categories
5. **Environment switcher** in breadcrumb nav — dropdown in the top bar, canvas reloads per environment
6. **Tab-based service detail** — Deployments, Variables, Metrics, Settings as horizontal tabs
7. **Status dots** — green/yellow/red dots next to service names everywhere
8. **Dark theme first** — the entire developer tool market expects dark mode default

### Nice-to-Have Patterns
1. **View toggle** (Architecture/List) on the projects dashboard
2. **Customizable observability dashboard** with block-based layout
3. **Log explorer** with timeline histogram, search, and time range
4. **Settings section nav** (sticky right-rail) for long settings pages
5. **Keyboard shortcuts** for search focus
6. **Illustrated empty states** with clear CTAs

### Lucity-Specific Adaptations
- Railway and Lucity share the **same model**: project > environments > services. Railway's canvas is environment-scoped — switching environments reloads the canvas with that environment's services and their environment-specific URLs/state.
- Railway's "Database" tab with built-in table browser is specific to their managed DB. Lucity will delegate to operator CRDs (CNPG) — the detail view should focus on **connection info, status, and backups** rather than table browsing.
- Railway's "Config-as-code" concept maps to Lucity's **ejectability** story — this could be a prominent feature in Settings.
- The GitOps repo is an **internal implementation detail** in Lucity. It should be abstracted away from the user. Only when the user chooses to eject does the GitOps structure become relevant — at that point, the ejection flow surfaces it.
