// Package permission defines the standard vocabulary (relations and objects) used in the authorization system.
// It is shared across services to prevent hardcoded strings and typos when interacting with OpenFGA adapters.
package permission

const (
	RelationCanPublishPost = "can_publish_post"
	RelationCanEdit        = "can_edit"
	RelationCanArchive     = "can_archive"
	RelationCanSuspend     = "can_suspend"
)

const (
	ObjectSystemCommunity = "system:community"
)
