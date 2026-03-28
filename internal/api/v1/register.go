package v1

import (
	"ValorantAPI/internal/api/v1/riot"
	"ValorantAPI/internal/api/v1/users"
	"ValorantAPI/internal/api/v1/valorant"
	"ValorantAPI/internal/deps"
	"ValorantAPI/internal/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(
	router *gin.RouterGroup,
	deps *deps.Deps,
) {
	securedGroup := router.Group("/")
	securedGroup.Use(middleware.Auth(deps.Cfg.Security.Secret))

	usersHandler := users.NewHandler(deps)

	// Публичные маршруты пользователей
	usersPublic := router.Group("/users")
	{
		usersPublic.POST("/register", usersHandler.Register)
		usersPublic.POST("/login", usersHandler.Login)
		// Refresh НЕ использует middleware аутентификации - access token истек к моменту вызова
		usersPublic.POST("/refresh", usersHandler.Refresh)
	}

	// Защищенные маршруты пользователей
	usersSecured := securedGroup.Group("/users")
	{
		usersSecured.GET("/me", usersHandler.Me)
		usersSecured.POST("/logout", usersHandler.Logout)
		usersSecured.GET("/accounts", usersHandler.GetAccounts)
	}

	// Маршруты Riot (все защищены)
	riotHandler := riot.NewHandler(deps)
	riotGroup := securedGroup.Group("/riot")
	{
		// Устаревший OAuth callback (клиент отправляет токены напрямую)
		riotGroup.POST("/callback", riotHandler.RiotAuthCallback)
		// Прямой вход с учетными данными Riot (имя пользователя + пароль)
		riotGroup.POST("/login", riotHandler.RiotLogin)
		riotGroup.POST("/login/mfa", riotHandler.RiotLoginMFA)
		riotGroup.POST("/login/captcha", riotHandler.RiotLoginCaptcha)
		// OAuth поток на основе URL (вставьте URL из браузера)
		riotGroup.GET("/auth/url", riotHandler.GetAuthURL)
		riotGroup.POST("/auth/submit-url", riotHandler.SubmitRedirectURL)
	}

	// Маршруты Valorant (все защищены)
	valorantHandler := valorant.NewHandler(deps)
	valorantGroup := securedGroup.Group("/valorant")
	{
		valorantGroup.GET("/store/:puuid", valorantHandler.GetUserStore)
		valorantGroup.GET("/wallet/:puuid", valorantHandler.GetWallet)
		valorantGroup.GET("/mmr/:puuid", valorantHandler.GetMMR)
		valorantGroup.GET("/matches/:puuid", valorantHandler.GetMatchHistory)
		valorantGroup.GET("/account/:puuid", valorantHandler.GetAccount)
	}
}
