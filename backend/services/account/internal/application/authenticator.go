package application

import (
	"context"
	"fmt"

	"github.com/cubelitblade/community-v2/backend/services/account/internal/domain"
	"github.com/cubelitblade/community-v2/backend/services/account/internal/domain/port"
)

type Authenticator struct {
	reader port.AccountReader
}

func NewAuthenticator(reader port.AccountReader) *Authenticator {
	return &Authenticator{reader: reader}
}

func (a *Authenticator) Authenticate(ctx context.Context, username, password string) (*account.Account, error) {
	u, err := account.NewUsername(username)
	if err != nil {
		return nil, account.ErrInvalidCredentials
	}

	acc, err := a.reader.FindByUsername(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("find by username: %w", err)
	}

	if !acc.VerifyPassword(password) {
		return nil, account.ErrInvalidCredentials
	}

	return acc, nil
}
