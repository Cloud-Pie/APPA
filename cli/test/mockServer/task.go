package main

type Status int
type Task struct {
	ID int
	Status
}

const (
	InQueue Status = iota
	Running
	Completed
	Failed
)
