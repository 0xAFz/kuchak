package cmd

import (
	"context"
	"fmt"
	"kuchak/internal/api"
	"kuchak/internal/config"
	"kuchak/internal/repository"
	"kuchak/internal/repository/postgres"
	"kuchak/internal/repository/redis"
	"kuchak/internal/service"
	"os"
	"os/signal"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func Serve() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	config.LoadConfig()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	pgxSession, err := postgres.NewPostgresSession()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect postgres")
	}

	err = pgxSession.Ping(context.Background())
	if err != nil {
		log.Fatal().Err(err).Msg("failed to ping postgres")
	}

	redisClient, err := redis.NewRedisClient(fmt.Sprintf("%s:%s", config.AppConfig.RedisHost, config.AppConfig.RedisPort))
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect redis")
	}

	URLRedisRepository := repository.NewURLRedisRepository(redisClient)
	URLPostgresRepository := repository.NewURLPostgresRepository(pgxSession)
	accountPostgresRepository := repository.NewAccountPostgresRepository(pgxSession)
	accountRedisRepository := repository.NewAccountRedisRepository(redisClient)
	rateLimitRepository := repository.NewRateLimiterRepository(redisClient)

	app := service.NewApp(
		service.NewAccountPostgresService(accountPostgresRepository),
		service.NewURLPostgresService(URLPostgresRepository),
		service.NewAccountRedisService(accountRedisRepository),
		service.NewURLRedisService(URLRedisRepository),
		service.NewRateLimitService(rateLimitRepository),
		service.NewEmailService(config.AppConfig.SmtpHost, config.AppConfig.SmtpPort, config.AppConfig.SmtpUsername, config.AppConfig.SmtpPassword, config.AppConfig.SmtpUsername),
	)

	wa := api.NewWebApp(config.AppConfig.ServerAddr, config.AppConfig.AppURL, app)

	go func() {
		log.Fatal().Err(wa.Start())
	}()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	defer wa.Shutdown(shutdownCtx)

	log.Info().Msg("Server is up and running...")
	<-ctx.Done()
	log.Info().Msg("Shutting down the server...")
}
