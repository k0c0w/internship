package pvz

import (
	"avito/internal/domain"
	"avito/internal/usecases"
	"context"
	"time"
)

type (
	GetPVZListReportsDTO struct {
		ReceptionStartTimeUTC *time.Time
		ReceptionEndTimeUTC   *time.Time
		Page                  int
		Limit                 int
	}

	GetPVZListUseCaseArgs struct {
		GetPVZListReportsDTO
		usecases.AuthenticationArgs
		domain.PVZRepository
		domain.PVZReportAggregateRepository
	}
)

func GetPVZListReportsUseCase(ctx context.Context, args GetPVZListUseCaseArgs) ([]*domain.PVZReportAggregate, error) {
	auth := args.AuthenticationArgs
	dto := args.GetPVZListReportsDTO
	if _, accessError := auth.ValidatePrivelegies(ctx); accessError != nil {
		return nil, accessError
	}

	dto.fixArgsIfNeeded()
	filter := domain.SearchPVZReportAggregateFilter{
		Page:                  dto.Page,
		Limit:                 dto.Limit,
		ReceptionStartTimeUTC: dto.ReceptionStartTimeUTC,
		ReceptionEndTimeUTC:   dto.ReceptionEndTimeUTC,
	}

	reports, err := args.PVZReportAggregateRepository.FindAllByFilter(ctx, filter)
	if err != nil {
		return nil, err
	}

	return reports, err
}

func (args *GetPVZListReportsDTO) fixArgsIfNeeded() {
	if args.ReceptionEndTimeUTC != nil && args.ReceptionStartTimeUTC != nil && args.ReceptionEndTimeUTC.Compare(*args.ReceptionStartTimeUTC) <= 0 {
		args.ReceptionEndTimeUTC = nil
	}

	if args.Limit <= 0 {
		args.Limit = 10
	}

	if args.Page < 1 {
		args.Page = 1
	}
}
