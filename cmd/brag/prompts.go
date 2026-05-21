package main

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/edwelker/eddie-brag/internal/brag"
)

func promptBucket() string {
	var bucket string

	fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("👤 Staff Software Engineer - Impact Framework")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Println("Staff engineers demonstrate impact across 4 dimensions.")
	fmt.Println("Balance your portfolio - aim for diverse evidence, not just execution.")
	fmt.Println()
	fmt.Println("  Delivery - Execution")
	fmt.Println("    Shipped code, infrastructure deployments, and bug fixes")
	fmt.Println("    Examples: merging application PRs, migrating databases, resolving API")
	fmt.Println("    performance bottlenecks, deploying new services")
	fmt.Println()
	fmt.Println("  Architecture - Discovery")
	fmt.Println("    Navigating ambiguity, system planning, and tech debt analysis")
	fmt.Println("    Examples: auditing legacy systems, writing RFCs or design docs, defining")
	fmt.Println("    technical roadmaps, evaluating new tools")
	fmt.Println()
	fmt.Println("  Process - Velocity")
	fmt.Println("    Systemic solutions, CI/CD, automation, and tooling")
	fmt.Println("    Examples: fixing dead QA GitHub Actions pipelines, building Slack PR")
	fmt.Println("    reporting bots, cutting build times, writing runbooks to reduce toil")
	fmt.Println()
	fmt.Println("  Leadership - Multiplication")
	fmt.Println("    Glue work, mentoring, cross-team mediation, and unblocking peers")
	fmt.Println("    Examples: unblocking the QA team's test rewrites, leading blameless")
	fmt.Println("    postmortems, conducting deep code reviews, resolving architectural disputes")
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	prompt := &survey.Select{
		Message: "Select bucket:",
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
			if err := survey.AskOne(confirmPrompt, &confirm); err != nil {
				return ""
			}
			if !confirm {
				continue
			}
		}
		break
	}

	return evidence
}

func promptEvidenceWithOptions() string {
	fmt.Println("\n⚠️  Evidence missing. This entry will be harder to verify during reviews.")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  1. Add evidence URL now")
	fmt.Println("  2. Add placeholder (e.g., \"Slack thread - need link\")")
	fmt.Println("  3. Skip (will prompt during brag review)")
	fmt.Println()

	var choice string
	choicePrompt := &survey.Select{
		Message: "Choose option:",
		Options: []string{"Add URL now", "Add placeholder", "Skip"},
		Default: "Add URL now",
	}
	if err := survey.AskOne(choicePrompt, &choice); err != nil {
		return ""
	}

	switch choice {
	case "Add URL now":
		return promptEvidenceWithValidation()
	case "Add placeholder":
		var placeholder string
		placeholderPrompt := &survey.Input{
			Message: "Placeholder text:",
			Default: "Slack thread - need link",
		}
		if err := survey.AskOne(placeholderPrompt, &placeholder); err != nil {
			return ""
		}
		return placeholder
	case "Skip":
		return ""
	}

	return ""
}

