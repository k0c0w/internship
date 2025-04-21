package storage

import (
	"avito/internal/domain"
	postgresql "avito/pkg/database"
)

type Repositories struct {
	domain.UserRepository
	domain.PVZRepository
	domain.PVZReportAggregateRepository
	domain.ReceptionInfoRepository
	domain.ProductRepository
}

func NewRepositories(client postgresql.Client) Repositories {
	return Repositories{
		UserRepository:               NewUserRepository(client),
		PVZRepository:                NewPVZRepository(client),
		PVZReportAggregateRepository: NewPVZReportAggregateRepository(client),
		ReceptionInfoRepository:      NewReceptionInfoRepository(client),
		ProductRepository:            NewProductRepository(client),
	}
}
