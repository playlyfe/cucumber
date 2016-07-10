package core

import (
	"regexp"

	"github.com/cucumber/gherkin-go"

	"github.com/playlyfe/cucumber/utils"
)

var outlinePattern = regexp.MustCompile("(<[^>]+>)")

type Pickle struct {
	Tags     []string
	Steps    []*PickleStep
	Feature  *gherkin.Feature
	FilePath string
}

type PickleStep struct {
	//Arguments  []*argument
	//Expression *CucumberExpression
	Step *gherkin.Step
	Text string
}

type Compiler struct {
	stepDefinitions []*StepDefinition
	transformLookup map[string]*Transform
}

func (c *Cucumber) compileFeatureFile(doc *featureFile) ([]*Pickle, error) {
	pickles := []*Pickle{}
	var background *gherkin.Background
	for _, child := range doc.document.Feature.Children {
		switch node := child.(type) {
		case *gherkin.Background:
			background = node
		case *gherkin.Scenario:
			pickle := &Pickle{
				Tags:     []string{},
				Steps:    []*PickleStep{},
				Feature:  doc.document.Feature,
				FilePath: doc.path,
			}

			// Add Feature Tags to the Pickle
			for _, tag := range doc.document.Feature.Tags {
				pickle.Tags = utils.SetAdd(pickle.Tags, tag.Name)
			}

			// Add Scenario Tags to the Pickle
			for _, tag := range node.Tags {
				pickle.Tags = utils.SetAdd(pickle.Tags, tag.Name)
			}

			if background != nil {
				pickle.Steps = append(pickle.Steps, c.compileSteps(background.Steps)...)
			}

			pickle.Steps = append(pickle.Steps, c.compileSteps(node.Steps)...)

			pickles = append(pickles, pickle)

		case *gherkin.ScenarioOutline:
			for _, example := range node.Examples {
				columnLookup := map[string]int{}
				for index, header := range example.TableHeader.Cells {
					columnLookup[header.Value] = index
				}
				for _, row := range example.TableBody {
					pickle := &Pickle{
						Tags:     []string{},
						Steps:    []*PickleStep{},
						Feature:  doc.document.Feature,
						FilePath: doc.path,
					}

					// Add Feature Tags to the Pickle
					for _, tag := range doc.document.Feature.Tags {
						pickle.Tags = utils.SetAdd(pickle.Tags, tag.Name)
					}

					// Add Scenario Tags to the Pickle
					for _, tag := range node.Tags {
						pickle.Tags = utils.SetAdd(pickle.Tags, tag.Name)
					}

					if background != nil {
						pickle.Steps = append(pickle.Steps, c.compileSteps(background.Steps)...)
					}

					pickle.Steps = append(pickle.Steps, c.compileStepOutlines(node.Steps, columnLookup, row)...)

					pickles = append(pickles, pickle)
				}
			}

		}
	}
	return pickles, nil
}

func (c *Cucumber) compileSteps(steps []*gherkin.Step) []*PickleStep {
	pickleSteps := []*PickleStep{}
	for _, step := range steps {
		pickleSteps = append(pickleSteps, &PickleStep{
			Step: step,
			Text: step.Text,
		})
	}
	return pickleSteps
}

func (c *Cucumber) compileStepOutlines(steps []*gherkin.Step, columnLookup map[string]int, row *gherkin.TableRow) []*PickleStep {
	pickleSteps := []*PickleStep{}
	for _, step := range steps {
		text := ""
		var matchIndexes [][]int

		matchIndexes = outlinePattern.FindAllStringSubmatchIndex(step.Text, -1)
		lastIndex := 0
		for _, matchIndex := range matchIndexes {
			if len(matchIndex) <= 2 {
				continue
			}
			text += step.Text[lastIndex:matchIndex[2]]
			text += row.Cells[columnLookup[step.Text[matchIndex[2]:matchIndex[3]]]].Value
			lastIndex = matchIndex[3]
		}
		text += step.Text[lastIndex:]

		pickleSteps = append(pickleSteps, &PickleStep{
			Step: step,
			Text: text,
		})
	}

	return pickleSteps
}

