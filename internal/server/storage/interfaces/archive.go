package storage

type Archive interface {
	restoreFrom() error
	parkTo() error
	Start() error
}
