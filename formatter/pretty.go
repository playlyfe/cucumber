package formatter

import (
	"fmt"
	"strings"

	"github.com/cucumber/gherkin-go"
	"github.com/fatih/color"

	"github.com/playlyfe/cucumber/core"
)

var colorFeature = color.New(color.FgWhite).Add(color.Bold).SprintFunc()
var colorDescription = color.New(color.FgWhite).SprintFunc()
var colorScenario = color.New(color.FgWhite).Add(color.Bold).SprintFunc()
var colorComment = color.New(color.FgWhite).SprintFunc()

var colorUndefined = color.New(color.FgYellow).SprintFunc()
var colorPending = color.New(color.FgYellow).SprintFunc()
var colorPendingParam = color.New(color.FgYellow).Add(color.Bold).SprintFunc()
var colorSkipped = color.New(color.FgCyan).SprintFunc()
var colorSkippedParam = color.New(color.FgCyan).Add(color.Bold).SprintFunc()
var colorFailed = color.New(color.FgRed).SprintFunc()
var colorFailedParam = color.New(color.FgRed).Add(color.Bold).SprintFunc()
var colorPassed = color.New(color.FgGreen).SprintFunc()
var colorPassedParam = color.New(color.FgGreen).Add(color.Bold).SprintFunc()
var colorTag = color.New(color.FgCyan).SprintFunc()

type prettyFormatter struct {
	filePath        string
	lineLength      int
	currentTestCase *core.TestCase
}

func NewPrettyFormatter() func(event *core.Event) {
	formatter := &prettyFormatter{}
	return func(event *core.Event) {
		formatter.handleEvent(event)
	}
}

func (p *prettyFormatter) handleEvent(event *core.Event) {
	switch event.Name {
	case core.TestCaseStarting:
		testCase := event.Data.(*core.TestCase)
		p.filePath = testCase.Pickle.FilePath
		p.currentTestCase = testCase
		p.feature(testCase.Pickle.Feature)
	case core.TestStepFinished:
		testStep := event.Data.(*core.TestStep)
		p.step(testStep)
	case core.TestCaseFinished:
		fmt.Printf("\n")
	case core.TestRunFinished:
	}
}

func (p *prettyFormatter) feature(node *gherkin.Feature) {
	p.tags(node.Tags, "")
	fmt.Printf("%s\n\n", colorFeature("Feature: "+node.Name))
}
func (p *prettyFormatter) step(testStep *core.TestStep) {
	line := fmt.Sprintf("%d", testStep.PickleStep.Step.Location.Line)
	colorFn := colorPending
	colorParamFn := colorPendingParam
	switch testStep.Result {
	case core.PassedResult:
		colorFn = colorPassed
		colorParamFn = colorPassedParam
	case core.FailedResult:
		colorFn = colorFailed
		colorParamFn = colorFailedParam
	case core.SkippedResult:
		colorFn = colorSkipped
		colorParamFn = colorSkippedParam
	}

	text := colorFn(testStep.PickleStep.Step.Keyword)
	var matchIndexes [][]int
	if testStep.StepDefinition != nil {
		matchIndexes = testStep.StepDefinition.Expression.Regexp.FindAllStringSubmatchIndex(testStep.PickleStep.Text, -1)
		lastIndex := 0
		for _, matchIndex := range matchIndexes {
			if len(matchIndex) <= 2 {
				continue
			}
			text += colorFn(testStep.PickleStep.Text[lastIndex:matchIndex[2]])
			text += colorParamFn(testStep.PickleStep.Text[matchIndex[2]:matchIndex[3]])
			lastIndex = matchIndex[3]
		}
		text += colorFn(testStep.PickleStep.Text[lastIndex:])
	} else {
		text += colorFn(testStep.PickleStep.Text)
	}
	fmt.Printf("    ")
	fmt.Printf(text)
	//for count := 0; count < p.lineLength-testStep.LineLength; count++ {
	//	fmt.Printf(" ")
	//}
	fmt.Printf("  %s:%s\n", colorComment("# "+p.filePath), colorComment(line))
	if docString, ok := testStep.PickleStep.Step.Argument.(*gherkin.DocString); ok {
		fmt.Printf("    %s\n", colorParamFn(docString.Delimitter))
		lines := strings.Split(docString.Content, "\n")
		for _, line := range lines {
			fmt.Printf("    %s\n", colorParamFn(line))
		}
		fmt.Printf("    %s\n", colorParamFn(docString.Delimitter))
	}
}

func (p *prettyFormatter) tags(tags []*gherkin.Tag, indent string) {
	for index, tag := range tags {
		if index > 0 {
			fmt.Printf(" ")
		} else {
			fmt.Printf(indent)
		}
		fmt.Printf("%s", colorTag(tag.Name))
		if index == len(tags)-1 {
			fmt.Printf("\n")
		}
	}
}
