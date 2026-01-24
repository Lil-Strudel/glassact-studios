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

CREATE TRIGGER update_dealerships_updated_at 
    BEFORE UPDATE ON dealerships 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER increment_dealerships_version 
    BEFORE UPDATE ON dealerships 
    FOR EACH ROW EXECUTE FUNCTION increment_version_column();

CREATE TRIGGER update_users_updated_at 
    BEFORE UPDATE ON users 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER increment_users_version 
    BEFORE UPDATE ON users 
    FOR EACH ROW EXECUTE FUNCTION increment_version_column();

CREATE TRIGGER update_accounts_updated_at 
    BEFORE UPDATE ON accounts 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER increment_accounts_version 
    BEFORE UPDATE ON accounts 
    FOR EACH ROW EXECUTE FUNCTION increment_version_column();

CREATE TRIGGER update_catalog_items_updated_at 
    BEFORE UPDATE ON catalog_items 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER increment_catalog_items_version 
    BEFORE UPDATE ON catalog_items 
    FOR EACH ROW EXECUTE FUNCTION increment_version_column();

CREATE TRIGGER update_projects_updated_at 
    BEFORE UPDATE ON projects 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER increment_projects_version 
    BEFORE UPDATE ON projects 
    FOR EACH ROW EXECUTE FUNCTION increment_version_column();

CREATE TRIGGER update_inlays_updated_at 
    BEFORE UPDATE ON inlays 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER increment_inlays_version 
    BEFORE UPDATE ON inlays 
    FOR EACH ROW EXECUTE FUNCTION increment_version_column();

CREATE TRIGGER update_inlay_catalog_infos_updated_at 
    BEFORE UPDATE ON inlay_catalog_infos 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER increment_inlay_catalog_infos_version 
    BEFORE UPDATE ON inlay_catalog_infos 
    FOR EACH ROW EXECUTE FUNCTION increment_version_column();

CREATE TRIGGER update_inlay_custom_infos_updated_at 
    BEFORE UPDATE ON inlay_custom_infos 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER increment_inlay_custom_infos_version 
    BEFORE UPDATE ON inlay_custom_infos 
    FOR EACH ROW EXECUTE FUNCTION increment_version_column();

CREATE TRIGGER update_inlay_chats_updated_at 
    BEFORE UPDATE ON inlay_chats 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER increment_inlay_chats_version 
    BEFORE UPDATE ON inlay_chats 
    FOR EACH ROW EXECUTE FUNCTION increment_version_column();

CREATE TRIGGER update_inlay_proofs_updated_at 
    BEFORE UPDATE ON inlay_proofs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER increment_inlay_proofs_version 
    BEFORE UPDATE ON inlay_proofs 
    FOR EACH ROW EXECUTE FUNCTION increment_version_column();
