-- GlassAct Studios - Complete Database Schema
-- This migration creates all tables for the ecommerce platform

--------------------------------------------------------------------------------
-- EXTENSIONS
--------------------------------------------------------------------------------

CREATE EXTENSION IF NOT EXISTS citext;
CREATE EXTENSION IF NOT EXISTS postgis;

--------------------------------------------------------------------------------
-- ENUMS
--------------------------------------------------------------------------------

CREATE TYPE notification_event_type AS ENUM (
    'proof_ready',
    'proof_approved',
    'proof_declined',
    'order_placed',
    'inlay_step_changed',
    'inlay_blocked',
    'inlay_unblocked',
    'project_shipped',
    'project_delivered',
    'invoice_sent',
    'payment_received',
    'chat_message'
);

--------------------------------------------------------------------------------
-- REFERENCE TABLES
--------------------------------------------------------------------------------

CREATE TABLE price_groups (
    id SERIAL PRIMARY KEY,
    uuid UUID DEFAULT gen_random_uuid() UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    base_price_cents INTEGER NOT NULL,
    description TEXT,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    version INTEGER NOT NULL DEFAULT 1
);

--------------------------------------------------------------------------------
-- DEALERSHIPS & DEALERSHIP USERS
--------------------------------------------------------------------------------

CREATE TABLE dealerships (
    id SERIAL PRIMARY KEY,
    uuid UUID DEFAULT gen_random_uuid() UNIQUE NOT NULL,
    name TEXT NOT NULL,
    street TEXT NOT NULL,
    street_ext TEXT NOT NULL DEFAULT '',
    city TEXT NOT NULL,
    state TEXT NOT NULL,
    postal_code TEXT NOT NULL,
    country TEXT NOT NULL,
    location GEOGRAPHY(Point, 4326) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    version INTEGER NOT NULL DEFAULT 1
);

CREATE TABLE dealership_users (
    id SERIAL PRIMARY KEY,
    uuid UUID DEFAULT gen_random_uuid() UNIQUE NOT NULL,
    dealership_id INTEGER NOT NULL REFERENCES dealerships ON DELETE RESTRICT,
    name TEXT NOT NULL,
    email CITEXT UNIQUE NOT NULL,
    avatar TEXT NOT NULL DEFAULT '',
    role VARCHAR(255) NOT NULL CHECK (role IN ('viewer', 'submitter', 'approver', 'admin')),
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    version INTEGER NOT NULL DEFAULT 1
);

CREATE INDEX idx_dealership_users_dealership ON dealership_users(dealership_id);

CREATE TABLE dealership_accounts (
    id SERIAL PRIMARY KEY,
    uuid UUID DEFAULT gen_random_uuid() UNIQUE NOT NULL,
    dealership_user_id INTEGER NOT NULL REFERENCES dealership_users ON DELETE CASCADE,
    type VARCHAR(255) NOT NULL,
    provider VARCHAR(255) NOT NULL,
    provider_account_id VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    version INTEGER NOT NULL DEFAULT 1
);

CREATE INDEX idx_dealership_accounts_user ON dealership_accounts(dealership_user_id);

CREATE TABLE dealership_tokens (
    hash BYTEA PRIMARY KEY,
    dealership_user_id INTEGER NOT NULL REFERENCES dealership_users ON DELETE CASCADE,
    expiry TIMESTAMPTZ NOT NULL,
    scope TEXT NOT NULL
);

CREATE INDEX idx_dealership_tokens_user ON dealership_tokens(dealership_user_id);

--------------------------------------------------------------------------------
-- INTERNAL USERS (GLASSACT STAFF)
--------------------------------------------------------------------------------

CREATE TABLE internal_users (
    id SERIAL PRIMARY KEY,
    uuid UUID DEFAULT gen_random_uuid() UNIQUE NOT NULL,
    name TEXT NOT NULL,
    email CITEXT UNIQUE NOT NULL,
    avatar TEXT NOT NULL DEFAULT '',
    role VARCHAR(255) NOT NULL CHECK (role IN ('designer', 'production', 'billing', 'admin')),
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    version INTEGER NOT NULL DEFAULT 1
);

