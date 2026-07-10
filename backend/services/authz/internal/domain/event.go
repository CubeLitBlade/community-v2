package domain

const (
	EventAccountCreated  = "account.created"
	EventPostPublished   = "post.published"
)

type Event interface {
	Type() string
}

type AccountCreated struct {
	Tuples []Tuple
}

func (AccountCreated) Type() string {
	return EventAccountCreated
}

type PostPublished struct {
	Tuples []Tuple
}

func (PostPublished) Type() string {
	return EventPostPublished
}
