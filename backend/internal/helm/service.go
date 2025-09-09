package helm

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"

	"mcp-backend/internal/config"
)

type Service struct{ cfg config.Config }

func NewService(cfg config.Config) *Service { return &Service{cfg: cfg} }

// Install or upgrade a release for an MCP server using Helm SDK
func (s *Service) UpsertRelease(releaseName string, valuesYAML string, namespace string) error {
	if namespace == "" {
		namespace = s.cfg.HelmNamespace
	}

	settings := cli.New()
	if s.cfg.KubeConfigPath != "" {
		settings.KubeConfig = s.cfg.KubeConfigPath
	}

	var cfg action.Configuration
	if err := cfg.Init(settings.RESTClientGetter(), namespace, "secrets", logrus.Debugf); err != nil {
		return fmt.Errorf("helm init failed: %w", err)
	}

	chart, err := loader.Load(s.cfg.HelmChartPath)
	if err != nil {
		return fmt.Errorf("load chart failed: %w", err)
	}

	vals := map[string]interface{}{}
	if valuesYAML != "" {
		if err := yaml.Unmarshal([]byte(valuesYAML), &vals); err != nil {
			return fmt.Errorf("values parse failed: %w", err)
		}
	}

	up := action.NewUpgrade(&cfg)
	up.Namespace = namespace
	up.Install = true // upgrade --install semantics

	if _, err := up.Run(releaseName, chart, vals); err != nil {
		return fmt.Errorf("helm upgrade/install failed: %w", err)
	}
	return nil
}

func (s *Service) UninstallRelease(releaseName string, namespace string) error {
	if namespace == "" {
		namespace = s.cfg.HelmNamespace
	}
	settings := cli.New()
	if s.cfg.KubeConfigPath != "" {
		settings.KubeConfig = s.cfg.KubeConfigPath
	}
	var cfg action.Configuration
	if err := cfg.Init(settings.RESTClientGetter(), namespace, "secrets", logrus.Debugf); err != nil {
		return fmt.Errorf("helm init failed: %w", err)
	}
	un := action.NewUninstall(&cfg)
	if _, err := un.Run(releaseName); err != nil {
		return fmt.Errorf("helm uninstall failed: %w", err)
	}
	return nil
}

// RenderValues maps arbitrary map[string]interface{} to YAML for Helm values.
func (s *Service) RenderValues(conf map[string]interface{}) (string, error) {
	b, err := yaml.Marshal(conf)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
