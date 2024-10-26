package cmd

import (
	"context"
	"kuchak/internal/api"
	"kuchak/internal/config"
	"kuchak/internal/repository"
	"kuchak/internal/repository/postgres"
	"kuchak/internal/repository/redis"
	"kuchak/internal/service"
	"log"
	"os"
	"os/signal"
	"time"
)

func Serve() {
	config.LoadConfig()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	pgxSession, err := postgres.NewPostgresSession()
	if err != nil {
		log.Fatalf("Failed to connect postgres: %v", err)
	}

	err = pgxSession.Ping(context.Background())
	if err != nil {
		log.Fatalf("Failed to ping postgres: %v", err)
	}

	redisClient, err := redis.NewRedisClient(config.AppConfig.RedisHost)
	if err != nil {
		log.Fatal(err)
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

	wa := api.NewWebApp(config.AppConfig.ServerAddr, app)

	go func() {
		log.Fatal(wa.Start())
	}()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	defer wa.Shutdown(shutdownCtx)

	log.Println("Server is up and running...")
	<-ctx.Done()
	log.Println("Shutting down the server...")
}
