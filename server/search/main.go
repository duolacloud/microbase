package main

import (
	"context"

	"github.com/duolacloud/microbase/cmd"
	"github.com/duolacloud/microbase/server/search/bootstrap"
	"github.com/urfave/cli/v2"
)

func main() {
	cmd.Run(func(c *cli.Context) error {
		app := bootstrap.Setup(c)
		return app.Start(context.Background())
	}, nil)
}
