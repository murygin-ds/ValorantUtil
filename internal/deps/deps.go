package deps

import (
	"ValorantAPI/internal/config"
	"ValorantAPI/internal/domain/asset"
	domainmatch "ValorantAPI/internal/domain/match"
	"ValorantAPI/internal/domain/user"
	"ValorantAPI/internal/domain/valorant"
	"ValorantAPI/internal/logger"
	"ValorantAPI/internal/riot"
	"ValorantAPI/internal/riot/assets"
	"ValorantAPI/internal/riot/auth"
	"ValorantAPI/internal/storage/postgres"
	"context"
	"errors"
	"log"
	"net/http"

	redisClient "ValorantAPI/internal/storage/redis"
	redisstorage "ValorantAPI/internal/storage/redis"

	"github.com/redis/go-redis/v9"
)

// ErrSessionExpired возвращается NewRiotClient, когда срок действия токена доступа Riot истек
var ErrSessionExpired = errors.New("riot session expired")

type Deps struct {
	HTTPClient *http.Client
	Cfg        *config.Config
	Logging    *logger.Logger

	RedisClient     *redis.Client
	SessionRepo     *redisstorage.SessionRepo
	AuthTokenRepo   *redisstorage.AuthTokenRepo
	PendingAuthRepo *redisstorage.PendingAuthRepo
	OAuthStateRepo  *redisstorage.OAuthStateRepo
	StorefrontRepo  *redisstorage.StorefrontRepo
	MatchesRepo     *redisstorage.MatchesRepo
	PlayerNamesRepo *redisstorage.PlayerNamesRepo
	AccountRepo     *redisstorage.AccountRepo
	AccountMetaRepo *redisstorage.AccountMetaRepo

	UserSrv     *user.Service
	ValorantSrv *valorant.Service
	AssetSrv    *asset.Service
	MatchSrv    *domainmatch.Service

	AssetsClient *assets.Client
}

func New() *Deps {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal(err)
	}

	httpClient := &http.Client{
		Timeout: cfg.HTTPClient.Timeout,
		Transport: &http.Transport{
			MaxIdleConns:        cfg.HTTPClient.Transport.MaxIdleConns,
			IdleConnTimeout:     cfg.HTTPClient.Transport.IdleConnTimeout,
			DisableCompression:  cfg.HTTPClient.Transport.DisableCompression,
			TLSHandshakeTimeout: cfg.HTTPClient.Transport.TLSHandshakeTimeout,
		},
	}

	logging := logger.New(cfg.Logger.Level, cfg.Logger.FilePath)

	redisConn, err := redisClient.NewRedisClient(context.Background(), cfg.Redis)
	if err != nil {
		logging.Fatalw("failed to create redis client", "err", err)
	}
	if err := redisConn.Ping(context.Background()).Err(); err != nil {
		logging.Fatalw("failed to ping redis", "err", err)
	}
	sessionRepo := redisstorage.NewSessionRepo(redisConn)
	authTokenRepo := redisstorage.NewAuthTokenRepo(redisConn)
	pendingAuthRepo := redisstorage.NewPendingAuthRepo(redisConn)
	oauthStateRepo := redisstorage.NewOAuthStateRepo(redisConn)
	storefrontRepo := redisstorage.NewStorefrontRepo(redisConn)
	matchesRepo := redisstorage.NewMatchesRepo(redisConn)
	playerNamesRepo := redisstorage.NewPlayerNamesRepo(redisConn)
	accountRepo := redisstorage.NewAccountRepo(redisConn)
	accountMetaRepo := redisstorage.NewAccountMetaRepo(redisConn)

	pool, err := postgres.NewPostgresPool(context.Background(), cfg.Postgres)
	if err != nil {
		log.Fatal(err)
	}
	if err := pool.Ping(context.Background()); err != nil {
		logging.Fatalw("failed to ping postgres", "err", err)
	}
	userRepo := postgres.NewUserRepo(pool)
	userSrv := user.NewService(userRepo)
	valorantRepo := postgres.NewValorantAccountRepo(pool)
	valorantSrv := valorant.NewService(valorantRepo)
	assetRepo := postgres.NewAssetRepo(pool)
	assetsClient := assets.NewClient(cfg.Riot.AssetsAPIBaseURL, httpClient)
	assetSrv := asset.NewService(assetRepo, assetsClient, logging)
	matchRepo := postgres.NewMatchRepo(pool)
	matchSrv := domainmatch.NewService(matchRepo, logging)

	return &Deps{
		HTTPClient:      httpClient,
		Cfg:             cfg,
		Logging:         logging,
		RedisClient:     redisConn,
		SessionRepo:     sessionRepo,
		AuthTokenRepo:   authTokenRepo,
		PendingAuthRepo: pendingAuthRepo,
		OAuthStateRepo:  oauthStateRepo,
		StorefrontRepo:  storefrontRepo,
		MatchesRepo:     matchesRepo,
		PlayerNamesRepo: playerNamesRepo,
		AccountRepo:     accountRepo,
		AccountMetaRepo: accountMetaRepo,
		UserSrv:         userSrv,
		ValorantSrv:     valorantSrv,
		AssetsClient:    assetsClient,
		AssetSrv:        assetSrv,
		MatchSrv:        matchSrv,
	}
}

func (d *Deps) NewRiotClient(session *auth.SessionData) (*riot.Client, error) {
	client, err := riot.NewClient(
		d.HTTPClient,
		session.AccessToken,
		session.PUUID,
		session.Region,
		session.Shard,
	)
	if err != nil {
		if errors.Is(err, auth.ErrAccessTokenExpired) {
			_ = d.SessionRepo.DeleteSession(context.Background(), session.PUUID)
			return nil, ErrSessionExpired
		}
		return nil, err
	}
	return client, nil
}
