package http

import (
	"avito/internal/config"
	"avito/internal/domain"
	"avito/internal/storage"
	pvz "avito/internal/usecases/pvz"
	"avito/internal/usecases/reception"
	"avito/internal/usecases/users"
	jwt "avito/pkg/authorization"
	"context"
	"embed"
	"fmt"
	netHttp "net/http"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/labstack/echo/v4"
)

//go:embed swagger.html
//go:embed openapi.codegen_input.yaml
var swaggerUI embed.FS

type (
	Dependencies struct {
		domain.AuthorizationService[jwt.JWT]
		storage.Repositories
		jwt.JWTManager
	}
	server struct {
		e      *echo.Echo
		config config.HTTPConfig
	}
	httpRequestHandlers struct {
		deps Dependencies
	}
)

func NewHTTPServer(dependencies Dependencies, config config.HTTPConfig) *server {
	e := echo.New()
	e.Use(BearerTokenMiddleware())
	RegisterHandlers(e, NewStrictHandler(
		httpRequestHandlers{deps: dependencies},
		[]StrictMiddlewareFunc{},
	))

	if config.IncludeSwagger {
		e.GET("/swagger/*", echo.WrapHandler(netHttp.StripPrefix("/swagger/", netHttp.FileServer(netHttp.FS(swaggerUI)))))
	}

	return &server{
		e:      e,
		config: config,
	}
}

func (s *server) Start() error {
	cfg := s.config

	listenerURL := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	s.e.Logger.Info(fmt.Sprintf("Starting server on %s", listenerURL))
	return s.e.Start(listenerURL)
}

func (s *server) Stop(ctx context.Context) error {
	return s.e.Shutdown(ctx)
}

func (h httpRequestHandlers) PostDummyLogin(ctx context.Context, request PostDummyLoginRequestObject) (PostDummyLoginResponseObject, error) {
	var (
		dummyUserId string
	)

	switch request.Body.Role {
	case PostDummyLoginJSONBodyRole(UserRoleModerator):
		dummyUserId = "0196521c-b2a9-7a04-88be-16ec981d104b"
	case PostDummyLoginJSONBodyRole(UserRoleEmployee):
		dummyUserId = "0196521c-873e-77fd-b244-bdd0c13c72ab"
	default:
		return PostDummyLogin400JSONResponse{
			Message: fmt.Sprintf("unknown role: %s", request.Body.Role),
		}, nil
	}

	// bypass authorization
	jwt, err := h.deps.JWTManager.GenerateToken(dummyUserId)
	if err != nil {
		return nil, err
	}

	return PostDummyLogin200JSONResponse(jwt), nil
}

func (h httpRequestHandlers) PostLogin(ctx context.Context, request PostLoginRequestObject) (PostLoginResponseObject, error) {
	args := users.LoginUserUseCaseArgs{
		AuthorizationService: h.deps.AuthorizationService,
		User: users.LoginUserDTO{
			Email:    string(request.Body.Email),
			Password: request.Body.Password,
		},
	}

	token, err := users.LoginUserUseCase(ctx, args)

	if err != nil {
		return PostLogin401JSONResponse{
			Message: domain.BadUserCredentialError,
		}, nil
	}

	return PostLogin200JSONResponse(token), nil
}

func (h httpRequestHandlers) PostProducts(ctx context.Context, request PostProductsRequestObject) (PostProductsResponseObject, error) {
	args := reception.AddProductToCurrentReceptionAtPVZArgs{
		AuthenticationArgs:      h.authArgs(ctx),
		ReceptionInfoRepository: h.deps.ReceptionInfoRepository,
		PVZRepository:           h.deps.PVZRepository,
		ProductRepository:       h.deps.ProductRepository,
		PVZ: reception.AddProductToCurrentReceptionAtPVZDTO{
			PVZID:           request.Body.PvzId,
			ProductCategory: string(request.Body.Type),
		},
	}

	product, err := reception.AddProductToCurrentReceptinoAtPVZUseCase(ctx, args)

	if err != nil {
		if domain.IsAccessError(err) {
			return PostProducts403JSONResponse{
				Message: err.Error(),
			}, nil
		}

		return PostProducts400JSONResponse{
			Message: err.Error(),
		}, nil
	}

	return PostProducts201JSONResponse{
		DateTime:    &product.CreationTimeUTC,
		Id:          &product.ID,
		ReceptionId: product.ReceptionID,
		Type:        productType(product.Category),
	}, nil
}

