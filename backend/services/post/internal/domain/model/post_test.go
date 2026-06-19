package model_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/cubelitblade/community-v2/backend/services/post/internal/adapter/driven/persistence/ent/post"
	"github.com/cubelitblade/community-v2/backend/services/post/internal/domain"
	"github.com/cubelitblade/community-v2/backend/services/post/internal/domain/model"
)

var (
	postID              = int64(1)
	authorID            = int64(2)
	blank               = "  "
	title               = "Lorem Ipsum"
	newTitle            = "Pangram"
	untrimmedTitle      = "  " + title + "   "
	untrimmedNewTitle   = "  " + newTitle + "   "
	content             = "Lorem ipsum dolor sit amet, consectetur adipiscing elit."
	newContent          = "The quick brown fox jumps over the lazy dog."
	untrimmedContent    = "  " + content + "   "
	untrimmedNewContent = "  " + newContent + "   "
	before              = time.Date(2026, 6, 11, 0, 0, 0, 0, time.UTC)
	now                 = time.Date(2026, 6, 12, 0, 0, 0, 0, time.UTC)
)

func TestNewPost(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		id            int64
		authorID      int64
		title         *string
		content       string
		now           time.Time
		wantID        int64
		wantAuthorID  int64
		wantTitle     *string
		wantContent   string
		wantStatus    model.Status
		wantCreatedAt time.Time
		wantUpdatedAt time.Time
		wantErr       error
	}{
		{
			name:          "happy_path/nil_title",
			id:            postID,
			authorID:      authorID,
			title:         nil,
			content:       content,
			now:           now,
			wantID:        postID,
			wantAuthorID:  authorID,
			wantTitle:     nil,
			wantContent:   content,
			wantStatus:    model.StatusPublished,
			wantCreatedAt: now,
			wantUpdatedAt: now,
		},
		{
			name:          "happy_path/valid_title",
			id:            postID,
			authorID:      authorID,
			title:         &title,
			content:       content,
			now:           now,
			wantID:        postID,
			wantAuthorID:  authorID,
			wantTitle:     &title,
			wantContent:   content,
			wantStatus:    model.StatusPublished,
			wantCreatedAt: now,
			wantUpdatedAt: now,
		},
		{
			name:          "rule/blank_title_cleaned_to_nil",
			id:            postID,
			authorID:      authorID,
			title:         &blank,
			content:       content,
			now:           now,
			wantID:        postID,
			wantAuthorID:  authorID,
			wantTitle:     nil,
			wantContent:   content,
			wantStatus:    model.StatusPublished,
			wantCreatedAt: now,
			wantUpdatedAt: now,
		},
		{
			name:     "validate/zero_id",
			id:       0,
			authorID: authorID,
			title:    nil,
			content:  content,
			now:      now,
			wantErr:  domain.ErrIDInvalid,
		},
		{
			name:     "validate/zero_author_id",
			id:       postID,
			authorID: 0,
			title:    nil,
			content:  content,
			now:      now,
			wantErr:  domain.ErrIDInvalid,
		},
		{
			name:     "validate/blank_content",
			id:       postID,
			authorID: authorID,
			title:    nil,
			content:  "",
			now:      now,
			wantErr:  domain.ErrContentBlank,
		},
		{
			name:          "smoke/trim_untrimmed_inputs",
			id:            postID,
			authorID:      authorID,
			title:         &untrimmedTitle,
			content:       untrimmedContent,
			now:           now,
			wantID:        postID,
			wantAuthorID:  authorID,
			wantTitle:     &title,
			wantContent:   content,
			wantStatus:    model.StatusPublished,
			wantCreatedAt: now,
			wantUpdatedAt: now,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, gotErr := model.NewPost(tt.id, tt.authorID, tt.title, tt.content, tt.now)
			if gotErr != nil {
				assert.ErrorIs(t, gotErr, tt.wantErr)
				return
			}

			if tt.wantErr != nil {
				t.Fatal("NewPost() succeeded unexpectedly")
			}

			assert.Equal(t, tt.wantID, got.ID())
			assert.Equal(t, tt.wantAuthorID, got.AuthorID())
			assert.Equal(t, tt.wantTitle, got.Title())
			assert.Equal(t, tt.wantContent, got.Content())
			assert.Equal(t, tt.wantStatus, got.Status())
			assert.Equal(t, tt.wantCreatedAt, got.CreatedAt())
			assert.Equal(t, tt.wantUpdatedAt, got.UpdatedAt())
		})
	}
}

