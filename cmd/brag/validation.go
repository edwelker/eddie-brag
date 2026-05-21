package main

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"
)

// validateEnrichment checks bucket-specific required fields and warns user
// Returns true if validation passes or user confirms to proceed anyway
func validateEnrichment(bucket, businessMetric, peerRecognition string) bool {
	var warnings []string

	switch bucket {
	case "Process":
		if businessMetric == "" {
			warnings = append(warnings, "Process entries should have a business metric (impact measurement)")
		}
	case "Leadership":
		if peerRecognition == "" {
			warnings = append(warnings, "Leadership entries should have peer recognition (social proof)")
		}
	}

	if len(warnings) == 0 {
		return true
	}

	// Show warnings
	fmt.Println("\n⚠️  Quality check:")
	for _, warning := range warnings {
		fmt.Printf("  • %s\n", warning)
	}
	fmt.Println()

	// Ask if they want to continue anyway
	var proceed bool
	confirmPrompt := &survey.Confirm{
		Message: "Continue with incomplete enrichment? (Not recommended for promotion/review prep)",
		Default: false,
	}
	if err := survey.AskOne(confirmPrompt, &proceed); err != nil {
		return false
	}

	return proceed
}

// promptRequiredBusinessMetric loops until user provides a value or explicitly skips
func promptRequiredBusinessMetric(bucket string) string {
	for {
		metric := promptBusinessMetric(bucket)
		if metric != "" {
			return metric
		}

		// Empty response - confirm they want to skip
		var skip bool
		confirmPrompt := &survey.Confirm{
			Message: "Business metric is important for Process entries. Skip anyway?",
			Default: false,
		}
		if err := survey.AskOne(confirmPrompt, &skip); err != nil {
			return ""
		}
		if skip {
			return ""
		}
		// Loop back to re-prompt
	}
}

// promptRequiredPeerRecognition loops until user provides a value or explicitly skips
func promptRequiredPeerRecognition() string {
	for {
		recognition := promptPeerRecognition()
		if recognition != "" {
			return recognition
		}

		// Empty response - confirm they want to skip
		var skip bool
		confirmPrompt := &survey.Confirm{
			Message: "Peer recognition is important for Leadership entries. Skip anyway?",
			Default: false,
		}
		if err := survey.AskOne(confirmPrompt, &skip); err != nil {
			return ""
		}
		if skip {
			return ""
		}
		// Loop back to re-prompt
	}
}
