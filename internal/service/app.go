package service

type App struct {
	AccountPostgres *AccountPostgresService
	URLPostgres     *URLPostgresService
	AccountRedis    *AccountRedisService
	URLRedis        *URLRedisService
	RateLimit       *RateLimitService
	EmailSender     *EmailService
}

func NewApp(
	AccountPostgres *AccountPostgresService,
	URLPostgres *URLPostgresService,
	AccountRedis *AccountRedisService,
	URLRedis *URLRedisService,
	RateLimit *RateLimitService,
	EmailSender *EmailService,
) *App {
	return &App{AccountPostgres: AccountPostgres, URLPostgres: URLPostgres, AccountRedis: AccountRedis, URLRedis: URLRedis, RateLimit: RateLimit, EmailSender: EmailSender}
}
