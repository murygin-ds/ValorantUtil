package riot

import (
	"ValorantAPI/internal/deps"
	"ValorantAPI/internal/http/response"
	"ValorantAPI/internal/riot/auth"
	redisstorage "ValorantAPI/internal/storage/redis"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	deps *deps.Deps
}

func NewHandler(deps *deps.Deps) *Handler {
	return &Handler{deps: deps}
}

// RiotAuthCallback godoc
// @Summary Привязка Riot аккаунта через OAuth
// @Description Принимает токены от клиентского OAuth-flow и привязывает аккаунт
// @Tags Riot
// @Accept json
// @Produce json
// @Param request body riotCallbackRequest true "OAuth токены"
// @Success 200 {object} linkRiotAccountResponse "Аккаунт успешно привязан"
// @Failure 400 {object} linkRiotAccountResponse "Некорректный запрос"
// @Failure 500 {object} linkRiotAccountResponse "Ошибка сервера"
// @Security CookieAuth
// @Router /v1/riot/callback [post]
func (h *Handler) RiotAuthCallback(c *gin.Context) {
	var req riotCallbackRequest
	var resp linkRiotAccountResponse

	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Error = &response.ErrorResponse{Message: "Invalid body"}
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	session, err := auth.NewClient(h.deps.HTTPClient).BuildSessionFromTokens(c.Request.Context(), req.AccessToken, req.IDToken)
	if err != nil {
		resp.Error = &response.ErrorResponse{
			Message: "Failed to build Riot session",
			Details: err.Error(),
		}
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	if err := h.saveAccountAndSession(c, session, c.GetInt64("user_id")); err != nil {
		resp.Error = &response.ErrorResponse{
			Message: "Failed to save account",
			Details: err.Error(),
		}
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	resp.Success = true
	resp.PUUID = session.PUUID
	resp.Region = session.Region
	c.JSON(http.StatusOK, resp)
}

// RiotLogin godoc
// @Summary Прямой вход через Riot аккаунт
// @Description Авторизует пользователя на серверах Riot по логину и паролю.
// Если сервер Riot требует капчу или 2FA - возвращает challenge со статусом и session_id для продолжения.
// @Tags Riot
// @Accept json
// @Produce json
// @Param request body riotLoginRequest true "Данные входа в Riot"
// @Success 200 {object} linkRiotAccountResponse "Аккаунт успешно привязан"
// @Success 200 {object} riotLoginChallengeResponse "Требуется капча или 2FA"
// @Failure 400 {object} linkRiotAccountResponse "Некорректный запрос"
// @Failure 500 {object} linkRiotAccountResponse "Ошибка сервера"
// @Security CookieAuth
// @Router /v1/riot/login [post]
func (h *Handler) RiotLogin(c *gin.Context) {
	var req riotLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, linkRiotAccountResponse{
			Response: response.Response{Error: &response.ErrorResponse{
				Message: "Invalid body",
			}},
		})
		return
	}

	riotAuth := auth.NewClient(h.deps.HTTPClient)

	if err := riotAuth.InitAuthCookies(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, linkRiotAccountResponse{
			Response: response.Response{Error: &response.ErrorResponse{
				Message: "Failed to initialize Riot session", Details: err.Error()}},
		})
		return
	}

	redirectURI, err := riotAuth.Authorize(c.Request.Context(), req.Username, req.Password, "")
	if err != nil {
		h.handleAuthChallengeOrError(c, riotAuth, err)
		return
	}

	h.completeAndRespond(c, riotAuth, redirectURI, c.GetInt64("user_id"))
}

// RiotLoginMFA godoc
// @Summary Подтверждение 2FA кода для входа в Riot
// @Description Принимает код из email/authenticator для завершения Riot-авторизации
// @Tags Riot
// @Accept json
// @Produce json
// @Param request body riotLoginMFARequest true "session_id из предыдущего шага и код 2FA"
// @Success 200 {object} linkRiotAccountResponse "Аккаунт успешно привязан"
// @Success 200 {object} riotLoginChallengeResponse "Неверный код - повторите попытку"
// @Failure 400 {object} linkRiotAccountResponse "Некорректный запрос"
// @Failure 401 {object} linkRiotAccountResponse "Сессия не найдена или истекла"
// @Failure 500 {object} linkRiotAccountResponse "Ошибка сервера"
// @Security CookieAuth
// @Router /v1/riot/login/mfa [post]
func (h *Handler) RiotLoginMFA(c *gin.Context) {
	var req riotLoginMFARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, linkRiotAccountResponse{
			Response: response.Response{Error: &response.ErrorResponse{Message: "Invalid body"}},
		})
		return
	}

	cookies, err := h.deps.PendingAuthRepo.Get(c.Request.Context(), req.SessionID)
	if err != nil {
		h.respondPendingAuthError(c, err)
		return
	}

	riotAuth := auth.NewClientWithCookies(h.deps.HTTPClient, cookies)

	redirectURI, err := riotAuth.SubmitMFA(c.Request.Context(), req.Code)
	if err != nil {
		// Сохранить обновленные куки, если выдается другое подтверждение
		_ = h.deps.PendingAuthRepo.Save(c.Request.Context(), req.SessionID, riotAuth.GetSessionCookies())
		h.handleAuthChallengeOrError(c, riotAuth, err)
		return
	}

	_ = h.deps.PendingAuthRepo.Delete(c.Request.Context(), req.SessionID)
	h.completeAndRespond(c, riotAuth, redirectURI, c.GetInt64("user_id"))
}

