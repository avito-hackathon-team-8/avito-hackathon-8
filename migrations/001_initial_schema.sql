CREATE TABLE IF NOT EXISTS users (
    id uuid PRIMARY KEY,
    email varchar(255) NOT NULL UNIQUE,
    verified boolean NOT NULL DEFAULT false,
    interests jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz,
    updated_at timestamptz
);

CREATE TABLE IF NOT EXISTS pets (
    id uuid PRIMARY KEY,
    user_id uuid NOT NULL UNIQUE,
    name varchar(35) NOT NULL DEFAULT '',
    level integer NOT NULL DEFAULT 1,
    leaves bigint NOT NULL DEFAULT 0,
    created_at timestamptz,
    updated_at timestamptz
);

CREATE TABLE IF NOT EXISTS level_rewards (
    id uuid PRIMARY KEY,
    user_id uuid NOT NULL,
    level smallint NOT NULL,
    title varchar(128) NOT NULL,
    description text NOT NULL,
    category varchar(32) NOT NULL,
    claim_expires_at timestamptz,
    claimed_at timestamptz,
    created_at timestamptz,
    updated_at timestamptz,
    UNIQUE (user_id, level),
    UNIQUE (user_id, id)
);

CREATE TABLE IF NOT EXISTS chest_openings (
    id uuid PRIMARY KEY,
    user_id uuid NOT NULL,
    leaves_spent bigint NOT NULL,
    opened_at timestamptz NOT NULL,
    created_at timestamptz,
    UNIQUE (user_id, id)
);

CREATE TABLE IF NOT EXISTS rewards (
    id uuid PRIMARY KEY,
    user_id uuid NOT NULL,
    level_reward_id uuid,
    chest_opening_id uuid,
    title text NOT NULL,
    category varchar(32) NOT NULL,
    source varchar(32) NOT NULL,
    expires_at timestamptz NOT NULL,
    redeemed_at timestamptz,
    created_at timestamptz NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_rewards_level_reward_id ON rewards (level_reward_id) WHERE level_reward_id IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_rewards_chest_opening_id ON rewards (chest_opening_id) WHERE chest_opening_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_rewards_active ON rewards (user_id, expires_at, redeemed_at);
CREATE INDEX IF NOT EXISTS idx_rewards_category ON rewards (category);

CREATE TABLE IF NOT EXISTS weekly_login_claims (
    id uuid PRIMARY KEY,
    user_id uuid NOT NULL,
    claim_date date NOT NULL,
    reward_leaves integer NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    UNIQUE (user_id, claim_date)
);

CREATE TABLE IF NOT EXISTS user_logins (
    id uuid PRIMARY KEY,
    user_id uuid NOT NULL,
    activity_date date NOT NULL,
    created_at timestamptz NOT NULL,
    UNIQUE (user_id, activity_date)
);

CREATE TABLE IF NOT EXISTS otps (
    id uuid PRIMARY KEY,
    user_id uuid NOT NULL UNIQUE,
    code_hash varchar(255) NOT NULL,
    failed_attempts integer NOT NULL DEFAULT 0,
    expires_at timestamptz NOT NULL,
    created_at timestamptz
);

CREATE TABLE IF NOT EXISTS external_events (
    id uuid PRIMARY KEY,
    user_id uuid NOT NULL,
    type varchar(64) NOT NULL,
    count integer NOT NULL,
    occurred_at timestamptz NOT NULL,
    payload_hash char(64) NOT NULL,
    created_at timestamptz NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_external_events_user_id ON external_events (user_id);
CREATE INDEX IF NOT EXISTS idx_external_events_occurred_at ON external_events (occurred_at);

CREATE TABLE IF NOT EXISTS user_game_states (
    user_id uuid PRIMARY KEY,
    pet_level integer NOT NULL DEFAULT 1,
    leaf_balance bigint NOT NULL DEFAULT 0,
    updated_at timestamptz NOT NULL
);

CREATE TABLE IF NOT EXISTS leaf_transactions (
    id uuid PRIMARY KEY,
    user_id uuid NOT NULL,
    amount bigint NOT NULL,
    reason varchar(32) NOT NULL,
    operation_key varchar(160) NOT NULL UNIQUE,
    occurred_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_leaf_period ON leaf_transactions (user_id, occurred_at);
CREATE INDEX IF NOT EXISTS idx_leaf_transactions_reason ON leaf_transactions (reason);

CREATE TABLE IF NOT EXISTS daily_task_definitions (
    id uuid PRIMARY KEY,
    code varchar(64) NOT NULL UNIQUE,
    title text NOT NULL,
    slot integer NOT NULL,
    type varchar(64) NOT NULL,
    target_count integer NOT NULL,
    reward integer NOT NULL,
    unlock_level integer NOT NULL,
    categories jsonb NOT NULL DEFAULT '[]'::jsonb,
    active boolean NOT NULL DEFAULT true
);
CREATE INDEX IF NOT EXISTS idx_daily_task_definitions_slot ON daily_task_definitions (slot);
CREATE INDEX IF NOT EXISTS idx_daily_task_definitions_type ON daily_task_definitions (type);

CREATE TABLE IF NOT EXISTS user_daily_tasks (
    id uuid PRIMARY KEY,
    user_id uuid NOT NULL,
    task_definition_id uuid NOT NULL,
    day date NOT NULL,
    status varchar(16) NOT NULL,
    current_count integer NOT NULL DEFAULT 0,
    completed_at timestamptz,
    claimed_at timestamptz,
    expires_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    UNIQUE (user_id, day, task_definition_id)
);
CREATE INDEX IF NOT EXISTS idx_user_daily_tasks_completed_at ON user_daily_tasks (completed_at);
CREATE INDEX IF NOT EXISTS idx_user_daily_tasks_claimed_at ON user_daily_tasks (claimed_at);
CREATE INDEX IF NOT EXISTS idx_user_daily_tasks_expires_at ON user_daily_tasks (expires_at);

CREATE TABLE IF NOT EXISTS leaderboard_entries (
    period_start date NOT NULL,
    user_id uuid NOT NULL,
    leaves bigint NOT NULL,
    rank bigint NOT NULL,
    calculated_at timestamptz NOT NULL,
    PRIMARY KEY (period_start, user_id)
);
CREATE INDEX IF NOT EXISTS idx_leaderboard_entries_rank ON leaderboard_entries (rank);

CREATE TABLE IF NOT EXISTS leaderboard_seasons (
    period_start date PRIMARY KEY,
    finalized_at timestamptz
);

CREATE TABLE IF NOT EXISTS job_runs (
    job_name varchar(128) NOT NULL,
    run_day date NOT NULL,
    ran_at timestamptz NOT NULL,
    PRIMARY KEY (job_name, run_day)
);