CREATE TABLE internal_accounts (
    id SERIAL PRIMARY KEY,
    uuid UUID DEFAULT gen_random_uuid() UNIQUE NOT NULL,
    internal_user_id INTEGER NOT NULL REFERENCES internal_users ON DELETE CASCADE,
    type VARCHAR(255) NOT NULL,
    provider VARCHAR(255) NOT NULL,
    provider_account_id VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    version INTEGER NOT NULL DEFAULT 1
);

CREATE INDEX idx_internal_accounts_user ON internal_accounts(internal_user_id);

CREATE TABLE internal_tokens (
    hash BYTEA PRIMARY KEY,
    internal_user_id INTEGER NOT NULL REFERENCES internal_users ON DELETE CASCADE,
    expiry TIMESTAMPTZ NOT NULL,
    scope TEXT NOT NULL
);

CREATE INDEX idx_internal_tokens_user ON internal_tokens(internal_user_id);

--------------------------------------------------------------------------------
-- CATALOG
--------------------------------------------------------------------------------

CREATE TABLE catalog_items (
    id SERIAL PRIMARY KEY,
    uuid UUID DEFAULT gen_random_uuid() UNIQUE NOT NULL,
    catalog_code VARCHAR(255) UNIQUE NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    category VARCHAR(255) NOT NULL,
    default_width DOUBLE PRECISION NOT NULL,
    default_height DOUBLE PRECISION NOT NULL,
    min_width DOUBLE PRECISION NOT NULL,
    min_height DOUBLE PRECISION NOT NULL,
    default_price_group_id INTEGER NOT NULL REFERENCES price_groups ON DELETE RESTRICT,
    svg_url TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    version INTEGER NOT NULL DEFAULT 1
);

CREATE INDEX idx_catalog_items_category ON catalog_items(category);
CREATE INDEX idx_catalog_items_active ON catalog_items(is_active) WHERE is_active = true;

CREATE TABLE catalog_item_tags (
    id SERIAL PRIMARY KEY,
    catalog_item_id INTEGER NOT NULL REFERENCES catalog_items ON DELETE CASCADE,
    tag VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(catalog_item_id, tag)
);

CREATE INDEX idx_catalog_item_tags_tag ON catalog_item_tags(tag);

--------------------------------------------------------------------------------
-- PROJECTS
--------------------------------------------------------------------------------

CREATE TABLE projects (
    id SERIAL PRIMARY KEY,
    uuid UUID DEFAULT gen_random_uuid() UNIQUE NOT NULL,
    dealership_id INTEGER NOT NULL REFERENCES dealerships ON DELETE RESTRICT,
    name TEXT NOT NULL,
    status VARCHAR(255) NOT NULL DEFAULT 'draft' CHECK (status IN (
        'draft',
        'designing',
        'pending-approval',
        'approved',
        'ordered',
        'in-production',
        'shipped',
        'delivered',
        'invoiced',
        'completed',
        'cancelled'
    )),
    ordered_at TIMESTAMPTZ,
    ordered_by INTEGER REFERENCES dealership_users,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    version INTEGER NOT NULL DEFAULT 1
);

CREATE INDEX idx_projects_dealership ON projects(dealership_id);
CREATE INDEX idx_projects_status ON projects(status);

--------------------------------------------------------------------------------
-- INLAYS
--------------------------------------------------------------------------------

CREATE TABLE inlays (
    id SERIAL PRIMARY KEY,
    uuid UUID DEFAULT gen_random_uuid() UNIQUE NOT NULL,
    project_id INTEGER NOT NULL REFERENCES projects ON DELETE RESTRICT,
    name TEXT NOT NULL,
    type VARCHAR(255) NOT NULL CHECK (type IN ('catalog', 'custom')),
    preview_url TEXT NOT NULL DEFAULT '',
    approved_proof_id INTEGER,
    manufacturing_step VARCHAR(255) CHECK (manufacturing_step IS NULL OR manufacturing_step IN (
        'ordered', 'materials-prep', 'cutting', 'fire-polish', 'packaging', 'shipped', 'delivered'
    )),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    version INTEGER NOT NULL DEFAULT 1
);

CREATE INDEX idx_inlays_project ON inlays(project_id);
CREATE INDEX idx_inlays_manufacturing_step ON inlays(manufacturing_step) WHERE manufacturing_step IS NOT NULL;

