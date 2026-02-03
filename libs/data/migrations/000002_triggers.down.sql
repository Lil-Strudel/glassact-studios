-- GlassAct Studios - Drop Triggers

-- Invoice Line Items
DROP TRIGGER IF EXISTS update_invoice_line_items_updated_at ON invoice_line_items;
DROP TRIGGER IF EXISTS increment_invoice_line_items_version ON invoice_line_items;

-- Invoices
DROP TRIGGER IF EXISTS update_invoices_updated_at ON invoices;
DROP TRIGGER IF EXISTS increment_invoices_version ON invoices;

-- Project Chats
DROP TRIGGER IF EXISTS update_project_chats_updated_at ON project_chats;
DROP TRIGGER IF EXISTS increment_project_chats_version ON project_chats;

-- Inlay Blockers
DROP TRIGGER IF EXISTS update_inlay_blockers_updated_at ON inlay_blockers;
DROP TRIGGER IF EXISTS increment_inlay_blockers_version ON inlay_blockers;

-- Inlay Milestones
DROP TRIGGER IF EXISTS update_inlay_milestones_updated_at ON inlay_milestones;
DROP TRIGGER IF EXISTS increment_inlay_milestones_version ON inlay_milestones;

-- Inlay Proofs
DROP TRIGGER IF EXISTS update_inlay_proofs_updated_at ON inlay_proofs;
DROP TRIGGER IF EXISTS increment_inlay_proofs_version ON inlay_proofs;

-- Inlay Chats
DROP TRIGGER IF EXISTS update_inlay_chats_updated_at ON inlay_chats;
DROP TRIGGER IF EXISTS increment_inlay_chats_version ON inlay_chats;

-- Inlay Custom Infos
DROP TRIGGER IF EXISTS update_inlay_custom_infos_updated_at ON inlay_custom_infos;
DROP TRIGGER IF EXISTS increment_inlay_custom_infos_version ON inlay_custom_infos;

-- Inlay Catalog Infos
DROP TRIGGER IF EXISTS update_inlay_catalog_infos_updated_at ON inlay_catalog_infos;
DROP TRIGGER IF EXISTS increment_inlay_catalog_infos_version ON inlay_catalog_infos;

-- Inlays
DROP TRIGGER IF EXISTS update_inlays_updated_at ON inlays;
DROP TRIGGER IF EXISTS increment_inlays_version ON inlays;

-- Projects
DROP TRIGGER IF EXISTS update_projects_updated_at ON projects;
DROP TRIGGER IF EXISTS increment_projects_version ON projects;

-- Catalog Items
DROP TRIGGER IF EXISTS update_catalog_items_updated_at ON catalog_items;
DROP TRIGGER IF EXISTS increment_catalog_items_version ON catalog_items;

-- Internal Accounts
DROP TRIGGER IF EXISTS update_internal_accounts_updated_at ON internal_accounts;
DROP TRIGGER IF EXISTS increment_internal_accounts_version ON internal_accounts;

-- Internal Users
DROP TRIGGER IF EXISTS update_internal_users_updated_at ON internal_users;
DROP TRIGGER IF EXISTS increment_internal_users_version ON internal_users;

-- Dealership Accounts
DROP TRIGGER IF EXISTS update_dealership_accounts_updated_at ON dealership_accounts;
DROP TRIGGER IF EXISTS increment_dealership_accounts_version ON dealership_accounts;

-- Dealership Users
DROP TRIGGER IF EXISTS update_dealership_users_updated_at ON dealership_users;
DROP TRIGGER IF EXISTS increment_dealership_users_version ON dealership_users;

-- Dealerships
DROP TRIGGER IF EXISTS update_dealerships_updated_at ON dealerships;
DROP TRIGGER IF EXISTS increment_dealerships_version ON dealerships;

-- Price Groups
DROP TRIGGER IF EXISTS update_price_groups_updated_at ON price_groups;
DROP TRIGGER IF EXISTS increment_price_groups_version ON price_groups;

-- Functions
DROP FUNCTION IF EXISTS increment_version_column();
DROP FUNCTION IF EXISTS update_updated_at_column();
