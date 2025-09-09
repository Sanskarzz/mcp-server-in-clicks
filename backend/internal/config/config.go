package config

import (
	"os"
)

type Config struct {
	MongoURI       string
	MongoDB        string
	HelmNamespace  string
	HelmChartPath  string
	KubeConfigPath string
}

func Load() Config {
	return Config{
		MongoURI:       env("MONGO_URI", "mongodb://localhost:27017"),
		MongoDB:        env("MONGO_DB", "mcp"),
		HelmNamespace:  env("HELM_NAMESPACE", "mcp"),
		HelmChartPath:  env("HELM_CHART_PATH", "../mcp-server-template/deploy/helm"),
		KubeConfigPath: env("KUBECONFIG", ""),
	}
}

func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
