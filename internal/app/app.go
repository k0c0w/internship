package app

import (
	grpc_profile "avito/internal/api/grpc-profile"
	http_profile "avito/internal/api/http"
	"avito/internal/config"
	services "avito/internal/services"
	"avito/internal/storage"
	jwt "avito/pkg/authorization"
	postgresql "avito/pkg/database"
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func Run(configPath string) {
	cfg, err := config.InitConfig(configPath)
	if err != nil {
		log.Fatalln(err)
		return
	}

	postgresClient, err := postgresql.NewClient(cfg.PostgresConfig)
	if err != nil {
		log.Fatalln(err)
		return
	}

	repositories := storage.NewRepositories(postgresClient)
	jwtManager := jwt.NewJWTManager(cfg.AuthConfig.JWTConfig.Sign, cfg.AuthConfig.JWTConfig.Issuer, cfg.AuthConfig.JWTConfig.TokenTTL)
	authService := services.NewAuthorizationService(*jwtManager, repositories.UserRepository)

	httpDeps := http_profile.Dependencies{
		AuthorizationService: authService,
		JWTManager:           *jwtManager,
		Repositories:         repositories,
	}

	grpcDeps := grpc_profile.Dependencies{
		AuthorizationService:         authService,
		PVZRepository:                repositories.PVZRepository,
		PVZReportAggregateRepository: repositories.PVZReportAggregateRepository,
		JWTManager:                   *jwtManager,
	}

	httpServer := http_profile.NewHTTPServer(httpDeps, cfg.HTTPConfig)
	grpcServer := grpc_profile.NewGRPCServer(grpcDeps, cfg.GRPCConfig)

	go func() {
		err := httpServer.Start()
		if err != nil {
			log.Fatalln(err)
		}
	}()

	go func() {
		err := grpcServer.Start()
		if err != nil {
			log.Fatalln(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	<-quit
	log.Println("stopping server")
	err = httpServer.Stop(context.TODO())
	if err != nil {
		log.Println(err)
	}

	log.Println("closing connections")
	postgresClient.Close()
	log.Println("shutting down")
}
