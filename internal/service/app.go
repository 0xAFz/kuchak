package service

type App struct {
	AccountPostgres *AccountPostgresService
	URLPostgres     *URLPostgresService
	AccountRedis    *AccountRedisService
	URLRedis        *URLRedisService
}

func NewApp(
	AccountPostgres *AccountPostgresService,
	URLPostgres *URLPostgresService,
	AccountRedis *AccountRedisService,
	URLRedis *URLRedisService,
) *App {
	return &App{AccountPostgres: AccountPostgres, URLPostgres: URLPostgres, AccountRedis: AccountRedis, URLRedis: URLRedis}
}
