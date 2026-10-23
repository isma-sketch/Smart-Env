package ui

import (
	"fmt"
	"strings"

	"github.com/manifoldco/promptui"
)

// PromptValue prompts for an env value. If secret, masks input.
func PromptValue(key, example, defaultVal string, secret bool) (string, error) {
	label := key
	if example != "" && example != defaultVal {
		label = fmt.Sprintf("%s (example: %s)", key, example)
	}
	p := promptui.Prompt{
		Label:   label,
		Default: defaultVal,
	}
	if secret {
		p.Mask = '*'
	}
	// validate non-empty? allow empty
	res, err := p.Run()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(res), nil
}

// Confirm asks yes/no.
func Confirm(label string, defaultYes bool) (bool, error) {
	p := promptui.Prompt{
		Label:     label,
		IsConfirm: true,
	}
	if defaultYes {
		p.Default = "y"
	} else {
		p.Default = "n"
	}
	res, err := p.Run()
	if err != nil {
		// EOF or interrupt -> default
		return defaultYes, nil
	}
	r := strings.ToLower(strings.TrimSpace(res))
	return r == "y" || r == "yes", nil
}

// SelectOption presents a list.
func SelectOption(label string, items []string) (int, string, error) {
	p := promptui.Select{
		Label: label,
		Items: items,
	}
	i, res, err := p.Run()
	if err != nil {
		return -1, "", err
	}
	return i, res, nil
}

// PromptGenerate asks whether to generate. Returns action: "generate", "manual", "skip".
func PromptGenerate(key string) (string, error) {
	_, res, err := SelectOption(fmt.Sprintf("Resolve %s", key), []string{"Generate secure value", "Enter manually", "Skip (leave empty)"})
	if err != nil {
		return "", err
	}
	switch res {
	case "Generate secure value":
		return "generate", nil
	case "Enter manually":
		return "manual", nil
	default:
		return "skip", nil
	}
}
