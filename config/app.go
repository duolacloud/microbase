package config

type AppConfig struct {
	Env           string
	GraphqlSchema string
}

func NewAppConfig() *AppConfig {
	return &AppConfig{
		Env:           Env("APP_ENV", "production"),
		GraphqlSchema: Env("GRAPHQL_SCHEMA", "schema.graphql"),
	}
}
