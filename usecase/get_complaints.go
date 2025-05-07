package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/logger"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/repository"

	"github.com/sirupsen/logrus"
)

type GetComplaint struct {
	complaintRepo repository.ComplaintRepository
	logger        *logger.LogrusLogger
}

func NewGetComplaintUseCase(complaintRepo repository.ComplaintRepository, logger *logger.LogrusLogger) (*GetComplaint, error) {

	return &GetComplaint{complaintRepo: complaintRepo, logger: logger}, nil
}

func (uc *GetComplaint) GetAllComplaints() ([]model.ComplaintWithLogins, error) {
	uc.logger.Info("GetAllComplaints")

	complaints, err := uc.complaintRepo.GetAllComplaints(context.Background())
	if err != nil {
		uc.logger.Error("CreateComplaint", "complaints", complaints, "error", err)
	} else {
		uc.logger.WithFields(&logrus.Fields{"complaints": complaints})
	}
	return complaints, err
}
