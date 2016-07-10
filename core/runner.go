package core

import (
	"fmt"
	"strings"
)

type Runner struct {
	world        interface{}
	pendingSteps map[string]*TestStep
	testCases    []*TestCase
	bus          *EventBus
}

func (r *Runner) ExecuteAllTestCases() error {
	r.bus.RegisterHandler(TestStepFinished, func(event *Event) {
		testStep := event.Data.(*TestStep)
		if testStep.Result == PendingResult {
			r.pendingSteps[testStep.Text] = testStep
		}
	})
	r.bus.RegisterHandler(TestRunFinished, func(event *Event) {
		if len(r.pendingSteps) > 0 {
			fmt.Printf("You can implement the missing steps with the snippets below:\n\n")
			for _, step := range r.pendingSteps {
				if len(step.Arguments) == 0 {
					fmt.Printf("%s(%q, func(world interface{}) error {\n    // Write your step definition here\n    return nil\n})\n\n", strings.TrimSpace(step.PickleStep.Step.Keyword), step.Text)
				} else {
					fmt.Printf("%s(%q, func(world interface{}, text string) error {\n    // Write your step definition here\n    println(text)\n    return nil\n})\n\n", strings.TrimSpace(step.PickleStep.Step.Keyword), step.Text)
				}
			}
		}
	})
	r.bus.Broadcast(TestRunStarting, nil)
	for _, testCase := range r.testCases {
		err := testCase.Execute(r.bus)
		if err != nil {
			return err
		}
	}
	r.bus.Broadcast(TestRunFinished, nil)
	return nil
}

func NewRunner(world interface{}, testCases []*TestCase, bus *EventBus) *Runner {
	return &Runner{
		pendingSteps: map[string]*TestStep{},
		world:        world,
		testCases:    testCases,
		bus:          bus,
	}
}
