-- GlassAct Studios - Triggers
-- Auto-update updated_at and increment version on all relevant tables

--------------------------------------------------------------------------------
-- TRIGGER FUNCTIONS
--------------------------------------------------------------------------------

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE OR REPLACE FUNCTION increment_version_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.version = OLD.version + 1;
    RETURN NEW;
END;
$$ language 'plpgsql';

--------------------------------------------------------------------------------
-- PRICE GROUPS
--------------------------------------------------------------------------------

CREATE TRIGGER update_price_groups_updated_at 
    BEFORE UPDATE ON price_groups 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER increment_price_groups_version 
    BEFORE UPDATE ON price_groups 
    FOR EACH ROW EXECUTE FUNCTION increment_version_column();

--------------------------------------------------------------------------------
-- DEALERSHIPS
--------------------------------------------------------------------------------

CREATE TRIGGER update_dealerships_updated_at 
    BEFORE UPDATE ON dealerships 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER increment_dealerships_version 
    BEFORE UPDATE ON dealerships 
    FOR EACH ROW EXECUTE FUNCTION increment_version_column();

--------------------------------------------------------------------------------
-- DEALERSHIP USERS
--------------------------------------------------------------------------------

CREATE TRIGGER update_dealership_users_updated_at 
    BEFORE UPDATE ON dealership_users 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER increment_dealership_users_version 
    BEFORE UPDATE ON dealership_users 
    FOR EACH ROW EXECUTE FUNCTION increment_version_column();

--------------------------------------------------------------------------------
-- DEALERSHIP ACCOUNTS
--------------------------------------------------------------------------------

CREATE TRIGGER update_dealership_accounts_updated_at 
    BEFORE UPDATE ON dealership_accounts 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER increment_dealership_accounts_version 
    BEFORE UPDATE ON dealership_accounts 
    FOR EACH ROW EXECUTE FUNCTION increment_version_column();

--------------------------------------------------------------------------------
-- INTERNAL USERS
--------------------------------------------------------------------------------

CREATE TRIGGER update_internal_users_updated_at 
    BEFORE UPDATE ON internal_users 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER increment_internal_users_version 
    BEFORE UPDATE ON internal_users 
    FOR EACH ROW EXECUTE FUNCTION increment_version_column();

--------------------------------------------------------------------------------
-- INTERNAL ACCOUNTS
--------------------------------------------------------------------------------

CREATE TRIGGER update_internal_accounts_updated_at 
    BEFORE UPDATE ON internal_accounts 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER increment_internal_accounts_version 
    BEFORE UPDATE ON internal_accounts 
    FOR EACH ROW EXECUTE FUNCTION increment_version_column();

--------------------------------------------------------------------------------
-- CATALOG ITEMS
--------------------------------------------------------------------------------

CREATE TRIGGER update_catalog_items_updated_at 
    BEFORE UPDATE ON catalog_items 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER increment_catalog_items_version 
    BEFORE UPDATE ON catalog_items 
    FOR EACH ROW EXECUTE FUNCTION increment_version_column();

--------------------------------------------------------------------------------
-- PROJECTS
--------------------------------------------------------------------------------

CREATE TRIGGER update_projects_updated_at 
    BEFORE UPDATE ON projects 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER increment_projects_version 
    BEFORE UPDATE ON projects 
    FOR EACH ROW EXECUTE FUNCTION increment_version_column();

--------------------------------------------------------------------------------
-- INLAYS
--------------------------------------------------------------------------------

CREATE TRIGGER update_inlays_updated_at 
    BEFORE UPDATE ON inlays 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER increment_inlays_version 
    BEFORE UPDATE ON inlays 
    FOR EACH ROW EXECUTE FUNCTION increment_version_column();

--------------------------------------------------------------------------------
-- INLAY CATALOG INFOS
--------------------------------------------------------------------------------

CREATE TRIGGER update_inlay_catalog_infos_updated_at 
    BEFORE UPDATE ON inlay_catalog_infos 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER increment_inlay_catalog_infos_version 
    BEFORE UPDATE ON inlay_catalog_infos 
    FOR EACH ROW EXECUTE FUNCTION increment_version_column();

--------------------------------------------------------------------------------
-- INLAY CUSTOM INFOS
--------------------------------------------------------------------------------

CREATE TRIGGER update_inlay_custom_infos_updated_at 
    BEFORE UPDATE ON inlay_custom_infos 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER increment_inlay_custom_infos_version 
    BEFORE UPDATE ON inlay_custom_infos 
    FOR EACH ROW EXECUTE FUNCTION increment_version_column();

--------------------------------------------------------------------------------
-- INLAY CHATS
--------------------------------------------------------------------------------

CREATE TRIGGER update_inlay_chats_updated_at 
    BEFORE UPDATE ON inlay_chats 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER increment_inlay_chats_version 
    BEFORE UPDATE ON inlay_chats 
    FOR EACH ROW EXECUTE FUNCTION increment_version_column();

--------------------------------------------------------------------------------
-- INLAY PROOFS
--------------------------------------------------------------------------------

CREATE TRIGGER update_inlay_proofs_updated_at 
    BEFORE UPDATE ON inlay_proofs 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER increment_inlay_proofs_version 
    BEFORE UPDATE ON inlay_proofs 
    FOR EACH ROW EXECUTE FUNCTION increment_version_column();

--------------------------------------------------------------------------------
-- INLAY MILESTONES
--------------------------------------------------------------------------------

CREATE TRIGGER update_inlay_milestones_updated_at 
    BEFORE UPDATE ON inlay_milestones 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER increment_inlay_milestones_version 
    BEFORE UPDATE ON inlay_milestones 
    FOR EACH ROW EXECUTE FUNCTION increment_version_column();

--------------------------------------------------------------------------------
-- INLAY BLOCKERS
--------------------------------------------------------------------------------

CREATE TRIGGER update_inlay_blockers_updated_at 
    BEFORE UPDATE ON inlay_blockers 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER increment_inlay_blockers_version 
    BEFORE UPDATE ON inlay_blockers 
    FOR EACH ROW EXECUTE FUNCTION increment_version_column();

--------------------------------------------------------------------------------
-- PROJECT CHATS
--------------------------------------------------------------------------------

CREATE TRIGGER update_project_chats_updated_at 
    BEFORE UPDATE ON project_chats 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER increment_project_chats_version 
    BEFORE UPDATE ON project_chats 
    FOR EACH ROW EXECUTE FUNCTION increment_version_column();

--------------------------------------------------------------------------------
-- INVOICES
--------------------------------------------------------------------------------

CREATE TRIGGER update_invoices_updated_at 
    BEFORE UPDATE ON invoices 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER increment_invoices_version 
    BEFORE UPDATE ON invoices 
    FOR EACH ROW EXECUTE FUNCTION increment_version_column();

--------------------------------------------------------------------------------
-- INVOICE LINE ITEMS
--------------------------------------------------------------------------------

CREATE TRIGGER update_invoice_line_items_updated_at 
    BEFORE UPDATE ON invoice_line_items 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER increment_invoice_line_items_version 
    BEFORE UPDATE ON invoice_line_items 
    FOR EACH ROW EXECUTE FUNCTION increment_version_column();
