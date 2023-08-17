package zanzigo

type Executer interface {
	// Exec(...fns)
}

type SequentialExecuter struct{}

func NewSequentialExecuter() *SequentialExecuter {
	return &SequentialExecuter{}
}
