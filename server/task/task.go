package task

type Status int
type Task struct {
	ID string
	Status
}

const (
	InQueue Status = iota
	Running
	Completed
	Failed
)
