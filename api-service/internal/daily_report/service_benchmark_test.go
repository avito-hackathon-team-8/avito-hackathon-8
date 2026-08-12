package daily_report

import (
	"fmt"
	"testing"
	"time"

	"github.com/avito-hackathon-team-8/avito-hackathon-8/api-service/internal/models"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func BenchmarkLeafActivityForDay(b *testing.B) {
	db, err := gorm.Open(sqlite.Open("file:"+uuid.NewString()+"?mode=memory&cache=shared"), &gorm.Config{})

	if err != nil {
		b.Fatal(err)
	}

	if err := db.AutoMigrate(&models.User{}, &models.Pet{}, &models.LeafTransaction{}); err != nil {
		b.Fatal(err)
	}

	user := models.User{Email: fmt.Sprintf("%s@example.com", uuid.NewString())}

	if err := db.Create(&user).Error; err != nil {
		b.Fatal(err)
	}

	pet := models.Pet{UserID: user.ID, Level: 1}

	if err := db.Create(&pet).Error; err != nil {
		b.Fatal(err)
	}

	day := time.Date(2026, 8, 12, 0, 0, 0, 0, time.UTC)
	rows := make([]models.LeafTransaction, 10000)

	for index := range rows {
		rows[index] = models.LeafTransaction{UserID: user.ID, Amount: 1, Reason: models.LeafReasonTaskReward, OperationKey: fmt.Sprintf("bench:%d", index), OccurredAt: day.Add(time.Duration(index) * time.Millisecond)}
	}

	if err := db.CreateInBatches(rows, 500).Error; err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for range b.N {
		if _, _, _, err := leafActivityForDay(db, user.ID, day, day.AddDate(0, 0, 1)); err != nil {
			b.Fatal(err)
		}
	}
}
