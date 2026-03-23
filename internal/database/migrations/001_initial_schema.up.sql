-- CairnPost initial schema
-- Multi-tenant CRM: org → users, contacts, companies, deals, activities, tasks

BEGIN;

-- Extensions
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Organizations (tenancy boundary)
CREATE TABLE orgs (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name       TEXT NOT NULL,
    slug       TEXT UNIQUE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Users
CREATE TABLE users (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id     UUID NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    name       TEXT NOT NULL,
    email      TEXT NOT NULL,
    role       TEXT NOT NULL DEFAULT 'member', -- admin | member
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(org_id, email)
);

-- Companies
CREATE TABLE companies (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id     UUID NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    name       TEXT NOT NULL,
    address    TEXT,
    website    TEXT,
    notes      TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Contacts
CREATE TABLE contacts (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id     UUID NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    first_name TEXT NOT NULL,
    last_name  TEXT,
    email      TEXT,
    phone      TEXT,
    tags       TEXT[] NOT NULL DEFAULT '{}',
    company_id UUID REFERENCES companies(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Deals
CREATE TABLE deals (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id     UUID NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    title      TEXT NOT NULL,
    stage      TEXT NOT NULL DEFAULT 'New Lead', -- plain text, user-configurable
    value      NUMERIC,
    contact_id UUID NOT NULL REFERENCES contacts(id) ON DELETE CASCADE,
    company_id UUID REFERENCES companies(id) ON DELETE SET NULL,
    closed_at  TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Activities
CREATE TABLE activities (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id      UUID NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    type        TEXT NOT NULL, -- note | call | email | sms | site_visit
    body        TEXT NOT NULL,
    contact_id  UUID NOT NULL REFERENCES contacts(id) ON DELETE CASCADE,
    deal_id     UUID REFERENCES deals(id) ON DELETE SET NULL,
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Tasks
CREATE TABLE tasks (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id      UUID NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    title       TEXT NOT NULL,
    due_date    DATE,
    done        BOOLEAN NOT NULL DEFAULT FALSE,
    contact_id  UUID REFERENCES contacts(id) ON DELETE SET NULL,
    deal_id     UUID REFERENCES deals(id) ON DELETE SET NULL,
    assigned_to UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_users_org       ON users(org_id);
CREATE INDEX idx_contacts_org    ON contacts(org_id);
CREATE INDEX idx_companies_org   ON companies(org_id);
CREATE INDEX idx_deals_org       ON deals(org_id);
CREATE INDEX idx_deals_contact   ON deals(contact_id);
CREATE INDEX idx_deals_stage     ON deals(org_id, stage);
CREATE INDEX idx_activities_org  ON activities(org_id);
CREATE INDEX idx_activities_contact ON activities(contact_id);
CREATE INDEX idx_activities_deal ON activities(deal_id);
CREATE INDEX idx_tasks_org       ON tasks(org_id);
CREATE INDEX idx_tasks_assigned  ON tasks(assigned_to);
CREATE INDEX idx_tasks_due       ON tasks(org_id, due_date) WHERE NOT done;
CREATE INDEX idx_tasks_contact   ON tasks(contact_id);
CREATE INDEX idx_tasks_deal      ON tasks(deal_id);

COMMIT;
