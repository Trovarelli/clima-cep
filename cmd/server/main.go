package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"clima-cep/internal/config"
	"clima-cep/internal/infra/viacep"
	"clima-cep/internal/infra/weatherapi"
	"clima-cep/internal/usecase"
	"clima-cep/internal/web"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg := config.LoadConfig()

	if cfg.WeatherAPIKey == "" {
		logger.Warn("ATENCAO: WEATHER_API_KEY nao informada. A consulta de clima falhara ate que a chave seja configurada.")
	}

	viaCEPClient := viacep.NewClient(nil, "")
	weatherClient := weatherapi.NewClient(nil, "", cfg.WeatherAPIKey)
	weatherUseCase := usecase.NewWeatherByCEPUseCase(viaCEPClient, weatherClient)
	handler := web.NewHandler(weatherUseCase, logger)

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           handler.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		logger.Info("servidor HTTP inicializado", "porta", cfg.Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("falha ao iniciar servidor HTTP", "erro", err)
			os.Exit(1)
		}
	}()

	<-stop
	logger.Info("desligando servidor HTTP de forma graciosa...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("falha durante desligamento gracioso", "erro", err)
		_ = server.Close()
	}

	logger.Info("servidor encerrado com sucesso.")
}
