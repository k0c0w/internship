package grpc_profile

import (
	"avito/internal/config"
	"avito/internal/domain"
	"avito/internal/usecases"
	"avito/internal/usecases/pvz"
	jwt "avito/pkg/authorization"
	context "context"
	"fmt"
	"log"
	"net"
	"time"

	grpc "google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Dependencies struct {
	domain.PVZRepository
	domain.PVZReportAggregateRepository
	domain.AuthorizationService[jwt.JWT]
	jwt.JWTManager
}

type gRPCSerrverWrapper struct {
	cfg  config.GRPCConfig
	deps Dependencies
}

type gRPCServer struct {
	deps Dependencies
	UnimplementedPVZReportServiceServer
}

func NewGRPCServer(deps Dependencies, cfg config.GRPCConfig) *gRPCSerrverWrapper {
	return &gRPCSerrverWrapper{
		cfg:  cfg,
		deps: deps,
	}
}

func (s *gRPCSerrverWrapper) Start() error {

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.cfg.Port))
	if err != nil {
		return err
	}

	serverRegistar := grpc.NewServer()
	service := &gRPCServer{
		deps: s.deps,
	}

	RegisterPVZReportServiceServer(serverRegistar, service)
	log.Println("starting grpc server")
	if err := serverRegistar.Serve(listener); err != nil {
		return err
	}

	return nil
}

func (s *gRPCServer) GetPVZReport(ctx context.Context, request *PVZReportRequest) (*PVZReportResponse, error) {
	token := byPassAuthorization(s.deps.JWTManager)

	var startTime *time.Time
	if request.StartDate != nil && request.StartDate.IsValid() {
		time := request.StartDate.AsTime()
		startTime = &time
	}

	var endTime *time.Time
	if request.EndDate != nil && request.EndDate.IsValid() {
		time := request.EndDate.AsTime()
		endTime = &time
	}

	args := pvz.GetPVZListUseCaseArgs{
		GetPVZListReportsDTO: pvz.GetPVZListReportsDTO{
			ReceptionStartTimeUTC: startTime,
			ReceptionEndTimeUTC:   endTime,
			Page:                  int(request.Page),
			Limit:                 int(request.Limit),
		},
		AuthenticationArgs: usecases.AuthenticationArgs{
			AuthorizationService: s.deps.AuthorizationService,
			JWT:                  token,
		},
		PVZRepository:                s.deps.PVZRepository,
		PVZReportAggregateRepository: s.deps.PVZReportAggregateRepository,
	}

	reports, err := pvz.GetPVZListReportsUseCase(ctx, args)

	if err != nil {
		return &PVZReportResponse{
			Error: err.Error(),
		}, nil
	}

	reportMessages := make([]*PVZReportAggregate, len(reports))
	for i, report := range reports {
		pvz := report.PVZ
		pvzId := pvz.ID.String()
		receptions := make([]*Reception, len(report.Receptions))
		for j, reception := range report.Receptions {
			receptionId := reception.Information.ID.String()
			products := make([]*Product, len(reception.Products))
			for k, product := range reception.Products {
				products[k] = &Product{
					Id:              product.ID.String(),
					ReceptionId:     receptionId,
					CreationTimeUtc: timestamppb.New(product.CreationTimeUTC),
					Category:        productType(product.Category),
				}
			}
			receptions[j] = &Reception{
				Reception: &ReceptionInfo{
					Id:              receptionId,
					PvzId:           pvzId,
					CreationTimeUtc: timestamppb.New(reception.Information.CreationTimeUTC),
					Status:          receptionStatus(reception.Information.Status),
				},
				Products: &ProductList{
					Values: products,
				},
			}
		}

		reportMessages[i] = &PVZReportAggregate{
			Pvz: &PVZ{
				Id:              pvzId,
				CreationTimeUtc: timestamppb.New(pvz.CreationTimeUTC),
				City: &City{
					Name: pvz.City.Name,
				},
			},
			Receptions: &ReceptionList{
				Values: receptions,
			},
		}
	}

	return &PVZReportResponse{
		Reports: &PVZReportAggregateList{
			Values: reportMessages,
		},
	}, nil
}

func byPassAuthorization(m jwt.JWTManager) jwt.JWT {
	const clientId string = "0196521c-873e-77fd-b244-bdd0c13c72ab"

	token, _ := m.GenerateToken(clientId)

	return token
}

func receptionStatus(status domain.ReceptionStatus) string {
	switch status {
	case domain.CloseProductAcceptanceStatus:
		return "close"
	case domain.InProggressProductAcceptanceStatus:
		return "in_progress"
	default:
		log.Println("fall of possible ReceptionStatuses")
		return ""
	}
}

func productType(category domain.ProductCategory) string {
	switch category {
	case domain.ElectronicsProductCategory:
		return "электроника"
	case domain.ClothesProductCategory:
		return "одежда"
	case domain.ShoesProductCategory:
		return "обувь"
	default:
		log.Println("fall out of known product categories")
		return ""
	}
}
