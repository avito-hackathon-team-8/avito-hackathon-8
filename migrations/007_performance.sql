CREATE TABLE IF NOT EXISTS leaderboard_totals (
    period_start date NOT NULL,
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    leaves bigint NOT NULL DEFAULT 0,
    updated_at timestamptz NOT NULL,
    PRIMARY KEY (period_start, user_id),
    CONSTRAINT chk_leaderboard_totals_leaves CHECK (leaves >= 0)
);

INSERT INTO leaderboard_totals (period_start, user_id, leaves, updated_at)
SELECT date_trunc('month', occurred_at)::date, user_id,
       SUM(amount), MAX(occurred_at)
FROM leaf_transactions
WHERE amount > 0
GROUP BY date_trunc('month', occurred_at)::date, user_id
ON CONFLICT (period_start, user_id) DO UPDATE
SET leaves = EXCLUDED.leaves, updated_at = EXCLUDED.updated_at;

CREATE INDEX IF NOT EXISTS idx_rewards_user_created
    ON rewards (user_id, created_at);
CREATE INDEX IF NOT EXISTS idx_user_logins_user_activity
    ON user_logins (user_id, activity_date);
CREATE INDEX IF NOT EXISTS idx_leaf_transactions_user_time_reason
    ON leaf_transactions (user_id, occurred_at, reason);
CREATE INDEX IF NOT EXISTS idx_user_daily_tasks_user_completed_claimed
    ON user_daily_tasks (user_id, completed_at, claimed_at);
