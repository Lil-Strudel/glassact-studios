--------------------------------------------------------------------------------
-- INSTALLATION KITS: per-inlay -> per-project
--
-- A kit covers the install materials for every inlay in a project, so it is one
-- flat charge on the project rather than an add-on repeated per inlay.
--------------------------------------------------------------------------------

ALTER TABLE projects
    ADD COLUMN installation_kit BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN installation_kit_price_cents INTEGER;

-- A project wanted a kit if any of its inlays did.
UPDATE projects p SET installation_kit = true
WHERE EXISTS (
    SELECT 1 FROM inlays i WHERE i.project_id = p.id AND i.installation_kit
);

-- Preserve what each existing order was actually charged rather than re-pricing
-- history at the new flat rate.
UPDATE projects p SET installation_kit_price_cents = s.total
FROM (
    SELECT project_id, SUM(installation_kit_price_cents) AS total
    FROM order_snapshots
    GROUP BY project_id
) s
WHERE s.project_id = p.id AND s.total > 0;

ALTER TABLE inlays DROP COLUMN installation_kit;

ALTER TABLE order_snapshots
    DROP COLUMN installation_kit,
    DROP COLUMN installation_kit_price_cents;

--------------------------------------------------------------------------------
-- CHAT: per-inlay -> per-project
--
-- project_chats already existed but was never wired up. It becomes the single
-- thread for a project; inlay_id tags a message with the inlay it is about so
-- the UI can link straight to it.
--------------------------------------------------------------------------------

ALTER TABLE project_chats
    ADD COLUMN inlay_id INTEGER REFERENCES inlays ON DELETE SET NULL,
    ADD COLUMN legacy_inlay_chat_id INTEGER;

CREATE INDEX idx_project_chats_inlay ON project_chats(inlay_id) WHERE inlay_id IS NOT NULL;

-- project_chats was scoped to text/image/system; it now carries proof events too.
ALTER TABLE project_chats DROP CONSTRAINT project_chats_message_type_check;
ALTER TABLE project_chats ADD CONSTRAINT project_chats_message_type_check CHECK (
    message_type IN ('text', 'image', 'proof_sent', 'proof_approved', 'proof_declined', 'system')
);

ALTER TABLE project_chats DROP CONSTRAINT project_chats_sender_check;
ALTER TABLE project_chats ADD CONSTRAINT project_chats_sender_check CHECK (
    (dealership_user_id IS NOT NULL AND internal_user_id IS NULL) OR
    (dealership_user_id IS NULL AND internal_user_id IS NOT NULL) OR
    (dealership_user_id IS NULL AND internal_user_id IS NULL AND message_type IN ('system', 'proof_sent', 'proof_approved', 'proof_declined'))
);

INSERT INTO project_chats (
    project_id, inlay_id, dealership_user_id, internal_user_id,
    message_type, message, attachment_url, created_at, updated_at,
    legacy_inlay_chat_id
)
SELECT i.project_id, c.inlay_id, c.dealership_user_id, c.internal_user_id,
       c.message_type, c.message, c.attachment_url, c.created_at, c.updated_at,
       c.id
FROM inlay_chats c
JOIN inlays i ON i.id = c.inlay_id
ORDER BY c.created_at, c.id;

-- Repoint proof -> chat links before the old table goes away.
ALTER TABLE inlay_proofs DROP CONSTRAINT inlay_proofs_sent_in_chat_id_fkey;

UPDATE inlay_proofs p SET sent_in_chat_id = pc.id
FROM project_chats pc
WHERE pc.legacy_inlay_chat_id = p.sent_in_chat_id;

ALTER TABLE inlay_proofs ADD CONSTRAINT inlay_proofs_sent_in_chat_id_fkey
    FOREIGN KEY (sent_in_chat_id) REFERENCES project_chats;

ALTER TABLE project_chats DROP COLUMN legacy_inlay_chat_id;

DROP TABLE inlay_chats;
