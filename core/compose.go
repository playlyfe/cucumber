package core

import (
	"fmt"
	"strings"

	"github.com/cucumber/gherkin-go"

	"github.com/playlyfe/cucumber/utils"
)

func (c *Cucumber) composeScenario(pickle *Pickle, requiredTags []string) (*TestCase, error) {
	// filter tags
	if c.matchTags(requiredTags, pickle.Tags) {
		testCase := &TestCase{
			BeforeHooks: []BeforeHook{},
			AfterHooks:  []AfterHook{},
			Pickle:      pickle,
			Steps:       []*TestStep{},
		}

		for _, pickleStep := range pickle.Steps {
			testStep := &TestStep{
				PickleStep: pickleStep,
				Text:       c.compileStepDefinitionText(pickleStep.Text),
			}

			// match step definitions
			for _, item := range c.stepDefinitions {
				match, arguments, err := item.Expression.match(pickleStep.Text)
				if err != nil {
					return nil, err
				}
				if docString, ok := pickleStep.Step.Argument.(*gherkin.DocString); ok {
					arguments = append(arguments, &argument{
						transformedValue: docString.Content,
					})
				}

				testStep.Arguments = append([]*argument{
					&argument{
						transformedValue: c.World,
					},
				}, arguments...)
				if match {
					testStep.StepDefinition = item
					break
				}
			}

			testCase.Steps = append(testCase.Steps, testStep)
		}

		// filter and add hooks
		for _, hook := range c.beforeHooks {
			if c.matchTags(hook.tags, pickle.Tags) {
				testCase.BeforeHooks = append(testCase.BeforeHooks, hook.fn.(BeforeHook))
			}
		}

		for _, hook := range c.afterHooks {
			if c.matchTags(hook.tags, pickle.Tags) {
				testCase.AfterHooks = append(testCase.AfterHooks, hook.fn.(AfterHook))
			}
		}

		return testCase, nil
	}
	return nil, nil
}

func (c *Cucumber) compileStepDefinitionText(expression string) string {
	sb := ""
	matches := paramPattern.FindAllStringSubmatch(expression, -1)
	matchIndexes := paramPattern.FindAllStringSubmatchIndex(expression, -1)
	lastIndex := 0
	for index := 0; index < len(matches); index++ {
		matchIndex := matchIndexes[index]
		text := expression[lastIndex:matchIndex[0]]
		captureRegexp := fmt.Sprintf("\"{arg%d}\"", index)
		lastIndex = matchIndex[1]
		sb += text
		sb += captureRegexp
	}
	sb += expression[lastIndex:]
	return strings.TrimSpace(sb)
}

func (c *Cucumber) matchTags(requiredTags []string, tags []string) bool {
	match := true
	for _, andTags := range requiredTags {
		orMatch := false
		orTags := strings.Split(andTags, ",")
		for _, orTag := range orTags {
			if utils.SetExists(tags, orTag) {
				orMatch = true
				break
			}
		}
		if !orMatch {
			match = false
			break
		}

	}
	return match
}
