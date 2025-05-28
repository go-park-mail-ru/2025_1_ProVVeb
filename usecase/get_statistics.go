package usecase

import (
	"time"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/logger"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/repository"
	"github.com/sirupsen/logrus"
)

type GetStatisticsCompl struct {
	complaintRepo repository.ComplaintRepository
	logger        *logger.LogrusLogger
}

func NewGetStatisticsComplUseCase(
	complaintRepo repository.ComplaintRepository,
	logger *logger.LogrusLogger,
) (*GetStatisticsCompl, error) {
	if complaintRepo == nil || logger == nil {
		return nil, model.ErrGetActiveQueriesUC
	}
	return &GetStatisticsCompl{complaintRepo: complaintRepo, logger: logger}, nil
}

func (g *GetStatisticsCompl) GetStatistics(useFrom bool, from time.Time, useTo bool, to time.Time) (model.ComplaintStats, error) {
	g.logger.Info("GetStatistics")

	Stats, err := g.complaintRepo.GetStatistics(useFrom, from, useTo, to)
	if err != nil {
		g.logger.Error("GetStatistics", "Stats", Stats, "error", err)
	} else {
		g.logger.WithFields(&logrus.Fields{"Stats": Stats})
	}

	return Stats, nil
}
