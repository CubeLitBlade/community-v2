//go:build integration

package persistence_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cubelitblade/community-v2/backend/services/post/internal/adapter/driven/persistence"
	"github.com/cubelitblade/community-v2/backend/services/post/internal/domain/model"
)

func TestIntegrationPostWorkflow(t *testing.T) {
	ctx := context.Background()

	client := newClient(ctx, t)

	client.AccountProfile.Create().SetID(authorID).SetUsername(username).SetDisplayName(displayName).Save(ctx)

	writer := &persistence.PostSaver{}
	reader := &persistence.PostReader{}

	t.Run("publish and find by id", func(t *testing.T) {
		post, err := model.NewPost(postID, authorID, &title, content, fixedTime)
		if err != nil {
			t.Fatalf("Failed to create post: %v", err)
		}

		err = writer.Save(ctx, *post)
		if err != nil {
			t.Fatalf("Save() failed: %v", err)
		}

		got, err := reader.Get(ctx, client, postID)
		if err != nil {
			t.Fatalf("Get() failed: %v", err)
		}

		assert.Equal(t, post.ID(), got.ID())
		assert.Equal(t, post.AuthorID(), got.AuthorID())
		assert.Equal(t, post.Title(), got.Title())
		assert.Equal(t, post.Content(), got.Content())
		assert.Equal(t, post.Status(), got.Status())

		assert.True(t, post.CreatedAt().Equal(got.CreatedAt()))
		assert.True(t, post.UpdatedAt().Equal(got.UpdatedAt()))
	})
}
