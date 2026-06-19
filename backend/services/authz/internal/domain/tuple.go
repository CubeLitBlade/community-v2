package domain

// Tuple represents an authorization relationship between a subject and an object.
type Tuple struct {
	User     string
	Relation string
	Object   string
}