CREATE TABLE inlay_catalog_infos (
    id SERIAL PRIMARY KEY,
    uuid UUID DEFAULT gen_random_uuid() UNIQUE NOT NULL,
    inlay_id INTEGER NOT NULL UNIQUE REFERENCES inlays ON DELETE CASCADE,
    catalog_item_id INTEGER NOT NULL REFERENCES catalog_items ON DELETE RESTRICT,
    customization_notes TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    version INTEGER NOT NULL DEFAULT 1
);

CREATE TABLE inlay_custom_infos (
    id SERIAL PRIMARY KEY,
    uuid UUID DEFAULT gen_random_uuid() UNIQUE NOT NULL,
    inlay_id INTEGER NOT NULL UNIQUE REFERENCES inlays ON DELETE CASCADE,
    description TEXT NOT NULL,
    requested_width DOUBLE PRECISION NOT NULL,
    requested_height DOUBLE PRECISION NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    version INTEGER NOT NULL DEFAULT 1
);

CREATE TABLE inlay_custom_reference_images (
    id SERIAL PRIMARY KEY,
    uuid UUID DEFAULT gen_random_uuid() UNIQUE NOT NULL,
    inlay_custom_info_id INTEGER NOT NULL REFERENCES inlay_custom_infos ON DELETE CASCADE,
    image_url TEXT NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_inlay_custom_reference_images_info ON inlay_custom_reference_images(inlay_custom_info_id);

--------------------------------------------------------------------------------
-- INLAY CHATS (Design-phase conversation per inlay)
--------------------------------------------------------------------------------

CREATE TABLE inlay_chats (
    id SERIAL PRIMARY KEY,
    uuid UUID DEFAULT gen_random_uuid() UNIQUE NOT NULL,
    inlay_id INTEGER NOT NULL REFERENCES inlays ON DELETE CASCADE,
    dealership_user_id INTEGER REFERENCES dealership_users ON DELETE SET NULL,
    internal_user_id INTEGER REFERENCES internal_users ON DELETE SET NULL,
    message_type VARCHAR(255) NOT NULL DEFAULT 'text' CHECK (message_type IN (
        'text', 'image', 'proof_sent', 'proof_approved', 'proof_declined', 'system'
    )),
    message TEXT NOT NULL,
    attachment_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    version INTEGER NOT NULL DEFAULT 1,
    CONSTRAINT inlay_chats_sender_check CHECK (
        (dealership_user_id IS NOT NULL AND internal_user_id IS NULL) OR
        (dealership_user_id IS NULL AND internal_user_id IS NOT NULL) OR
        (dealership_user_id IS NULL AND internal_user_id IS NULL AND message_type IN ('system', 'proof_sent', 'proof_approved', 'proof_declined'))
    )
);

CREATE INDEX idx_inlay_chats_inlay ON inlay_chats(inlay_id);
CREATE INDEX idx_inlay_chats_created ON inlay_chats(inlay_id, created_at);

--------------------------------------------------------------------------------
-- INLAY PROOFS (Versioned designs with approval tracking)
--------------------------------------------------------------------------------

CREATE TABLE inlay_proofs (
    id SERIAL PRIMARY KEY,
    uuid UUID DEFAULT gen_random_uuid() UNIQUE NOT NULL,
    inlay_id INTEGER NOT NULL REFERENCES inlays ON DELETE RESTRICT,
    version_number INTEGER NOT NULL,
    design_asset_url TEXT NOT NULL,
    width DOUBLE PRECISION NOT NULL,
    height DOUBLE PRECISION NOT NULL,
    price_group_id INTEGER REFERENCES price_groups,
    price_cents INTEGER,
    scale_factor DOUBLE PRECISION NOT NULL DEFAULT 1.0,
    color_overrides JSONB NOT NULL DEFAULT '{}',
    status VARCHAR(255) NOT NULL DEFAULT 'pending' CHECK (status IN (
        'pending', 'approved', 'declined', 'superseded'
    )),
    approved_at TIMESTAMPTZ,
    approved_by INTEGER REFERENCES dealership_users,
    declined_at TIMESTAMPTZ,
    declined_by INTEGER REFERENCES dealership_users,
    decline_reason TEXT,
    sent_in_chat_id INTEGER REFERENCES inlay_chats NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    version INTEGER NOT NULL DEFAULT 1,
    UNIQUE(inlay_id, version_number)
);

CREATE INDEX idx_inlay_proofs_inlay ON inlay_proofs(inlay_id);
CREATE INDEX idx_inlay_proofs_status ON inlay_proofs(status);

ALTER TABLE inlays ADD CONSTRAINT inlays_approved_proof_fk 
    FOREIGN KEY (approved_proof_id) REFERENCES inlay_proofs ON DELETE SET NULL;

--------------------------------------------------------------------------------
-- MANUFACTURING WORKFLOW
--------------------------------------------------------------------------------

CREATE TABLE inlay_milestones (
    id SERIAL PRIMARY KEY,
    uuid UUID DEFAULT gen_random_uuid() UNIQUE NOT NULL,
    inlay_id INTEGER NOT NULL REFERENCES inlays ON DELETE RESTRICT,
    step VARCHAR(255) NOT NULL CHECK (step IN (
        'ordered', 'materials-prep', 'cutting', 'fire-polish', 'packaging', 'shipped', 'delivered'
    )),
    event_type VARCHAR(255) NOT NULL CHECK (event_type IN ('entered', 'exited', 'reverted')),
    performed_by INTEGER REFERENCES internal_users NOT NULL,
    event_time TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    version INTEGER NOT NULL DEFAULT 1
);

CREATE INDEX idx_inlay_milestones_inlay ON inlay_milestones(inlay_id);
CREATE INDEX idx_inlay_milestones_step ON inlay_milestones(step);
CREATE INDEX idx_inlay_milestones_event_time ON inlay_milestones(inlay_id, event_time);

CREATE TABLE inlay_blockers (
    id SERIAL PRIMARY KEY,
    uuid UUID DEFAULT gen_random_uuid() UNIQUE NOT NULL,
    inlay_id INTEGER NOT NULL REFERENCES inlays ON DELETE RESTRICT,
    blocker_type VARCHAR(255) NOT NULL CHECK (blocker_type IN ('soft', 'hard')),
    reason TEXT NOT NULL,
    step_blocked VARCHAR(255) NOT NULL,
    created_by INTEGER REFERENCES internal_users,
    resolved_at TIMESTAMPTZ,
    resolved_by INTEGER REFERENCES internal_users,
    resolution_notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    version INTEGER NOT NULL DEFAULT 1
);

CREATE INDEX idx_inlay_blockers_inlay ON inlay_blockers(inlay_id);
CREATE INDEX idx_inlay_blockers_active ON inlay_blockers(inlay_id) WHERE resolved_at IS NULL;

--------------------------------------------------------------------------------
-- PROJECT CHATS (Manufacturing-phase discussion)
--------------------------------------------------------------------------------

CREATE TABLE project_chats (
    id SERIAL PRIMARY KEY,
    uuid UUID DEFAULT gen_random_uuid() UNIQUE NOT NULL,
    project_id INTEGER NOT NULL REFERENCES projects ON DELETE RESTRICT,
    dealership_user_id INTEGER REFERENCES dealership_users ON DELETE SET NULL,
    internal_user_id INTEGER REFERENCES internal_users ON DELETE SET NULL,
    message_type VARCHAR(255) NOT NULL DEFAULT 'text' CHECK (message_type IN ('text', 'image', 'system')),
    message TEXT NOT NULL,
    attachment_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    version INTEGER NOT NULL DEFAULT 1,
    CONSTRAINT project_chats_sender_check CHECK (
        (dealership_user_id IS NOT NULL AND internal_user_id IS NULL) OR
        (dealership_user_id IS NULL AND internal_user_id IS NOT NULL) OR
        (dealership_user_id IS NULL AND internal_user_id IS NULL AND message_type = 'system')
    )
);

CREATE INDEX idx_project_chats_project ON project_chats(project_id);
CREATE INDEX idx_project_chats_created ON project_chats(project_id, created_at);

--------------------------------------------------------------------------------
-- ORDER SNAPSHOTS (Immutable pricing at order time)
--------------------------------------------------------------------------------

CREATE TABLE order_snapshots (
    id SERIAL PRIMARY KEY,
    uuid UUID DEFAULT gen_random_uuid() UNIQUE NOT NULL,
    project_id INTEGER NOT NULL REFERENCES projects ON DELETE RESTRICT,
    inlay_id INTEGER NOT NULL UNIQUE REFERENCES inlays ON DELETE RESTRICT,
    proof_id INTEGER NOT NULL REFERENCES inlay_proofs,
    price_group_id INTEGER NOT NULL,
    price_cents INTEGER NOT NULL,
    width DOUBLE PRECISION NOT NULL,
    height DOUBLE PRECISION NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_order_snapshots_project ON order_snapshots(project_id);

--------------------------------------------------------------------------------
-- INVOICING
--------------------------------------------------------------------------------

CREATE TABLE invoices (
    id SERIAL PRIMARY KEY,
    uuid UUID DEFAULT gen_random_uuid() UNIQUE NOT NULL,
    project_id INTEGER NOT NULL UNIQUE REFERENCES projects ON DELETE RESTRICT,
    invoice_number VARCHAR(255) UNIQUE NOT NULL,
    subtotal_cents INTEGER NOT NULL,
    tax_cents INTEGER NOT NULL DEFAULT 0,
    total_cents INTEGER NOT NULL,
    status VARCHAR(255) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'sent', 'paid', 'void')),
    sent_at TIMESTAMPTZ,
    sent_to_email TEXT,
    paid_at TIMESTAMPTZ,
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    version INTEGER NOT NULL DEFAULT 1
);

