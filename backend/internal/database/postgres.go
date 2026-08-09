package database

import (
	"github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Open(databaseURL string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})

	if err != nil {
		return nil, err
	}

	if err := migrateDailyTaskDefinitionType(db); err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(
		&models.User{},
		&models.Pet{},
		&models.LevelReward{},
		&models.ChestOpening{},
		&models.Reward{},
		&models.WeeklyLoginClaim{},
		&models.UserLogin{},
		&models.ExternalEvent{},
		&models.UserGameState{},
		&models.LeafTransaction{},
		&models.DailyTaskDefinition{},
		&models.UserDailyTask{},
		&models.LeaderboardEntry{},
		&models.LeaderboardSeason{},
		&models.JobRun{},
	); err != nil {
		return nil, err
	}
	if err := ensureRewardOwnerConstraints(db); err != nil {
		return nil, err
	}

	return db, nil
}

func ensureRewardOwnerConstraints(db *gorm.DB) error {
	constraints := []string{
		`DO $$
		BEGIN
			IF EXISTS (
				SELECT 1
				FROM pg_constraint
				WHERE conname = 'fk_rewards_level_reward'
					AND conrelid = 'rewards'::regclass
					AND pg_get_constraintdef(oid) LIKE 'FOREIGN KEY (user_id, level_reward_id) REFERENCES level_rewards(user_id, id) ON DELETE RESTRICT%'
			) THEN
				RETURN;
			END IF;

			ALTER TABLE rewards DROP CONSTRAINT IF EXISTS fk_rewards_level_reward;
			ALTER TABLE rewards
				ADD CONSTRAINT fk_rewards_level_reward
				FOREIGN KEY (user_id, level_reward_id)
				REFERENCES level_rewards(user_id, id)
				ON DELETE RESTRICT;
		END $$`,
		`DO $$
		BEGIN
			IF EXISTS (
				SELECT 1
				FROM pg_constraint
				WHERE conname = 'fk_rewards_chest_opening'
					AND conrelid = 'rewards'::regclass
					AND pg_get_constraintdef(oid) LIKE 'FOREIGN KEY (user_id, chest_opening_id) REFERENCES chest_openings(user_id, id) ON DELETE RESTRICT%'
			) THEN
				RETURN;
			END IF;

			ALTER TABLE rewards DROP CONSTRAINT IF EXISTS fk_rewards_chest_opening;
			ALTER TABLE rewards
				ADD CONSTRAINT fk_rewards_chest_opening
				FOREIGN KEY (user_id, chest_opening_id)
				REFERENCES chest_openings(user_id, id)
				ON DELETE RESTRICT;
		END $$`,
	}

	for _, constraint := range constraints {
		if err := db.Exec(constraint).Error; err != nil {
			return err
		}
	}

	return nil
}
func migrateDailyTaskDefinitionType(db *gorm.DB) error {
	if !db.Migrator().HasTable(&models.DailyTaskDefinition{}) {
		return nil
	}

	if err := db.Exec(`ALTER TABLE daily_task_definitions ADD COLUMN IF NOT EXISTS type varchar(64)`).Error; err != nil {
		return err
	}

	if err := db.Exec(`UPDATE daily_task_definitions SET type = code WHERE type IS NULL`).Error; err != nil {
		return err
	}

	return db.Exec(`ALTER TABLE daily_task_definitions ALTER COLUMN type SET NOT NULL`).Error
}
