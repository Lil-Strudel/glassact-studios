--------------------------------------------------------------------------------
-- DEALERSHIP CONTACT & PRODUCTION PREFERENCES
--------------------------------------------------------------------------------

ALTER TABLE dealerships
    DROP COLUMN phone,
    DROP COLUMN sandblast_file_format;

--------------------------------------------------------------------------------
-- PAYMENT TIMING: three-way stage marker -> boolean
--
-- Both pre-payment stages collapse back onto the single boolean; which of the
-- two it was is not recoverable.
--------------------------------------------------------------------------------

ALTER TABLE dealerships
    ADD COLUMN requires_payment_before_shipping BOOLEAN NOT NULL DEFAULT false;

UPDATE dealerships SET requires_payment_before_shipping = true
WHERE payment_timing IN ('pre-manufacturing', 'pre-shipping');

ALTER TABLE dealerships DROP COLUMN payment_timing;