func (h httpRequestHandlers) GetPvz(ctx context.Context, request GetPvzRequestObject) (GetPvzResponseObject, error) {
	params := request.Params

	var (
		page  int
		limit int
	)

	if params.Page == nil {
		page = 1
	} else {
		page = *params.Page
	}

	if params.Limit == nil {
		limit = 10
	} else {
		limit = *params.Limit
	}

	args := pvz.GetPVZListUseCaseArgs{
		GetPVZListReportsDTO: pvz.GetPVZListReportsDTO{
			ReceptionStartTimeUTC: params.StartDate,
			ReceptionEndTimeUTC:   params.EndDate,
			Page:                  page,
			Limit:                 limit,
		},
		AuthenticationArgs:           h.authArgs(ctx),
		PVZRepository:                h.deps.PVZRepository,
		PVZReportAggregateRepository: h.deps.PVZReportAggregateRepository,
	}

	reports, err := pvz.GetPVZListReportsUseCase(ctx, args)

	if err != nil {
		return nil, err
	}

	response := make(GetPvz200JSONResponse, len(reports))

	for i, report := range reports {
		receptions := make([]struct {
			Products  *[]Product "json:\"products,omitempty\""
			Reception *Reception "json:\"reception,omitempty\""
		}, len(report.Receptions))
		for j, reception := range report.Receptions {
			products := make([]Product, len(reception.Products))
			for k, product := range reception.Products {
				products[k] = Product{
					DateTime:    &product.CreationTimeUTC,
					Id:          &product.ID,
					ReceptionId: reception.Information.ID,
					Type:        productType(product.Category),
				}
			}
			receptions[j] = struct {
				Products  *[]Product "json:\"products,omitempty\""
				Reception *Reception "json:\"reception,omitempty\""
			}{
				Products: &products,
				Reception: &Reception{
					Id:       &reception.Information.ID,
					DateTime: reception.Information.CreationTimeUTC,
					PvzId:    reception.Information.PVZID,
					Status:   receptionStatus(reception.Information.Status),
				},
			}
		}

		response[i] = struct {
			Pvz        *PVZ "json:\"pvz,omitempty\""
			Receptions *[]struct {
				Products  *[]Product "json:\"products,omitempty\""
				Reception *Reception "json:\"reception,omitempty\""
			} "json:\"receptions,omitempty\""
		}{
			Pvz: &PVZ{
				City:             PVZCity(report.PVZ.City.Name),
				Id:               &report.PVZ.ID,
				RegistrationDate: &report.PVZ.CreationTimeUTC,
			},
			Receptions: &receptions,
		}
	}
	return response, nil

}

func (h httpRequestHandlers) PostPvz(ctx context.Context, request PostPvzRequestObject) (PostPvzResponseObject, error) {
	args := pvz.CreatePVZUseCaseArgs{
		AuthenticationArgs: h.authArgs(ctx),
		PVZRepository:      h.deps.PVZRepository,
		PVZ: pvz.CreatePVZDTO{
			PVZCity:          string(request.Body.City),
			PVZID:            request.Body.Id,
			RegistrationTime: request.Body.RegistrationDate,
		},
	}

	createdPVZ, err := pvz.CreatePVZUseCase(ctx, args)

	if err != nil {
		msg := err.Error()
		if domain.IsAccessError(err) {
			return PostPvz403JSONResponse{
				Message: msg,
			}, nil
		}

		return PostPvz400JSONResponse{
			Message: msg,
		}, nil
	}

	return PostPvz201JSONResponse{
		Id:               (*openapi_types.UUID)(&createdPVZ.ID),
		City:             PVZCity(createdPVZ.City.Name),
		RegistrationDate: &createdPVZ.CreationTimeUTC,
	}, nil
}

