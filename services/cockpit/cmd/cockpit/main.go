package main

import (
	"context"
	_ "embed"
	"flag"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"slices"
	"strings"

	"github.com/wailsapp/wails/v3/pkg/application"

	"github.com/zeitlos/lucity/pkg/logger"
	"github.com/zeitlos/lucity/services/cockpit/web"
)

//go:embed appicon.png
var appIcon []byte

func main() {
	environment := flag.String("env", "dev", "platform environment to open first, dev or prod")
	deployments := flag.String("deployments", "../../deployments", "directory with the lucity-<env> secrets and values")
	flag.Parse()

	logger.Setup("info")

	if !slices.Contains(knownEnvironments, *environment) {
		slog.Error("-env must be dev or prod")
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	assets, err := fs.Sub(web.Dist, "dist")

	if err != nil {
		slog.Error("load web assets", "error", err)
		os.Exit(1)
	}

	mux := http.NewServeMux()
	mux.Handle("/graphql", newEnvironments(ctx, *deployments, *environment))
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

	mux.HandleFunc("POST /open", func(w http.ResponseWriter, r *http.Request) {
		url, err := io.ReadAll(io.LimitReader(r.Body, 2048))

		if err != nil || !strings.HasPrefix(string(url), "https://github.com/") {
			http.Error(w, "only github.com links open", http.StatusBadRequest)
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
