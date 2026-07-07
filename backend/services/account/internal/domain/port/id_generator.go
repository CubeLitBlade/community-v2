package port

type IDGenerator interface {
	NextID() (int64, error)
}
