package users

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	usersapp "github.com/vaaxooo/xbackend/internal/modules/users/application"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/apple"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/challenge"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/google"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/link"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/login"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/password"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/profile"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/refresh"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/register"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/session"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/telegram"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/twofactor"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/verification"
	"github.com/vaaxooo/xbackend/internal/modules/users/domain"
	"github.com/vaaxooo/xbackend/internal/modules/users/public"
	phttp "github.com/vaaxooo/xbackend/internal/platform/http"
	"github.com/vaaxooo/xbackend/internal/platform/http/users/dto"
	"github.com/vaaxooo/xbackend/internal/platform/httputil"
	"github.com/vaaxooo/xbackend/internal/platform/log"
)

type stubLogger struct{}

func (stubLogger) Debug(context.Context, string, ...any) {}
func (stubLogger) Info(context.Context, string, ...any)  {}
func (stubLogger) Warn(context.Context, string, ...any)  {}
func (stubLogger) Error(_ context.Context, msg string, err error, args ...any) {
	fmt.Println("log:", msg, "error:", err, "args:", args)
}

// Ensure stubLogger satisfies the interface at compile time.
var _ log.Logger = stubLogger{}

type fakeService struct {
	registerOut login.Output
	registerErr error

	confirmOut login.Output
	confirmErr error

	loginOut login.Output
	loginErr error

	telegramOut login.Output
	telegramErr error
	googleOut   login.Output
	googleErr   error
	appleOut    login.Output
	appleErr    error

	refreshOut refresh.Output
	refreshErr error

	requestEmailErr  error
	requestResetErr  error
	resetPasswordErr error

	twoFactorSetupOut twofactor.SetupOutput
	twoFactorSetupErr error
	twoFactorConfirm  error
	twoFactorDisable  error

	listSessionsOut  session.Output
	listSessionsErr  error
	revokeSessionErr error
	revokeOthersErr  error

	getOut profile.Output
	getErr error

	updateOut profile.Output
	updateErr error

	changePasswordErr error

	linkOut link.Output
	linkErr error

	challengeOut login.Output
	challengeErr error
}

func (f *fakeService) Register(context.Context, register.Input) (login.Output, error) {
	return f.registerOut, f.registerErr
}
func (f *fakeService) ConfirmEmail(context.Context, verification.ConfirmEmailInput) (login.Output, error) {
	return f.confirmOut, f.confirmErr
}
func (f *fakeService) Login(context.Context, login.Input) (login.Output, error) {
	return f.loginOut, f.loginErr
}
func (f *fakeService) LoginWithTelegram(context.Context, telegram.Input) (login.Output, error) {
	return f.telegramOut, f.telegramErr
}
func (f *fakeService) LoginWithGoogle(context.Context, google.Input) (login.Output, error) {
	return f.googleOut, f.googleErr
}
func (f *fakeService) LoginWithApple(context.Context, apple.Input) (login.Output, error) {
	return f.appleOut, f.appleErr
}
func (f *fakeService) Refresh(context.Context, refresh.Input) (refresh.Output, error) {
	return f.refreshOut, f.refreshErr
}
func (f *fakeService) RequestEmailConfirmation(context.Context, verification.RequestEmailInput) error {
	return f.requestEmailErr
}
func (f *fakeService) RequestPasswordReset(context.Context, verification.RequestPasswordResetInput) error {
	return f.requestResetErr
}
func (f *fakeService) ResetPassword(context.Context, verification.ResetPasswordInput) error {
	return f.resetPasswordErr
}
func (f *fakeService) SetupTwoFactor(context.Context, twofactor.SetupInput) (twofactor.SetupOutput, error) {
	return f.twoFactorSetupOut, f.twoFactorSetupErr
}
func (f *fakeService) ConfirmTwoFactor(context.Context, twofactor.ConfirmInput) error {
	return f.twoFactorConfirm
}
func (f *fakeService) DisableTwoFactor(context.Context, twofactor.DisableInput) error {
	return f.twoFactorDisable
}
func (f *fakeService) GetMe(context.Context, profile.GetInput) (profile.Output, error) {
	return f.getOut, f.getErr
}
func (f *fakeService) UpdateProfile(context.Context, profile.UpdateInput) (profile.Output, error) {
	return f.updateOut, f.updateErr
}
func (f *fakeService) ChangePassword(context.Context, password.ChangeInput) error {
	return f.changePasswordErr
}
func (f *fakeService) LinkProvider(context.Context, link.Input) (link.Output, error) {
	return f.linkOut, f.linkErr
}

