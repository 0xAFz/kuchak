package service

type App struct {
	AccountPostgres *AccountPostgresService
	URLPostgres     *URLPostgresService
	AccountRedis    *AccountRedisService
	URLRedis        *URLRedisService
	RateLimit       *RateLimitService
}

func NewApp(
	AccountPostgres *AccountPostgresService,
	URLPostgres *URLPostgresService,
	AccountRedis *AccountRedisService,
	URLRedis *URLRedisService,
	RateLimit *RateLimitService,
) *App {
	return &App{AccountPostgres: AccountPostgres, URLPostgres: URLPostgres, AccountRedis: AccountRedis, URLRedis: URLRedis, RateLimit: RateLimit}
}
