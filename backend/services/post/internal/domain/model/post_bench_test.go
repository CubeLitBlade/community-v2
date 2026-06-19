package model

import (
	"testing"
	"time"
)

func BenchmarkCleanString_NonBlank(b *testing.B) {
	input := "  some valid content with spaces  "

	for b.Loop() {
		cleanString(&input)
	}
}

func BenchmarkCleanString_Blank(b *testing.B) {
	input := "     "

	for b.Loop() {
		cleanString(&input)
	}
}

func BenchmarkNewPost_NoTitle(b *testing.B) {
	now := time.Now()

	for b.Loop() {
		_, _ = NewPost(1, 2, nil, "This is a valid post content.", now)
	}
}

func BenchmarkNewPost_WithTitle(b *testing.B) {
	title := "My First Post"
	now := time.Now()

	for b.Loop() {
		_, _ = NewPost(1, 2, &title, "This is a valid post content.", now)
	}
}

func BenchmarkPost_Edit_ContentOnly(b *testing.B) {
	title := "Original Title"
	p, _ := NewPost(1, 2, &title, "Original content.", time.Now())

	newContent := "Updated content here."
	now := time.Now()

	for b.Loop() {
		p.Edit(nil, &newContent, now)
	}
}
