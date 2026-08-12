package testutil

import "github.com/google/uuid"

type DailyReportNotifierMock struct{}

func (DailyReportNotifierMock) Notify(uuid.UUID) {}
