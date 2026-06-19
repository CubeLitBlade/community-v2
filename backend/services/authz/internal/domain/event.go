package domain

const (
	EventAccountCreated = "account.created"
	EventAccountUpdated = "account.updated"
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
