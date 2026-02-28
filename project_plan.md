# CairnPost — Project Plan

> Forms → Contacts → CRM. Sort the gold from the gravel.

**Domain:** cairnpost.com
**Parent Company:** Firefly Software (Helena, MT)
**Tech Stack:** Go, htmx, Alpine.js, Tailwind CSS, PostgreSQL, Postmark (email)

---

## Vision

CairnPost is a simple, affordable tool that helps small businesses — especially owner-operators and service businesses in the Midwest — collect customer inquiries through forms and turn them into a usable contact list. It is the first product in a planned suite of small business tools (inspired by 37signals) that share a common contact backbone.

The core insight: existing form builders (Jotform, Typeform, Tally, Fillout) treat form submissions as data to dump into a spreadsheet. None of them treat a submission as the beginning of a customer relationship. CairnPost does. Every form submission is a potential contact. The business owner triages submissions — accepting real inquiries into their contact list and discarding spam — building a clean, validated customer list over time.

### Product Philosophy

- **For whom:** Owner-operators and small service businesses (1-10 people) in trades, services, and local retail.
- **Core belief:** Forms are conversations with customers, not data collection exercises. Every form submission is the beginning (or continuation) of a relationship.
- **What it is NOT:** Not a survey tool, not a marketing tool, not an enterprise tool.
- **Design approach:** Opinionated. One good default is better than ten mediocre options. Convention over configuration.

### Platform Roadmap

CairnPost is the wedge tool. The contact record is the backbone that future services hang off of:

1. **Forms + Contacts (v1)** — this plan
2. **Simple Invoicing** — send an invoice to a contact, payment history in their record
3. **Appointment Booking** — book a time with a contact, appointment in their record
4. **Messaging** — send a text or email to a contact, message in their record

Each new feature adds a new type of history entry on the contact detail page and a new action button. The data model is designed for this from day one.

---

## Competitive Landscape

| Tool | Free Tier | First Paid Tier | Strength | Gap |
|------|-----------|-----------------|----------|-----|
| Jotform | 5 forms, 100 submissions/mo | $34/mo | Massive template library, integrations | Dated UI, expensive, no CRM |
| Typeform | 10 submissions/mo | $25/mo | Beautiful conversational UX | Extremely restrictive free tier, no CRM |
| Tally | Unlimited forms & submissions | $29/mo (Pro) | Most generous free tier | No CRM, optimized for tech-savvy users |
| Fillout | Unlimited forms, 1,000 submissions/mo | $15/mo | Good design, Airtable/Notion integration | No CRM, many features gated |
| Cognito Forms | 100 entries/mo | $15-19/mo | Strong logic/calculations | Generic design, no CRM |
| Google Forms | Unlimited, free | N/A | Free, ubiquitous | Ugly, unprofessional, no CRM |

**CairnPost's angle:** None of these tools know who your customers are. They collect data. CairnPost builds relationships. The form is just the door. The CRM is the house.

### Pricing Strategy

**Free tier (genuinely useful):**

- 3 active forms
- Unlimited submissions
- Contact list with full history
- Hosted form pages
- Email notifications on submission
- CairnPost branding on forms

**Paid tier (~$5-8/month):**

- Unlimited forms
- Custom branding (logo, colors)
- Signature field
- Custom domain for hosted pages
- SMS notifications
- No CairnPost branding

The free tier covers "contact us," "request a quote," and "customer intake" — enough for most small businesses to run on indefinitely. Upgrade triggers are vanity (branding), professionalism (signatures, custom domain), or growth (more than 3 forms).

---

## Data Model

### Entity Relationship Overview

```
users ──┐
        ├── org_members ──── organizations
        │                        │
        │                   ┌────┴────┐
        │                   │         │
        │                 forms    contacts
        │                   │         │
        │                   └────┬────┘
        │                        │
        │                   submissions
        │
```

### Schema

