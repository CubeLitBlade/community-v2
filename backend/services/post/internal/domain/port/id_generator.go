package port

type IDGenerator interface {
	Next() (int64, error)
}