func (h httpRequestHandlers) PostPvzPvzIdCloseLastReception(ctx context.Context, request PostPvzPvzIdCloseLastReceptionRequestObject) (PostPvzPvzIdCloseLastReceptionResponseObject, error) {
	args := reception.CloseLastOpenedReceptionAtPVZArgs{
		AuthenticationArgs:      h.authArgs(ctx),
		ReceptionInfoRepository: h.deps.ReceptionInfoRepository,
		PVZRepository:           h.deps.PVZRepository,
		PVZ: reception.CloseLastOpenedReceptionAtPVZDTO{
			PVZID: request.PvzId,
		},
	}

	reception, err := reception.CloseLastOpenedReceptionUseCase(ctx, args)

	if err != nil {
		msg := err.Error()
		if domain.IsAccessError(err) {
			return PostPvzPvzIdCloseLastReception403JSONResponse{
				Message: msg,
			}, nil
		}

		return PostPvzPvzIdCloseLastReception400JSONResponse{
			Message: msg,
		}, nil
	}

	return PostPvzPvzIdCloseLastReception200JSONResponse{
		Id:       &reception.ID,
		DateTime: reception.CreationTimeUTC,
		PvzId:    reception.PVZID,
		Status:   receptionStatus(reception.Status),
	}, nil
}

func (h httpRequestHandlers) PostPvzPvzIdDeleteLastProduct(ctx context.Context, request PostPvzPvzIdDeleteLastProductRequestObject) (PostPvzPvzIdDeleteLastProductResponseObject, error) {
	args := reception.DeleteLastProductFromCurrentReceptionAtPVZArgs{
		AuthenticationArgs:      h.authArgs(ctx),
		ReceptionInfoRepository: h.deps.ReceptionInfoRepository,
		PVZRepository:           h.deps.PVZRepository,
		ProductRepository:       h.deps.ProductRepository,
		PVZ: reception.DeleteLastProductFromCurrentReceptionAtPVZDTO{
			PVZID: request.PvzId,
		},
	}

	err := reception.DeleteLastProductFromCurrentReceptionAtPVZUseCase(ctx, args)

	if err != nil {
		if domain.IsAccessError(err) {
			return PostPvzPvzIdDeleteLastProduct403JSONResponse{
				Message: err.Error(),
			}, nil
		}

		return PostPvzPvzIdDeleteLastProduct400JSONResponse{
			Message: err.Error(),
		}, nil
	}

	return PostPvzPvzIdDeleteLastProduct200Response{}, nil
}

func (h httpRequestHandlers) PostReceptions(ctx context.Context, request PostReceptionsRequestObject) (PostReceptionsResponseObject, error) {
	args := reception.CreateNewReceptionArgs{
		AuthenticationArgs:      h.authArgs(ctx),
		ReceptionInfoRepository: h.deps.ReceptionInfoRepository,
		PVZRepository:           h.deps.PVZRepository,
		PVZ: reception.CreateNewReceptionAtPVZDTO{
			PVZID: request.Body.PvzId,
		},
	}

	reception, err := reception.CreateNewReceptionUseCase(ctx, args)

	if err != nil {
		msg := err.Error()
		if domain.IsAccessError(err) {
			return PostReceptions403JSONResponse{
				Message: msg,
			}, nil
		}

		return PostReceptions400JSONResponse{
			Message: msg,
		}, nil
	}

	return PostReceptions201JSONResponse{
		Id:       &reception.ID,
		DateTime: reception.CreationTimeUTC,
		PvzId:    reception.PVZID,
		Status:   receptionStatus(reception.Status),
	}, nil
}

func (h httpRequestHandlers) PostRegister(ctx context.Context, request PostRegisterRequestObject) (PostRegisterResponseObject, error) {
	args := users.RegisterUserUseCaseArgs{
		AuthorizationService: h.deps.AuthorizationService,
		User: users.RegisterUserDTO{
			Email:    string(request.Body.Email),
			Password: request.Body.Password,
			Role:     string(request.Body.Role),
		},
	}

	user, err := users.RegisterUserUseCase(ctx, args)

	if err != nil {
		return PostRegister400JSONResponse{
			Message: err.Error(),
		}, nil
	}
	return PostRegister201JSONResponse{
		Email: openapi_types.Email(user.Email),
		Id:    &user.ID,
		Role:  role(user.UserRole.ID),
	}, nil
}
