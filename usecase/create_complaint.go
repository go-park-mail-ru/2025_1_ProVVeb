package usecase

import (
	"github.com/go-park-mail-ru/2025_1_ProVVeb/logger"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/repository"

	"github.com/sirupsen/logrus"
)

type CreateComplaint struct {
	complaintRepo repository.ComplaintRepository
	logger        *logger.LogrusLogger
}

func NewCreateComplaintUseCase(complaintRepo repository.ComplaintRepository, logger *logger.LogrusLogger) (*CreateComplaint, error) {

	return &CreateComplaint{complaintRepo: complaintRepo, logger: logger}, nil
}

func (uc *CreateComplaint) CreateComplaint(complaint_by int, complaint_on int, ComplaintType string, text string) error {
	uc.logger.Info("CreateComplaint", "complaint_by", complaint_by, "complaint_on", complaint_on)

	err := uc.complaintRepo.CreateComplaint(complaint_by, complaint_on, ComplaintType, text)
	if err != nil {
		uc.logger.Error("CreateComplaint", "complaint_by", complaint_by, "complaint_on", complaint_on, "error", err)
	} else {
		uc.logger.WithFields(&logrus.Fields{"complaint_by": complaint_by, "complaint_on": complaint_on})
	}
	return err
}
