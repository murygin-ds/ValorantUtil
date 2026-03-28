package users

import (
	"ValorantAPI/internal/deps"
	"ValorantAPI/internal/domain/user"
	"ValorantAPI/internal/http/response"
	"ValorantAPI/internal/pkg/hash"
	"ValorantAPI/internal/pkg/jwt"
	"ValorantAPI/internal/storage/postgres"
	redisstorage "ValorantAPI/internal/storage/redis"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	deps *deps.Deps
}

func NewHandler(deps *deps.Deps) *Handler {
	return &Handler{deps: deps}
}

// Register godoc
// @Summary Регистрация нового пользователя
// @Description Создает нового пользователя с указанным логином и паролем, устанавливает cookies авторизации
// @Tags Users
// @Accept json
// @Produce json
// @Param request body authUserRequest true "Данные для регистрации пользователя"
// @Success 201 {object} authUserResponse "Пользователь успешно зарегистрирован"
// @Failure 400 {object} authUserResponse "Некорректные данные в запросе"
// @Failure 500 {object} authUserResponse "Ошибка сервера при создании пользователя"
// @Router /v1/users/register [post]
func (h *Handler) Register(c *gin.Context) {
	var resp authUserResponse
	var req authUserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Error = &response.ErrorResponse{Message: "Invalid body"}
		c.JSON(http.StatusBadRequest, resp)
		return
	}
	u := user.User{
		Login:    req.Login,
		Password: req.Password,
	}
	if err := h.deps.UserSrv.CreateUser(c.Request.Context(), &u); err != nil {
		if errors.Is(err, postgres.ErrLoginAlreadyTaken) {
			resp.Error = &response.ErrorResponse{Message: "Login is already taken"}
			c.JSON(http.StatusBadRequest, resp)
			return
		}
		resp.Error = &response.ErrorResponse{Message: "Failed to create user", Details: err.Error()}
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	accessToken, err := jwt.Generate(u.ID, h.deps.Cfg.Security.Secret, accessTokenDuration)
	if err != nil {
		resp.Error = &response.ErrorResponse{Message: "Failed to generate token"}
		c.JSON(http.StatusInternalServerError, resp)
		return
	}
	refreshUUID, err := redisstorage.GenerateRefreshTokenUUID()
	if err != nil {
		resp.Error = &response.ErrorResponse{Message: "Failed to generate refresh token"}
		c.JSON(http.StatusInternalServerError, resp)
		return
	}
	if err := h.deps.AuthTokenRepo.SaveRefreshToken(c.Request.Context(), refreshUUID, u.ID); err != nil {
		resp.Error = &response.ErrorResponse{Message: "Failed to save refresh token"}
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	setAuthCookies(c.Writer, accessToken, refreshUUID)
	resp.Success = true
	resp.Login = u.Login
	c.JSON(http.StatusCreated, resp)
}

// Login godoc
// @Summary Аутентификация пользователя
// @Description Выполняет вход пользователя по логину и паролю, устанавливает cookies авторизации
// @Tags Users
// @Accept json
// @Produce json
// @Param request body authUserRequest true "Данные для входа пользователя"
// @Success 200 {object} authUserResponse "Пользователь успешно авторизован"
// @Failure 400 {object} authUserResponse "Некорректные данные в запросе"
// @Failure 401 {object} authUserResponse "Неверный логин или пароль"
// @Failure 500 {object} authUserResponse "Ошибка сервера"
// @Router /v1/users/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req authUserRequest
	var resp authUserResponse

	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Error = &response.ErrorResponse{Message: "Invalid body"}
		c.JSON(http.StatusBadRequest, resp)
		return
	}
	u := &user.User{
		Login:    req.Login,
		Password: req.Password,
	}
	if err := h.deps.UserSrv.GetUserByLogin(c.Request.Context(), u); err != nil {
		if errors.Is(err, postgres.ErrUserNotFound) {
			resp.Error = &response.ErrorResponse{Message: "Invalid login or password"}
			c.JSON(http.StatusUnauthorized, resp)
			return
		}
		resp.Error = &response.ErrorResponse{Message: "Failed to get user", Details: err.Error()}
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	if !hash.CheckPassword(req.Password, u.Password) {
		resp.Error = &response.ErrorResponse{Message: "Invalid login or password"}
		c.JSON(http.StatusUnauthorized, resp)
		return
	}

	accessToken, err := jwt.Generate(u.ID, h.deps.Cfg.Security.Secret, accessTokenDuration)
	if err != nil {
		resp.Error = &response.ErrorResponse{Message: "Failed to generate token"}
		c.JSON(http.StatusInternalServerError, resp)
		return
	}
	refreshUUID, err := redisstorage.GenerateRefreshTokenUUID()
	if err != nil {
		resp.Error = &response.ErrorResponse{Message: "Failed to generate refresh token"}
		c.JSON(http.StatusInternalServerError, resp)
		return
	}
	if err := h.deps.AuthTokenRepo.SaveRefreshToken(c.Request.Context(), refreshUUID, u.ID); err != nil {
		resp.Error = &response.ErrorResponse{Message: "Failed to save refresh token"}
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	setAuthCookies(c.Writer, accessToken, refreshUUID)
	resp.Success = true
	resp.Login = u.Login
	c.JSON(http.StatusOK, resp)
}

// Logout godoc
// @Summary Выход из системы
// @Description Удаляет refresh токен из Redis и очищает cookies авторизации
// @Tags Users
// @Produce json
// @Success 200 {object} response.Response "Выход выполнен успешно"
// @Security CookieAuth
// @Router /v1/users/logout [post]
func (h *Handler) Logout(c *gin.Context) {
	if refreshUUID, err := c.Cookie("refresh_token"); err == nil && refreshUUID != "" {
		_ = h.deps.AuthTokenRepo.DeleteRefreshToken(c.Request.Context(), refreshUUID)
	}
	clearAuthCookies(c.Writer)
	c.JSON(http.StatusOK, response.Response{Success: true})
}

// Refresh godoc
// @Summary Обновление access токена
// @Description Обновляет access токен используя refresh токен из cookie. Ротирует refresh токен.
// @Tags Users
// @Produce json
// @Success 200 {object} response.Response "Токен успешно обновлен"
// @Failure 401 {object} response.Response "Refresh токен отсутствует или недействителен"
// @Failure 500 {object} response.Response "Ошибка сервера"
// @Router /v1/users/refresh [post]
func (h *Handler) Refresh(c *gin.Context) {
	refreshUUID, err := c.Cookie("refresh_token")
	if err != nil || refreshUUID == "" {
		c.JSON(http.StatusUnauthorized, response.Response{
			Error: &response.ErrorResponse{Message: "Refresh token is required"},
		})
		return
	}

	userID, err := h.deps.AuthTokenRepo.GetRefreshToken(c.Request.Context(), refreshUUID)
	if err != nil {
		if errors.Is(err, redisstorage.ErrRefreshTokenNotFound) {
			c.JSON(http.StatusUnauthorized, response.Response{
				Error: &response.ErrorResponse{Message: "Invalid or expired refresh token"},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, response.Response{
			Error: &response.ErrorResponse{Message: "Failed to validate refresh token"},
		})
		return
	}

	// Ротация: удаляем старый токен, выдаем новый
	_ = h.deps.AuthTokenRepo.DeleteRefreshToken(c.Request.Context(), refreshUUID)
	newRefreshUUID, err := redisstorage.GenerateRefreshTokenUUID()
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Response{
			Error: &response.ErrorResponse{Message: "Failed to generate refresh token"},
		})
		return
	}
	if err := h.deps.AuthTokenRepo.SaveRefreshToken(c.Request.Context(), newRefreshUUID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, response.Response{
			Error: &response.ErrorResponse{Message: "Failed to save refresh token"},
		})
		return
	}

	accessToken, err := jwt.Generate(userID, h.deps.Cfg.Security.Secret, accessTokenDuration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Response{
			Error: &response.ErrorResponse{Message: "Failed to generate token"},
		})
		return
	}

	setAuthCookies(c.Writer, accessToken, newRefreshUUID)
	c.JSON(http.StatusOK, response.Response{Success: true})
}

