--------------------------------------------------------------------------------
-- MANUFACTURING LADDER: manufacturing -> cutting + fire-polish
--
-- Mirrors the up migration: the constraints come off before the remap, since
-- neither value set covers both sides of it. Which half of the merged step an
-- inlay was actually in is not recoverable, so everything collapses onto
-- 'cutting' — no row comes back as 'fire-polish'.
--------------------------------------------------------------------------------

ALTER TABLE inlays DROP CONSTRAINT inlays_manufacturing_step_check;
ALTER TABLE inlay_milestones DROP CONSTRAINT inlay_milestones_step_check;

UPDATE inlays SET manufacturing_step = 'cutting'
WHERE manufacturing_step = 'manufacturing';

UPDATE inlay_milestones SET step = 'cutting'
WHERE step = 'manufacturing';

UPDATE inlay_updates SET step = 'cutting'
WHERE step = 'manufacturing';

ALTER TABLE inlays ADD CONSTRAINT inlays_manufacturing_step_check CHECK (
    manufacturing_step IS NULL OR manufacturing_step IN (
        'ordered', 'materials-prep', 'cutting', 'fire-polish', 'packaging', 'ready-to-ship'
    )
);

ALTER TABLE inlay_milestones ADD CONSTRAINT inlay_milestones_step_check CHECK (
    step IN ('ordered', 'materials-prep', 'cutting', 'fire-polish', 'packaging', 'ready-to-ship')
);
