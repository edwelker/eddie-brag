package main

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/edwelker/eddie-brag/internal/brag"
)

func promptBucket() string {
	var bucket string
	prompt := &survey.Select{
		Message: "Work Context:",
		Options: []string{"Delivery", "Architecture", "Process", "Leadership"},
	}
	if err := survey.AskOne(prompt, &bucket); err != nil {
		fmt.Printf("Error: %v\n", err)
		return ""
	}
	return bucket
}

func promptDescription() string {
	var description string
	prompt := &survey.Input{
		Message: "Description:",
	}
	if err := survey.AskOne(prompt, &description, survey.WithValidator(survey.Required)); err != nil {
		fmt.Printf("Error: %v\n", err)
		return ""
	}
	return description
}

func promptEvidenceWithValidation() string {
	var evidence string
	client := brag.NewDefaultHTTPClient()

	for {
		prompt := &survey.Input{
			Message: "Evidence URL:",
		}
		if err := survey.AskOne(prompt, &evidence, survey.WithValidator(survey.Required)); err != nil {
			fmt.Printf("Error: %v\n", err)
			return ""
		}

		// Validate URL
		valid, err := brag.ValidateURL(evidence, client)
		if err != nil || !valid {
			var confirm bool
			confirmPrompt := &survey.Confirm{
				Message: fmt.Sprintf("Warning: URL validation failed (%v). Continue anyway?", err),
				Default: false,
			}
			survey.AskOne(confirmPrompt, &confirm)
			if !confirm {
				continue
			}
		}
		break
	}

	return evidence
}

func promptHoursSaved() *float64 {
	var hoursInput string

	for {
		prompt := &survey.Input{
			Message: "Hours Saved (e.g., 1.5, 90m, 2h, or 0 for none):",
		}
		if err := survey.AskOne(prompt, &hoursInput); err != nil {
			return nil
		}

		if hoursInput == "" {
			return nil
		}

		hours, err := brag.ParseHoursInput(hoursInput)
		if err != nil {
			fmt.Printf("Invalid input: %v\n", err)
			continue
		}

		if hours < 0 {
			fmt.Println("Hours saved must be >= 0.")
			continue
		}

		return &hours
	}
}

func promptBusinessMetric() string {
	var metric string
	prompt := &survey.Input{
		Message: "Business Metric (press Enter to skip):",
	}
	survey.AskOne(prompt, &metric)
	return metric
}

func promptStrategicAlign() string {
	var align string
	prompt := &survey.Input{
		Message: "Strategic Alignment (press Enter to skip):",
	}
	survey.AskOne(prompt, &align)
	return align
}

func promptPeerRecognition() string {
	var recognition string
	prompt := &survey.Input{
		Message: "Peer Recognition (press Enter to skip):",
	}
	survey.AskOne(prompt, &recognition)
	return recognition
}