// Me godoc
// @Summary Информация о текущем пользователе
// @Description Возвращает данные авторизованного пользователя
// @Tags Users
// @Produce json
// @Success 200 {object} meResponse "Данные пользователя"
// @Failure 500 {object} meResponse "Ошибка сервера"
// @Security CookieAuth
// @Router /v1/users/me [get]
func (h *Handler) Me(c *gin.Context) {
	var resp meResponse

	userID := c.GetInt64("user_id")
	u := &user.User{ID: userID}
	if err := h.deps.UserSrv.GetUserByID(c.Request.Context(), u); err != nil {
		resp.Error = &response.ErrorResponse{Message: "Failed to get user", Details: err.Error()}
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	resp.Success = true
	resp.ID = u.ID
	resp.Login = u.Login
	c.JSON(http.StatusOK, resp)
}

// GetAccounts godoc
// @Summary Получение списка аккаунтов пользователя
// @Description Возвращает список аккаунтов Valorant, принадлежащих авторизованному пользователю.
// Поля game_name и tag_line заполняются из Redis-кэша (TTL 7 дней, обновляется при просмотре матчей).
// Поля tier и rr заполняются из Redis-кэша (TTL 1 час, обновляются при вызове GET /v1/valorant/mmr/{puuid}).
// Если данные еще не закешированы - поля будут пустыми / нулевыми.
// @Tags Users
// @Produce json
// @Param limit query int false "Максимальное количество записей" default(100)
// @Param offset query int false "Смещение для пагинации" default(0)
// @Success 200 {object} getAccountsResponse "Список аккаунтов успешно получен"
// @Failure 500 {object} getAccountsResponse "Ошибка сервера"
// @Security CookieAuth
// @Router /v1/users/accounts [get]
func (h *Handler) GetAccounts(c *gin.Context) {
	var resp getAccountsResponse

	userID := c.GetInt64("user_id")
	limit, err := strconv.Atoi(c.Query("limit"))
	if err != nil {
		limit = 100
	}
	offset, err := strconv.Atoi(c.Query("offset"))
	if err != nil {
		offset = 0
	}

	accounts, err := h.deps.ValorantSrv.GetAccountsList(c.Request.Context(), int(userID), limit, offset)
	if err != nil {
		resp.Error = &response.ErrorResponse{Message: "Failed to get accounts list", Details: err.Error()}
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	ctx := c.Request.Context()
	dtos := make([]accountDTO, 0, len(accounts))
	for _, acc := range accounts {
		dto := accountDTO{
			ID:     acc.ID,
			PUUID:  acc.PUUID,
			Region: acc.Region,
			Shard:  acc.Shard,
		}

		if name, nameErr := h.deps.PlayerNamesRepo.Get(ctx, acc.PUUID); nameErr == nil {
			if parts := strings.SplitN(name, "#", 2); len(parts) == 2 {
				dto.GameName = parts[0]
				dto.TagLine = parts[1]
			}
		}

		if meta, metaErr := h.deps.AccountMetaRepo.Get(ctx, acc.PUUID); metaErr == nil {
			dto.Tier = meta.Tier
			dto.RR = meta.RR
		}

		dtos = append(dtos, dto)
	}

	resp.Success = true
	resp.Accounts = dtos
	c.JSON(http.StatusOK, resp)
}
