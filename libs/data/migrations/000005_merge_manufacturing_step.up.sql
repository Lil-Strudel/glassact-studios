--------------------------------------------------------------------------------
-- MANUFACTURING LADDER: cutting + fire-polish -> manufacturing
--
-- Production never tracked cutting and fire-polish separately: an inlay moves
-- through both as one continuous operation, so the split only ever produced two
-- kanban columns and two milestone events that said the same thing. They
-- collapse into a single 'manufacturing' step.
--
-- The old constraints are dropped before the remap, because neither the old nor
-- the new value set covers both sides of it. inlay_updates.step has no CHECK
-- constraint of its own, so it is remapped alongside them purely to keep dead
-- step names off the timeline.
--------------------------------------------------------------------------------

ALTER TABLE inlays DROP CONSTRAINT inlays_manufacturing_step_check;
ALTER TABLE inlay_milestones DROP CONSTRAINT inlay_milestones_step_check;

UPDATE inlays SET manufacturing_step = 'manufacturing'
WHERE manufacturing_step IN ('cutting', 'fire-polish');

UPDATE inlay_milestones SET step = 'manufacturing'
WHERE step IN ('cutting', 'fire-polish');

UPDATE inlay_updates SET step = 'manufacturing'
WHERE step IN ('cutting', 'fire-polish');

ALTER TABLE inlays ADD CONSTRAINT inlays_manufacturing_step_check CHECK (
    manufacturing_step IS NULL OR manufacturing_step IN (
        'ordered', 'materials-prep', 'manufacturing', 'packaging', 'ready-to-ship'
    )
);

ALTER TABLE inlay_milestones ADD CONSTRAINT inlay_milestones_step_check CHECK (
    step IN ('ordered', 'materials-prep', 'manufacturing', 'packaging', 'ready-to-ship')
);
