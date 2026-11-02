package main

import (
	"context"
	_ "embed"
	"errors"
	"flag"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/joho/godotenv"
	"github.com/wailsapp/wails/v3/pkg/application"

	"github.com/zeitlos/lucity/pkg/logger"
	"github.com/zeitlos/lucity/services/cockpit/web"
)

//go:embed appicon.png
var appIcon []byte

func main() {
	environment := flag.String("env", "", "platform environment to open first, dev or prod (default: the last one opened, else dev)")
	deployments := flag.String("deployments", "../../deployments", "directory with the lucity-<env> secrets and values")
	flag.Parse()

	logger.Setup("info")

	if err := godotenv.Load(); err != nil && !errors.Is(err, fs.ErrNotExist) {
		slog.Error("load .env", "error", err)
		os.Exit(1)
	}

	if *environment != "" && !slices.Contains(knownEnvironments, *environment) {
		slog.Error("-env must be dev or prod")
		os.Exit(1)
	}

	configDirectory, err := os.UserConfigDir()

	if err != nil {
		slog.Error("find config directory", "error", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if flag.Arg(0) == "mcp" {
		if err := serveMCP(ctx, *deployments); err != nil {
			slog.Error("serve mcp", "error", err)
			os.Exit(1)
		}

		return
	}

	assets, err := fs.Sub(web.Dist, "dist")

	if err != nil {
		slog.Error("load web assets", "error", err)
		os.Exit(1)
	}

	mux := http.NewServeMux()
	mux.Handle("/graphql", newEnvironments(ctx, *deployments, filepath.Join(configDirectory, "lucity-cockpit", "environment"), *environment))
	mux.Handle("/", application.AssetFileServerFS(assets))

	app := application.New(application.Options{
		Name: "Cockpit",
		Icon: appIcon,
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
		Assets: application.AssetOptions{
			Handler: mux,
		},
	})

	links := []string{"https://github.com/", "https://dashboard.stripe.com/"}

	for _, name := range knownEnvironments {
		if base, err := analyticsBase(filepath.Join(*deployments, "lucity-"+name)); err == nil && base != "" {
			links = append(links, base+"/")
		}
	}

	mux.HandleFunc("POST /open", func(w http.ResponseWriter, r *http.Request) {
		url, err := io.ReadAll(io.LimitReader(r.Body, 2048))
		allowed := slices.ContainsFunc(links, func(prefix string) bool {
			return strings.HasPrefix(string(url), prefix)
		})

		if err != nil || !allowed {
			http.Error(w, "only GitHub, Stripe and analytics links open", http.StatusBadRequest)
			return
		}

		if err := app.Browser.OpenURL(string(url)); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:     "Cockpit",
		Width:     1440,
		Height:    900,
		MinWidth:  1024,
		MinHeight: 640,
		URL:       "/",
		Mac: application.MacWindow{
			TitleBar:                application.MacTitleBarHiddenInset,
			InvisibleTitleBarHeight: 52,
		},
	})

	if err := app.Run(); err != nil {
		slog.Error("run", "error", err)
		os.Exit(1)
	}
}