func TestPost_Edit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string // description of this test case
		// Named input parameters for receiver constructor.
		ctitle *string
		// Named input parameters for target function.
		title   *string
		content *string

		wantTitle     *string
		wantContent   string
		wantUpdatedAt time.Time
	}{
		{
			name:          "no_change/all_nil",
			ctitle:        nil,
			title:         nil,
			content:       nil,
			wantTitle:     nil,
			wantContent:   content,
			wantUpdatedAt: before,
		},
		{
			name:          "no_change/ignore_both_blank_noise",
			ctitle:        &title,
			title:         &blank,
			content:       &blank,
			wantTitle:     &title,
			wantContent:   content,
			wantUpdatedAt: before,
		},
		{
			name:          "edit/both_title_and_content",
			ctitle:        &title,
			title:         &newTitle,
			content:       &newContent,
			wantTitle:     &newTitle,
			wantContent:   newContent,
			wantUpdatedAt: now,
		},
		{
			name:          "reject/add_title_to_non_title_post",
			ctitle:        nil,
			title:         &newTitle,
			content:       &blank,
			wantTitle:     nil,
			wantContent:   content,
			wantUpdatedAt: before,
		},
		{
			name:          "edit/content_on_non_title_post",
			ctitle:        nil,
			title:         &blank,
			content:       &newContent,
			wantTitle:     nil,
			wantContent:   newContent,
			wantUpdatedAt: now,
		},
		{
			name:          "edit/content_only",
			ctitle:        &title,
			title:         nil,
			content:       &newContent,
			wantTitle:     &title,
			wantContent:   newContent,
			wantUpdatedAt: now,
		},
		{
			name:          "no_change/ignore_blank_content",
			ctitle:        &title,
			title:         nil,
			content:       &blank,
			wantTitle:     &title,
			wantContent:   content,
			wantUpdatedAt: before,
		},
		{
			name:          "no_change/ignore_blank_title",
			ctitle:        &title,
			title:         &blank,
			content:       nil,
			wantTitle:     &title,
			wantContent:   content,
			wantUpdatedAt: before,
		},
		{
			name:          "edit/title_only",
			ctitle:        &title,
			title:         &newTitle,
			content:       nil,
			wantTitle:     &newTitle,
			wantContent:   content,
			wantUpdatedAt: now,
		},
		{
			name:          "smoke/trim_untrimmed_inputs",
			ctitle:        &title,
			title:         &untrimmedNewTitle,
			content:       &untrimmedNewContent,
			wantTitle:     &newTitle,
			wantContent:   newContent,
			wantUpdatedAt: now,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p, err := model.NewPost(postID, authorID, tt.ctitle, content, before)
			if err != nil {
				t.Fatalf("could not construct receiver type: %v", err)
			}

			p.Edit(tt.title, tt.content, now)

			assert.Equal(t, postID, p.ID())
			assert.Equal(t, authorID, p.AuthorID())
			assert.Equal(t, post.StatusPublished, p.Status())
			assert.Equal(t, before, p.CreatedAt())

			assert.Equal(t, tt.wantTitle, p.Title())
			assert.Equal(t, tt.wantContent, p.Content())
			assert.Equal(t, tt.wantUpdatedAt, p.UpdatedAt())
		})
	}
}
