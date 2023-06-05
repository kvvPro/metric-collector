package storage

type Storage interface {
	Update(t string, n string, v string) error
}
