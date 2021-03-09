package source

import (
	"github.com/duolacloud/microbase/config"
	"github.com/duolacloud/microbase/logger"
	"github.com/micro/go-micro/v2/config/encoder/yaml"
	"github.com/micro/go-micro/v2/config/source"
	"github.com/urfave/cli/v2"
	apollo "github.com/xxxmicro/go-micro-apollo-plugin"
)

func NewSourceProvider(c *cli.Context) source.Source {
	address := c.String("apollo_address")
	if len(address) == 0 {
		address = config.Env("APOLLO_ADDRESS", "")
	}

	if len(address) == 0 {
		logger.Fatal("need config address")
		return nil
	}

	namespace := c.String("apollo_namespace")
	if len(namespace) == 0 {
		namespace = config.Env("APOLLO_NAMESPACE", "application")
	}

	appId := c.String("apollo_app_id")
	if len(appId) == 0 {
		appId = config.Env("APOLLO_APP_ID", "")
	}

	cluster := c.String("apollo_cluster")
	if len(cluster) == 0 {
		cluster = config.Env("APOLLO_CLUSTER", "dev")
	}

	backupConfigPath := config.Env("BACKUP_CONFIG_PATH", "./")

	e := yaml.NewEncoder()
	return apollo.NewSource(
		apollo.WithAddress(address),
		apollo.WithNamespace(namespace),
		apollo.WithAppId(appId),
		apollo.WithCluster(cluster),
		apollo.WithBackupConfigPath(backupConfigPath),
		source.WithEncoder(e),
	)
}
