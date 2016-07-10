package main

import (
	"os"
	"strconv"
	"strings"

	"github.com/playlyfe/cucumber/core"
	"github.com/playlyfe/cucumber/formatter"
)

func main() {
	featureFilesPath := "features"
	if len(os.Args) > 1 {
		featureFilesPath = strings.TrimSuffix(os.Args[1], "/")
	}
	cucumber := NewCucumber()
	cucumber.AddOuputFormatter(formatter.NewPrettyFormatter())
	cucumber.Execute(&core.ExecuteParams{
		FeaturesPath: featureFilesPath,
	})
}

func NewCucumber() *core.Cucumber {
	c := core.NewCucumber()
	c.AddTransform("int", &core.Transform{
		CaptureRegexp: "-?\\d+",
		Transformer: func(value string) (interface{}, error) {
			return strconv.ParseInt(value, 10, 64)
		},
	})
	c.AddTransform("float", &core.Transform{
		CaptureRegexp: "-?\\d*\\.?\\d+",
		Transformer: func(value string) (interface{}, error) {
			return strconv.ParseFloat(value, 64)
		},
	})
	c.AddTransform("string", &core.Transform{
		CaptureRegexp: ".+",
		Transformer: func(value string) (interface{}, error) {
			return value, nil
		},
	})
	return c
}
