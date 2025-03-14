// prompt.go
package pkg

import (
	"github.com/AlecAivazis/survey/v2"
)

func AskQuestion(question string) string {
	var answer string
	prompt := &survey.Input{Message: question}
	survey.AskOne(prompt, &answer)
	return answer
}

func InfoText(message string) string {
	var name string
	prompt := &survey.Input{
		Message: message,
	}
	survey.AskOne(prompt, &name, survey.WithIcons(func(icons *survey.IconSet) {
		icons.Question.Text = "✓"
	}))
	return name
}

func MultiSelect(message string, options []string) []string {
	selectedOptions := []string{}
	prompt := &survey.MultiSelect{
		Message: message,
		Options: options,
	}
	survey.AskOne(prompt, &selectedOptions)
	return selectedOptions
}
