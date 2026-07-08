package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	postv1 "github.com/cubelitblade/community-v2/backend/gen/go/post/v1"
	"github.com/cubelitblade/community-v2/backend/pkg/metadata"
	"github.com/cubelitblade/community-v2/backend/services/post/internal/application"
	"github.com/cubelitblade/community-v2/backend/services/post/internal/domain"
)

var _ postv1.PostServiceServer = (*PostServiceServer)(nil)

var errorMap = map[error]error{
	domain.ErrContentBlank:     status.Error(codes.InvalidArgument, "content is blank"),
	domain.ErrIDInvalid:        status.Error(codes.InvalidArgument, "invalid id"),
	domain.ErrPermissionDenied: status.Error(codes.PermissionDenied, "permission denied"),
}

type PostServiceServer struct {
	postv1.UnimplementedPostServiceServer

	publishPost *application.Publisher
}

func NewPostServiceServer(publishPost *application.Publisher) *PostServiceServer {
	return &PostServiceServer{
		UnimplementedPostServiceServer: postv1.UnimplementedPostServiceServer{},
		publishPost:                    publishPost,
	}
}

func (s *PostServiceServer) Publish(ctx context.Context, req *postv1.PublishRequest) (*postv1.PublishResponse, error) {
	accountID, err := metadata.AccountIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "missing account id")
	}

	id, err := s.publishPost.Publish(ctx, accountID, req.Title, req.Content)
	if err != nil {
		gerr, ok := errorMap[err]
		if !ok {
			//nolint:wrapcheck // should return grpc standard error for server to handle
			return nil, status.Error(codes.Internal, "unexpected error")
		}

		return nil, gerr
	}

	return &postv1.PublishResponse{
		PostId: id,
	}, nil
}
