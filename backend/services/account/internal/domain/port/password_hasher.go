package port

import (
	"github.com/cubelitblade/community-v2/backend/services/account/internal/domain"
)

type PasswordHasher interface {
	Hash(plaintext string) (account.PasswordHash, error)
}
