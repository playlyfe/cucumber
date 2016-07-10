package core

import (
	//"fmt"
	"regexp"
	"strings"
)

var variablePattern = regexp.MustCompile("\\{([^}:]+)(:([^}]+))?}")
var paramPattern = regexp.MustCompile("\"[^\"]+\"")

type argument struct {
	offset           int
	value            string
	transformedValue interface{}
}

type CucumberExpression struct {
	transforms []*Transform
	Rawexp     string
	Regexp     *regexp.Regexp
}

func (e *CucumberExpression) match(text string) (bool, []*argument, error) {
	arguments := []*argument{}

	matches := e.Regexp.FindAllStringSubmatch(text, -1)
	if len(matches) == 0 {
		return e.Regexp.MatchString(text), nil, nil
	}
	for index, match := range matches[0] {
		if index < 1 {
			continue
		}
		offset := strings.Index(text, match)
		transformedValue, err := e.transforms[index-1].Transformer(match)
		if err != nil {
			return true, nil, err
		}
		arguments = append(arguments, &argument{
			offset:           offset,
			value:            match,
			transformedValue: transformedValue,
		})
	}
	return true, arguments, nil
}

func newCucumberExpression(expression string, targetTypes []string, transformLookup map[string]*Transform) *CucumberExpression {
	e := &CucumberExpression{}
	sb := "^"
	typeNameIndex := 0
	matches := variablePattern.FindAllStringSubmatch(expression, -1)
	matchIndexes := variablePattern.FindAllStringSubmatchIndex(expression, -1)
	lastIndex := 0
	for index := 0; index < len(matches); index++ {
		matchIndex := matchIndexes[index]
		expressionTypeName := matches[index][3]
		targetType := ""
		if len(targetTypes) <= typeNameIndex {
			targetType = ""
		} else {
			targetType = targetTypes[typeNameIndex]
			typeNameIndex++
		}
		var transform *Transform
		if expressionTypeName != "" {
			transform = transformLookup[expressionTypeName]
		} else if targetType != "" {
			transform = transformLookup[targetType]
		} else {
			transform = transformLookup["string"]
		}
		e.transforms = append(e.transforms, transform)

		text := regexp.QuoteMeta(expression[lastIndex:matchIndex[0]])
		captureRegexp := "(" + transform.CaptureRegexp + ")"
		lastIndex = matchIndex[1]
		sb += text
		sb += captureRegexp
	}
	sb += regexp.QuoteMeta(expression[lastIndex:])
	sb += "$"
	e.Rawexp = sb
	e.Regexp = regexp.MustCompile(sb)
	return e
}
