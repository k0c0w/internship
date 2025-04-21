package domain_test

import (
	"avito/internal/domain"
	"context"
	"math"
	"math/rand"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

var random *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))
var ctx context.Context = context.TODO()

func TestGrantModeratorRole_ShouldGrantModeratorRole(t *testing.T) {
	user := domain.User{
		ID:       uuid.MustParse("0196587b-1ca2-7ee7-83ef-9c37c806c968"),
		Email:    "test@test.test",
		Password: "sdf",
		UserRole: domain.UserRole{
			ID:   domain.ClientUserRoleID,
			Name: domain.ClientUserRoleName,
		},
	}

	//act
	domain.GrantModeratorRole(&user)

	require.Equal(t, domain.ModeratorUserRoleID, user.UserRole.ID)
	require.Equal(t, domain.ModeratorUserRoleName, user.UserRole.Name)
}

func TestNewUser(t *testing.T) {
	email := "example.com@example.com"
	password := email
	initialRole := domain.UserRole{
		ID:   domain.ClientUserRoleID,
		Name: domain.ClientUserRoleName,
	}

	user, err := domain.NewUser(email, password)

	require.NoError(t, err)
	require.Equal(t, email, user.Email)
	require.Equal(t, password, user.Password)
	require.Equal(t, initialRole, user.UserRole)
}

func TestPVZCreateNewReception_ShouldCreateReception(t *testing.T) {
	pvz := getPVZ(t)
	receptionRepositoryFake := NewFakeReceptionInfoRepository(t)
	timeBeforeRun := time.Now().UTC()

	// act
	reception, err := pvz.CreateNewReception(ctx, receptionRepositoryFake)

	// assert
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, reception.ID)
	require.NotEqual(t, uuid.Nil, reception.PVZID)
	require.LessOrEqual(t, timeBeforeRun, reception.CreationTimeUTC)
	require.Equal(t, domain.InProggressProductAcceptanceStatus, reception.Status)
	addedReception, exists := receptionRepositoryFake.Receptions[reception.ID]

	require.True(t, exists)
	require.Equal(t, reception, addedReception)
}

func TestPVZCreateNewReception_ShouldReturnError_WhenOtherReceptionIsOpened(t *testing.T) {
	pvz := getPVZ(t)
	receptionRepositoryFake := NewFakeReceptionInfoRepository(t)
	recption := getValidReception(t, pvz, domain.InProggressProductAcceptanceStatus)
	receptionRepositoryFake.Receptions[recption.ID] = recption
	expectedErrorMsg := domain.AnotherOpenedReceptionError

	// act
	_, err := pvz.CreateNewReception(ctx, receptionRepositoryFake)

	// assert
	require.Error(t, err)
	require.Equal(t, expectedErrorMsg, err.Error())
}

func TestPVZCurrentReception_ShouldReturnLastOpenedReception(t *testing.T) {
	pvz := getPVZ(t)
	receptionRepositoryFake := NewFakeReceptionInfoRepository(t)
	reception := getValidReception(t, pvz, domain.InProggressProductAcceptanceStatus)
	receptionRepositoryFake.Receptions[reception.ID] = reception

	//act
	currentReception, err := pvz.CurrentReception(ctx, receptionRepositoryFake)

	require.NoError(t, err)
	require.Equal(t, reception, currentReception)
}

func TestPVZCurrentReception_ShouldReturnError_WhenNoReceptions(t *testing.T) {
	pvz := getPVZ(t)
	receptionRepositoryFake := NewFakeReceptionInfoRepository(t)
	expectedErrorMsg := domain.AllReceptionsAreClosed

	//act
	_, err := pvz.CurrentReception(ctx, receptionRepositoryFake)

	require.Error(t, err)
	require.Equal(t, expectedErrorMsg, err.Error())
}

func TestReceptionInfoIsCompleted_ShouldReturnCorrectValue(t *testing.T) {
	id := uuid.MustParse("351182e5-54af-4b7a-b45e-5fbd186f8503")
	time := time.Now()

	testCases := []struct {
		name          string
		reception     domain.ReceptionInfo
		expectedValue bool
	}{
		{
			name: "false on in_progress",
			reception: domain.ReceptionInfo{
				ID:              id,
				PVZID:           id,
				CreationTimeUTC: time,
				Status:          domain.InProggressProductAcceptanceStatus,
			},
			expectedValue: false,
		},
		{
			name: "true on close",
			reception: domain.ReceptionInfo{
				ID:              id,
				PVZID:           id,
				CreationTimeUTC: time,
				Status:          domain.CloseProductAcceptanceStatus,
			},
			expectedValue: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.reception.IsCompleted()
			require.Equal(t, tc.expectedValue, result)
		})
	}
}

func TestReceptionInfoClose_ShouldChangeReceptionStatus(t *testing.T) {
	id := uuid.MustParse("351182e5-54af-4b7a-b45e-5fbd186f8503")
	reception := domain.ReceptionInfo{
		ID:              id,
		PVZID:           id,
		CreationTimeUTC: time.Now(),
		Status:          domain.InProggressProductAcceptanceStatus,
	}

	err := reception.Close()

	require.NoError(t, err)
	require.Equal(t, domain.CloseProductAcceptanceStatus, reception.Status)
}

