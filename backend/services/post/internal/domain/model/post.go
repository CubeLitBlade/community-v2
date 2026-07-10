// Package domain comment to insert
package model

import (
	"strings"
	"time"

	"github.com/cubelitblade/community-v2/backend/services/post/internal/domain"
)

type Post struct {
	id, authorID         int64
	title                *string
	content              string
	status               Status
	createdAt, updatedAt time.Time
}

func NewPost(id, authorID int64, title *string, content string, now time.Time) (*Post, error) {
	if id <= 0 || authorID <= 0 {
		return nil, domain.ErrIDInvalid
	}

	content = strings.TrimSpace(content)
	if content == "" {
		return nil, domain.ErrContentBlank
	}

	return &Post{
		id:        id,
		authorID:  authorID,
		title:     cleanString(title),
		content:   content,
		status:    StatusPublished,
		createdAt: now,
		updatedAt: now,
	}, nil
}

func Unmarshal(
	id, authorID int64, title *string, content string, status Status, createdAt, updatedAt time.Time,
) *Post {
	return &Post{
		id:        id,
		authorID:  authorID,
		title:     title,
		content:   content,
		status:    status,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
}

func (p *Post) Edit(title *string, content *string, now time.Time) {
	if p.title != nil && title != nil {
		cleanedTitle := cleanString(title)
		if cleanedTitle != nil {
			p.title = cleanedTitle
			p.updatedAt = now
		}
	}

	cleanedContent := cleanString(content)
	if cleanedContent != nil {
		p.content = *cleanedContent
		p.updatedAt = now
	}
}

func (p *Post) Archive() {
	p.status = StatusArchived
}

func (p *Post) Publish() {
	p.status = StatusPublished
}

func (p *Post) ID() int64 {
	return p.id
}

func (p *Post) AuthorID() int64 {
	return p.authorID
}

func (p *Post) Title() *string {
	return p.title
}

// TitleString returns the title as a plain string, or empty string if nil.
func (p *Post) TitleString() string {
	if p.title == nil {
		return ""
	}
	return *p.title
}

func (p *Post) Content() string {
	return p.content
}

func (p *Post) Status() Status {
	return p.status
}

func (p *Post) CreatedAt() time.Time {
	return p.createdAt
}

func (p *Post) UpdatedAt() time.Time {
	return p.updatedAt
}

func cleanString(str *string) *string {
	if str == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*str)
	if trimmed == "" {
		return nil
	}

	return &trimmed
}
