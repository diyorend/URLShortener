package storage

type URLStorage interface {
	Save(short string, long string) error
	Get(short string) (string, error)
}

