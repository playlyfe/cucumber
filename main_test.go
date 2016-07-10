package main

import (
	"github.com/playlyfe/cucumber/core"
	"github.com/playlyfe/cucumber/formatter"
	"testing"
)

var cucumber *core.Cucumber

func TestMain(m *testing.M) {
	cucumber = NewCucumber()
	cucumber.AddOuputFormatter(formatter.NewPrettyFormatter())

	Given := cucumber.Step()
	//And := Given
	//When := Given

	cucumber.AddTransform("date", &core.Transform{
		CaptureRegexp: "\\d+-\\d+-\\d+",
		Transformer: func(value string) (interface{}, error) {
			return value, nil
		},
	})

	Given("a scenario with:", func(world interface{}, _ string) error {
		// Write your step definition here
		//println(text)
		return nil
	})

	Given("the following feature:", func(world interface{}, _ string) error {
		// Write your step definition here
		//println(text)
		return nil
	})

	m.Run()
}

func TestCucumberTCKFeatures(t *testing.T) {
	cucumber.Execute(&core.ExecuteParams{
		FeaturesPath: "features/core.feature",
	})
}
