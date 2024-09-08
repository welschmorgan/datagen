package utils_test

import (
	"testing"
	"time"

	"github.com/welschmorgan/datagen/pkg/utils"
)

const NumTasks = 10_000

func TestBatchProcessor(t *testing.T) {
	startTime := time.Now()
	tasks := []*utils.Task[any, int]{}
	for i := range NumTasks {
		tasks = append(tasks, utils.NewTask(func(v any) int {
			time.Sleep(5 * time.Microsecond)
			return i
		}))
	}
	s := utils.NewScheduler(100, tasks...)
	res := s.Run()
	if len(res) != len(tasks) {
		t.Errorf("invalid number of result returned, expected %d but got %d", len(tasks), len(res))
	}
	t.Logf("Final list: %d results in %s", len(res), time.Since(startTime))
	// time.Sleep(5 * time.Second)
}
