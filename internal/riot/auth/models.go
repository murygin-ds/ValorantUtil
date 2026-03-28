package auth

import "net/http"

type SessionData struct {
	AccessToken      string
	IDToken          string
	EntitlementToken string
	PUUID            string
	Region           string
	Shard            string
	Version          string
	Cookies          []*http.Cookie
}

type playerInfoResponse struct {
	Sub string `json:"sub"` // puuid игрока
}

type entitlementResponse struct {
	EntitlementsToken string `json:"entitlements_token"`
}

// authRequestBody используется для первоначальной отправки учетных данных.
type authRequestBody struct {
	Type         string       `json:"type"`
	Language     string       `json:"language"`
	Remember     bool         `json:"remember"`
	RiotIdentity riotIdentity `json:"riot_identity"`
}

type riotIdentity struct {
	Captcha  string `json:"captcha"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// mfaRequestBody используется для отправки кода 2FA.
type mfaRequestBody struct {
	Type           string `json:"type"`
	Code           string `json:"code"`
	RememberDevice bool   `json:"rememberDevice"`
}

// authRequestResponse представляет все возможные ответы от эндпоинта авторизации Riot.
type authRequestResponse struct {
	Type        string           `json:"type"`
	Success     *authSuccess     `json:"response,omitempty"` // случай "success"
	Error       string           `json:"error,omitempty"`
	Multifactor *multifactorInfo `json:"multifactor,omitempty"` // случай "multifactor"
	Captcha     *captchaInfo     `json:"captcha,omitempty"`     // случай "auth"
	Country     string           `json:"country,omitempty"`
}

type authSuccess struct {
	Mode       string     `json:"mode"`
	Parameters authParams `json:"parameters"`
}

type authParams struct {
	URI string `json:"uri"` // redirect URL с токенами во фрагменте
}

// multifactorInfo возвращается, когда требуется 2FA.
type multifactorInfo struct {
	Email                 string `json:"email"`
	Method                string `json:"method"`
	MultiFactorCodeLength int    `json:"multiFactorCodeLength"`
	MfaVersion            string `json:"mfaVersion"`
}

// captchaInfo возвращается, когда требуется hcaptcha.
type captchaInfo struct {
	Type     string        `json:"type"`
	HCaptcha *hcaptchaInfo `json:"hcaptcha,omitempty"`
}

type hcaptchaInfo struct {
	Key  string `json:"key"`
	Data string `json:"data"`
}

type riotGeoBody struct {
	IDToken string `json:"id_token"`
}

type riotGeoResponse struct {
	Token      string     `json:"token"`
	Affinities affinities `json:"affinities"`
}

type affinities struct {
	Live string `json:"live"`
	PBE  string `json:"pbe"`
}

// AuthChallengeType описывает тип challenge, выданного сервером авторизации Riot.
type AuthChallengeType string

const (
	ChallengeCaptcha AuthChallengeType = "captcha"
	ChallengeMFA     AuthChallengeType = "mfa"
)

// AuthChallenge возвращается вызывающей стороне, когда процесс авторизации требует дополнительного ввода.
type AuthChallenge struct {
	Type AuthChallengeType
	// Поля MFA
	Email      string
	Method     string
	CodeLength int
	// Поля капчи
	HCaptchaKey string
}