```sql
-- Auth & Users
CREATE TABLE users (
    id              UUID PRIMARY KEY,
    email           TEXT UNIQUE NOT NULL,
    name            TEXT NOT NULL,
    password_hash   TEXT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Tenants
CREATE TABLE organizations (
    id              UUID PRIMARY KEY,
    name            TEXT NOT NULL,
    slug            TEXT UNIQUE NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- User ↔ Organization membership
CREATE TABLE org_members (
    id              UUID PRIMARY KEY,
    organization_id UUID NOT NULL REFERENCES organizations,
    user_id         UUID NOT NULL REFERENCES users,
    role            TEXT NOT NULL DEFAULT 'owner',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(organization_id, user_id)
);

-- Form templates
CREATE TABLE forms (
    id              UUID PRIMARY KEY,
    organization_id UUID NOT NULL REFERENCES organizations,
    title           TEXT NOT NULL,
    slug            TEXT NOT NULL,
    fields          JSONB NOT NULL DEFAULT '[]',
    status          TEXT NOT NULL DEFAULT 'active',
    settings        JSONB NOT NULL DEFAULT '{}',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(organization_id, slug)
);

-- Raw form submissions
CREATE TABLE submissions (
    id              UUID PRIMARY KEY,
    form_id         UUID NOT NULL REFERENCES forms,
    contact_id      UUID REFERENCES contacts,
    status          TEXT NOT NULL DEFAULT 'pending',
    data            JSONB NOT NULL,
    meta            JSONB NOT NULL DEFAULT '{}',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    reviewed_at     TIMESTAMPTZ
);

-- Validated contacts (the CRM seed)
CREATE TABLE contacts (
    id              UUID PRIMARY KEY,
    organization_id UUID NOT NULL REFERENCES organizations,
    email           TEXT,
    phone           TEXT,
    name            TEXT,
    tags            TEXT[] NOT NULL DEFAULT '{}',
    source_form_id  UUID REFERENCES forms,
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(organization_id, email),
    UNIQUE(organization_id, phone)
);

-- Indexes
CREATE INDEX idx_submissions_form_status ON submissions(form_id, status);
CREATE INDEX idx_contacts_org_email ON contacts(organization_id, email);
CREATE INDEX idx_contacts_org_phone ON contacts(organization_id, phone);
CREATE INDEX idx_submissions_contact ON submissions(contact_id);
```

### Design Decisions

- **submissions.data is JSONB** — each form has different fields, so values are stored as a JSON blob keyed by field ID. The field definitions in `forms.fields` provide the schema for interpreting this data.
- **contact_id on submissions is nullable** — null while pending, set when promoted. A contact can have many submissions across many forms.
- **Deduplication** — unique constraints on `(organization_id, email)` and `(organization_id, phone)`. On promotion, check for existing contact; link if found, create if not.
- **No hard deletes on submissions** — discarded submissions get `status = 'discarded'` and stay for 30 days, enabling undo.
- **tags on contacts** — column exists as `text[]`, unused in v1. Reserved for future use without requiring a migration.
- **Single user in v1** — users/org_members tables exist and all auth checks go through org_members, but the UI assumes one user per org. Multi-user support added later without restructuring.

### Field Definition Structure

The `forms.fields` JSONB column stores an ordered array of field definitions:

```json
[
    {"id": "f1", "type": "name",     "label": "Your Name",        "required": true},
    {"id": "f2", "type": "email",    "label": "Email",            "required": true},
    {"id": "f3", "type": "phone",    "label": "Phone",            "required": false},
    {"id": "f4", "type": "text",     "label": "Subject",          "required": false},
    {"id": "f5", "type": "textarea", "label": "How can we help?", "required": false},
    {"id": "f6", "type": "select",   "label": "Service Needed",   "required": true,
                 "options": ["Repair", "Install", "Quote"], "display": "dropdown"},
    {"id": "f7", "type": "checkbox", "label": "Agree to terms",   "required": true},
    {"id": "f8", "type": "date",     "label": "Preferred date",   "required": false}
]
```

**Field types (8 total):**

| Type | Purpose | Magic? |
|------|---------|--------|
| `name` | Contact name | Yes — maps to `contacts.name` |
| `email` | Contact email | Yes — maps to `contacts.email` |
| `phone` | Contact phone | Yes — maps to `contacts.phone` |
| `text` | Short text input | No |
| `textarea` | Long text input | No |
| `select` | Pick one from options | No |
| `checkbox` | Boolean toggle | No |
| `date` | Date picker | No |

**Magic fields** (`name`, `email`, `phone`) are used for contact creation and deduplication on promotion. A form can have at most one of each magic type, enforced in the form builder UI.

