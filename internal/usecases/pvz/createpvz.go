package pvz

import (
	"avito/internal/domain"
	"avito/internal/usecases"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

const CouldNotCreatePVZError string = "an error while PVZ creation"

type CreatePVZUseCaseArgs struct {
	usecases.AuthenticationArgs
	domain.PVZRepository

	PVZ CreatePVZDTO
}

type CreatePVZDTO struct {
	PVZID            *uuid.UUID
	PVZCity          string
	RegistrationTime *time.Time
}

func CreatePVZUseCase(ctx context.Context, args CreatePVZUseCaseArgs) (domain.PVZ, error) {
	createPVZDTO := args.PVZ
	auth := args.AuthenticationArgs

	if _, accessError := auth.ValidatePrivelegies(ctx, domain.ModeratorUserRoleID); accessError != nil {
		return domain.PVZ{}, accessError
	}

	location, err := toDomainCity(createPVZDTO.PVZCity)
	if err != nil {
		return domain.PVZ{}, err
	}

	if createPVZDTO.RegistrationTime == nil {
		return domain.PVZ{}, errors.New("registration time is required for creation")
	}

	if createPVZDTO.PVZID == nil || *createPVZDTO.PVZID == uuid.Nil {
		return domain.PVZ{}, errors.New(usecases.IdIsRequiredArgError)
	}

	// pvz, err := domain.NewPVZ(location)
	pvz := domain.PVZ{
		ID:              *createPVZDTO.PVZID,
		CreationTimeUTC: *createPVZDTO.RegistrationTime,
		City:            location,
	}

	// todo add unit of work here
	err = args.PVZRepository.Add(ctx, pvz)

	return pvz, err
}

func toDomainCity(cityName string) (city domain.City, err error) {
	switch cityName {
	case "Казань", "казань", "КАЗАНЬ":
		city, err = domain.NewCity(domain.KazanCityID)

	case "Москва", "москва", "МОСКВА":
		city, err = domain.NewCity(domain.MoscowCityID)
	case "Санкт-Петербург", "санкт-петербург", "САНКТ-ПЕТЕРБУРГ":
		city, err = domain.NewCity(domain.SaintPetersburgID)
	default:
		err = errors.New(domain.UnknownCityError)
	}

	return
}
