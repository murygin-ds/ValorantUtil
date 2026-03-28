package valorant

import (
	"ValorantAPI/internal/deps"
	domainmatch "ValorantAPI/internal/domain/match"
	"ValorantAPI/internal/http/response"
	riotmatch "ValorantAPI/internal/riot/match"
	"ValorantAPI/internal/riot/mmr"
	"ValorantAPI/internal/riot/store"
	"ValorantAPI/internal/riot/wallet"
	redisstorage "ValorantAPI/internal/storage/redis"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	deps *deps.Deps
}

func NewHandler(deps *deps.Deps) *Handler {
	return &Handler{deps: deps}
}

// GetUserStore godoc
// @Summary Получить ежедневный магазин пользователя
// @Description Получает информацию о ежедневном магазине для конкретного пользователя по его PUUID
// @Tags Valorant
// @Produce json
// @Param force query bool false "Принудительное обновление (игнорировать кэш)"
// @Param puuid path string true "UUID игрока"
// @Success 200 {object} getDailyStoreResponse "Магазин успешно получен"
// @Failure 401 {object} getDailyStoreResponse "Не авторизован - сессия истекла или недействительна"
// @Failure 500 {object} getDailyStoreResponse "Внутренняя ошибка сервера"
// @Security CookieAuth
// @Router /v1/valorant/store/{puuid} [get]
func (h *Handler) GetUserStore(c *gin.Context) {
	var resp getDailyStoreResponse

	puuid := c.Param("puuid")

	force := c.Query("force") == "true"
	if force {
		h.invalidateStorefrontCache(c, puuid)
	}

	if !force {
		// Получаем данные из кэша
		if cached, err := h.deps.StorefrontRepo.Get(c.Request.Context(), puuid); err == nil {
			if jsonErr := json.Unmarshal(cached, &resp); jsonErr == nil {
				c.JSON(http.StatusOK, resp)
				return
			}
		}
	}

	session, err := h.deps.SessionRepo.GetSession(c.Request.Context(), puuid)
	if err != nil {
		resp.Error = &response.ErrorResponse{Message: "Login to your Riot account again"}
		c.JSON(http.StatusUnauthorized, resp)
		return
	}

	riotClient, err := h.deps.NewRiotClient(session)
	if err != nil {
		if errors.Is(err, deps.ErrSessionExpired) {
			resp.Error = &response.ErrorResponse{Message: "Login to your Riot account again"}
			c.JSON(http.StatusUnauthorized, resp)
		} else {
			resp.Error = &response.ErrorResponse{Message: "Failed to create Riot client", Details: err.Error()}
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	storeFront, err := store.NewClient(riotClient).GetStorefront(c.Request.Context())
	if err != nil {
		resp.Error = &response.ErrorResponse{Message: "Failed to retrieve storefront", Details: err.Error()}
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	resp.Store = h.enrichOffers(c, storeFront.SkinsPanelLayout.SingleItemStoreOffers)
	resp.Accessories = h.enrichAccessoryOffers(c, storeFront.AccessoryStore.AccessoryStoreOffers)
	resp.Bundles = h.enrichBundles(c, storeFront.FeaturedBundle)
	resp.Success = true

	// Кешируем обогащенный ответ на время жизни самого короткого раздела витрины.
	if ttl := storefrontTTL(storeFront); ttl > 0 {
		if data, jsonErr := json.Marshal(resp); jsonErr == nil {
			if saveErr := h.deps.StorefrontRepo.Save(c.Request.Context(), puuid, data, ttl); saveErr != nil {
				h.deps.Logging.Warnw("failed to cache storefront", "puuid", puuid, "err", saveErr)
			}
		}
	}

	c.JSON(http.StatusOK, resp)
}

// GetWallet godoc
// @Summary Получить баланс кошелька игрока
// @Description Возвращает количество Valorant Points (VP), Radianite Points и Kingdom Credits игрока
// @Tags Valorant
// @Produce json
// @Param puuid path string true "UUID игрока"
// @Success 200 {object} getWalletResponse "Баланс успешно получен"
// @Failure 401 {object} getWalletResponse "Не авторизован - сессия истекла"
// @Failure 500 {object} getWalletResponse "Внутренняя ошибка сервера"
// @Security CookieAuth
// @Router /v1/valorant/wallet/{puuid} [get]
func (h *Handler) GetWallet(c *gin.Context) {
	var resp getWalletResponse

	puuid := c.Param("puuid")

	session, err := h.deps.SessionRepo.GetSession(c.Request.Context(), puuid)
	if err != nil {
		resp.Error = &response.ErrorResponse{Message: "Login to your Riot account again"}
		c.JSON(http.StatusUnauthorized, resp)
		return
	}

	riotClient, err := h.deps.NewRiotClient(session)
	if err != nil {
		if errors.Is(err, deps.ErrSessionExpired) {
			resp.Error = &response.ErrorResponse{Message: "Login to your Riot account again"}
			c.JSON(http.StatusUnauthorized, resp)
		} else {
			resp.Error = &response.ErrorResponse{Message: "Failed to create Riot client", Details: err.Error()}
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	walletData, err := wallet.NewClient(riotClient).GetWallet(c.Request.Context())
	if err != nil {
		resp.Error = &response.ErrorResponse{Message: "Failed to retrieve wallet", Details: err.Error()}
		c.JSON(http.StatusInternalServerError, resp)
		return
	}
	resp.Success = true
	resp.Wallet = walletData
	c.JSON(http.StatusOK, resp)
}

// GetMMR godoc
// @Summary Получить MMR и ранг игрока
// @Description Возвращает информацию о ранге, рейтинге и последнем конкурентном обновлении игрока
// @Tags Valorant
// @Produce json
// @Param puuid path string true "UUID игрока"
// @Success 200 {object} getMMRResponse "MMR успешно получен"
// @Failure 401 {object} getMMRResponse "Не авторизован - сессия истекла"
// @Failure 500 {object} getMMRResponse "Внутренняя ошибка сервера"
// @Security CookieAuth
// @Router /v1/valorant/mmr/{puuid} [get]
func (h *Handler) GetMMR(c *gin.Context) {
	var resp getMMRResponse

	puuid := c.Param("puuid")

	session, err := h.deps.SessionRepo.GetSession(c.Request.Context(), puuid)
	if err != nil {
		resp.Error = &response.ErrorResponse{Message: "Login to your Riot account again"}
		c.JSON(http.StatusUnauthorized, resp)
		return
	}

	riotClient, err := h.deps.NewRiotClient(session)
	if err != nil {
		if errors.Is(err, deps.ErrSessionExpired) {
			resp.Error = &response.ErrorResponse{Message: "Login to your Riot account again"}
			c.JSON(http.StatusUnauthorized, resp)
		} else {
			resp.Error = &response.ErrorResponse{Message: "Failed to create Riot client", Details: err.Error()}
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	mmrData, err := mmr.NewClient(riotClient).GetMMR(c.Request.Context())
	if err != nil {
		resp.Error = &response.ErrorResponse{Message: "Failed to retrieve MMR", Details: err.Error()}
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	last := mmrData.LatestCompetitiveUpdate
	resp.Success = true
	resp.MMR = &mmrInfo{
		Tier:         last.TierAfterUpdate,
		RankedRating: last.RankedRatingAfterUpdate,
		LastMatchID:  last.MatchID,
		LastMapID:    last.MapID,
		RRChange:     last.RankedRatingEarned,
	}

	if saveErr := h.deps.AccountMetaRepo.Save(c.Request.Context(), puuid, redisstorage.AccountMeta{
		Tier: last.TierAfterUpdate,
		RR:   last.RankedRatingAfterUpdate,
	}); saveErr != nil {
		h.deps.Logging.Warnw("failed to cache account rank", "puuid", puuid, "err", saveErr)
	}

	c.JSON(http.StatusOK, resp)
}

// GetMatchHistory godoc
// @Summary История матчей игрока
// @Description Возвращает последние матчи со статистикой игроков и дуэлями.
// При force=true игнорирует кэш и синхронизирует новые матчи с Riot API.
// @Tags Valorant
// @Produce json
// @Param puuid path string true "UUID игрока"
// @Param force query bool false "Принудительное обновление (игнорировать кэш)"
// @Success 200 {object} getMatchHistoryResponse
// @Failure 401 {object} getMatchHistoryResponse
// @Failure 500 {object} getMatchHistoryResponse
// @Security CookieAuth
// @Router /v1/valorant/matches/{puuid} [get]
func (h *Handler) GetMatchHistory(c *gin.Context) {
	var resp getMatchHistoryResponse
	puuid := c.Param("puuid")
	force := c.Query("force") == "true"

	session, err := h.deps.SessionRepo.GetSession(c.Request.Context(), puuid)
	if err != nil {
		resp.Error = &response.ErrorResponse{Message: "Login to your Riot account again"}
		c.JSON(http.StatusUnauthorized, resp)
		return
	}

	riotClient, err := h.deps.NewRiotClient(session)
	if err != nil {
		if errors.Is(err, deps.ErrSessionExpired) {
			resp.Error = &response.ErrorResponse{Message: "Login to your Riot account again"}
			c.JSON(http.StatusUnauthorized, resp)
		} else {
			resp.Error = &response.ErrorResponse{Message: "Failed to create Riot client", Details: err.Error()}
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	// force=true: синхронизируем новые матчи из Riot API, затем возвращаем все.
	// force=false: отдаем из PostgreSQL кеша (данные матчей хранятся постоянно).
	var matches []domainmatch.Match
	if force {
		matches, err = h.deps.MatchSrv.SyncAndGet(c.Request.Context(), riotmatch.NewClient(riotClient), puuid)
	} else {
		matches, err = h.deps.MatchSrv.GetCached(c.Request.Context(), puuid)
	}
	if err != nil {
		resp.Error = &response.ErrorResponse{Message: "Failed to retrieve match history", Details: err.Error()}
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	resp.Success = true
	resp.Matches = make([]matchDTO, 0, len(matches))
	for _, m := range matches {
		dto := toMatchDTO(m)
		h.enrichMatchPlayers(c, riotClient, dto.Players)
		resp.Matches = append(resp.Matches, dto)
	}

	c.JSON(http.StatusOK, resp)
}

// GetAccount godoc
// @Summary Получить скины Valorant аккаунта
// @Description Возвращает список скинов игрока с именем и иконкой.
// Кэшируется постоянно (скины можно только добавить, не убрать).
// При force=true игнорирует кэш и обновляет данные.
// @Tags Valorant
// @Produce json
// @Param puuid path string true "UUID игрока"
// @Param force query bool false "Принудительное обновление"
// @Success 200 {object} getAccountResponse
// @Failure 401 {object} getAccountResponse
// @Failure 500 {object} getAccountResponse
// @Security CookieAuth
// @Router /v1/valorant/account/{puuid} [get]
func (h *Handler) GetAccount(c *gin.Context) {
	var resp getAccountResponse
	puuid := c.Param("puuid")
	force := c.Query("force") == "true"

	// Постоянный кеш - скины можно только добавлять, убрать нельзя.
	if !force {
		if cached, err := h.deps.AccountRepo.Get(c.Request.Context(), puuid); err == nil {
			if jsonErr := json.Unmarshal(cached, &resp); jsonErr == nil {
				c.JSON(http.StatusOK, resp)
				return
			}
		}
	}

	session, err := h.deps.SessionRepo.GetSession(c.Request.Context(), puuid)
	if err != nil {
		resp.Error = &response.ErrorResponse{Message: "Login to your Riot account again"}
		c.JSON(http.StatusUnauthorized, resp)
		return
	}

	riotClient, err := h.deps.NewRiotClient(session)
	if err != nil {
		if errors.Is(err, deps.ErrSessionExpired) {
			resp.Error = &response.ErrorResponse{Message: "Login to your Riot account again"}
			c.JSON(http.StatusUnauthorized, resp)
		} else {
			resp.Error = &response.ErrorResponse{Message: "Failed to create Riot client", Details: err.Error()}
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	ctx := c.Request.Context()
	resp.Skins = h.resolveSkins(c, riotClient)
	resp.Success = true

	if data, jsonErr := json.Marshal(resp); jsonErr == nil {
		if saveErr := h.deps.AccountRepo.Save(ctx, puuid, data); saveErr != nil {
			h.deps.Logging.Warnw("failed to cache account skins", "puuid", puuid, "err", saveErr)
		}
	}

	c.JSON(http.StatusOK, resp)
}