func TestReceptionInfoClose_ShouldReturnError_WhenAlreadyClosed(t *testing.T) {
	id := uuid.MustParse("351182e5-54af-4b7a-b45e-5fbd186f8503")
	reception := domain.ReceptionInfo{
		ID:              id,
		PVZID:           id,
		CreationTimeUTC: time.Now(),
		Status:          domain.CloseProductAcceptanceStatus,
	}

	err := reception.Close()

	require.Error(t, err)
	require.Equal(t, domain.ReceptionIsAlreadyClosedError, err.Error())
}

func TestReceptionInfoAddNewProduct_ShouldCreateAndAddProduct(t *testing.T) {
	id := uuid.MustParse("351182e5-54af-4b7a-b45e-5fbd186f8503")
	reception := domain.ReceptionInfo{
		ID:              id,
		PVZID:           id,
		CreationTimeUTC: time.Now(),
		Status:          domain.InProggressProductAcceptanceStatus,
	}
	expectedProductCategory := domain.ElectronicsProductCategory
	productRepositoryFake := NewFakeProductRepository(t)
	timeBeforeRun := time.Now().UTC()

	// act
	addedProduct, err := reception.AddNewProduct(ctx, expectedProductCategory, productRepositoryFake)

	// assert
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, addedProduct.ID)
	require.NotEqual(t, uuid.Nil, addedProduct.ReceptionID)
	require.LessOrEqual(t, timeBeforeRun, addedProduct.CreationTimeUTC)
	require.Equal(t, expectedProductCategory, addedProduct.Category)

	storageProduct, exists := productRepositoryFake.Products[addedProduct.ID]
	require.True(t, exists)
	require.Equal(t, addedProduct, storageProduct)
}

func TestReceptionInfoAddNewProduct_ShouldReturnError_WhenClosed(t *testing.T) {
	id := uuid.MustParse("351182e5-54af-4b7a-b45e-5fbd186f8503")
	reception := domain.ReceptionInfo{
		ID:              id,
		PVZID:           id,
		CreationTimeUTC: time.Now(),
		Status:          domain.CloseProductAcceptanceStatus,
	}
	expectedProductCategory := domain.ElectronicsProductCategory
	productRepositoryFake := NewFakeProductRepository(t)
	expectedErrorMsg := domain.ReceptionIsAlreadyClosedError

	// act
	_, err := reception.AddNewProduct(ctx, expectedProductCategory, productRepositoryFake)

	// assert
	require.Error(t, err)
	require.Equal(t, expectedErrorMsg, err.Error())

	require.Equal(t, 0, len(productRepositoryFake.Products))
}

func TestReceptionInfoRemoveLastProduct_ShouldRemoveLastAddedProduct(t *testing.T) {
	id := uuid.MustParse("351182e5-54af-4b7a-b45e-5fbd186f8503")
	reception := domain.ReceptionInfo{
		ID:              id,
		PVZID:           id,
		CreationTimeUTC: time.Now(),
		Status:          domain.InProggressProductAcceptanceStatus,
	}
	product1Category := domain.ElectronicsProductCategory
	product2Category := domain.ShoesProductCategory
	productRepositoryFake := NewFakeProductRepository(t)
	var product1, product2 domain.Product

	// act
	product1, _ = reception.AddNewProduct(ctx, product1Category, productRepositoryFake)
	time.Sleep(time.Duration(1 * time.Second))
	product2, _ = reception.AddNewProduct(ctx, product2Category, productRepositoryFake)
	removedProduct, err := reception.RemoveLastProduct(ctx, productRepositoryFake)

	// assert
	require.NoError(t, err)
	require.Equal(t, product2.Category, removedProduct.Category)

	storageProduct, exists := productRepositoryFake.Products[product1.ID]
	require.True(t, exists)
	require.Equal(t, product1.Category, storageProduct.Category)
}

func TestReceptionInfoRemoveLastProduct_ShouldReturnError_WhenReceptionIsClosed(t *testing.T) {
	id := uuid.MustParse("351182e5-54af-4b7a-b45e-5fbd186f8503")
	reception := domain.ReceptionInfo{
		ID:              id,
		PVZID:           id,
		CreationTimeUTC: time.Now(),
		Status:          domain.InProggressProductAcceptanceStatus,
	}
	productCategory := domain.ElectronicsProductCategory
	productRepositoryFake := NewFakeProductRepository(t)
	expectedErrorMsg := domain.ReceptionIsAlreadyClosedError

	// act
	_, _ = reception.AddNewProduct(ctx, productCategory, productRepositoryFake)
	reception.Close()
	_, err := reception.RemoveLastProduct(ctx, productRepositoryFake)

	// assert
	require.Error(t, err)
	require.Equal(t, expectedErrorMsg, err.Error())
}

