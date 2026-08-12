CREATE TABLE IF NOT EXISTS pet_states (
    user_id uuid PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    happiness double precision NOT NULL DEFAULT 100,
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT chk_pet_states_happiness CHECK (happiness BETWEEN 0 AND 100)
);

CREATE TABLE IF NOT EXISTS pet_state_actions (
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    action_type varchar(16) NOT NULL,
    next_available_at timestamptz NOT NULL,
    PRIMARY KEY (user_id, action_type),
    CONSTRAINT chk_pet_state_action_type CHECK (action_type IN ('STROKE', 'FEED'))
);

CREATE TABLE IF NOT EXISTS pet_state_outbox (
    event_id uuid PRIMARY KEY,
    topic varchar(128) NOT NULL,
    event_key varchar(128) NOT NULL,
    payload jsonb NOT NULL,
    created_at timestamptz NOT NULL,
    published_at timestamptz
);
CREATE INDEX IF NOT EXISTS idx_pet_state_outbox_pending
    ON pet_state_outbox (created_at) WHERE published_at IS NULL;

ALTER TABLE weekly_login_claims
    ADD COLUMN IF NOT EXISTS base_reward_leaves integer,
    ADD COLUMN IF NOT EXISTS happiness_snapshot double precision,
    ADD COLUMN IF NOT EXISTS happiness_multiplier double precision;

UPDATE weekly_login_claims
SET base_reward_leaves = reward_leaves
WHERE base_reward_leaves IS NULL;

ALTER TABLE weekly_login_claims
    ALTER COLUMN base_reward_leaves SET NOT NULL;
