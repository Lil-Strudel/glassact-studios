-- Lossy: the SB-BACKGROUNDS catalog rows the up migration deleted are not
-- recreated, since the catalog columns they need (price group, dimensions,
-- manifest) were never carried over. Restore them from a catalog snapshot.

UPDATE project_stone_designs d
SET layout = jsonb_set(d.layout, '{elements}', (
    SELECT COALESCE(jsonb_agg(
        CASE
            WHEN el ? 'background_uuid'
                THEN (el - 'background_uuid') || jsonb_build_object('catalog_item_uuid', el -> 'background_uuid')
            ELSE el
        END
        ORDER BY ordinality
    ), '[]'::jsonb)
    FROM jsonb_array_elements(d.layout -> 'elements') WITH ORDINALITY AS e(el, ordinality)
))
WHERE EXISTS (
    SELECT 1 FROM jsonb_array_elements(d.layout -> 'elements') el WHERE el ? 'background_uuid'
);

DROP TABLE sandblast_background_related_items;
DROP TABLE sandblast_backgrounds;
