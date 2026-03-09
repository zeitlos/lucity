package main

import (
	"log/slog"
	"os"

	"github.com/kelseyhightower/envconfig"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/zeitlos/lucity/pkg/deployer"
	"github.com/zeitlos/lucity/pkg/graceful"
	"github.com/zeitlos/lucity/pkg/logger"
	cashiergrpc "github.com/zeitlos/lucity/services/cashier/grpc"
	cashierhttp "github.com/zeitlos/lucity/services/cashier/http"
	stripelib "github.com/zeitlos/lucity/services/cashier/stripe"
)

type Config struct {
	Port        string `envconfig:"PORT" default:"9005"`
	WebhookPort string `envconfig:"WEBHOOK_PORT" default:"9006"`
	LogLevel    string `envconfig:"LOG_LEVEL" default:"info"`

	DeployerAddr string `envconfig:"DEPLOYER_ADDR" default:"localhost:9003"`

	StripeSecretKey     string `envconfig:"STRIPE_SECRET_KEY" required:"true"`
	StripeWebhookSecret string `envconfig:"STRIPE_WEBHOOK_SECRET" required:"true"`

	HobbyPriceID    string `envconfig:"STRIPE_HOBBY_PRICE_ID" required:"true"`
	ProPriceID      string `envconfig:"STRIPE_PRO_PRICE_ID" required:"true"`
	EcoCPUPriceID   string `envconfig:"STRIPE_ECO_CPU_PRICE_ID" required:"true"`
	EcoMemPriceID   string `envconfig:"STRIPE_ECO_MEM_PRICE_ID" required:"true"`
	EcoDiskPriceID  string `envconfig:"STRIPE_ECO_DISK_PRICE_ID" required:"true"`
	ProdCPUPriceID  string `envconfig:"STRIPE_PROD_CPU_PRICE_ID" required:"true"`
	ProdMemPriceID  string `envconfig:"STRIPE_PROD_MEM_PRICE_ID" required:"true"`
	ProdDiskPriceID string `envconfig:"STRIPE_PROD_DISK_PRICE_ID" required:"true"`
}

func main() {
	var config Config
	if err := envconfig.Process("", &config); err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	logger.Setup(config.LogLevel)

	// Connect to deployer for workspace metadata
	deployerConn, err := grpc.NewClient(config.DeployerAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		slog.Error("failed to connect to deployer", "error", err, "addr", config.DeployerAddr)
		os.Exit(1)
	}
	defer deployerConn.Close()

	deployerClient := deployer.NewDeployerServiceClient(deployerConn)

	// Stripe client
	prices := stripelib.PriceConfig{
		HobbyPriceID:    config.HobbyPriceID,
		ProPriceID:      config.ProPriceID,
		EcoCPUPriceID:   config.EcoCPUPriceID,
		EcoMemPriceID:   config.EcoMemPriceID,
		EcoDiskPriceID:  config.EcoDiskPriceID,
		ProdCPUPriceID:  config.ProdCPUPriceID,
		ProdMemPriceID:  config.ProdMemPriceID,
		ProdDiskPriceID: config.ProdDiskPriceID,
	}
	stripeClient := stripelib.NewClient(config.StripeSecretKey, prices)

	// gRPC server
	svc := cashiergrpc.NewServer(stripeClient, deployerClient)
	grpcServer := cashiergrpc.NewGRPCServer(":"+config.Port, svc)

	// Stripe webhook HTTP server
	webhookHandler := stripelib.NewWebhookHandler(config.StripeWebhookSecret, stripeClient, svc)
	httpServer := cashierhttp.NewServer(config.WebhookPort, webhookHandler)

	ctx, cancel := graceful.Context()
	defer cancel()

	graceful.Serve(ctx, grpcServer, httpServer)
}
