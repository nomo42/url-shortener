package interfaces

type Storage interface {
	WriteValue(key, value string)
	ReadValue(key string) (string, bool)
	ExistenceCheck(key string) bool
	Clear()
	Close() error
}
