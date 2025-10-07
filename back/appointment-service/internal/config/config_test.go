package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testConfig struct {
	Server struct {
		Host string `koanf:"host"`
		Port int    `koanf:"port"`
	} `koanf:"server"`

	DB struct {
		User string `koanf:"user"`
		Pass string `koanf:"pass"`
	} `koanf:"db"`
}

func TestMain(m *testing.M) {
	cleanupEnvs("TestMain")
	os.Exit(m.Run())
}

func TestLoadConfig_OnlyEnvVars(t *testing.T) {
	t.Setenv("SHED_DATABASE__HOST", "localhost")
	t.Setenv("SHED_DATABASE__PORT", "5432")
	t.Setenv("SHED_APP_NAME", "test-app")

	cfg, err := LoadConfig()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "localhost", cfg.String("database.host"))
	assert.Equal(t, "5432", cfg.String("database.port"))
	assert.Equal(t, "test-app", cfg.String("app_name"))
}

func TestLoadConfig_UnmarshalEnv(t *testing.T) {
	t.Setenv("SHED_SERVER__HOST", "localhost")
	t.Setenv("SHED_SERVER__PORT", "8080")
	t.Setenv("SHED_DB__USER", "admin")
	t.Setenv("SHED_DB__PASS", "secret")

	k, err := LoadConfig()
	require.NoError(t, err)
	require.NotNil(t, k)

	var cfg testConfig
	err = k.Unmarshal("", &cfg)
	require.NoError(t, err)

	require.Equal(t, "localhost", cfg.Server.Host)
	require.Equal(t, 8080, cfg.Server.Port)
	require.Equal(t, "admin", cfg.DB.User)
	require.Equal(t, "secret", cfg.DB.Pass)
}

func TestLoadConfig_OnlyYamlFile(t *testing.T) {
	yamlContent := `
database:
  host: yaml-host
  port: 3306
app:
  name: yaml-app
  timeout: 30
`
	yamlPath := createTempYaml(t, yamlContent)
	t.Setenv("SHED_CONFIG_PATH", yamlPath)

	cfg, err := LoadConfig()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "yaml-host", cfg.String("database.host"))
	assert.Equal(t, int64(3306), cfg.Int64("database.port"))
	assert.Equal(t, "yaml-app", cfg.String("app.name"))
	assert.Equal(t, int64(30), cfg.Int64("app.timeout"))
}

func TestLoadConfig_EnvOverridesYaml(t *testing.T) {
	yamlContent := `
database:
  host: yaml-host
  port: 3306
app:
  name: yaml-app
`
	yamlPath := createTempYaml(t, yamlContent)
	t.Setenv("SHED_CONFIG_PATH", yamlPath)

	// override YAML
	t.Setenv("SHED_DATABASE__HOST", "env-host")
	t.Setenv("SHED_APP__NAME", "env-app")

	cfg, err := LoadConfig()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "env-host", cfg.String("database.host"))
	assert.Equal(t, "env-app", cfg.String("app.name"))
	assert.Equal(t, int64(3306), cfg.Int64("database.port"))
}

func TestLoadConfig_NestedEnvVars(t *testing.T) {
	t.Setenv("SHED_SERVER__HTTP__HOST", "0.0.0.0")
	t.Setenv("SHED_SERVER__HTTP__PORT", "8080")
	t.Setenv("SHED_SERVER__GRPC__PORT", "9090")

	cfg, err := LoadConfig()
	require.NoError(t, err)

	assert.Equal(t, "0.0.0.0", cfg.String("server.http.host"))
	assert.Equal(t, "8080", cfg.String("server.http.port"))
	assert.Equal(t, "9090", cfg.String("server.grpc.port"))
}

func TestLoadConfig_InvalidYamlPath(t *testing.T) {
	t.Setenv("SHED_CONFIG_PATH", "/nonexistent/path/config.yaml")

	cfg, err := LoadConfig()
	assert.Error(t, err)
	assert.Nil(t, cfg)
}

func TestLoadConfig_InvalidYamlContent(t *testing.T) {
	yamlContent := `
database:
  host: invalid
- this is broken yaml
  port: not valid
`
	yamlPath := createTempYaml(t, yamlContent)
	t.Setenv("SHED_CONFIG_PATH", yamlPath)

	cfg, err := LoadConfig()
	assert.Error(t, err)
	assert.Nil(t, cfg)
}

func TestLoadConfig_EmptyConfig(t *testing.T) {
	cfg, err := LoadConfig()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "", cfg.String("nonexistent.key"))
}

func TestLoadConfig_SpecialCharacters(t *testing.T) {
	t.Setenv("SHED_APP__SECRET", "my$ecret!@#")
	t.Setenv("SHED_APP__URL", "https://example.com/path?query=value")

	cfg, err := LoadConfig()
	require.NoError(t, err)

	assert.Equal(t, "my$ecret!@#", cfg.String("app.secret"))
	assert.Equal(t, "https://example.com/path?query=value", cfg.String("app.url"))
}

func TestLoadConfig_NumericValues(t *testing.T) {
	yamlContent := `
app:
  port: 8080
  timeout: 30
  ratio: 0.75
  enabled: true
`
	yamlPath := createTempYaml(t, yamlContent)
	t.Setenv("SHED_CONFIG_PATH", yamlPath)

	cfg, err := LoadConfig()
	require.NoError(t, err)

	assert.Equal(t, int64(8080), cfg.Int64("app.port"))
	assert.Equal(t, int64(30), cfg.Int64("app.timeout"))
	assert.Equal(t, 0.75, cfg.Float64("app.ratio"))
	assert.Equal(t, true, cfg.Bool("app.enabled"))
}

func TestLoadConfig_ComplexYamlStructure(t *testing.T) {
	yamlContent := `
database:
  primary:
    host: primary-db
    port: 5432
    credentials:
      username: admin
      password: secret
  replica:
    host: replica-db
    port: 5433
`
	yamlPath := createTempYaml(t, yamlContent)
	t.Setenv("SHED_CONFIG_PATH", yamlPath)

	cfg, err := LoadConfig()
	require.NoError(t, err)

	assert.Equal(t, "primary-db", cfg.String("database.primary.host"))
	assert.Equal(t, int64(5432), cfg.Int64("database.primary.port"))
	assert.Equal(t, "admin", cfg.String("database.primary.credentials.username"))
	assert.Equal(t, "secret", cfg.String("database.primary.credentials.password"))
	assert.Equal(t, "replica-db", cfg.String("database.replica.host"))
}

func cleanupEnvs(label string) {
	for _, env := range os.Environ() {
		if len(env) > 5 && env[:5] == "SHED_" {
			pair := strings.SplitN(env, "=", 2)
			os.Unsetenv(pair[0])
			fmt.Printf("%s: unset %s\n", label, pair[0])
		}
	}
}

func createTempYaml(t *testing.T, content string) string {
	t.Helper()

	tmpDir := t.TempDir()
	yamlPath := filepath.Join(tmpDir, "config.yaml")

	err := os.WriteFile(yamlPath, []byte(content), 0644)
	require.NoError(t, err)

	return yamlPath
}
