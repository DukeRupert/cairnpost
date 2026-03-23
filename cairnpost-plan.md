# CairnPost — CRM Product Plan

> A lean, handcrafted CRM for small service businesses. Part of the Firefly Software product suite.

---

## Product Context

CairnPost is a focused CRM targeting small service businesses running a lead → quote → close pipeline — contractors, consultants, tradespeople, and similar relationship-driven businesses. It is designed to be the contact and deal layer shared across the Firefly product suite, with other products (Skalkaho, CairnPost forms) reading and writing to the same data via a shared `org_id`.

**Naming:** CairnPost (existing Firefly product name, expanded in scope from forms-to-CRM to full CRM platform)

**Stack:** Go + htmx + Alpine.js + Tailwind CSS (HTMX for most interactions, Svelte islands for high-interactivity views like the pipeline kanban)

**Infrastructure:** Self-hosted, Hetzner VPS, Docker Compose, PostgreSQL, Caddy

---

## Target Customer

- Primary: Helena/MT small businesses (Firefly's existing client base)
- Secondary: General small service business market (productized SaaS)
- Business motions: service/project-based sales, lead → quote → close pipeline, repeat customer relationships
- Team size: 1–10 people, no dedicated CRM admin

---

## Feature Tiers

### Core (ship in v1)

| Feature | Description |
|---|---|
| Contact management | Name, phone, email, tags, custom fields, linked company |
| Activity log | Timestamped notes, calls, emails, site visits per contact |
| Task & follow-up reminders | Due dates linked to contacts or deals |
| Deal / pipeline stages | User-configurable stages, value tracking |
| Email logging | BCC address to attach emails to contact records |
| Daily agenda digest | Postmark email: tasks due + pipeline summary. User configures frequency and exempt days. |
| Search & filter | Full-text and field-level search across contacts/deals |
| Contact grouping / tags | Segment by type, status, source, etc. |

### Useful (v1 or early v2)

| Feature | Description |
|---|---|
| Web lead capture form | Embeddable form that creates contact + deal records (CairnPost integration) |
| Calendar sync | Google / Outlook calendar integration for tasks/events |
| Company / account records | Group contacts under a business entity |
| Basic reporting | Pipeline value, open deals, overdue tasks |
| Team access / roles | Multi-user with admin/member permission levels |
| CSV import / export | Migrate in from spreadsheets, export to anything |

### Explicitly excluded (v1)

- Email sequences / marketing automation
- AI lead scoring or predictive analytics
- Built-in VoIP / phone calling
- Revenue forecasting
- Quote / invoice builder (Skalkaho owns this)
- Social media integration

---

## Data Model

### Entities

**`contact`**
```
id            uuid PK
org_id        uuid FK
first_name    text
last_name     text
email         text
phone         text
tags          text[]
company_id    uuid FK (nullable)
created_at    timestamptz
updated_at    timestamptz
```

**`company`**
```
id            uuid PK
org_id        uuid FK
name          text
address       text
website       text
notes         text
created_at    timestamptz
updated_at    timestamptz
```

**`deal`**
```
id            uuid PK
org_id        uuid FK
title         text
stage         text          -- plain text, not enum (user-configurable stages)
value         numeric
contact_id    uuid FK
company_id    uuid FK (nullable)
closed_at     timestamptz (nullable)
created_at    timestamptz
updated_at    timestamptz
```

**`activity`**
```
id            uuid PK
org_id        uuid FK
type          text          -- note | call | email | sms | site_visit
body          text
contact_id    uuid FK
deal_id       uuid FK (nullable)
user_id       uuid FK
occurred_at   timestamptz
created_at    timestamptz
```

**`task`**
```
id            uuid PK
org_id        uuid FK
title         text
due_date      date
done          bool
contact_id    uuid FK (nullable)
deal_id       uuid FK (nullable)
assigned_to   uuid FK
created_at    timestamptz
updated_at    timestamptz
```

**`user`**
```
id            uuid PK
org_id        uuid FK
name          text
email         text
role          text          -- admin | member
created_at    timestamptz
```

**`org`**
```
id            uuid PK
name          text
slug          text (unique)
created_at    timestamptz
```

### Relationships

- `org` → `user` (1:many)
- `org` → `contact`, `company`, `deal`, `activity`, `task` (1:many — row-level tenancy)
- `company` → `contact` (1:many)
- `contact` → `deal` (1:many)
- `contact` → `activity` (1:many)
- `contact` → `task` (1:many)
- `deal` → `activity` (1:many)
- `deal` → `task` (1:many)
- `user` → `activity` (logs, 1:many)
- `user` → `task` (assigned, 1:many)

### Design notes

- `activity.type` is a text enum stored as string — extensible without migration
- `deal.stage` is plain text, not a DB enum — stages are user-configurable per org
- `activity.deal_id` and `task.deal_id` are nullable — contact-only interactions are valid without a deal context
- `org_id` on every table is the multi-tenancy boundary — enables cross-product integration via shared `org_id`

---

## User Workflow

### Daily rhythm (primary loop)

```
Daily digest email (morning)
  → Today view: overdue + due today tasks
      → New lead (web form or manual)    → Contact record → Deal stage
      → Log interaction (call/email/note) → Activity log  → Deal stage
  → Deal / pipeline stage
      → Task / follow-up (due date)
      → Won / closed → becomes customer record
          → repeat customer loop back to Log interaction
```

### Integration points

- **CairnPost forms** → submits to `/api/leads` → creates `contact` + `deal` with `org_id`
- **Skalkaho** (quoting) → reads `contact`/`company` by `org_id`, writes back a deal stage update when quote is sent/accepted
- **Future products** → same `org_id` pattern, read/write contacts and activities via internal API

---

## UX Principles

1. **Today-first navigation** — the default view answers "what do I need to do right now", not "show me all my data"
2. **Capture speed** — logging a call or note must be ≤3 taps/clicks from anywhere in the app
3. **One screen per object** — contacts, deals, and companies each have a single detail page with a unified timeline; no modal-within-modal
4. **No required fields except name** — friction kills adoption; every field beyond a name is optional
5. **Pipeline is visual but not primary** — kanban view exists, but the list/table view is the workhorse for most users

## Navigation / Sidebar Behavior

- **Narrow (< ~900px):** icon only, 52px wide
- **Hover (any width):** tooltip label appears on `mouseenter`, 200ms delay
- **Wide (≥ ~900px):** expands to icon + label, ~160px wide — triggered by CSS container query or JS resize observer
- Implementation: CSS `width` transition on sidebar; Alpine.js `x-data` for hover/expanded state; labels use `opacity` + `translateX` so they don't affect layout mid-transition
- Active state: filled icon + brand green (`#2D4A3E`) background pill, muted green label text

---

## Pages (planned)

| Route | Description |
|---|---|
| `/` | Today view — tasks due, overdue, pipeline summary |
| `/contacts` | Contact list with search and filter |
| `/contacts/:id` | Contact detail — info, timeline, tasks, linked deals |
| `/companies` | Company list |
| `/companies/:id` | Company detail — contacts, deals, timeline |
| `/deals` | Deal list (table view, filterable by stage) |
| `/deals/pipeline` | Kanban pipeline view (Svelte island) |
| `/deals/:id` | Deal detail — contact, stage, activities, tasks |
| `/tasks` | All tasks — filterable by due date, assignee, contact |
| `/settings` | Pipeline stages, custom fields, team members, integrations |

---

## Decisions

| Question | Decision |
|---|---|
| Pipeline stages configurable? | Ship with sensible defaults (`New Lead`, `Estimate Sent`, `Follow-up`, `Won`, `Lost`). User-configurable stage management deferred to v2. |
| Daily digest delivery | Postmark. User configures frequency (daily/weekly) and exempt days (e.g. no Sunday sends). |
| Email logging (BCC) | Deferred. v1 uses manual activity log entry. BCC magic address or forward integration is a v2 feature. |
| Multi-org / SaaS | Single-org first — Firefly internal use. `org` table and `org_id` FK on all entities present from day one to make the SaaS upgrade non-destructive. No org switcher or signup flow in v1. |
| Mobile | Responsive web first. PWA wrapper deferred.

## Open Questions

- [ ] Mobile: responsive web first, or plan for a PWA wrapper early?