/*

func NewCompiler(stepDefinitions []*stepDefinition, transformLookup map[string]*Transform) *Compiler {
	c := &Compiler{
		stepDefinitions: stepDefinitions,
		transformLookup: transformLookup,
	}
	return c
}

func (c *Compiler) compile(files []*featureFile) ([]*TestCase, error) {
	testCases := []*TestCase{}
	for _, file := range files {
		documentTestCases := c.compileDocument(file)
		testCases = append(testCases, documentTestCases...)
	}
	return testCases
}

func (c *Compiler) compileDocument(file *featureFile) []*TestCase {
	testCases := []*TestCase{}
	document := file.document
	var currentBackground *gherkin.Background

	for _, item := range document.Feature.Children {
		switch child := item.(type) {
		case *gherkin.Background:
			currentBackground = child
		case *gherkin.Scenario:
			testCase := &TestCase{
				FeatureNode:    document.Feature,
				BackgroundNode: currentBackground,
				FilePath:       file.path,
			}
			if currentBackground != nil {
				testCase.BackgroundSteps, testCase.LongestBackgroundLine = c.compileSteps(currentBackground.Steps, currentBackground)
			}
			testCase.ScenarioSteps, testCase.LongestScenarioLine = c.compileSteps(child.Steps, child)
			testCases = append(testCases, testCase)
		case *gherkin.ScenarioOutline:

		}
	}
	return testCases
}

func (c *Compiler) compileSteps(steps []*gherkin.Step, scenarioNode interface{}) ([]*TestStep, int) {
	testSteps := []*TestStep{}
	longestLine := 0
	for index, step := range steps {
		testStep := &TestStep{
			StepNode: step,
		}
		testStep.LineLength = utf8.RuneCountInString(step.Keyword + step.Text)
		testStep.Text = c.compileExpression(step.Text)
		if testStep.LineLength > longestLine {
			longestLine = testStep.LineLength
		}
		if index == 0 {
			testStep.ScenarioNode = scenarioNode
		} else {
			testStep.ScenarioNode = nil
		}
		for _, item := range c.stepDefinitions {
			match, arguments, err := item.expression.match(testStep.Text)
			if err != nil {
				panic(err)
			}
			if docString, ok := step.Argument.(*gherkin.DocString); ok {
				arguments = append(arguments, &argument{
					transformedValue: docString.Content,
				})
			}
			testStep.Arguments = arguments
			if match {
				testStep.Expression = item.expression
				testStep.StepDefinition = item.fn
				break
			}
		}
		testSteps = append(testSteps, testStep)
	}
	return testSteps, longestLine
}

func (c *Compiler) compileExpression(expression string) string {
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

func (c *Compiler) compileStepDefintion(fn interface{}) error {

}

func (c *Compiler) compileOutlineSteps(steps []*gherkin.Step, scenarioNode interface{}, columnLookup map[string]int, row *gherkin.TableRow) ([]*TestStep, int) {
	testSteps := []*TestStep{}
	longestLine := 0
	for index, step := range steps {
		testStep := &TestStep{
			StepNode: step,
		}
		text := ""
		var matchIndexes [][]int

		matchIndexes = outlinePattern.FindAllStringSubmatchIndex(step.Text, -1)
		lastIndex := 0
		for _, matchIndex := range matchIndexes {
			if len(matchIndex) <= 2 {
				continue
			}
			text += step.Text[lastIndex:matchIndex[2]]
			text += row.Cells[columnLookup[step.Text[matchIndex[2]:matchIndex[3]]]].Value
			lastIndex = matchIndex[3]
		}
		text += step.Text[lastIndex:]

		testStep.LineLength = utf8.RuneCountInString(step.Keyword + text)
		testStep.Text = c.compileExpression(text)

		if testStep.LineLength > longestLine {
			longestLine = testStep.LineLength
		}
		if index == 0 {
			testStep.ScenarioNode = scenarioNode
		} else {
			testStep.ScenarioNode = nil
		}
		for _, item := range c.stepDefinitions {
			match, arguments, err := item.expression.match(testStep.Text)
			if err != nil {
				panic(err)
			}
			if docString, ok := step.Argument.(*gherkin.DocString); ok {
				arguments = append(arguments, &argument{
					transformedValue: docString.Content,
				})
			}
			testStep.Arguments = arguments
			if match {
				testStep.Expression = item.expression
				testStep.StepDefinition = item.fn
				break
			}
		}
		testSteps = append(testSteps, testStep)
	}
	return testSteps, longestLine
}
*/
