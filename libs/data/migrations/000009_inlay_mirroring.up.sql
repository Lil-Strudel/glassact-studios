--------------------------------------------------------------------------------
-- INLAY MIRRORING
--
-- A dealership can mirror an inlay left-to-right or top-to-bottom. The flags
-- are an orientation instruction for production, not a new design: they never
-- reset approval or change the price, and design downloads come out mirrored
-- so the piece cannot be cut the wrong way round.
--------------------------------------------------------------------------------

ALTER TABLE inlays
    ADD COLUMN mirror_x BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN mirror_y BOOLEAN NOT NULL DEFAULT false;
