package cmd

import (
	"context"
	"fmt"
	"kuchak/internal/config"
	"kuchak/internal/repository/postgres"
	"kuchak/internal/repository/redis"
	"log"
)

func Serve() {
	config.LoadConfig()
	pgxSession, err := postgres.NewPostgresSession()
	if err != nil {
		log.Fatalf("Failed to connect postgres: %v", err)
	}
	err = pgxSession.Ping(context.Background())
	if err != nil {
		fmt.Println("Failed to get ping from postgres")
	}

	client, err := redis.NewRedisClient(config.AppConfig.RedisHost)
	if err != nil {
		log.Fatal(err)
	}
	cmd := client.B().Set().Key("key").Value("value").Build()
	err = client.Do(context.Background(), cmd).Error()
	if err != nil {
		fmt.Println("Failed to set item in redis")
	}

	fmt.Println("Server is up and running...")
}