CREATE INDEX idx_invoices_status ON invoices(status);

CREATE TABLE invoice_line_items (
    id SERIAL PRIMARY KEY,
    uuid UUID DEFAULT gen_random_uuid() UNIQUE NOT NULL,
    invoice_id INTEGER NOT NULL REFERENCES invoices ON DELETE CASCADE,
    inlay_id INTEGER REFERENCES inlays,
    description TEXT NOT NULL,
    quantity INTEGER NOT NULL DEFAULT 1,
    unit_price_cents INTEGER NOT NULL,
    total_cents INTEGER NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    version INTEGER NOT NULL DEFAULT 1
);

CREATE INDEX idx_invoice_line_items_invoice ON invoice_line_items(invoice_id);

--------------------------------------------------------------------------------
-- NOTIFICATIONS
--------------------------------------------------------------------------------

CREATE TABLE notifications (
    id SERIAL PRIMARY KEY,
    uuid UUID DEFAULT gen_random_uuid() UNIQUE NOT NULL,
    dealership_user_id INTEGER REFERENCES dealership_users ON DELETE CASCADE,
    internal_user_id INTEGER REFERENCES internal_users ON DELETE CASCADE,
    event_type notification_event_type NOT NULL,
    title TEXT NOT NULL,
    body TEXT NOT NULL,
    project_id INTEGER REFERENCES projects ON DELETE SET NULL,
    inlay_id INTEGER REFERENCES inlays ON DELETE SET NULL,
    read_at TIMESTAMPTZ,
    email_sent_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT notifications_recipient_check CHECK (
        (dealership_user_id IS NOT NULL AND internal_user_id IS NULL) OR
        (dealership_user_id IS NULL AND internal_user_id IS NOT NULL)
    )
);

