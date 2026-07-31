--------------------------------------------------------------------------------
-- CHAT: per-project -> per-inlay
--------------------------------------------------------------------------------

CREATE TABLE inlay_chats (
    id SERIAL PRIMARY KEY,
    uuid UUID DEFAULT gen_random_uuid() UNIQUE NOT NULL,
    inlay_id INTEGER NOT NULL REFERENCES inlays ON DELETE CASCADE,
    dealership_user_id INTEGER REFERENCES dealership_users ON DELETE SET NULL,
    internal_user_id INTEGER REFERENCES internal_users ON DELETE SET NULL,
    message_type VARCHAR(255) NOT NULL DEFAULT 'text' CHECK (message_type IN (
        'text', 'image', 'proof_sent', 'proof_approved', 'proof_declined', 'system'
    )),
    message TEXT NOT NULL,
    attachment_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    version INTEGER NOT NULL DEFAULT 1,
    CONSTRAINT inlay_chats_sender_check CHECK (
        (dealership_user_id IS NOT NULL AND internal_user_id IS NULL) OR
        (dealership_user_id IS NULL AND internal_user_id IS NOT NULL) OR
        (dealership_user_id IS NULL AND internal_user_id IS NULL AND message_type IN ('system', 'proof_sent', 'proof_approved', 'proof_declined'))
    )
);

CREATE INDEX idx_inlay_chats_inlay ON inlay_chats(inlay_id);
CREATE INDEX idx_inlay_chats_created ON inlay_chats(inlay_id, created_at);

CREATE TRIGGER update_inlay_chats_updated_at
    BEFORE UPDATE ON inlay_chats
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER increment_inlay_chats_version
    BEFORE UPDATE ON inlay_chats
    FOR EACH ROW EXECUTE FUNCTION increment_version_column();

-- Only inlay-tagged messages can go back; project-wide messages have no home
-- in the per-inlay model and are dropped.
ALTER TABLE project_chats ADD COLUMN legacy_project_chat_id INTEGER;

INSERT INTO inlay_chats (
    inlay_id, dealership_user_id, internal_user_id,
    message_type, message, attachment_url, created_at, updated_at
)
SELECT pc.inlay_id, pc.dealership_user_id, pc.internal_user_id,
       pc.message_type, pc.message, pc.attachment_url, pc.created_at, pc.updated_at
FROM project_chats pc
WHERE pc.inlay_id IS NOT NULL
ORDER BY pc.created_at, pc.id;

-- Match the copies back to their source rows so proofs can be repointed.
UPDATE project_chats pc SET legacy_project_chat_id = ic.id
FROM inlay_chats ic
WHERE pc.inlay_id = ic.inlay_id
  AND pc.created_at = ic.created_at
  AND pc.message = ic.message
  AND pc.message_type = ic.message_type;

ALTER TABLE inlay_proofs DROP CONSTRAINT inlay_proofs_sent_in_chat_id_fkey;

UPDATE inlay_proofs p SET sent_in_chat_id = pc.legacy_project_chat_id
FROM project_chats pc
WHERE pc.id = p.sent_in_chat_id;

ALTER TABLE inlay_proofs ADD CONSTRAINT inlay_proofs_sent_in_chat_id_fkey
    FOREIGN KEY (sent_in_chat_id) REFERENCES inlay_chats;

ALTER TABLE project_chats DROP COLUMN legacy_project_chat_id;

DELETE FROM project_chats WHERE inlay_id IS NOT NULL;

ALTER TABLE project_chats DROP CONSTRAINT project_chats_sender_check;
ALTER TABLE project_chats ADD CONSTRAINT project_chats_sender_check CHECK (
    (dealership_user_id IS NOT NULL AND internal_user_id IS NULL) OR
    (dealership_user_id IS NULL AND internal_user_id IS NOT NULL) OR
    (dealership_user_id IS NULL AND internal_user_id IS NULL AND message_type = 'system')
);

ALTER TABLE project_chats DROP CONSTRAINT project_chats_message_type_check;
ALTER TABLE project_chats ADD CONSTRAINT project_chats_message_type_check CHECK (
    message_type IN ('text', 'image', 'system')
);

DROP INDEX idx_project_chats_inlay;
ALTER TABLE project_chats DROP COLUMN inlay_id;

--------------------------------------------------------------------------------
-- INSTALLATION KITS: per-project -> per-inlay
--------------------------------------------------------------------------------

ALTER TABLE inlays ADD COLUMN installation_kit BOOLEAN NOT NULL DEFAULT false;

ALTER TABLE order_snapshots
    ADD COLUMN installation_kit BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN installation_kit_price_cents INTEGER NOT NULL DEFAULT 0;

-- The project-level choice fans back out to every inlay it covered.
UPDATE inlays i SET installation_kit = true
FROM projects p
WHERE p.id = i.project_id AND p.installation_kit;

UPDATE order_snapshots os SET installation_kit = true
FROM projects p
WHERE p.id = os.project_id AND p.installation_kit;

ALTER TABLE projects
    DROP COLUMN installation_kit,
    DROP COLUMN installation_kit_price_cents;
