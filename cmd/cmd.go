package cmd

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

var defaultFlags []cli.Flag

func init() {
	defaultFlags = []cli.Flag{
		&cli.StringSliceFlag{
			Name:    "registry_address",
			Usage:   "registry_address",
			EnvVars: []string{"MICRO_REGISTRY_ADDRESS"},
		},
		&cli.StringFlag{
			Name:    "service_name",
			Usage:   "service_name",
			EnvVars: []string{"MICRO_SERVICE_NAME"},
		},
		&cli.StringFlag{
			Name:  "apollo_namespace",
			Usage: "apollo_namespace",
		},
		&cli.StringFlag{
			Name:  "apollo_address",
			Usage: "apollo_address",
		},
		&cli.StringFlag{
			Name:  "apollo_app_id",
			Usage: "apollo_app_id",
		},
		&cli.StringFlag{
			Name:  "apollo_cluster",
			Usage: "apollo_cluster",
		},
		&cli.StringFlag{
			Name:    "prometheus_addr",
			Usage:   "prometheus_addr",
			EnvVars: []string{"PROMETHEUS_ADDR"},
			Value:   ":16627",
		},
	}
}

func Run(cb func(c *cli.Context) error, flags []cli.Flag) {
	f := defaultFlags
	for _, flag := range flags {
		f = append(f, flag)
	}

	app := &cli.App{
		Flags:  f,
		Action: cb,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
