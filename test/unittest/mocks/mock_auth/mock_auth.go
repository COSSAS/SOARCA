package mock_auth

import (
	"context"
	"soarca/pkg/auth"

	"github.com/stretchr/testify/mock"
)

type MockAuthenticator struct {
	mock.Mock
}

func (authenticator *MockAuthenticator) Authenticate(ctx context.Context, credential string) (*auth.Principal, error) {
	args := authenticator.Called(ctx, credential)
	principal, _ := args.Get(0).(*auth.Principal)
	return principal, args.Error(1)
}

type MockIdentityProvider struct {
	mock.Mock
}

func (provider *MockIdentityProvider) Name() string {
	args := provider.Called()
	return args.String(0)
}

func (provider *MockIdentityProvider) Handles() auth.Match {
	args := provider.Called()
	return args.Get(0).(auth.Match)
}

func (provider *MockIdentityProvider) Authenticate(ctx context.Context, credential auth.Credential) (*auth.Principal, error) {
	args := provider.Called(ctx, credential)
	principal, _ := args.Get(0).(*auth.Principal)
	return principal, args.Error(1)
}
