package cmd

import (
	"context"
	"fmt"
	"kuchak/internal/config"
	"kuchak/internal/entity"
	"kuchak/internal/repository"
	"kuchak/internal/repository/postgres"
	"kuchak/internal/repository/redis"
	"log"
	"time"
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

	URLRedisRepository := repository.NewURLRedisRepository(client)

	expiry := time.Now().Add(5 * time.Minute)

	url := entity.URL{
		ID:          1,
		ShortURL:    "xyz",
		OriginalURL: "https://domain.tld",
		UserID:      1,
		ClickCount:  100,
		ExpiryDate:  &expiry,
		CreatedAt:   time.Now(),
	}

	err = URLRedisRepository.Save(context.Background(), url)
	if err != nil {
		log.Fatal(err)
	}

	u, err := URLRedisRepository.ByShortURL(context.Background(), "xyz")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(u)

	fmt.Println("Server is up and running...")
}
