package interfaces

type Storage interface {
	WriteValue(key, value string)
	ReadValue(key string) (string, error)
	ExistenceCheck(key string) bool
	Clear()
}
