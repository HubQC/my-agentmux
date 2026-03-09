package wizard

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/cqi/my_agentmux/internal/config"
)

type StartResult struct {
	Name      string
	AgentType string
	WorkDir   string
}

// NewStartForm creates a new huh.Form for the TUI to use.
func NewStartForm(res *StartResult, presets []string) *huh.Form {
	res.WorkDir = "./"

	options := make([]huh.Option[string], len(presets))
	for i, p := range presets {
		options[i] = huh.NewOption(p, p)
	}

	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("What should we name this agent?").
				Value(&res.Name).
				Validate(func(str string) error {
					if str == "" {
						return fmt.Errorf("name cannot be empty")
					}
					return nil
				}),
			huh.NewSelect[string]().
				Title("Which preset or template do you want to use?").
				Options(options...).
				Value(&res.AgentType),
			huh.NewInput().
				Title("Where should this agent run?").
				Value(&res.WorkDir),
		),
	)
}

// RunStartWizard launches an interactive CLI form to collect agent launch options.
func RunStartWizard(cfg *config.Config, presets []string) (*StartResult, error) {
	var res StartResult
	res.WorkDir = "./"

	options := make([]huh.Option[string], len(presets))
	for i, p := range presets {
		options[i] = huh.NewOption(p, p)
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("What should we name this agent?").
				Value(&res.Name).
				Validate(func(str string) error {
					if str == "" {
						return fmt.Errorf("name cannot be empty")
					}
					return nil
				}),
			huh.NewSelect[string]().
				Title("Which preset or template do you want to use?").
				Options(options...).
				Value(&res.AgentType),
			huh.NewInput().
				Title("Where should this agent run?").
				Value(&res.WorkDir),
		),
	)

	err := form.Run()
	if err != nil {
		return nil, err
	}

	return &res, nil
}
