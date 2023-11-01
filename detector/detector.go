package detector

import (
	"context"
	"fmt"
	"sync"
)

// Detector support sync and async method to detect target
type Detector[T, R any] interface {
	// Start Detector
	Start() error

	// Stop Detector, it will wait for all detect target finished
	Stop() error
	// Detect detect single target and return result sync
	Detect(T) Result[T, R]
	Detects() chan<- Task[T]
	Results() <-chan Result[T, R]
}

// Task may contain many targets
type Task[T any] struct {
	Targets []T
}

type Result[T, R any] struct {
	Detect T
	Data   R
	Error  error
}

type RunnerController struct {
	ctx        context.Context
	cancelFunc context.CancelFunc
}

type Controller struct {
	sync.RWMutex
	controls map[string]*RunnerController
}

func (controller *Controller) Add(name string, runner *RunnerController) error {
	controller.Lock()
	defer controller.Unlock()
	var _, exist = controller.controls[name]
	if exist {
		return fmt.Errorf("runner name %s already exist", name)
	}
	controller.controls[name] = runner
	return nil
}

func (controller *Controller) Get(name string) (*RunnerController, bool) {
	controller.RLock()
	defer controller.RUnlock()
	var runner, exist = controller.controls[name]
	return runner, exist
}

func (controller *Controller) RunningCount() int {
	controller.RLock()
	defer controller.RUnlock()
	return len(controller.controls)
}

func NewController() *Controller {
	return &Controller{
		controls: make(map[string]*RunnerController),
	}
}