func (f *fakeService) ChallengeStatus(context.Context, challenge.StatusInput) (login.Output, error) {
	return f.challengeOut, f.challengeErr
}

func (f *fakeService) VerifyChallengeTOTP(context.Context, challenge.VerifyTOTPInput) (login.Output, error) {
	return f.challengeOut, f.challengeErr
}

func (f *fakeService) ResendChallengeEmail(context.Context, challenge.ResendEmailInput) (login.Output, error) {
	return f.challengeOut, f.challengeErr
}

func (f *fakeService) ConfirmChallengeEmail(context.Context, challenge.ConfirmEmailInput) (login.Output, error) {
	return f.challengeOut, f.challengeErr
}

func (f *fakeService) ListSessions(context.Context, session.ListInput) (session.Output, error) {
	return f.listSessionsOut, f.listSessionsErr
}

func (f *fakeService) RevokeSession(context.Context, session.RevokeInput) error {
	return f.revokeSessionErr
}

func (f *fakeService) RevokeOtherSessions(context.Context, session.RevokeOthersInput) error {
	return f.revokeOthersErr
}

type fakeTokenParser struct {
	userID    string
	sessionID string
	err       error
}

func (f *fakeTokenParser) Parse(string) (string, error) { return f.userID, f.err }
func (f *fakeTokenParser) Issue(string, string, time.Duration) (string, error) {
	return "token", nil
}
func (f *fakeTokenParser) Verify(string) (public.AuthContext, error) {
	if f.err != nil {
		return public.AuthContext{}, f.err
	}
	return public.AuthContext{UserID: f.userID, SessionID: f.sessionID}, nil
}

type noopUseCase[Cmd any, Resp any] struct{}

func (noopUseCase[Cmd, Resp]) Execute(ctx context.Context, cmd Cmd) (Resp, error) {
	var zero Resp
	return zero, nil
}

type stubHasher struct{}

func (stubHasher) Hash(context.Context, string) (string, error)  { return "hash", nil }
func (stubHasher) Compare(context.Context, string, string) error { return nil }

type stubTelegramUseCase struct{}

func (stubTelegramUseCase) Execute(context.Context, telegram.Input) (login.Output, error) {
	return login.Output{}, nil
}

func newTestServer(svc usersapp.Service, tp *fakeTokenParser) *httptest.Server {
	router := phttp.NewRouter(phttp.RouterDeps{Logger: stubLogger{}, Timeout: time.Second}, func(r chi.Router) {
		RegisterV1(r, svc, tp)
	})
	return httptest.NewServer(router)
}

func TestRegisterEndpoint(t *testing.T) {
	svc := &fakeService{registerOut: login.Output{UserID: "id", DisplayName: "User", AccessToken: "acc", RefreshToken: "ref"}}
	server := newTestServer(svc, &fakeTokenParser{})
	defer server.Close()

	body, _ := json.Marshal(map[string]string{"email": "user@example.com", "password": "password123", "display_name": "User"})
	resp, err := http.Post(server.URL+"/api/v1/auth/register", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("http error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 201, got %d: %s", resp.StatusCode, string(body))
	}
	payload := decodeBody[dto.LoginResponse](t, resp)
	if payload.AccessToken != "acc" || payload.RefreshToken != "ref" {
		t.Fatalf("unexpected tokens: %+v", payload)
	}
}

func TestRegisterEndpointValidationError(t *testing.T) {
	svc := &fakeService{registerErr: domain.ErrInvalidDisplayName}
	server := newTestServer(svc, &fakeTokenParser{})
	defer server.Close()

	body, _ := json.Marshal(map[string]string{"email": "user@example.com", "password": "password123"})

	resp, err := http.Post(server.URL+"/api/v1/auth/register", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("http error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 400, got %d: %s", resp.StatusCode, string(body))
	}

	decodeBody[httputil.ErrorBody](t, resp)
}

func TestLoginUnauthorized(t *testing.T) {
	svc := &fakeService{loginErr: domain.ErrInvalidCredentials}
	server := newTestServer(svc, &fakeTokenParser{})
	defer server.Close()

	body, _ := json.Marshal(map[string]string{"email": "user@example.com", "password": "bad"})
	resp, _ := http.Post(server.URL+"/api/v1/auth/login", "application/json", bytes.NewReader(body))
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
	defer resp.Body.Close()

	decodeBody[httputil.ErrorBody](t, resp)
}

