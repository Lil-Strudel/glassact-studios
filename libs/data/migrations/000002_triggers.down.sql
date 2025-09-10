DROP TRIGGER IF EXISTS update_inlay_custom_infos_updated_at ON inlay_custom_infos;
DROP TRIGGER IF EXISTS update_inlay_catalog_infos_updated_at ON inlay_catalog_infos;
DROP TRIGGER IF EXISTS update_inlays_updated_at ON inlays;
DROP TRIGGER IF EXISTS update_projects_updated_at ON projects;
DROP TRIGGER IF EXISTS update_catalog_items_updated_at ON catalog_items;
DROP TRIGGER IF EXISTS update_accounts_updated_at ON accounts;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP TRIGGER IF EXISTS update_dealerships_updated_at ON dealerships;

DROP FUNCTION IF EXISTS update_updated_at_column();
