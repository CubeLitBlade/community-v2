package hasher

import (
	"github.com/cubelitblade/community-v2/backend/services/account/internal/domain"
)

type BcryptHasher struct{}

func NewBcryptHasher() *BcryptHasher {
	return &BcryptHasher{}
}

func (BcryptHasher) Hash(plaintext string) (account.PasswordHash, error) {
	return account.HashPassword(plaintext)
}
