--------------------------------------------------------------------------------
-- STONE DESIGNS
--
-- One design per project: the stone a dealership's inlays will be set into,
-- with the inlays, sandblast backgrounds and lettering arranged on it. It is
-- a visual aid for the dealership and GlassAct staff, not part of the price,
-- and is frozen once the project leaves draft.
--
-- The stone itself gets columns so it can be queried and constrained. The
-- arrangement is JSONB because its elements are only ever read and written as
-- a whole by the designer; the API validates its shape on every write.
--------------------------------------------------------------------------------

CREATE TABLE project_stone_designs (
    id SERIAL PRIMARY KEY,
    uuid UUID DEFAULT gen_random_uuid() UNIQUE NOT NULL,
    project_id INTEGER UNIQUE NOT NULL REFERENCES projects ON DELETE CASCADE,
    stone_shape TEXT NOT NULL CHECK (stone_shape IN (
        'flat-top', 'serpentine', 'half-serpentine', 'oval-top',
        'cathedral', 'heart', 'flat-marker', 'slant'
    )),
    stone_width DOUBLE PRECISION NOT NULL CHECK (stone_width > 0 AND stone_width <= 240),
    stone_height DOUBLE PRECISION NOT NULL CHECK (stone_height > 0 AND stone_height <= 240),
    granite_key TEXT NOT NULL,
    granite_custom_hex TEXT CHECK (granite_custom_hex ~ '^#[0-9a-fA-F]{6}$'),
    layout JSONB NOT NULL DEFAULT '{"elements": [], "groups": []}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    version INTEGER NOT NULL DEFAULT 1
);

CREATE TRIGGER update_project_stone_designs_updated_at
    BEFORE UPDATE ON project_stone_designs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER increment_project_stone_designs_version
    BEFORE UPDATE ON project_stone_designs
    FOR EACH ROW EXECUTE FUNCTION increment_version_column();
