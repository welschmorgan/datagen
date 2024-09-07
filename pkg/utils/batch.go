package utils

import (
	"fmt"
	"log/slog"
	"sync"
)

type Handler[I any, O any] func(I) O

type Task[I any, O any] struct {
	Id      int
	Data    I
	Handler Handler[I, O]
}

func NewTask[I any, O any](fn Handler[I, O]) *Task[I, O] {
	return &Task[I, O]{
		Id:      -1,
		Handler: fn,
	}
}

func NewTaskWithData[I any, O any](fn Handler[I, O], data I) *Task[I, O] {
	return &Task[I, O]{
		Id:      -1,
		Handler: fn,
		Data:    data,
	}
}

type Result[O any] struct {
	Value O
}

type Worker[I any, O any] struct {
	Id       int
	NumTasks int
	Result   chan Result[O]
	// Schedule chan *Task[I, O]
	Tasks []*Task[I, O]
}

type Scheduler[I any, O any] struct {
	NumWorkers          int
	EffectiveNumWorkers int
	Workers             []*Worker[I, O]
	Tasks               []*Task[I, O]
	Debug               bool
	WorkersDone         sync.WaitGroup
}

func NewScheduler[I any, O any](numWorkers int, tasks ...*Task[I, O]) *Scheduler[I, O] {
	s := &Scheduler[I, O]{
		NumWorkers:          numWorkers,
		EffectiveNumWorkers: numWorkers,
		Tasks:               []*Task[I, O]{},
	}
	if len(tasks) > 0 {
		s.Schedule(tasks...)
	}
	return s
}

func NewSchedulerSingleHandler[I any, O any](numWorkers int, h Handler[I, O], data ...I) *Scheduler[I, O] {
	s := &Scheduler[I, O]{
		NumWorkers:          numWorkers,
		EffectiveNumWorkers: numWorkers,
		Tasks:               []*Task[I, O]{},
	}
	if len(data) > 0 {
		tasks := []*Task[I, O]{}
		for _, item := range data {
			tasks = append(tasks, NewTaskWithData(h, item))
		}
		s.Schedule(tasks...)
	}
	return s
}

func (s *Scheduler[I, O]) Schedule(tasks ...*Task[I, O]) {
	s.Tasks = append(s.Tasks, tasks...)
}

func (s *Scheduler[I, O]) Run() []Result[O] {
	s.spreadTasks()
	go s.spawnWorkers()
	final := make(chan Result[O], len(s.Tasks))
	go s.aggregateResults(final)
	s.debug("Waiting for final results ...")
	ret := []Result[O]{}
	for res := range final {
		ret = append(ret, res)
	}
	return ret
}

func (s *Scheduler[I, O]) debug(msg string, args ...any) {
	if s.Debug {
		slog.Debug(fmt.Sprintf(msg, args...))
	}
}

func (s *Scheduler[I, O]) spreadTasks() {
	s.Workers = []*Worker[I, O]{}
	chanSize := 1
	s.EffectiveNumWorkers = s.NumWorkers
	if len(s.Tasks) >= s.NumWorkers {
		chanSize = int(float32(len(s.Tasks)) / float32(s.NumWorkers))
	} else {
		s.EffectiveNumWorkers = len(s.Tasks)
	}
	for workerId := range s.EffectiveNumWorkers {
		s.Workers = append(s.Workers, &Worker[I, O]{
			Id:       workerId,
			NumTasks: chanSize,
			Result:   make(chan Result[O], chanSize),
			// Schedule: make(chan *Task[I, O], chanSize),
			Tasks: []*Task[I, O]{},
		})
	}
	s.debug("Schedule %d tasks on %d workers", len(s.Tasks), s.EffectiveNumWorkers)

	for i, task := range s.Tasks {
		task.Id = i
		workerId := task.Id % len(s.Workers)
		worker := s.Workers[workerId]
		worker.Tasks = append(worker.Tasks, task)
	}
	for _, worker := range s.Workers {
		worker.NumTasks = len(worker.Tasks)
	}
}

func (s *Scheduler[I, O]) spawnWorkers() {
	s.WorkersDone.Add(len(s.Workers))
	for _, worker := range s.Workers {
		go s.executeWorkerTasks(worker)
	}
}

func (s *Scheduler[I, O]) executeWorkerTasks(worker *Worker[I, O]) {
	defer func() {
		s.debug("Worker %d is done", worker.Id)
		s.WorkersDone.Done()
	}()
	s.debug("Executing %d tasks on worker %d", worker.NumTasks, worker.Id)
	for _, task := range worker.Tasks {
		// fmt.Printf("\x1b[2Kexec task %d\r", task.Id)
		worker.Result <- Result[O]{Value: task.Handler(task.Data)}
	}
	close(worker.Result)
}

func (s *Scheduler[I, O]) aggregateResults(final chan Result[O]) {
	s.debug("Waiting for %d workers to collect results", len(s.Workers))
	s.WorkersDone.Wait()
	for _, worker := range s.Workers {
		for res := range worker.Result {
			final <- res
		}
	}
	close(final)
}
