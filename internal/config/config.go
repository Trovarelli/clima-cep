package config

import (
	"bufio"
	"os"
	"strings"
)

type Config struct {
	Port          string
	WeatherAPIKey string
}

func LoadConfig() *Config {
	loadDotEnv(".env")

	port := os.Getenv("PORT")
	if strings.TrimSpace(port) == "" {
		port = "8080"
	}

	apiKey := os.Getenv("WEATHER_API_KEY")
	if strings.TrimSpace(apiKey) == "" {
		apiKey = os.Getenv("WEATHERAPI_KEY")
	}

	return &Config{
		Port:          port,
		WeatherAPIKey: strings.TrimSpace(apiKey),
	}
}

func loadDotEnv(filepath string) {
	file, err := os.Open(filepath)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			val = strings.Trim(val, `"'`)

			if os.Getenv(key) == "" {
				_ = os.Setenv(key, val)
			}
		}
	}
}
