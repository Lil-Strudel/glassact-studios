--------------------------------------------------------------------------------
-- PAYMENT TIMING: boolean -> three-way stage marker
--
-- Some dealerships are asked to pay before manufacturing starts, not just
-- before shipping, which a boolean cannot express. Like the boolean it
-- replaces, this gates nothing in the app: it only decides which informational
-- message a project shows to the dealership and to internal staff.
--------------------------------------------------------------------------------

ALTER TABLE dealerships
    ADD COLUMN payment_timing TEXT NOT NULL DEFAULT 'post-shipping'
        CHECK (payment_timing IN ('pre-manufacturing', 'pre-shipping', 'post-shipping'));

UPDATE dealerships SET payment_timing = 'pre-shipping'
WHERE requires_payment_before_shipping;

ALTER TABLE dealerships DROP COLUMN requires_payment_before_shipping;

--------------------------------------------------------------------------------
-- DEALERSHIP CONTACT & PRODUCTION PREFERENCES
--
-- sandblast_file_format is what the dealership wants the per-inlay sandblasting
-- file delivered in, surfaced to whoever uploads it. Existing dealerships take
-- the 'pdf' default.
--------------------------------------------------------------------------------

ALTER TABLE dealerships
    ADD COLUMN sandblast_file_format TEXT NOT NULL DEFAULT 'pdf'
        CHECK (sandblast_file_format IN ('pdf', 'svg', 'png', 'dxf')),
    -- Digits only; the UI is responsible for formatting it for display.
    ADD COLUMN phone TEXT NOT NULL DEFAULT ''
        CHECK (phone ~ '^[0-9]*$');
