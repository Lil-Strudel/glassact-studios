--------------------------------------------------------------------------------
-- SANDBLAST BACKGROUNDS
--
-- Sandblast masks for the stone face. They used to be catalog items in the
-- SB-BACKGROUNDS category, which dragged inlay-only baggage along (price
-- group, inch dimensions, glass/grout manifest) and meant every inlay path had
-- to refuse them. A background has no size of its own: it is scaled to the
-- stone when placed in the designer.
--------------------------------------------------------------------------------

CREATE TABLE sandblast_backgrounds (
    id SERIAL PRIMARY KEY,
    uuid UUID DEFAULT gen_random_uuid() UNIQUE NOT NULL,
    code TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    svg_url TEXT NOT NULL,
    tags TEXT[] NOT NULL DEFAULT '{}',
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    version INTEGER NOT NULL DEFAULT 1
);

CREATE INDEX idx_sandblast_backgrounds_active ON sandblast_backgrounds(is_active) WHERE is_active = true;

CREATE TRIGGER update_sandblast_backgrounds_updated_at
    BEFORE UPDATE ON sandblast_backgrounds
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER increment_sandblast_backgrounds_version
    BEFORE UPDATE ON sandblast_backgrounds
    FOR EACH ROW EXECUTE FUNCTION increment_version_column();

-- Catalog inlays an admin has picked as going well with a background (hearts
-- for a heart frame), offered first when that background is on a stone.
CREATE TABLE sandblast_background_related_items (
    id SERIAL PRIMARY KEY,
    background_id INTEGER NOT NULL REFERENCES sandblast_backgrounds ON DELETE CASCADE,
    catalog_item_id INTEGER NOT NULL REFERENCES catalog_items ON DELETE RESTRICT,
    sort_order INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (background_id, catalog_item_id)
);

CREATE INDEX idx_sandblast_background_related_items_background
    ON sandblast_background_related_items(background_id, sort_order);
CREATE INDEX idx_sandblast_background_related_items_item
    ON sandblast_background_related_items(catalog_item_id);

-- UUIDs are kept so stone designs that already place a background still
-- point at it.
INSERT INTO sandblast_backgrounds (uuid, code, name, description, svg_url, tags, is_active, created_at, updated_at)
SELECT
    c.uuid,
    c.catalog_code,
    c.name,
    c.description,
    c.svg_url,
    COALESCE(
        (SELECT array_agg(t.tag ORDER BY t.tag) FROM catalog_item_tags t WHERE t.catalog_item_id = c.id),
        '{}'
    ),
    c.is_active,
    c.created_at,
    c.updated_at
FROM catalog_items c
WHERE c.category = 'SB-BACKGROUNDS';

UPDATE project_stone_designs d
SET layout = jsonb_set(d.layout, '{elements}', (
    SELECT COALESCE(jsonb_agg(
        CASE
            WHEN el ? 'catalog_item_uuid'
                THEN (el - 'catalog_item_uuid') || jsonb_build_object('background_uuid', el -> 'catalog_item_uuid')
            ELSE el
        END
        ORDER BY ordinality
    ), '[]'::jsonb)
    FROM jsonb_array_elements(d.layout -> 'elements') WITH ORDINALITY AS e(el, ordinality)
))
WHERE EXISTS (
    SELECT 1 FROM jsonb_array_elements(d.layout -> 'elements') el WHERE el ? 'catalog_item_uuid'
);

-- A background ordered as an inlay before they were refused would block the
-- delete; it stays behind, retired, so that order's history is intact.
DELETE FROM catalog_items c
WHERE c.category = 'SB-BACKGROUNDS'
  AND NOT EXISTS (SELECT 1 FROM inlay_catalog_infos i WHERE i.catalog_item_id = c.id)
  AND NOT EXISTS (SELECT 1 FROM catalog_compound_members m WHERE m.catalog_item_id = c.id);

UPDATE catalog_items SET is_active = false
WHERE category = 'SB-BACKGROUNDS' AND is_active;
