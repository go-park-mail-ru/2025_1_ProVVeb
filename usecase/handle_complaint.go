package usecase

import (
	"github.com/go-park-mail-ru/2025_1_ProVVeb/logger"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/repository"
)

type HandleComplaint struct {
	QueryService repository.ComplaintRepository
	logger       *logger.LogrusLogger
}

func NewHandleComplaintUseCase(
	queryService repository.ComplaintRepository,
	logger *logger.LogrusLogger,
) (*HandleComplaint, error) {
	if queryService == nil || logger == nil {
		return nil, model.ErrGetActiveQueriesUC
	}
	return &HandleComplaint{QueryService: queryService, logger: logger}, nil
}

func (uc *HandleComplaint) HandleComplaint(complaint_id int, new_status int) error {
	uc.logger.Info("DeleteComplaint", "complaint_id", complaint_id)
	err := uc.QueryService.HandleComplaint(complaint_id, new_status)
	return err
}
