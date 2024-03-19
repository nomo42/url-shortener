package interfaces

type FileStorage interface {
	CreateRecord(hash, originalURL string) error
	Close() error
}
