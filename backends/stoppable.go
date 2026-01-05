package backends

type StoppableBackend interface {
	Stop() error
}
