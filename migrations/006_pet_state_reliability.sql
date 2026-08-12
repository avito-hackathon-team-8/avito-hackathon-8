ALTER TABLE pet_state_outbox
    ADD COLUMN IF NOT EXISTS lease_token uuid,
    ADD COLUMN IF NOT EXISTS lease_until timestamptz;

CREATE INDEX IF NOT EXISTS idx_pet_state_outbox_claimable
    ON pet_state_outbox (created_at)
    WHERE published_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_pet_state_outbox_retention
    ON pet_state_outbox (published_at)
    WHERE published_at IS NOT NULL;

CREATE TABLE IF NOT EXISTS pet_state_care_idempotency (
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    idempotency_key varchar(128) NOT NULL,
    action_type varchar(16) NOT NULL,
    response jsonb NOT NULL,
    created_at timestamptz NOT NULL,
    PRIMARY KEY (user_id, idempotency_key),
    CONSTRAINT chk_pet_state_idempotency_action_type CHECK (action_type IN ('STROKE', 'FEED'))
);

CREATE INDEX IF NOT EXISTS idx_pet_state_care_idempotency_created_at
    ON pet_state_care_idempotency (created_at);

CREATE INDEX IF NOT EXISTS idx_leaderboard_entries_period_rank
    ON leaderboard_entries (period_start, rank);
