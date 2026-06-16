package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/activecm/rita/v5/cmd"
	"github.com/activecm/rita/v5/config"
	zlog "github.com/activecm/rita/v5/logger"
	"github.com/activecm/rita/v5/metrics"
	"github.com/activecm/rita/v5/telemetry"
	"github.com/activecm/rita/v5/viewer"

	"github.com/joho/godotenv"
	"github.com/urfave/cli/v2"
)

// Version is populated by build flags with the current Git tag
var Version string

func main() {
	// set the version in config to make it more importable by other packages
	config.Version = Version

	// UNIX Time is faster and smaller than most timestamps
	// zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// Observability is configured from the process environment (not the .env app
	// config) so it can be driven by the container/orchestrator: OTEL_* enables
	// tracing, RITA_METRICS_ADDR enables the Prometheus + health endpoint.
	ctx := context.Background()
	obsLogger := zlog.GetLogger()

	shutdownTracer, err := telemetry.InitTracer(ctx, "rita", Version)
	if err != nil {
		obsLogger.Warn().Err(err).Msg("failed to initialize tracing; continuing without it")
		shutdownTracer = func(context.Context) error { return nil }
	}

	var metricsServer *metrics.Server
	if addr := os.Getenv("RITA_METRICS_ADDR"); addr != "" {
		metricsServer = metrics.NewServer(addr)
		metricsErrc := metricsServer.Start()
		go func() {
			if err := <-metricsErrc; err != nil {
				l := zlog.GetLogger()
				l.Error().Err(err).Str("addr", addr).Msg("metrics server error")
			}
		}()
		obsLogger.Info().Str("addr", addr).Msg("serving Prometheus metrics at /metrics and liveness at /healthz")
	}

	shutdownObservability := func() {
		if metricsServer != nil {
			_ = metricsServer.Shutdown(ctx)
		}
		_ = shutdownTracer(ctx)
	}

	app := &cli.App{
		EnableBashCompletion: true,
		Commands:             cmd.Commands(),
		Name:                 "RITA",
		Usage:                "Look for evil needles in big haystacks",
		UsageText:            "rita [-d] command [command options]",
		Version:              Version,
		Args:                 true,
		ExitErrHandler:       exitErrHandler,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:     "debug",
				Aliases:  []string{"d"},
				Usage:    "Run in debug mode",
				Value:    false, // default config file path
				Required: false,
			},
		},
		Before: func(cCtx *cli.Context) error {
			// set logger mode based on APP_ENV
			zlog.DebugMode = os.Getenv("APP_ENV") == "dev"

			// override APP_ENV if the --debug flag is set
			// *note that global flags must be placed before the subcommand when running in the CLI
			if cCtx.Bool("debug") {
				zlog.DebugMode = true
				viewer.DebugMode = true
			}

			// load environment variables from .env files
			// base .env file is required
			err := godotenv.Load("./.env")
			if err != nil {
				log.Fatal("Error loading .env file", err)
			}

			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		// flush metrics/traces before the process exits via Fatal (which skips defers)
		shutdownObservability()
		logger := zlog.GetLogger()
		logger.Fatal().Err(err).Send()
	}

	shutdownObservability()
}

// exitErrHandler implements cli.ExitErrHandlerFunc
func exitErrHandler(c *cli.Context, err error) {
	if err == nil {
		return
	}
	fmt.Fprintf(c.App.ErrWriter, "\n\n\t[!] %+v\n\n", err.Error())
	cli.OsExiter(1)

}
