package tasks

import "github.com/avito-hackathon-team-8/avito-hackathon-8/backend/internal/models"

var arrTasks = []models.TaskType{
	models.OpenNotificationsTaskType,
	models.AddToFavoritesTaskType,
	models.PublishListingTaskType,
	models.BoostListingTaskType,
	models.LeaveReviewTaskType,
	models.CompleteDealTaskType,
	models.OrderWithDeliveryTaskType,
}
