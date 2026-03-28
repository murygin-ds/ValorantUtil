package auth

import "fmt"

var (
	ErrMFARequired        = fmt.Errorf("MFA is required")
	ErrCaptchaRequired    = fmt.Errorf("captcha is required")
	ErrInvalidCredentials = fmt.Errorf("invalid credentials")
)

// ChallengeError возвращается, когда сервер авторизации требует капчу или MFA.
// Содержит данные challenge, необходимые для продолжения процесса.
type ChallengeError struct {
	Challenge AuthChallenge
}

func (e *ChallengeError) Error() string {
	switch e.Challenge.Type {
	case ChallengeMFA:
		return "MFA required"
	case ChallengeCaptcha:
		return "captcha required"
	default:
		return "auth challenge required"
	}
}
