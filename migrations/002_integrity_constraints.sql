DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_pets_user') THEN
        ALTER TABLE pets ADD CONSTRAINT fk_pets_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_level_rewards_user') THEN
        ALTER TABLE level_rewards ADD CONSTRAINT fk_level_rewards_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_chest_openings_user') THEN
        ALTER TABLE chest_openings ADD CONSTRAINT fk_chest_openings_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_rewards_user') THEN
        ALTER TABLE rewards ADD CONSTRAINT fk_rewards_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_rewards_level_reward_owner') THEN
        ALTER TABLE rewards ADD CONSTRAINT fk_rewards_level_reward_owner
            FOREIGN KEY (user_id, level_reward_id) REFERENCES level_rewards(user_id, id) ON DELETE RESTRICT;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_rewards_chest_opening_owner') THEN
        ALTER TABLE rewards ADD CONSTRAINT fk_rewards_chest_opening_owner
            FOREIGN KEY (user_id, chest_opening_id) REFERENCES chest_openings(user_id, id) ON DELETE RESTRICT;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_weekly_login_claims_user') THEN
        ALTER TABLE weekly_login_claims ADD CONSTRAINT fk_weekly_login_claims_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_user_logins_user') THEN
        ALTER TABLE user_logins ADD CONSTRAINT fk_user_logins_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_otps_user') THEN
        ALTER TABLE otps ADD CONSTRAINT fk_otps_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_external_events_user') THEN
        ALTER TABLE external_events ADD CONSTRAINT fk_external_events_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_user_game_states_user') THEN
        ALTER TABLE user_game_states ADD CONSTRAINT fk_user_game_states_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_leaf_transactions_user') THEN
        ALTER TABLE leaf_transactions ADD CONSTRAINT fk_leaf_transactions_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_user_daily_tasks_user') THEN
        ALTER TABLE user_daily_tasks ADD CONSTRAINT fk_user_daily_tasks_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_user_daily_tasks_definition') THEN
        ALTER TABLE user_daily_tasks ADD CONSTRAINT fk_user_daily_tasks_definition FOREIGN KEY (task_definition_id) REFERENCES daily_task_definitions(id) ON DELETE RESTRICT;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_leaderboard_entries_user') THEN
        ALTER TABLE leaderboard_entries ADD CONSTRAINT fk_leaderboard_entries_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'chk_pets_level') THEN
        ALTER TABLE pets ADD CONSTRAINT chk_pets_level CHECK (level BETWEEN 1 AND 10);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'chk_pets_leaves') THEN
        ALTER TABLE pets ADD CONSTRAINT chk_pets_leaves CHECK (leaves >= 0);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'chk_level_rewards_level') THEN
        ALTER TABLE level_rewards ADD CONSTRAINT chk_level_rewards_level CHECK (level BETWEEN 1 AND 10);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'chk_chest_openings_cost') THEN
        ALTER TABLE chest_openings ADD CONSTRAINT chk_chest_openings_cost CHECK (leaves_spent = 200);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'chk_rewards_origin') THEN
        ALTER TABLE rewards ADD CONSTRAINT chk_rewards_origin CHECK (
            (source = 'LEVEL' AND level_reward_id IS NOT NULL AND chest_opening_id IS NULL) OR
            (source = 'CHEST' AND level_reward_id IS NULL AND chest_opening_id IS NOT NULL) OR
            (source = 'LEADERBOARD' AND level_reward_id IS NULL AND chest_opening_id IS NULL)
        );
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'chk_external_events_count') THEN
        ALTER TABLE external_events ADD CONSTRAINT chk_external_events_count CHECK (count > 0);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'chk_daily_task_definitions_values') THEN
        ALTER TABLE daily_task_definitions ADD CONSTRAINT chk_daily_task_definitions_values
            CHECK (slot BETWEEN 1 AND 4 AND target_count > 0 AND reward > 0 AND unlock_level > 0);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'chk_user_daily_tasks_status') THEN
        ALTER TABLE user_daily_tasks ADD CONSTRAINT chk_user_daily_tasks_status
            CHECK (status IN ('LOCKED', 'IN_PROGRESS', 'COMPLETED', 'CLAIMED', 'EXPIRED'));
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'chk_user_daily_tasks_count') THEN
        ALTER TABLE user_daily_tasks ADD CONSTRAINT chk_user_daily_tasks_count CHECK (current_count >= 0);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'chk_weekly_login_reward') THEN
        ALTER TABLE weekly_login_claims ADD CONSTRAINT chk_weekly_login_reward CHECK (reward_leaves > 0);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'chk_leaderboard_entry_values') THEN
        ALTER TABLE leaderboard_entries ADD CONSTRAINT chk_leaderboard_entry_values CHECK (leaves >= 0 AND rank > 0);
    END IF;
END $$;
