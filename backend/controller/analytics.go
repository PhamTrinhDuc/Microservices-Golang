package controller

import (
	"strconv"
	"time"

	"backend/domain"
	"backend/internal/utils"

	"github.com/gin-gonic/gin"
)

type AnalyticsController struct {
	usecase domain.AnalyticsUsecase
}

func NewAnalyticsController(usecase domain.AnalyticsUsecase) *AnalyticsController {
	return &AnalyticsController{usecase: usecase}
}

func (ctl *AnalyticsController) GetAnalytics(ctx *gin.Context) {
	startDateStr := ctx.Query("start_date")
	endDateStr := ctx.Query("end_date")
	storeIDStr := ctx.Query("store_id")

	var start, end time.Time
	var err error

	if startDateStr != "" {
		start, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			utils.RespondBadRequest(ctx, "invalid start_date format, must be YYYY-MM-DD")
			return
		}
	} else {
		// Default to 30 days ago (beginning of the day)
		now := time.Now()
		start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, -30)
	}

	if endDateStr != "" {
		end, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			utils.RespondBadRequest(ctx, "invalid end_date format, must be YYYY-MM-DD")
			return
		}
		// Adjust end date to cover the entire day (up to 23:59:59)
		end = time.Date(end.Year(), end.Month(), end.Day(), 23, 59, 59, 999999999, end.Location())
	} else {
		// Default to now
		end = time.Now()
	}

	var storeID *int
	if storeIDStr != "" {
		id, err := strconv.Atoi(storeIDStr)
		if err != nil {
			utils.RespondBadRequest(ctx, "invalid store_id, must be an integer")
			return
		}
		storeID = &id
	}

	report, err := ctl.usecase.GetSummary(ctx.Request.Context(), start, end, storeID)
	if err != nil {
		utils.RespondInternalError(ctx, err.Error())
		return
	}

	utils.RespondOK(ctx, report)
}
