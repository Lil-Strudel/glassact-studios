-- GlassAct Studios - Rollback Schema
-- Drop all tables in reverse dependency order

DROP TABLE IF EXISTS internal_user_notification_prefs;
DROP TABLE IF EXISTS dealership_user_notification_prefs;
DROP TABLE IF EXISTS notifications;

DROP TABLE IF EXISTS invoice_line_items;
DROP TABLE IF EXISTS invoices;

DROP TABLE IF EXISTS order_snapshots;

DROP TABLE IF EXISTS project_chats;

DROP TABLE IF EXISTS inlay_blockers;
DROP TABLE IF EXISTS inlay_milestones;

-- Remove FK before dropping inlay_proofs
ALTER TABLE IF EXISTS inlays DROP CONSTRAINT IF EXISTS inlays_approved_proof_fk;

DROP TABLE IF EXISTS inlay_proofs;
DROP TABLE IF EXISTS inlay_chats;

DROP TABLE IF EXISTS inlay_custom_reference_images;
DROP TABLE IF EXISTS inlay_custom_infos;
DROP TABLE IF EXISTS inlay_catalog_infos;
DROP TABLE IF EXISTS inlays;

DROP TABLE IF EXISTS projects;

DROP TABLE IF EXISTS catalog_item_images;
DROP TABLE IF EXISTS catalog_item_tags;
DROP TABLE IF EXISTS catalog_items;

DROP TABLE IF EXISTS internal_tokens;
DROP TABLE IF EXISTS internal_accounts;
DROP TABLE IF EXISTS internal_users;

DROP TABLE IF EXISTS dealership_tokens;
DROP TABLE IF EXISTS dealership_accounts;
DROP TABLE IF EXISTS dealership_users;
DROP TABLE IF EXISTS dealerships;

DROP TABLE IF EXISTS price_groups;

DROP TYPE IF EXISTS notification_event_type;

DROP EXTENSION IF EXISTS postgis CASCADE;
DROP EXTENSION IF EXISTS citext;