// RiotLoginCaptcha godoc
// @Summary Отправка решения капчи для входа в Riot
// @Description Повторяет авторизацию с решенной hcaptcha
// @Tags Riot
// @Accept json
// @Produce json
// @Param request body riotLoginCaptchaRequest true "session_id, учетные данные и токен hcaptcha"
// @Success 200 {object} linkRiotAccountResponse "Аккаунт успешно привязан"
// @Success 200 {object} riotLoginChallengeResponse "Требуется 2FA после капчи"
// @Failure 400 {object} linkRiotAccountResponse "Некорректный запрос"
// @Failure 401 {object} linkRiotAccountResponse "Сессия не найдена или истекла"
// @Failure 500 {object} linkRiotAccountResponse "Ошибка сервера"
// @Security CookieAuth
// @Router /v1/riot/login/captcha [post]
func (h *Handler) RiotLoginCaptcha(c *gin.Context) {
	var req riotLoginCaptchaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, linkRiotAccountResponse{
			Response: response.Response{Error: &response.ErrorResponse{Message: "Invalid body"}},
		})
		return
	}

	cookies, err := h.deps.PendingAuthRepo.Get(c.Request.Context(), req.SessionID)
	if err != nil {
		h.respondPendingAuthError(c, err)
		return
	}

	riotAuth := auth.NewClientWithCookies(h.deps.HTTPClient, cookies)

	captchaToken := "hcaptcha " + req.CaptchaToken
	redirectURI, err := riotAuth.Authorize(c.Request.Context(), req.Username, req.Password, captchaToken)
	if err != nil {
		// Сохранить обновленные куки для последующих шагов (например, каптча пройдена -> MFA следующий)
		_ = h.deps.PendingAuthRepo.Save(c.Request.Context(), req.SessionID, riotAuth.GetSessionCookies())
		h.handleAuthChallengeOrError(c, riotAuth, err)
		return
	}

	_ = h.deps.PendingAuthRepo.Delete(c.Request.Context(), req.SessionID)
	h.completeAndRespond(c, riotAuth, redirectURI, c.GetInt64("user_id"))
}

// GetAuthURL godoc
// @Summary Получить ссылку для авторизации через Riot
// @Description Генерирует OAuth URL на страницу входа Riot. Откройте эту ссылку в браузере,
// пройдите авторизацию, а затем скопируйте URL из адресной строки (начинается с https://playvalorant.com/opt_in#...)
// и отправьте его через POST /v1/riot/auth/submit-url.
// @Tags Riot
// @Produce json
// @Success 200 {object} authURLResponse "OAuth ссылка сформирована"
// @Failure 500 {object} authURLResponse "Ошибка генерации ссылки"
// @Security CookieAuth
// @Router /v1/riot/auth/url [get]
func (h *Handler) GetAuthURL(c *gin.Context) {
	nonce, err := redisstorage.GenerateRefreshTokenUUID()
	if err != nil {
		c.JSON(http.StatusInternalServerError, authURLResponse{
			Response: response.Response{Error: &response.ErrorResponse{
				Message: "Failed to generate nonce",
			}},
		})
		return
	}

	authURL := buildRiotAuthURL(nonce)
	c.JSON(http.StatusOK, authURLResponse{
		Response: response.Response{Success: true},
		AuthURL:  authURL,
	})
}

// SubmitRedirectURL godoc
// @Summary Завершить OAuth-привязку через вставку URL
// @Description Принимает URL, скопированный из браузера после авторизации через Riot.
// Токены извлекаются из фрагмента URL и используются для привязки аккаунта Valorant.
// @Tags Riot
// @Accept json
// @Produce json
// @Param request body submitRedirectURLRequest true "URL из браузера после авторизации"
// @Success 200 {object} linkRiotAccountResponse "Аккаунт успешно привязан"
// @Failure 400 {object} linkRiotAccountResponse "Некорректный запрос или токен не найден"
// @Failure 500 {object} linkRiotAccountResponse "Ошибка сервера"
// @Security CookieAuth
// @Router /v1/riot/auth/submit-url [post]
func (h *Handler) SubmitRedirectURL(c *gin.Context) {
	var req submitRedirectURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, linkRiotAccountResponse{
			Response: response.Response{Error: &response.ErrorResponse{
				Message: "Invalid body",
			}},
		})
		return
	}

	accessToken, idToken, err := parseRedirectURL(req.RedirectURL)
	if err != nil {
		c.JSON(http.StatusBadRequest, linkRiotAccountResponse{
			Response: response.Response{Error: &response.ErrorResponse{
				Message: "Invalid redirect URL",
			}},
		})
		return
	}
	if accessToken == "" {
		c.JSON(http.StatusBadRequest, linkRiotAccountResponse{
			Response: response.Response{Error: &response.ErrorResponse{
				Message: "Токен не найден в URL. Убедитесь, что вы скопировали полный URL после входа в Riot.",
			}},
		})
		return
	}

	session, err := auth.NewClient(h.deps.HTTPClient).BuildSessionFromTokens(c.Request.Context(), accessToken, idToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, linkRiotAccountResponse{
			Response: response.Response{Error: &response.ErrorResponse{
				Message: "Failed to build Riot session",
				Details: err.Error(),
			}},
		})
		return
	}

	if err := h.saveAccountAndSession(c, session, c.GetInt64("user_id")); err != nil {
		c.JSON(http.StatusInternalServerError, linkRiotAccountResponse{
			Response: response.Response{Error: &response.ErrorResponse{
				Message: "Failed to save account",
				Details: err.Error(),
			}},
		})
		return
	}

	c.JSON(http.StatusOK, linkRiotAccountResponse{
		Response: response.Response{Success: true},
		PUUID:    session.PUUID,
		Region:   session.Region,
	})
}
