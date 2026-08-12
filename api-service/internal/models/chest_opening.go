package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const ChestOpeningLeavesCost int64 = 200

type ChestOpening struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;uniqueIndex:idx_chest_openings_user_id_id,priority:2"`
	UserID      uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_chest_openings_user_id_id,priority:1"`
	User        User      `gorm:"constraint:OnDelete:CASCADE"`
	LeavesSpent int64     `gorm:"not null;check:leaves_spent = 200"`
	OpenedAt    time.Time `gorm:"not null;index"`
	CreatedAt   time.Time
}

func (opening *ChestOpening) BeforeCreate(_ *gorm.DB) error {
	if opening.ID == uuid.Nil {
		opening.ID = uuid.New()
	}

	return nil
}