func promptHoursSaved() (*float64, string) {
	var hoursInput string

	for {
		prompt := &survey.Input{
			Message: "Hours Saved (e.g., 1.5, 90m, 2h, or 0 for none):",
		}
		if err := survey.AskOne(prompt, &hoursInput); err != nil {
			return nil, ""
		}

		if hoursInput == "" {
			return nil, ""
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

		// Prompt for calculation notes
		var calculation string
		calcPrompt := &survey.Input{
			Message: "Calculation (e.g., '585 builds × 3 min + 30 min local', press Enter to skip):",
		}
		if err := survey.AskOne(calcPrompt, &calculation); err != nil {
			return &hours, ""
		}

		return &hours, calculation
	}
}

func promptBusinessMetric(bucket string) string {
	var metric string

	// Show bucket-specific examples
	fmt.Println("\nBusiness Metric - What measurable outcome did this achieve?")
	switch bucket {
	case "Delivery":
		fmt.Println("  Examples:")
		fmt.Println("    • Shipped feature enabling $X ARR pipeline")
		fmt.Println("    • Resolved P0 blocker affecting X customers")
		fmt.Println("    • Deployed X services to production")
	case "Architecture":
		fmt.Println("  Examples:")
		fmt.Println("    • Reduced technical debt by X hours/sprint")
		fmt.Println("    • Enabled team to deprecate X legacy systems")
		fmt.Println("    • Unblocked X teams waiting on design decision")
	case "Process":
		fmt.Println("  Examples:")
		fmt.Println("    • Reduced deploy time from X to Y minutes")
		fmt.Println("    • Increased release frequency from X to Y per week")
		fmt.Println("    • Saved $X in compute/infrastructure costs")
		fmt.Println("    • Reduced incident count by X%")
	case "Leadership":
		fmt.Println("  Examples:")
		fmt.Println("    • Unblocked X engineers/teams")
		fmt.Println("    • Mentored X engineers to promotion")
		fmt.Println("    • Resolved X-week cross-team conflict")
		fmt.Println("    • Reduced onboarding time from X to Y weeks")
	}
	fmt.Println()

	prompt := &survey.Input{
		Message: "Your metric (press Enter to skip):",
	}
	if err := survey.AskOne(prompt, &metric); err != nil {
		return ""
	}
	return metric
}

func promptStrategicAlign() string {
	var choice string

	fmt.Println("\n💡 OKR = Objectives and Key Results (quarterly goals with measurable outcomes)")
	fmt.Println("   Example: 'Improve test reliability' with KR 'Reduce CI time from 18min to <10min'")
	fmt.Println()

	prompt := &survey.Select{
		Message: "Strategic Alignment - What org goal does this support?",
		Options: []string{
			"Developer velocity / feedback loops",
			"CI/build performance improvement",
			"Tech debt reduction",
			"Security / compliance",
			"Team enablement / knowledge sharing",
			"Process automation / tooling",
			"QA/testing infrastructure",
			"Operational excellence",
			"Skip (no specific alignment)",
		},
	}
	if err := survey.AskOne(prompt, &choice); err != nil {
		return ""
	}

	if choice == "Skip (no specific alignment)" {
		return ""
	}

	return choice
}

func promptPeerRecognition() string {
	var choice string

	prompt := &survey.Select{
		Message: "Peer Recognition - Who acknowledged this work?",
		Options: []string{
			"Team feedback / positive reactions",
			"Mentioned in Slack or retro",
			"Adopted by other teams",
			"Manager / leadership noticed",
			"None / don't remember",
		},
	}
	if err := survey.AskOne(prompt, &choice); err != nil {
		return ""
	}

	if choice == "None / don't remember" {
		return ""
	}

	// If they selected a category, prompt for details
	var details string
	detailPrompt := &survey.Input{
		Message: fmt.Sprintf("%s - provide details (Slack link, specifics):", choice),
	}
	if err := survey.AskOne(detailPrompt, &details); err != nil {
		return choice
	}

	if details != "" {
		return fmt.Sprintf("%s: %s", choice, details)
	}

	return choice
}

func promptDateOption() string {
	var choice string
	prompt := &survey.Select{
		Message: "When did this work happen?",
		Options: []string{"Today", "Specific date", "Week number", "Month number"},
	}
	if err := survey.AskOne(prompt, &choice); err != nil {
		return "Today"
	}
	return choice
}

func promptSpecificDate(label string) string {
	var dateStr string
	prompt := &survey.Input{
		Message: fmt.Sprintf("%s (YYYY-MM-DD):", label),
	}
	if err := survey.AskOne(prompt, &dateStr, survey.WithValidator(survey.Required)); err != nil {
		return ""
	}
	return dateStr
}

func promptWeekNumber() int {
	var weekStr string
	prompt := &survey.Input{
		Message: "Week number (relative to role start):",
	}
	if err := survey.AskOne(prompt, &weekStr, survey.WithValidator(survey.Required)); err != nil {
		return 0
	}
	var weekNum int
	if _, err := fmt.Sscanf(weekStr, "%d", &weekNum); err != nil {
		return 0
	}
	return weekNum
}

func promptMonthNumber() int {
	var monthStr string
	prompt := &survey.Input{
		Message: "Month number (relative to role start):",
	}
	if err := survey.AskOne(prompt, &monthStr, survey.WithValidator(survey.Required)); err != nil {
		return 0
	}
	var monthNum int
	if _, err := fmt.Sscanf(monthStr, "%d", &monthNum); err != nil {
		return 0
	}
	return monthNum
}

func promptEnrichNow() bool {
	var enrich bool
	prompt := &survey.Confirm{
		Message: "Add enrichment data now (hours saved, metrics)?",
		Default: false,
	}
	if err := survey.AskOne(prompt, &enrich); err != nil {
		return false
	}
	return enrich
}

func promptStatus() string {
	var status string

	fmt.Println("\nStatus - Is this work completed, ongoing, or proposed?")
	fmt.Println("  Completed: Work is shipped and verified")
	fmt.Println("  In Progress: Work is underway but not shipped")
	fmt.Println("  Proposed: Design/RFC created, awaiting approval or implementation")
	fmt.Println("  Abandoned: Work was started but not continued")
	fmt.Println()

	prompt := &survey.Select{
		Message: "Select status:",
		Options: []string{"Completed", "In Progress", "Proposed", "Abandoned"},
		Default: "Completed",
	}
	if err := survey.AskOne(prompt, &status); err != nil {
		return "Completed"
	}
	return status
}

func promptStatusUpdate(currentStatus string) string {
	var updateStatus bool

	updatePrompt := &survey.Confirm{
		Message: fmt.Sprintf("Status is currently '%s'. Update status?", currentStatus),
		Default: false,
	}
	if err := survey.AskOne(updatePrompt, &updateStatus); err != nil || !updateStatus {
		return ""
	}

	return promptStatus()
}
