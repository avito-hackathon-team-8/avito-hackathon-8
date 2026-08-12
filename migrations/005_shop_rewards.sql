ALTER TABLE rewards ADD COLUMN IF NOT EXISTS item_type varchar(64);
CREATE INDEX IF NOT EXISTS idx_rewards_item_type ON rewards (item_type);

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'chk_rewards_origin') THEN
        ALTER TABLE rewards DROP CONSTRAINT chk_rewards_origin;
    END IF;

    ALTER TABLE rewards ADD CONSTRAINT chk_rewards_origin CHECK (
        (source = 'LEVEL' AND level_reward_id IS NOT NULL AND chest_opening_id IS NULL AND item_type IS NULL) OR
        (source = 'CHEST' AND level_reward_id IS NULL AND chest_opening_id IS NOT NULL AND item_type IS NULL) OR
        (source = 'LEADERBOARD' AND level_reward_id IS NULL AND chest_opening_id IS NULL AND item_type IS NULL) OR
        (source = 'SHOP' AND level_reward_id IS NULL AND chest_opening_id IS NULL AND item_type IS NOT NULL)
    );
END $$;
