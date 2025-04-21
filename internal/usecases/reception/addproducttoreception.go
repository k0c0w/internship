package reception

import (
	"avito/internal/domain"
	"avito/internal/usecases"
	"context"
	"errors"

	"github.com/google/uuid"
)

type AddProductToCurrentReceptionAtPVZArgs struct {
	usecases.AuthenticationArgs
	domain.ReceptionInfoRepository
	domain.PVZRepository
	domain.ProductRepository

	PVZ AddProductToCurrentReceptionAtPVZDTO
}

type AddProductToCurrentReceptionAtPVZDTO struct {
	PVZID           uuid.UUID
	ProductCategory string
}

func AddProductToCurrentReceptinoAtPVZUseCase(ctx context.Context, args AddProductToCurrentReceptionAtPVZArgs) (domain.Product, error) {
	dto := args.PVZ
	auth := args.AuthenticationArgs
	if _, accessError := auth.ValidatePrivelegies(ctx, domain.ClientUserRoleID); accessError != nil {
		return domain.Product{}, accessError
	}

	pvz, category, argumentsErros := dto.validateArguments(ctx, args.PVZRepository)
	if argumentsErros != nil {
		return domain.Product{}, argumentsErros
	}

	//todo: unit of work here

	reception, err := pvz.CurrentReception(ctx, args.ReceptionInfoRepository)
	if err != nil {
		return domain.Product{}, err
	}

	product, addErr := reception.AddNewProduct(ctx, category, args.ProductRepository)

	return product, addErr
}

func (args *AddProductToCurrentReceptionAtPVZDTO) validateArguments(ctx context.Context, r domain.PVZRepository) (domain.PVZ, domain.ProductCategory, error) {
	receptionPVZID := args.PVZID

	if receptionPVZID == uuid.Nil {
		return domain.PVZ{}, 0, errors.New(usecases.IdIsRequiredArgError)
	}

	pvzId := domain.PVZID(receptionPVZID)
	pvz, err := r.FindById(ctx, pvzId)

	if err != nil {
		return pvz, 0, err
	}

	category, err := toDomainCategory(args.ProductCategory)

	return pvz, category, err
}

func toDomainCategory(productCategory string) (category domain.ProductCategory, err error) {
	switch productCategory {
	case "электроника", "ЭЛЕКТРОНИКА":
		category = domain.ElectronicsProductCategory

	case "одежда", "ОДЕЖДА":
		category = domain.ClothesProductCategory
	case "обувь", "ОБУВЬ":
		category = domain.ShoesProductCategory
	default:
		err = errors.New(domain.UnknownProductCategoryError)
	}

	return
}
