package usecase

import (
	"github.com/go-park-mail-ru/2025_1_ProVVeb/logger"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/repository"

	"github.com/sirupsen/logrus"
)

type FindComplaint struct {
	complaintRepo repository.ComplaintRepository
	logger        *logger.LogrusLogger
}

func NewFindComplaintUseCase(complaintRepo repository.ComplaintRepository, logger *logger.LogrusLogger) (*FindComplaint, error) {

	return &FindComplaint{complaintRepo: complaintRepo, logger: logger}, nil
}

func (uc *FindComplaint) FindComplaint(complaint_by int, name_by string, complaint_on int, name_on string, complaint_type string, status int) ([]model.ComplaintWithLogins, error) {
	uc.logger.Info("FindComplaint")

	complaints, err := uc.complaintRepo.FindComplaint(complaint_by, name_by, complaint_on, name_on, complaint_type, status)
	if err != nil {
		uc.logger.Error("FindComplaint", "complaints", complaints, "error", err)
	} else {
		uc.logger.WithFields(&logrus.Fields{"complaints": complaints}).Info("Found complaints")
	}
	return complaints, err
}
