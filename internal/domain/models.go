package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

type PVZ struct {
	ID              PVZID     `json:"id"`
	CreationTimeUTC time.Time `json:"creation_time_utc"`
	City            `json:"city"`
}

func (p *PVZ) CurrentReception(ctx context.Context, receptions ReceptionInfoRepository) (ReceptionInfo, error) {
	filter := SearchReceptionInfoFilter{
		PVZID:                  p.ID,
		DescendingDateOrdering: true,
		Status:                 InProggressProductAcceptanceStatus,
		Limit:                  1,
	}

	receptionList, err := receptions.FindAllByFilter(ctx, filter)

	if err != nil {
		return ReceptionInfo{}, err
	} else if len(receptionList) == 0 {
		return ReceptionInfo{}, errors.New(AllReceptionsAreClosed)
	}

	return receptionList[0], nil
}

func (p *PVZ) CreateNewReception(ctx context.Context, receptions ReceptionInfoRepository) (reception ReceptionInfo, err error) {
	reception, err = newReception(p.ID)
	if err != nil {
		return
	}

	filter := SearchReceptionInfoFilter{
		PVZID:                  p.ID,
		DescendingDateOrdering: true,
		Status:                 InProggressProductAcceptanceStatus,
		Limit:                  1,
	}

	receptionList, err := receptions.FindAllByFilter(ctx, filter)
	if err != nil {
		return
	} else if len(receptionList) > 0 {
		err = errors.New(AnotherOpenedReceptionError)
		return
	}

	err = receptions.Add(ctx, reception)
	return
}

const (
	ElectronicsProductCategory ProductCategory = 1
	ClothesProductCategory     ProductCategory = 2
	ShoesProductCategory       ProductCategory = 3
)

type Product struct {
	ID              ProductID       `json:"id"`
	ReceptionID     ReceptionID     `json:"reception_id"`
	CreationTimeUTC time.Time       `json:"creation_time_utc"`
	Category        ProductCategory `json:"category"`
}

func newProduct(parentReceptionId ReceptionID, category ProductCategory) (product Product, err error) {
	id, err := uuid.NewV7()
	if err != nil {
		return
	}

	product = Product{
		ID:              id,
		ReceptionID:     parentReceptionId,
		CreationTimeUTC: time.Now().UTC(),
		Category:        category,
	}
	return
}

const (
	InProggressProductAcceptanceStatus ReceptionStatus = 1
	CloseProductAcceptanceStatus       ReceptionStatus = 2
)

type ReceptionInfo struct {
	ID              ReceptionID     `json:"id"`
	PVZID           PVZID           `json:"pvz_id"`
	CreationTimeUTC time.Time       `json:"creation_time_utc"`
	Status          ReceptionStatus `json:"status"`
}

func newReception(pvzId PVZID) (reception ReceptionInfo, err error) {
	id, err := uuid.NewV7()
	if err != nil {
		return
	} else if pvzId == uuid.Nil {
		err = errors.New(InvalidIdStateError)
		return
	}
	reception = ReceptionInfo{
		ID:              id,
		PVZID:           pvzId,
		CreationTimeUTC: time.Now().UTC(),
		Status:          InProggressProductAcceptanceStatus,
	}

	return
}

func (r *ReceptionInfo) IsCompleted() bool {
	return r.Status == CloseProductAcceptanceStatus
}

func (r *ReceptionInfo) Close() error {
	if r.Status == CloseProductAcceptanceStatus {
		return errors.New(ReceptionIsAlreadyClosedError)
	}

	r.Status = CloseProductAcceptanceStatus

	return nil
}

func (r *ReceptionInfo) AddNewProduct(ctx context.Context, category ProductCategory, products ProductRepository) (product Product, err error) {
	if r.IsCompleted() {
		return Product{}, errors.New(ReceptionIsAlreadyClosedError)
	}

	product, err = newProduct(r.ID, category)
	if err != nil {
		return
	}

	err = products.Add(ctx, product)
	return
}

func (r *ReceptionInfo) RemoveLastProduct(ctx context.Context, products ProductRepository) (removedProduct Product, err error) {
	if r.IsCompleted() {
		return Product{}, errors.New(ReceptionIsAlreadyClosedError)
	}

	receptionProducts, err := products.FindAllByReceptionID(ctx, r.ID)

	if err != nil {
		return Product{}, err
	}

	if receptionProducts == nil || len(receptionProducts) == 0 {
		return Product{}, errors.New(ReceptionIsEmptyError)
	}

	lastAddedProductIndex := 0
	for i, product := range receptionProducts {
		if product.CreationTimeUTC.Compare(receptionProducts[lastAddedProductIndex].CreationTimeUTC) > 0 {
			lastAddedProductIndex = i
		}
	}

	removedProduct = *receptionProducts[lastAddedProductIndex]

	err = products.Remove(ctx, removedProduct)

	if err != nil {
		return Product{}, err
	}

	return removedProduct, nil
}

type User struct {
	ID       UserID
	Email    Email
	Password string
	UserRole
}

// new user is always a client
func NewUser(email Email, password string) (User, error) {
	userId, err := uuid.NewV7()

	if err != nil {
		return User{}, err
	}

	return User{
		ID:       userId,
		Email:    email,
		Password: password,
		UserRole: clientRole(),
	}, nil
}

func GrantModeratorRole(u *User) {
	u.UserRole = moderatorRole()
}

const (
	ClientUserRoleID    UserRoleID = 1
	ModeratorUserRoleID UserRoleID = 2
)

const (
	ClientUserRoleName    string = "client"
	ModeratorUserRoleName string = "moderator"
)

type UserRole struct {
	ID   UserRoleID
	Name string
}

func clientRole() UserRole {
	return UserRole{
		ID:   ClientUserRoleID,
		Name: ClientUserRoleName,
	}
}

func moderatorRole() UserRole {
	return UserRole{
		ID:   ModeratorUserRoleID,
		Name: ModeratorUserRoleName,
	}
}

const (
	KazanCityID       CityID = 1
	MoscowCityID      CityID = 2
	SaintPetersburgID CityID = 3
)

type City struct {
	ID   CityID `json:"id"`
	Name string `json:"name"`
}

func NewCity(cityID CityID) (City, error) {
	switch cityID {
	case KazanCityID:
		return City{
			ID:   KazanCityID,
			Name: "Казань",
		}, nil
	case MoscowCityID:
		return City{
			ID:   MoscowCityID,
			Name: "Москва",
		}, nil
	case SaintPetersburgID:
		return City{
			ID:   SaintPetersburgID,
			Name: "Caнкт-Петербург",
		}, nil
	default:
		return City{}, errors.New(UnknownCityError)
	}
}
