--------------------------------------------------------------------------------
-- COMPOUND INLAYS
--
-- A compound is an admin-authored arrangement of catalog items that looks good
-- together. Adding one to a project creates an ordinary stock inlay per member,
-- so ordering, proofs and manufacturing never need to know compounds exist;
-- the arrangement only survives as a group in the project's stone design.
--------------------------------------------------------------------------------

CREATE TABLE catalog_compounds (
    id SERIAL PRIMARY KEY,
    uuid UUID DEFAULT gen_random_uuid() UNIQUE NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    is_active BOOLEAN NOT NULL DEFAULT true,
    display_order INTEGER,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    version INTEGER NOT NULL DEFAULT 1
);

CREATE INDEX idx_catalog_compounds_active ON catalog_compounds(is_active) WHERE is_active = true;

-- center_x/center_y are inches from the compound's own center, so the group can
-- be dropped anywhere on a stone by adding the drop point.
CREATE TABLE catalog_compound_members (
    id SERIAL PRIMARY KEY,
    compound_id INTEGER NOT NULL REFERENCES catalog_compounds ON DELETE CASCADE,
    catalog_item_id INTEGER NOT NULL REFERENCES catalog_items ON DELETE RESTRICT,
    center_x DOUBLE PRECISION NOT NULL,
    center_y DOUBLE PRECISION NOT NULL,
    rotation_deg DOUBLE PRECISION NOT NULL DEFAULT 0,
    sort_order INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_catalog_compound_members_compound ON catalog_compound_members(compound_id, sort_order);
CREATE INDEX idx_catalog_compound_members_item ON catalog_compound_members(catalog_item_id);

CREATE TRIGGER update_catalog_compounds_updated_at
    BEFORE UPDATE ON catalog_compounds
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER increment_catalog_compounds_version
    BEFORE UPDATE ON catalog_compounds
    FOR EACH ROW EXECUTE FUNCTION increment_version_column();
