package domain

import (
	"context"
	"time"
)

const (
	UserDoesNotExistsError string = "user does not exists"
)

type UserRepository interface {
	FindByID(ctx context.Context, id UserID) (User, error)
	FindByEmail(ctx context.Context, email Email) (User, error)
	Add(ctx context.Context, user User) error
}

type (
	PVZRepository interface {
		Add(ctx context.Context, pvz PVZ) error
		FindById(ctx context.Context, id PVZID) (PVZ, error)
	}
)

type (
	PVZReportAggregateRepository interface {
		FindAllByFilter(ctx context.Context, filter SearchPVZReportAggregateFilter) ([]*PVZReportAggregate, error)
	}
	SearchPVZReportAggregateFilter struct {
		Page                  int
		Limit                 int
		ReceptionStartTimeUTC *time.Time
		ReceptionEndTimeUTC   *time.Time
	}
)

const (
	ReceptionDoesNotExistsError string = "reception does not exists"
)

type (
	ReceptionInfoRepository interface {
		FindAllByFilter(ctx context.Context, filter SearchReceptionInfoFilter) ([]ReceptionInfo, error)
		Update(ctx context.Context, reception ReceptionInfo) error
		Add(ctx context.Context, reception ReceptionInfo) error
	}

	SearchReceptionInfoFilter struct {
		PVZID                  PVZID
		DescendingDateOrdering bool
		Status                 ReceptionStatus
		Limit                  int
	}
)

type (
	ProductRepository interface {
		Add(ctx context.Context, product Product) error
		FindAllByReceptionID(ctx context.Context, receptionId ReceptionID) ([]*Product, error)
		Remove(ctx context.Context, product Product) error
	}
)
