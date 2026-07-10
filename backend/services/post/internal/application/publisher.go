package application

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	postv1 "github.com/cubelitblade/community-v2/backend/pkg/events/post/v1"
	"github.com/cubelitblade/community-v2/backend/services/post/internal/domain"
	"github.com/cubelitblade/community-v2/backend/services/post/internal/domain/model"
	"github.com/cubelitblade/community-v2/backend/services/post/internal/domain/port"
)

type Publisher struct {
	postSaver     port.PostSaver
	authz         port.Authorizer
	transactor    port.Transactor
	idgen         port.IDGenerator
	eventRecorder port.EventRecorder

	logger   *slog.Logger
	timeFunc func() time.Time
}

func NewPublisher(
	postSaver port.PostSaver, authz port.Authorizer, transactor port.Transactor, idgen port.IDGenerator,
	eventRecorder port.EventRecorder,
	logger *slog.Logger,
) *Publisher {
	return &Publisher{
		postSaver:     postSaver,
		authz:         authz,
		transactor:    transactor,
		idgen:         idgen,
		eventRecorder: eventRecorder,
		logger:        logger,
		timeFunc:      time.Now,
	}
}

func (p *Publisher) Publish(ctx context.Context, accountID int64, title *string, content string) (int64, error) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.Int64("account.id", accountID),
	)

	var err error
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
			span.RecordError(err)
		}
	}()

	if err := p.checkPermission(ctx, accountID, span); err != nil {
		return 0, err
	}

	id, err := p.idgen.Next()
	if err != nil {
		p.logger.ErrorContext(ctx, "failed to generate post id", slog.Any("error", err))
		return 0, fmt.Errorf("generate id: %w", err)
	}

	post, err := model.NewPost(id, accountID, title, content, p.timeFunc())
	if err != nil {
		p.logger.ErrorContext(ctx, "failed to create post", slog.Any("error", err))
		return 0, fmt.Errorf("create post: %w", err)
	}

	if err := p.transactor.InTx(ctx, func(ctx context.Context) error {
		return p.savePostAndOutbox(ctx, post)
	}); err != nil {
		p.logger.ErrorContext(ctx, "transaction error", slog.Any("error", err))
		return 0, fmt.Errorf("tx error: %w", err)
	}

	p.logger.InfoContext(ctx, "post published successfully",
		slog.Int64("post_id", post.ID()),
		slog.Int64("account_id", accountID),
	)
	span.SetAttributes(attribute.Int64("post.id", id))

	return id, nil
}

func (p *Publisher) savePostAndOutbox(ctx context.Context, post *model.Post) error {
	serr := p.savePost(ctx, post)
	werr := p.writeOutbox(ctx, post)

	return errors.Join(serr, werr)
}

func (p *Publisher) checkPermission(ctx context.Context, accountID int64, span trace.Span) error {
	ok, err := p.authz.CanPublishPost(ctx, accountID)
	if err != nil {
		p.logger.ErrorContext(ctx, "failed to check permission", slog.Any("error", err))
		span.SetStatus(codes.Error, "failed to check permission")

		return fmt.Errorf("check permission: %w", err)
	}

	if !ok {
		p.logger.WarnContext(ctx, "permission denied",
			slog.Int64("account_id", accountID),
			slog.String("permission", "can_publish_post"),
		)

		return domain.ErrPermissionDenied
	}

	p.logger.DebugContext(ctx, "permission check passed", slog.Int64("account_id", accountID))

	return nil
}

func (p *Publisher) savePost(ctx context.Context, post *model.Post) error {
	if err := p.postSaver.Save(ctx, *post); err != nil {
		p.logger.ErrorContext(ctx, "failed to save post", slog.Any("error", err))
		return fmt.Errorf("save post: %w", err)
	}

	return nil
}

func (p *Publisher) writeOutbox(ctx context.Context, post *model.Post) error {
	id, err := p.idgen.Next()
	if err != nil {
		return fmt.Errorf("generate outbox id: %w", err)
	}

	entry := &port.OutboxEntry{
		ID:            id,
		AggregateID:   post.ID(),
		AggregateType: postv1.AggregateType,
		Topic:         postv1.TopicPostPublished,
		EventType:     postv1.EventTypePostPublished,
		Payload: postv1.PostPublished{
			PostID:   strconv.FormatInt(post.ID(), 10),
			AuthorID: strconv.FormatInt(post.AuthorID(), 10),
			Title:    post.TitleString(),
		},
	}

	return p.eventRecorder.Record(ctx, entry)
}