**Select display modes:** The `select` type accepts an optional `"display"` property — either `"dropdown"` (default) or `"radio"`. Same data, different rendering. Guideline: 2-4 options suit radio buttons, 5+ suit a dropdown.

### Form Settings Structure

The `forms.settings` JSONB column:

```json
{
    "notification_email": "owner@example.com",
    "redirect_url": null,
    "thank_you_message": "Thanks! We'll be in touch soon."
}
```

---

## Submission Promotion Flow

The core mechanic of CairnPost: submissions arrive as raw, unvalidated data. The owner triages them into contacts or discards them as spam.

```
Form Submission arrives
       │
       ▼
  ┌──────────┐
  │  pending  │  ← appears in inbox
  └──────────┘
       │
  ┌────┴────┐
  ▼         ▼
accept    discard
  │         │
  ▼         ▼
Contact   status = 'discarded'
created   (retained 30 days, undo available)
or linked
```

### Acceptance Logic (per submission)

1. Extract magic field values from `submission.data` (name, email, phone)
2. Check for existing contact by email (if present): `SELECT FROM contacts WHERE organization_id = ? AND email = ?`
3. If no match, check by phone (if present)
4. If existing contact found: link submission (`submission.contact_id = contact.id`)
5. If no existing contact: create new contact, link submission
6. Set `submission.status = 'accepted'` and `submission.reviewed_at = NOW()`

This runs per-submission inside a transaction, but the API accepts arrays of IDs for bulk operations.

---

## UI Design

### Submission Inbox (Triage Queue)

This is NOT a communication inbox. It is a read-only triage queue. The purpose is to present each submission and allow the owner to rapidly classify it as a contact or spam. Speed is everything.

**Default view: pending submissions only.**

**Card layout — each submission is a card with three lines max:**

- **Line 1:** Identity — name, email, phone (the magic fields)
- **Line 2:** Context — which form, how long ago
- **Line 3:** Preview — truncated first non-identity field value

```
┌──────────────────────────────────────────┐
│ ☐  Jane Smith · jane@example.com         │
│    Request a Quote · 2 hours ago         │
│    "We need new flooring in our..."      │
├──────────────────────────────────────────┤
│ ☐  · asdf@spam.xyz                       │
│    Contact Us · 45 min ago               │
│    "Buy cheap viagra online..."          │
├──────────────────────────────────────────┤
│ ☐  Mike Torres · (406) 555-1234          │
│    Schedule Service · 3 hours ago        │
│    "Looking to get my furnace..."        │
└──────────────────────────────────────────┘

        [ ✓ Accept ]    [ ✗ Discard ]
```

**Interactions:**

- **Expand card** — click/tap a card to expand inline showing full submission data. Read-only.
- **Individual action** — accept or discard from the expanded card. Card slides out.
- **Bulk action** — checkboxes on each card, select-all toggle at top, action buttons at bottom. One POST, all selected cards clear.
- **Undo** — discarded submissions stay in DB. Brief "Undo" toast after discard. No complex undo UI.
- **Filtering** — status tabs (Pending / Accepted / Discarded) and a form dropdown. htmx swaps the card list on filter change.
- **No feedback on accept** — the card disappearing IS the feedback. Contact is in the contact list when the owner needs it.

**Mobile-first:**

- Full-width stacked cards
- Large touch targets (checkboxes, buttons)
- Sticky action bar at bottom of viewport
- 16px minimum font size

### Contact List

A simple searchable list of validated contacts.

```
┌──────────────────────────────────────────────────────┐
│  Contacts (47)                        [Search...]    │
├──────────────────────────────────────────────────────┤
│  Jane Smith     jane@example.com    (406) 555-8821   │
│  Request a Quote · 2 days ago                        │
│                                                      │
│  Mike Torres    mike.t@gmail.com    (406) 555-1234   │
│  Schedule Service · 3 days ago                       │
└──────────────────────────────────────────────────────┘
```

**Two lines per contact:**

- **Line 1:** Name, email, phone
- **Line 2:** Source form and when first created (answers "how do I know this person?")

**Search:** Single search box, searches across name, email, phone. Simple `ILIKE` query. No advanced filters.

**Sort:** Most recently created first. No sort options in v1.

**No pagination controls:** Load 50-100, "load more" at bottom if needed. Most small businesses will have a few hundred contacts at most.

