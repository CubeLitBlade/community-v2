package v1

// Profile is the response body for an account profile.
type Profile struct {
	ID          int64  `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"displayName"`
}
