package reception

import (
	"avito/internal/domain"
	"avito/internal/usecases"
	"context"
	"errors"

	"github.com/google/uuid"
)

type CloseLastOpenedReceptionAtPVZArgs struct {
	usecases.AuthenticationArgs
	domain.ReceptionInfoRepository
	domain.PVZRepository

	PVZ CloseLastOpenedReceptionAtPVZDTO
}

type CloseLastOpenedReceptionAtPVZDTO struct {
	PVZID uuid.UUID
}

func CloseLastOpenedReceptionUseCase(ctx context.Context, args CloseLastOpenedReceptionAtPVZArgs) (domain.ReceptionInfo, error) {
	createAtPVZID := args.PVZ.PVZID
	auth := args.AuthenticationArgs

	if _, accessErr := auth.ValidatePrivelegies(ctx); accessErr != nil {
		return domain.ReceptionInfo{}, accessErr
	} else if createAtPVZID == uuid.Nil {
		return domain.ReceptionInfo{}, errors.New(usecases.IdIsRequiredArgError)
	}

	pvz, err := args.PVZRepository.FindById(ctx, domain.PVZID(createAtPVZID))
	if err != nil {
		return domain.ReceptionInfo{}, err
	}

	reception, err := pvz.CurrentReception(ctx, args.ReceptionInfoRepository)
	if err != nil {
		return reception, err
	}

	err = reception.Close()
	if err != nil {
		return reception, err
	}

	err = args.ReceptionInfoRepository.Update(ctx, reception)

	return reception, err
}
