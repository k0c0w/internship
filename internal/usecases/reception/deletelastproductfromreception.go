package reception

import (
	"avito/internal/domain"
	"avito/internal/usecases"
	"context"
	"errors"

	"github.com/google/uuid"
)

type DeleteLastProductFromCurrentReceptionAtPVZArgs struct {
	usecases.AuthenticationArgs
	domain.ReceptionInfoRepository
	domain.PVZRepository
	domain.ProductRepository

	PVZ DeleteLastProductFromCurrentReceptionAtPVZDTO
}

type DeleteLastProductFromCurrentReceptionAtPVZDTO struct {
	PVZID uuid.UUID
}

func DeleteLastProductFromCurrentReceptionAtPVZUseCase(ctx context.Context, args DeleteLastProductFromCurrentReceptionAtPVZArgs) error {
	dto := args.PVZ
	auth := args.AuthenticationArgs
	if _, accessError := auth.ValidatePrivelegies(ctx, domain.ClientUserRoleID); accessError != nil {
		return accessError
	}

	pvz, argumentsErros := dto.validateArguments(ctx, args.PVZRepository)
	if argumentsErros != nil {
		return argumentsErros
	}

	//todo: unit of work here
	reception, err := pvz.CurrentReception(ctx, args.ReceptionInfoRepository)
	if err != nil {
		return err
	}

	_, err = reception.RemoveLastProduct(ctx, args.ProductRepository)

	return err
}

func (args *DeleteLastProductFromCurrentReceptionAtPVZDTO) validateArguments(ctx context.Context, r domain.PVZRepository) (domain.PVZ, error) {
	receptionPVZID := args.PVZID

	if receptionPVZID == uuid.Nil {
		return domain.PVZ{}, errors.New(usecases.IdIsRequiredArgError)
	}

	pvzId := domain.PVZID(receptionPVZID)
	pvz, err := r.FindById(ctx, pvzId)

	if err != nil {
		return pvz, err
	}

	return pvz, err
}