func TestNewCity_ShouldCreateCity(t *testing.T) {
	testCases := []struct {
		name          string
		cityId        domain.CityID
		expectedValue domain.City
	}{
		{
			name:   "Казань",
			cityId: domain.KazanCityID,
			expectedValue: domain.City{
				ID:   domain.KazanCityID,
				Name: "Казань",
			},
		},
		{
			name:   "Москва",
			cityId: domain.MoscowCityID,
			expectedValue: domain.City{
				ID:   domain.MoscowCityID,
				Name: "Москва",
			},
		},
		{
			name:   "Caнкт-Петербург",
			cityId: domain.SaintPetersburgID,
			expectedValue: domain.City{
				ID:   domain.SaintPetersburgID,
				Name: "Caнкт-Петербург",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			city, err := domain.NewCity(tc.cityId)

			require.NoError(t, err)
			require.Equal(t, tc.expectedValue, city)
		})
	}
}

func TestNewCity_ShouldReturnError_WhenUnknownCity(t *testing.T) {
	expectedErrorMsg := domain.UnknownCityError
	var wrongCityId domain.CityID = math.MinInt16

	_, err := domain.NewCity(wrongCityId)

	require.Error(t, err)
	require.Equal(t, expectedErrorMsg, err.Error())
}

func getValidReception(t *testing.T, pvz domain.PVZ, status domain.ReceptionStatus) domain.ReceptionInfo {
	t.Helper()

	start := time.Now().UTC()
	end := pvz.CreationTimeUTC

	if end.Before(start) {
		start, end = end, start
	}

	duration := end.Sub(start)

	randomDuration := time.Duration(random.Int63n(int64(duration)))

	reception := domain.ReceptionInfo{
		ID:              uuid.Must(uuid.NewV7()),
		PVZID:           pvz.ID,
		Status:          status,
		CreationTimeUTC: start.Add(randomDuration),
	}

	return reception
}

func getPVZ(t *testing.T) domain.PVZ {
	t.Helper()

	time, _ := time.Parse("2006-01-02T15:04:05.000Z", "2020-05-05T09:00:00.000Z")

	return domain.PVZ{
		ID:              uuid.MustParse("0196587b-1ca2-7ee7-83ef-9c37c806c968"),
		CreationTimeUTC: time,
		City:            domain.City{},
	}
}

type FakeReceptionInfoRepository struct {
	Receptions map[domain.ReceptionID]domain.ReceptionInfo
}

func NewFakeReceptionInfoRepository(t *testing.T) *FakeReceptionInfoRepository {
	t.Helper()
	return &FakeReceptionInfoRepository{
		Receptions: make(map[domain.ReceptionID]domain.ReceptionInfo),
	}
}

func (r *FakeReceptionInfoRepository) FindAllByFilter(ctx context.Context, filter domain.SearchReceptionInfoFilter) ([]domain.ReceptionInfo, error) {
	var results []domain.ReceptionInfo

	for _, reception := range r.Receptions {
		if reception.PVZID == filter.PVZID && (filter.Status == 0 || reception.Status == filter.Status) {
			results = append(results, reception)
		}
	}

	if filter.DescendingDateOrdering {
		for i := 0; i < len(results)-1; i++ {
			for j := i + 1; j < len(results); j++ {
				if results[i].CreationTimeUTC.Before(results[j].CreationTimeUTC) {
					results[i], results[j] = results[j], results[i]
				}
			}
		}
	} else {
		for i := 0; i < len(results)-1; i++ {
			for j := i + 1; j < len(results); j++ {
				if results[i].CreationTimeUTC.After(results[j].CreationTimeUTC) {
					results[i], results[j] = results[j], results[i]
				}
			}
		}
	}

	if filter.Limit > 0 && len(results) > filter.Limit {
		results = results[:filter.Limit]
	}

	return results, nil
}

func (r *FakeReceptionInfoRepository) Update(ctx context.Context, reception domain.ReceptionInfo) error {
	r.Receptions[reception.ID] = reception
	return nil
}

func (r *FakeReceptionInfoRepository) Add(ctx context.Context, reception domain.ReceptionInfo) error {
	r.Receptions[reception.ID] = reception
	return nil
}

type FakeProductRepository struct {
	Products map[domain.ReceptionID]domain.Product
}

func NewFakeProductRepository(t *testing.T) *FakeProductRepository {
	t.Helper()
	return &FakeProductRepository{
		Products: make(map[domain.ReceptionID]domain.Product),
	}
}

func (r *FakeProductRepository) Add(ctx context.Context, product domain.Product) error {
	r.Products[product.ID] = product
	return nil
}

func (r *FakeProductRepository) FindAllByReceptionID(ctx context.Context, receptionId domain.ReceptionID) ([]*domain.Product, error) {
	result := make([]*domain.Product, 0)
	for _, p := range r.Products {
		if p.ReceptionID == receptionId {
			result = append(result, &p)
		}
	}

	return result, nil
}

func (r *FakeProductRepository) Remove(ctx context.Context, product domain.Product) error {
	delete(r.Products, product.ID)

	return nil
}