**No bulk actions on contacts.** This is a reference tool, not a triage queue.

**No manual contact creation in v1.** Contacts only come from promoted submissions.

### Contact Detail Page

Full picture of one person's relationship with the business. Own page, not a side panel (simpler on mobile).

```
┌──────────────────────────────────────────────────────┐
│  ← Contacts                                         │
│                                                      │
│  Jane Smith                              [Edit]      │
│  jane@example.com · (406) 555-8821                   │
│                                                      │
│  Notes                                               │
│  ┌────────────────────────────────────────────────┐  │
│  │ Has two dogs, prefers mornings. Referred by    │  │
│  │ Mike Torres.                                   │  │
│  └────────────────────────────────────────────────┘  │
│                                                      │
│  History                                             │
│  ┌────────────────────────────────────────────────┐  │
│  │  Request a Quote · Feb 26, 2026                │  │
│  │  Service: Install                              │  │
│  │  "We need new flooring in our kitchen and      │  │
│  │   living room. About 800 sq ft..."             │  │
│  ├────────────────────────────────────────────────┤  │
│  │  Contact Us · Jan 14, 2026                     │  │
│  │  "Do you service the Helena area?"             │  │
│  └────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────┘
```

**Three sections:**

1. **Identity** — name, email, phone. Edit button for corrections/additions.
2. **Notes** — single free-text field. Editable inline, saves on blur. Not threaded, not a timeline. A sticky note.
3. **History** — reverse-chronological list of all form submissions from this contact. Read-only. This is where future features plug in: invoices, appointments, messages all become new history entry types.

### Form Builder

The form preview IS the builder. Single column, WYSIWYG-ish. No canvas, no grid.

**Field list — each field is an editable row:**

- Drag handle (≡) for reordering (SortableJS via Alpine)
- Label — editable inline
- Type badge — set at creation, not changeable after
- Required toggle (asterisk)
- Remove button (×)

**Adding a field:** "+ Add field" button expands an inline picker with 8 field types. Magic types (Name, Email, Phone) shown with distinct icons. If already present on form, grayed out.

**Editing a field:** Click to expand inline. Shows label, required toggle, and type-specific options (e.g., options list for select, display mode toggle for dropdown vs radio).

**Reordering:** Drag handles, SortableJS, htmx POST to persist new order.

**Form settings (behind a toggle):**

- Slug (auto-generated, editable)
- Status (active / paused)
- Notification email
- Redirect URL (optional)
- Thank you message

**Templates at creation:** When creating a new form, offer 3-4 starter templates (Blank, Contact Us, Request a Quote, Service Inquiry). Each is a pre-populated field list. Fully editable after selection.

**Left out of v1:** Multi-column layouts, conditional logic, file uploads, signature field, field-level help text, form duplication, custom CSS.

### Public Form Rendering

Server-rendered HTML. No SPA, no JavaScript framework. Fast on every connection.

**URL structure:** `forms.cairnpost.com/{org_slug}/{form_slug}`

**Layout:** Form centered on clean background, max-width ~540px. Organization name and form title at top. Fields stacked vertically. Submit button. "Powered by CairnPost" footer (free tier).

**Styling:** One good default. White/light background, readable font, good spacing, clear input borders. Mobile-friendly by default (single column). 16px minimum font size on inputs (prevents iOS zoom). No theme picker, no customization in v1.

**Validation:**

- Client-side: native HTML attributes (`required`, `type="email"`, `type="tel"`)
- Server-side: always validate. Re-render form with errors and preserved input on failure.

**Submission flow:** Standard PRG (Post/Redirect/Get). POST creates submission record, sends Postmark notification email, redirects to thank-you page.

**Embedding:** iframe embed snippet provided in the app. Direct link sharing (URL, QR code). No JavaScript embed SDK in v1.

**Spam prevention:**

