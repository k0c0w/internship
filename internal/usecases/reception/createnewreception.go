package reception

import (
	"avito/internal/domain"
	"avito/internal/usecases"
	"context"
	"errors"

	"github.com/google/uuid"
)

type CreateNewReceptionArgs struct {
	usecases.AuthenticationArgs
	domain.ReceptionInfoRepository
	domain.PVZRepository

	PVZ CreateNewReceptionAtPVZDTO
}

type CreateNewReceptionAtPVZDTO struct {
	PVZID uuid.UUID
}

func CreateNewReceptionUseCase(ctx context.Context, args CreateNewReceptionArgs) (domain.ReceptionInfo, error) {
	auth := args.AuthenticationArgs
	dto := args.PVZ
	if _, accessErr := auth.ValidatePrivelegies(ctx, domain.ClientUserRoleID); accessErr != nil {
		return domain.ReceptionInfo{}, accessErr
	}

	pvz, err := dto.validateArguments(ctx, args.PVZRepository)
	if err != nil {
		return domain.ReceptionInfo{}, err
	}

	//todo: unit of work here
	reception, err := pvz.CreateNewReception(ctx, args.ReceptionInfoRepository)

	return reception, err
}

func (args *CreateNewReceptionAtPVZDTO) validateArguments(ctx context.Context, p domain.PVZRepository) (domain.PVZ, error) {
	receptionPVZID := args.PVZID

	if receptionPVZID == uuid.Nil {
		return domain.PVZ{}, errors.New(usecases.IdIsRequiredArgError)
	}

	pvzId := domain.PVZID(receptionPVZID)
	pvz, err := p.FindById(ctx, pvzId)

	return pvz, err
}
