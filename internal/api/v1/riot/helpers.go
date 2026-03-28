package riot

import (
	"ValorantAPI/internal/domain/valorant"
	"ValorantAPI/internal/http/response"
	"ValorantAPI/internal/riot/auth"
	redisstorage "ValorantAPI/internal/storage/redis"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// saveAccountAndSession сохраняет Riot сессию в Redis и создает/обновляет Valorant аккаунт.
func (h *Handler) saveAccountAndSession(c *gin.Context, session *auth.SessionData, userID int64) error {
	if err := h.deps.SessionRepo.SaveSession(c.Request.Context(), session.PUUID, session); err != nil {
		return err
	}
	return h.deps.ValorantSrv.CreateAccount(c.Request.Context(), &valorant.Account{
		UserID: userID,
		PUUID:  session.PUUID,
		Region: session.Region,
		Shard:  session.Shard,
	})
}

// respondPendingAuthError отвечает 401, если сессия не найдена, или 500 при другой ошибке.
func (h *Handler) respondPendingAuthError(c *gin.Context, err error) {
	status := http.StatusInternalServerError
	if errors.Is(err, redisstorage.ErrPendingAuthNotFound) {
		status = http.StatusUnauthorized
	}
	c.JSON(status, linkRiotAccountResponse{
		Response: response.Response{Error: &response.ErrorResponse{Message: "Auth session not found or expired"}},
	})
}

// handleAuthChallengeOrError отправляет подходящий ответ для подтверждения или неожиданной ошибки.
func (h *Handler) handleAuthChallengeOrError(c *gin.Context, riotAuth *auth.Client, err error) {
	var challengeErr *auth.ChallengeError
	if !errors.As(err, &challengeErr) {
		c.JSON(http.StatusInternalServerError, linkRiotAccountResponse{
			Response: response.Response{Error: &response.ErrorResponse{
				Message: "Riot auth failed",
				Details: err.Error(),
			}},
		})
		return
	}

	sessionID, genErr := redisstorage.GenerateRefreshTokenUUID()
	if genErr != nil {
		c.JSON(http.StatusInternalServerError, linkRiotAccountResponse{
			Response: response.Response{Error: &response.ErrorResponse{
				Message: "Failed to create auth session",
			}},
		})
		return
	}
	if saveErr := h.deps.PendingAuthRepo.Save(c.Request.Context(), sessionID, riotAuth.GetSessionCookies()); saveErr != nil {
		c.JSON(http.StatusInternalServerError, linkRiotAccountResponse{
			Response: response.Response{Error: &response.ErrorResponse{
				Message: "Failed to save auth session",
			}},
		})
		return
	}

	ch := challengeErr.Challenge
	resp := riotLoginChallengeResponse{SessionID: sessionID}

	switch ch.Type {
	case auth.ChallengeMFA:
		resp.Status = "mfa_required"
		resp.Email = ch.Email
		resp.Method = ch.Method
		resp.CodeLength = ch.CodeLength
	case auth.ChallengeCaptcha:
		resp.Status = "captcha_required"
		resp.HCaptchaKey = ch.HCaptchaKey
	}

	c.JSON(http.StatusOK, resp)
}

// completeAndRespond завершение пайплайна авторизации и сохранения сессии Riot
func (h *Handler) completeAndRespond(c *gin.Context, riotAuth *auth.Client, redirectURI string, userID int64) {
	session, err := riotAuth.CompleteAuth(c.Request.Context(), redirectURI)
	if err != nil {
		c.JSON(http.StatusInternalServerError, linkRiotAccountResponse{
			Response: response.Response{Error: &response.ErrorResponse{
				Message: "Failed to complete auth",
				Details: err.Error(),
			}},
		})
		return
	}

	if err := h.saveAccountAndSession(c, session, userID); err != nil {
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
