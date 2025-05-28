package usecase

import (
	"github.com/go-park-mail-ru/2025_1_ProVVeb/logger"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/repository"
)

type DeleteComplaint struct {
	QueryService repository.ComplaintRepository
	logger       *logger.LogrusLogger
}

func NewDeleteComplaintUseCase(
	queryService repository.ComplaintRepository,
	logger *logger.LogrusLogger,
) (*DeleteComplaint, error) {
	if queryService == nil || logger == nil {
		return nil, model.ErrGetActiveQueriesUC
	}
	return &DeleteComplaint{QueryService: queryService, logger: logger}, nil
}

func (uc *DeleteComplaint) DeleteComplaint(complaint_id int) error {
	uc.logger.Info("DeleteComplaint", "complaint_id", complaint_id)
	err := uc.QueryService.DeleteComplaint(complaint_id)
	return err
}
