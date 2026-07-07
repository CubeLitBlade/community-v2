package application

import "github.com/cubelitblade/community-v2/backend/services/account/internal/domain"

type Profile struct {
	ID          int64
	Username    string
	DisplayName string
}

func NewProfile(acc *account.Account) *Profile {
	return &Profile{
		ID:          int64(acc.ID),
		Username:    acc.Username.Value(),
		DisplayName: acc.DisplayName,
	}
}
