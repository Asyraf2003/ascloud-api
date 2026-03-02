//go:build !legacy_postgres
// +build !legacy_postgres

package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudfrontkeyvaluestore"
	s3svc "github.com/aws/aws-sdk-go-v2/service/s3"
	sqssvc "github.com/aws/aws-sdk-go-v2/service/sqs"

	"example.com/your-api/internal/config"
	authwire "example.com/your-api/internal/modules/auth/wire"
	hostinghttp "example.com/your-api/internal/modules/hosting/transport/http"
	hostingwire "example.com/your-api/internal/modules/hosting/wire"
	dbplat "example.com/your-api/internal/platform/datastore/dynamodb"
	jwtp "example.com/your-api/internal/platform/token/jwt"
	"example.com/your-api/internal/transport/http/router"
	v1AuthRouter "example.com/your-api/internal/transport/http/router/v1/auth"
	"example.com/your-api/internal/transport/http/server"
)

func run() {
	appCfg := config.Load()

	addr := env("HTTP_ADDR", ":"+env("HTTP_PORT", appCfg.HTTPPort))
	service := env("SERVICE_NAME", "api")
	shutdownTimeout := envDuration("SHUTDOWN_TIMEOUT", 10*time.Second)

	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	log = log.With("service", service)

	authCfg := config.LoadAuth(appCfg.Env)
	if err := authCfg.Validate(); err != nil {
		log.Error("invalid auth config", "err", err)
		os.Exit(1)
	}

	v1AuthRouter.InitPolicy(authCfg)

	jwtv, err := jwtp.NewHMACVerifier(authCfg.JWT.Issuer, authCfg.JWT.Audience, authCfg.JWT.Secret)
	if err != nil {
		log.Error("jwt verifier init failed", "err", err)
		os.Exit(1)
	}
	router.SetAccessTokenVerifier(jwtv)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	region := env("AWS_REGION", "ap-southeast-3")

	ddb, err := dbplat.New(ctx, dbplat.ConfigFromEnv(region))
	if err != nil {
		log.Error("dynamodb init failed", "err", err)
		os.Exit(1)
	}

	if err := authwire.WireAuthGoogle(ddb, authCfg); err != nil {
		log.Error("wire auth google failed", "err", err)
		os.Exit(1)
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
	if err != nil {
		log.Error("aws config load failed", "err", err)
		os.Exit(1)
	}

	s3 := s3svc.NewFromConfig(awsCfg)
	sqs := sqssvc.NewFromConfig(awsCfg)
	kvs := cloudfrontkeyvaluestore.NewFromConfig(awsCfg)

	// Step 4: Upload pipeline (presign + head + enqueue)
	hostingSvc, err := hostingwire.WireHostingUploadPipeline(ddb, s3, sqs)
	if err != nil {
		log.Error("wire hosting upload pipeline failed", "err", err)
		os.Exit(1)
	}
	hostinghttp.SetUploadHandler(hostinghttp.NewUploadHandler(hostingSvc))

	// Step 7: Dashboard control-plane (sites + releases + rollback via KVS)
	dash, err := hostingwire.WireHostingDashboard(ddb, kvs)
	if err != nil {
		log.Error("wire hosting dashboard failed", "err", err)
		os.Exit(1)
	}
	hostinghttp.SetSiteHandler(hostinghttp.NewSiteHandler(dash))
	hostinghttp.SetReleaseHandler(hostinghttp.NewReleaseHandler(dash))
	hostinghttp.SetRollbackHandler(hostinghttp.NewRollbackHandler(dash))

	e := server.New(log, ddb, authCfg.Security.AllowedOrigins)
	router.Register(e)

	go func() {
		log.Info("starting", "addr", addr)
		if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
			log.Error("server start failed", "err", err)
		}
	}()

	ctx2, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	<-ctx2.Done()
	stop()

	log.Info("shutting down")
	sdCtx, cancel2 := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel2()

	if err := e.Shutdown(sdCtx); err != nil {
		log.Error("shutdown failed", "err", err)
		return
	}
	log.Info("shutdown complete")
}
