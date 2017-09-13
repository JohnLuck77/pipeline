package pipeline

import (
	"context"
	"fmt"

	"github.com/fatih/color"
)

// Result is returned by a step to dispatch data to the next step or stage
type Result struct {
	Error error
	// dispatch any type
	Data interface{}
	// dispatch key value pairs
	KeyVal map[string]interface{}
}

// Request is the result dispatched in a previous step.
type Request struct {
	stepContext *StepContext
	Data        interface{}
	KeyVal      map[string]interface{}
}

// Status logs the status line to the out channel
func (r *Request) Status(line string) {
	r.stepContext.Status(line)
}

// Step is the unit of work which can be concurrently or sequentially staged with other steps
type Step interface {
	out
	// Exec is invoked by the pipeline when it is run
	Exec(*Request) *Result
	// Cancel is invoked by the pipeline when one of the concurrent steps set Result{Error:err}
	Cancel()
}

type out interface {
	Status(line string)
	getCtx() *stepContextVal
	setCtx(ctx *stepContextVal)
}

type stepContextVal struct {
	pipelineKey string
	name        string
	index       int
	concurrent  bool
}

// StepContext type is embedded in types which need to statisfy the Step interface
type StepContext struct {
	ctx *stepContextVal
}

func (sc *StepContext) getCtx() *stepContextVal {
	return sc.ctx
}

func (sc *StepContext) setCtx(ctx *stepContextVal) {
	sc.ctx = ctx
}

// Status is used to log status from a step
func (sc *StepContext) Status(line string) {
	stepText := fmt.Sprintf("[step-%d]", sc.getCtx().index)
	blue := color.New(color.FgBlue).SprintFunc()
	line = blue(stepText) + "[" + sc.getCtx().name + "]: " + line
	send(sc.getCtx().pipelineKey, line)
}

type step struct {
	StepContext
	execFunc   func(context context.Context, r *Request) *Result
	cancelFunc context.CancelFunc
}

func (s *step) Exec(request *Request) *Result {

	request.stepContext = &s.StepContext
	var ctx context.Context
	ctx, s.cancelFunc = context.WithCancel(context.Background())
	return s.execFunc(ctx, request)

}

func (s *step) Cancel() {
	s.cancelFunc()
}

// NewStep creates a new step
func NewStep(exec func(context context.Context, r *Request) *Result) Step {
	return &step{
		execFunc: exec,
	}

}