func TestProfileEndpoints(t *testing.T) {
	svc := &fakeService{
		getOut:    profile.Output{UserID: "user", DisplayName: "User"},
		updateOut: profile.Output{UserID: "user", FirstName: "New"},
	}
	tp := &fakeTokenParser{userID: "user"}
	server := newTestServer(svc, tp)
	defer server.Close()

	req, _ := http.NewRequest(http.MethodGet, server.URL+"/api/v1/me", nil)
	req.Header.Set("Authorization", "Bearer token")
	resp, _ := http.DefaultClient.Do(req)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	defer resp.Body.Close()
	decodeBody[dto.UserProfileResponse](t, resp)

	updateBody, _ := json.Marshal(map[string]string{"first_name": "New"})
	req, _ = http.NewRequest(http.MethodPatch, server.URL+"/api/v1/me", bytes.NewReader(updateBody))
	req.Header.Set("Authorization", "Bearer token")
	resp, _ = http.DefaultClient.Do(req)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on update, got %d", resp.StatusCode)
	}
	defer resp.Body.Close()
	decodeBody[dto.UserProfileResponse](t, resp)
}

func TestChangePassword(t *testing.T) {
	svc := &fakeService{}
	server := newTestServer(svc, &fakeTokenParser{userID: "user"})
	defer server.Close()

	body, _ := json.Marshal(map[string]string{"current_password": "old", "new_password": "newpassword"})
	req, _ := http.NewRequest(http.MethodPost, server.URL+"/api/v1/auth/password/change", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer token")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("http error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", resp.StatusCode)
	}
}

func TestChangePasswordUnauthorized(t *testing.T) {
	svc := &fakeService{changePasswordErr: domain.ErrInvalidCredentials}
	server := newTestServer(svc, &fakeTokenParser{userID: "user"})
	defer server.Close()

	body, _ := json.Marshal(map[string]string{"current_password": "old", "new_password": "newpassword"})
	req, _ := http.NewRequest(http.MethodPost, server.URL+"/api/v1/auth/password/change", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer token")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("http error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
	decodeBody[httputil.ErrorBody](t, resp)
}

func TestRefreshUnauthorized(t *testing.T) {
	svc := &fakeService{refreshErr: domain.ErrRefreshTokenInvalid}
	server := newTestServer(svc, &fakeTokenParser{})
	defer server.Close()

	body, _ := json.Marshal(map[string]string{"refresh_token": "bad"})
	resp, _ := http.Post(server.URL+"/api/v1/auth/refresh", "application/json", bytes.NewReader(body))
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
	defer resp.Body.Close()
	decodeBody[httputil.ErrorBody](t, resp)
}

func TestLinkConflict(t *testing.T) {
	svc := &fakeService{linkErr: domain.ErrIdentityAlreadyLinked}
	tp := &fakeTokenParser{userID: "user"}
	server := newTestServer(svc, tp)
	defer server.Close()

	body, _ := json.Marshal(map[string]string{"provider": "github", "provider_user_id": "gh-1"})
	req, _ := http.NewRequest(http.MethodPost, server.URL+"/api/v1/auth/link", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer token")
	resp, _ := http.DefaultClient.Do(req)
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("expected 409, got %d", resp.StatusCode)
	}
	defer resp.Body.Close()
	decodeBody[httputil.ErrorBody](t, resp)
}

func TestRegisterEndpointTransactionalCommit(t *testing.T) {
	svc := &fakeService{registerOut: login.Output{UserID: "id", AccessToken: "access", RefreshToken: "refresh"}}
	server := newTestServer(svc, &fakeTokenParser{})
	defer server.Close()

	body, _ := json.Marshal(map[string]string{"email": "user@example.com", "password": "password123", "display_name": "User"})
	resp, err := http.Post(server.URL+"/api/v1/auth/register", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("http error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 201, got %d: %s", resp.StatusCode, string(body))
	}

	decodeBody[dto.LoginResponse](t, resp)
}

func TestRegisterEndpointRollbackOnFailure(t *testing.T) {
	svc := &fakeService{registerErr: errors.New("fail")}
	auth := &fakeTokenParser{}
	server := newTestServer(svc, auth)
	defer server.Close()

	body, _ := json.Marshal(map[string]string{"email": "user@example.com", "password": "password123", "display_name": "User"})
	resp, err := http.Post(server.URL+"/api/v1/auth/register", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("http error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected 500 on failure, got %d", resp.StatusCode)
	}

	decodeBody[httputil.ErrorBody](t, resp)

}

func decodeBody[T any](t *testing.T, resp *http.Response) T {
	t.Helper()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}

	var payload T
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("decode body: %v", err)
	}

	return payload
}
