package core

import (
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/cucumber/gherkin-go"
)

// Transform represents a cucumber transform expression
type Transform struct {
	CaptureRegexp string
	Transformer   func(value string) (interface{}, error)
}

// Cucumber is a new cucumber
type Cucumber struct {
	World           interface{}
	stepDefinitions []*StepDefinition
	transformLookup map[string]*Transform
	beforeAllHooks  []*hook
	afterAllHooks   []*hook
	beforeHooks     []*hook
	afterHooks      []*hook
	eventBus        *EventBus
}

type BeforeHook func(world interface{}) error
type AfterHook func(world interface{}) error
type AroundHook func(world interface{}, next func() error) error

type hook struct {
	tags []string
	fn   interface{}
}

type CucumberError struct {
	Name        string
	Description string
}

func (e *CucumberError) Error() string {
	return e.Name + ": " + e.Description
}

type StepDefinition struct {
	Expression *CucumberExpression
	Fn         interface{}
}

type file struct {
	path   string
	reader *os.File
}

type featureFile struct {
	path     string
	document *gherkin.GherkinDocument
}

// NewCucumber creates a new Cucumber
func NewCucumber() *Cucumber {
	return &Cucumber{
		stepDefinitions: []*StepDefinition{},
		transformLookup: map[string]*Transform{},
		eventBus:        NewEventBus(),
	}
}

func (c *Cucumber) AddTransform(typeName string, transform *Transform) {
	c.transformLookup[typeName] = transform
}

func (c *Cucumber) AddOuputFormatter(printHandler EventHandler) {
	c.eventBus.RegisterHandler(TestRunStarting, printHandler)
	c.eventBus.RegisterHandler(TestCaseStarting, printHandler)
	c.eventBus.RegisterHandler(TestStepStarting, printHandler)
	c.eventBus.RegisterHandler(TestStepFinished, printHandler)
	c.eventBus.RegisterHandler(TestCaseFinished, printHandler)
	c.eventBus.RegisterHandler(TestRunFinished, printHandler)
}

func (c *Cucumber) BeforeAll(fn BeforeHook) {
	c.beforeAllHooks = append(c.beforeAllHooks, &hook{
		fn: fn,
	})
}

func (c *Cucumber) AfterAll(fn AfterHook) {
	c.afterAllHooks = append(c.afterAllHooks, &hook{
		fn: fn,
	})
}

func (c *Cucumber) Before(fn BeforeHook, tags ...string) {
	c.beforeHooks = append(c.beforeHooks, &hook{
		tags: tags,
		fn:   fn,
	})
}

func (c *Cucumber) After(fn AfterHook, tags ...string) {
	c.afterHooks = append(c.afterHooks, &hook{
		tags: tags,
		fn:   fn,
	})
}

func (c *Cucumber) Step() func(text string, fn interface{}) {
	return func(text string, fn interface{}) {
		targetTypes := []string{}
		exp := newCucumberExpression(text, targetTypes, c.transformLookup)
		c.stepDefinitions = append(c.stepDefinitions, &StepDefinition{
			Fn:         fn,
			Expression: exp,
		})
	}
}

type ExecuteParams struct {
	FeaturesPath string
	Tags         []string
	Formatter    string
}

func (c *Cucumber) Execute(params *ExecuteParams) error {
	files, err := c.load(params.FeaturesPath)
	if err != nil {
		return err
	}

	featureFiles, err := c.parse(files)
	if err != nil {
		return err
	}

	pickles, err := c.compile(featureFiles)
	if err != nil {
		return err
	}

	testCases, err := c.compose(pickles, params.Tags)
	if err != nil {
		return err
	}

	runner := NewRunner(c.World, testCases, c.eventBus)

	err = runner.ExecuteAllTestCases()
	if err != nil {
		if cerr, ok := err.(*CucumberError); ok {
			println(cerr.Name)
			println(cerr.Description)
		} else {
			return err
		}
	}
	return nil
}

func (c *Cucumber) compile(featureFiles []*featureFile) ([]*Pickle, error) {
	pickles := []*Pickle{}
	for _, item := range featureFiles {
		filePickles, err := c.compileFeatureFile(item)
		if err != nil {
			return nil, err
		}
		pickles = append(pickles, filePickles...)
	}
	return pickles, nil
}

func (c *Cucumber) parse(files []*file) ([]*featureFile, error) {
	featureFiles := []*featureFile{}
	for _, item := range files {
		document, err := gherkin.ParseGherkinDocument(item.reader)
		if err != nil {
			return nil, err
		}
		featureFiles = append(featureFiles, &featureFile{
			path:     item.path,
			document: document,
		})
	}
	return featureFiles, nil
}

func (c *Cucumber) compose(pickles []*Pickle, tags []string) ([]*TestCase, error) {
	testCases := []*TestCase{}
	for _, pickle := range pickles {
		testCase, err := c.composeScenario(pickle, tags)
		if err != nil {
			return nil, err
		}
		if testCase != nil {
			testCases = append(testCases, testCase)
		}
	}
	testCount := len(testCases)
	if testCount > 0 {
		firstTestCase := testCases[0]
		lastTestCase := testCases[testCount-1]
		for _, hook := range c.beforeAllHooks {
			firstTestCase.BeforeHooks = append([]BeforeHook{hook.fn.(BeforeHook)}, firstTestCase.BeforeHooks...)
		}
		for _, hook := range c.afterAllHooks {
			lastTestCase.AfterHooks = append(lastTestCase.AfterHooks, hook.fn.(AfterHook))
		}
	}

	return testCases, nil
}

func (c *Cucumber) load(path string) ([]*file, error) {
	files := []*file{}

	fileInfo, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if fileInfo.IsDir() {
		fileInfos, err := ioutil.ReadDir(path)
		if err != nil {
			return nil, err
		}
		for _, fileInfo := range fileInfos {
			filePath := strings.TrimSuffix(path, string(os.PathSeparator)) + string(os.PathSeparator) + fileInfo.Name()
			if !fileInfo.IsDir() {
				if strings.HasSuffix(fileInfo.Name(), ".feature") {
					reader, err := os.Open(filePath)
					if err != nil {
						return nil, err
					}
					files = append(files, &file{
						reader: reader,
						path:   filePath,
					})
					log.Printf("loading feature %s", filePath)
				}
			} else {
				subDirFiles, err := c.load(filePath)
				if err != nil {
					return nil, err
				}
				files = append(files, subDirFiles...)
			}
		}
	} else {
		reader, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		files = append(files, &file{
			reader: reader,
			path:   path,
		})
		log.Printf("loading feature %s", path)
	}
	return files, nil
}
