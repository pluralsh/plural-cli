package bridge

import (
	"context"
	"errors"
	"time"
)

func NewAuthService(clients AuthClientFactory, pollInterval time.Duration) *AuthService {
	if pollInterval <= 0 {
		pollInterval = 2 * time.Second
	}
	return &AuthService{clients: clients, pollInterval: pollInterval}
}

func (s *AuthService) BeginDeviceLogin(ctx context.Context, endpoint string) (DeviceAuthorization, error) {
	authorization, err := s.clients.New(ctx, endpoint, "").DeviceLogin(ctx)
	if err != nil {
		return DeviceAuthorization{}, s.operationError(OperationDeviceLogin, err)
	}
	return authorization, nil
}

func (s *AuthService) AwaitDeviceLogin(ctx context.Context, endpoint, deviceToken string) (string, error) {
	client := s.clients.New(ctx, endpoint, "")
	for {
		credential, err := client.PollLoginToken(ctx, deviceToken)
		if err == nil {
			return credential, nil
		}

		timer := time.NewTimer(s.pollInterval)
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}
			return "", &Error{Code: ErrorCancelled, Operation: OperationPollLoginToken, Err: ctx.Err()}
		case <-timer.C:
		}
	}
}

func (s *AuthService) EstablishSession(
	ctx context.Context,
	endpoint, credential, serviceAccount string,
	persist bool,
	notify func(AuthEvent),
) (AuthSession, error) {
	client := s.clients.New(ctx, endpoint, credential)
	baseEmail, err := client.CurrentIdentity(ctx)
	if err != nil {
		return AuthSession{}, s.operationError(OperationCurrentIdentity, err)
	}
	if notify != nil {
		notify(AuthEvent{Kind: AuthEventIdentified, Email: baseEmail, Credential: credential})
	}

	result := AuthSession{BaseEmail: baseEmail, EffectiveEmail: baseEmail, Credential: credential}
	if serviceAccount != "" {
		impersonatedCredential, effectiveEmail, err := client.ImpersonateServiceAccount(ctx, serviceAccount)
		if err != nil {
			return AuthSession{}, s.operationError(OperationImpersonateServiceAccount, err)
		}
		result.EffectiveEmail = effectiveEmail
		result.Credential = impersonatedCredential
		result.Impersonated = true
		if notify != nil {
			notify(AuthEvent{Kind: AuthEventImpersonated, Email: effectiveEmail, Credential: impersonatedCredential})
		}
		client = s.clients.New(ctx, endpoint, impersonatedCredential)
		if !persist {
			return result, nil
		}
	}

	accessToken, err := client.GrabAccessToken(ctx)
	if err != nil {
		return AuthSession{}, s.operationError(OperationGrabAccessToken, err)
	}
	result.Credential = accessToken
	return result, nil
}

func (s *AuthService) operationError(operation Operation, err error) error {
	code := ErrorUnavailable
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		code = ErrorCancelled
	}
	return &Error{Code: code, Operation: operation, Err: err}
}
