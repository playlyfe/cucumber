package core

import (
	"fmt"
	"reflect"
)

type TestResult int

const (
	PassedResult  TestResult = 1
	PendingResult TestResult = 2
	FailedResult  TestResult = 3
	SkippedResult TestResult = 4
)

type TestCase struct {
	BeforeHooks []BeforeHook
	AfterHooks  []AfterHook
	Result      TestResult
	Steps       []*TestStep
	Pickle      *Pickle
}

type TestStep struct {
	Arguments      []*argument
	StepDefinition *StepDefinition
	Result         TestResult
	PickleStep     *PickleStep
	Text           string
}

func (t *TestCase) Execute(bus *EventBus) error {
	bus.Broadcast(TestCaseStarting, t)
	skipSteps := false
	for _, step := range t.Steps {
		bus.Broadcast(TestStepStarting, step)
		if skipSteps {
			step.Result = SkippedResult
		} else {
			if step.StepDefinition == nil {
				step.Result = PendingResult
				skipSteps = true
			} else {

				stepDefinitionFn := reflect.ValueOf(step.StepDefinition.Fn)
				stepDefinitionType := reflect.TypeOf(step.StepDefinition.Fn)
				if stepDefinitionType.Kind() != reflect.Func {
					return &CucumberError{
						Name:        "Invalid Step Definition",
						Description: "Step definition must be a function",
					}
				}

				if stepDefinitionType.NumIn() != len(step.Arguments) {
					return &CucumberError{
						Name:        "Step Definition Parameter Count Mismatch",
						Description: fmt.Sprintf("Step definition must contain %d arguments but found %d arguments", len(step.Arguments), stepDefinitionType.NumIn()),
					}
				}

				arguments := []reflect.Value{}
				argumentTypes := []reflect.Type{}
				for index, argument := range step.Arguments {
					argumentType := stepDefinitionType.In(index)
					if argument.transformedValue == nil {
						arguments = append(arguments, reflect.New(argumentType).Elem())
						argumentTypes = append(argumentTypes, argumentType)
					} else {
						arguments = append(arguments, reflect.ValueOf(argument.transformedValue))
						argumentTypes = append(argumentTypes, reflect.TypeOf(argument.transformedValue))
					}
				}

				if !(len(arguments) == stepDefinitionType.NumIn() || (stepDefinitionType.IsVariadic() && len(arguments) >= stepDefinitionType.NumIn()-1)) {
					typeList := ""
					for index, argumentType := range argumentTypes {
						if index > 0 {
							typeList += ", "
							typeList += argumentType.Name()
						} else {
							typeList += "interface{}"
						}

					}

					return &CucumberError{
						Name:        "Invalid arguments in Step Definition",
						Description: fmt.Sprintf("Step definition for:\n\n%s\n\nmust have a function with signature: func(%s) error", step.Text, typeList),
					}
				}

				results := stepDefinitionFn.Call(arguments)
				if !results[0].IsNil() {
					err := results[0].Interface().(error)
					println(err.Error())
					step.Result = FailedResult
					skipSteps = true
				} else {
					step.Result = PassedResult
				}
			}
		}
		bus.Broadcast(TestStepFinished, step)
	}
	bus.Broadcast(TestCaseFinished, t)
	return nil
}