CREATE INDEX idx_notifications_dealership_user ON notifications(dealership_user_id) 
    WHERE dealership_user_id IS NOT NULL;
CREATE INDEX idx_notifications_internal_user ON notifications(internal_user_id) 
    WHERE internal_user_id IS NOT NULL;
CREATE INDEX idx_notifications_unread_dealership ON notifications(dealership_user_id, created_at) 
    WHERE dealership_user_id IS NOT NULL AND read_at IS NULL;
CREATE INDEX idx_notifications_unread_internal ON notifications(internal_user_id, created_at) 
    WHERE internal_user_id IS NOT NULL AND read_at IS NULL;

CREATE TABLE dealership_user_notification_prefs (
    id SERIAL PRIMARY KEY,
    dealership_user_id INTEGER NOT NULL REFERENCES dealership_users ON DELETE CASCADE,
    event_type notification_event_type NOT NULL,
    email_enabled BOOLEAN NOT NULL DEFAULT true,
    UNIQUE(dealership_user_id, event_type)
);

CREATE TABLE internal_user_notification_prefs (
    id SERIAL PRIMARY KEY,
    internal_user_id INTEGER NOT NULL REFERENCES internal_users ON DELETE CASCADE,
    event_type notification_event_type NOT NULL,
    email_enabled BOOLEAN NOT NULL DEFAULT true,
    UNIQUE(internal_user_id, event_type)
);
