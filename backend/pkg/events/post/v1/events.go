package v1

const (
	// TopicPostPublished is the RabbitMQ topic for post published events.
	TopicPostPublished = "post.published.v1"

	// EventTypePostPublished is the CloudEvents type for post published events.
	EventTypePostPublished = "io.github.cubelitblade.post.published.v1"

	// AggregateType is the aggregate type identifier for the post service.
	AggregateType = "/community-v2/post-service"
)
