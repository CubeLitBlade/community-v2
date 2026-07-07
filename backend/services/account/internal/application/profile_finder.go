package application

import (
	"context"
	"fmt"

	"github.com/cubelitblade/community-v2/backend/services/account/internal/domain/port"
)

type ProfileFinderUseCase interface {
	Find(ctx context.Context, id int64) (*Profile, error)
}

type ProfileFinder struct {
	reader port.AccountReader
}

func NewProfileFinder(reader port.AccountReader) *ProfileFinder {
	return &ProfileFinder{reader: reader}
}

func (f *ProfileFinder) Find(ctx context.Context, id int64) (*Profile, error) {
	acc, err := f.reader.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("find by id: %w", err)
	}

	return NewProfile(acc), nil
}
