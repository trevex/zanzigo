package zanzigo

type Storage interface {
	Write(t Tuple) error
}
