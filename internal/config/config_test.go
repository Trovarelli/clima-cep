package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_Defaults(t *testing.T) {
	os.Unsetenv("PORT")
	os.Unsetenv("WEATHER_API_KEY")
	os.Unsetenv("WEATHERAPI_KEY")

	cfg := LoadConfig()
	if cfg.Port != "8080" {
		t.Errorf("Port default esperada '8080', obteve %q", cfg.Port)
	}
}

func TestLoadConfig_EnvVars(t *testing.T) {
	os.Setenv("PORT", "9090")
	os.Setenv("WEATHER_API_KEY", "test_key_123")
	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("WEATHER_API_KEY")
	}()

	cfg := LoadConfig()
	if cfg.Port != "9090" {
		t.Errorf("Port esperada '9090', obteve %q", cfg.Port)
	}
	if cfg.WeatherAPIKey != "test_key_123" {
		t.Errorf("WeatherAPIKey esperada 'test_key_123', obteve %q", cfg.WeatherAPIKey)
	}
}

func TestLoadConfig_DotEnvFallback(t *testing.T) {
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env")

	content := "PORT=7070\nWEATHER_API_KEY=file_key_456\n# comentario\n\n"
	if err := os.WriteFile(envPath, []byte(content), 0644); err != nil {
		t.Fatalf("Erro ao criar arquivo .env: %v", err)
	}

	os.Unsetenv("PORT")
	os.Unsetenv("WEATHER_API_KEY")

	loadDotEnv(envPath)

	if os.Getenv("PORT") != "7070" {
		t.Errorf("PORT lida do .env esperada '7070', obteve %q", os.Getenv("PORT"))
	}
	if os.Getenv("WEATHER_API_KEY") != "file_key_456" {
		t.Errorf("WEATHER_API_KEY lida do .env esperada 'file_key_456', obteve %q", os.Getenv("WEATHER_API_KEY"))
	}
}