- Honeypot field (hidden, bots fill it, humans don't)
- Time-based check (submitted < 3 seconds after load = bot)
- Rate limiting per IP (~10/hour)
- No CAPTCHA (hurts real users, inbox triage handles the rest)

---

## Email Notifications

Via Postmark. Sent inline with the submission POST handler.

**On new submission:**

```
Subject: New submission — {form_title}

You have a new submission on your "{form_title}" form.

Name: {name}
Email: {email}
Phone: {phone}

View it in your inbox:
https://cairnpost.com/app/submissions

— CairnPost
```

Shows identity fields in the email for quick gut-check from a phone. Call to action points to the inbox for triage.

---

## API Routes

### Public (no auth)

```
GET  /f/{org_slug}/{form_slug}          Render the public form page
POST /f/{org_slug}/{form_slug}          Submit the form (PRG → thank you page)
GET  /f/{org_slug}/{form_slug}/thanks   Thank you page after submission
```

### Auth

```
POST /auth/register       Create account + organization
POST /auth/login          Log in, set session cookie
POST /auth/logout         Clear session
```

### App (authenticated, org-scoped)

**Forms:**

```
GET    /app/forms                       List all forms
GET    /app/forms/new                   New form page (template picker)
POST   /app/forms                       Create a form
GET    /app/forms/{form_id}             Form builder/editor
PUT    /app/forms/{form_id}             Update form (title, settings, status)
DELETE /app/forms/{form_id}             Delete/archive a form
PUT    /app/forms/{form_id}/fields      Update full field list
GET    /app/forms/{form_id}/embed       Embed snippet / share link / QR code
```

**Submissions (inbox):**

```
GET    /app/submissions                 Inbox (default: pending, filterable)
GET    /app/submissions/{id}            Expanded submission detail (htmx partial)
POST   /app/submissions/accept          Bulk accept — body: {"ids": [...]}
POST   /app/submissions/discard         Bulk discard — body: {"ids": [...]}
POST   /app/submissions/undo            Undo discard — body: {"ids": [...]}
```

Query params for filtering:

```
?status=pending|accepted|discarded
?form_id={uuid}
```

**Contacts (CRM):**

```
GET    /app/contacts                    Contact list (searchable)
GET    /app/contacts/{id}               Contact detail (with history)
PUT    /app/contacts/{id}               Update contact (name, email, phone, notes)
DELETE /app/contacts/{id}               Delete a contact
```

Query params:

```
?q={search_term}          Searches name, email, phone
```

**Dashboard:**

```
GET    /app                             Redirects to /app/submissions (for now)
```

**22 routes total.**

### htmx Partial Pattern

Routes serve full pages or htmx partials based on request headers:

```go
if r.Header.Get("HX-Request") == "true" {
    // render just the partial (card list, contact list, expanded card, etc.)
} else {
    // render full page with layout
}
```

---

## Implementation Phases

### Phase 1 — Foundation

- Project scaffolding (Go modules, directory structure, Tailwind, templ)
- Database setup and migrations
- User registration and session-based auth
- Organization creation on signup

### Phase 2 — Form Builder

- Form CRUD (create, list, edit, delete)
- Field definition management (add, remove, reorder, edit)
- Template picker on creation
- Form settings (slug, status, notifications, thank you message)

### Phase 3 — Public Forms

- Form rendering from field definitions
- Client-side validation (native HTML)
- Server-side validation
- Submission creation and PRG flow
- Thank you page (custom message or redirect)
- Spam prevention (honeypot, timing, rate limit)
- Email notification via Postmark

### Phase 4 — Submission Inbox

- Inbox view with card layout
- Card expand/collapse (htmx)
- Bulk accept and discard
- Undo discard
- Status filtering (pending/accepted/discarded)
- Form filtering
- Contact creation/linking on accept (deduplication logic)

### Phase 5 — Contacts

- Contact list with search
- Contact detail page
- Submission history on contact detail
- Notes field (inline edit)
- Contact editing (name, email, phone)

### Phase 6 — Polish & Launch

- Embed snippet generation (iframe + direct link)
- QR code generation for form URLs
- Mobile UX testing and refinement
- "Powered by CairnPost" branding on free tier forms
- Landing page at cairnpost.com
- Stripe integration for paid tier

---

## What's Deliberately Left Out of V1

- Tags/labels on contacts (column exists, no UI)
- Export (CSV, etc.)
- Multi-user / team access (tables exist, no UI)
- Conditional logic on forms
- File upload fields
- Signature fields
- Multi-page forms
- Custom form styling / themes
- Form analytics / conversion tracking
- Manual contact creation
- Bulk actions on contacts
- Sorting options on contact list
- Advanced search / filters
- Integrations (Zapier, webhooks, etc.)
- API for third-party access
- CAPTCHA
- JavaScript embed SDK
