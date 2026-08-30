package reporter

import (
	api_model "soarca/internal/transport/http/schema"
	runstate_model "soarca/internal/runs/state"
)

const defaultRequestInterval int = 5

func parseRunStateEntry(entry runstate_model.RunEntry) (api_model.PlaybookRunReport, error) {
	playbookStatus := api_model.RunStatusEnum2String(entry.Status)

	playbookStatusText, err := api_model.GetRunStatusText(playbookStatus, api_model.ReportLevelPlaybook)
	if err != nil {
		return api_model.PlaybookRunReport{}, err
	}
	if entry.Error != nil {
		playbookStatusText = playbookStatusText + " - error: " + entry.Error.Error()
	}

	stepResults, err := parseRunStateSteps(entry.StepResults)
	if err != nil {
		return api_model.PlaybookRunReport{}, err
	}

	runReport := api_model.PlaybookRunReport{
		Type:            "run_status",
		Name:            entry.Name,
		Description:     entry.Description,
		RunId:           entry.RunId.String(),
		PlaybookId:      entry.PlaybookId,
		Started:         entry.Started,
		Ended:           entry.Ended,
		Status:          playbookStatus,
		StatusText:      playbookStatusText,
		StepResults:     stepResults,
		RequestInterval: defaultRequestInterval,
	}
	return runReport, nil
}

func parseRunStateSteps(stepEntries map[string]runstate_model.StepResult) (map[string]api_model.StepRunReport, error) {
	parsedEntries := map[string]api_model.StepRunReport{}
	for stepRunKey, stepEntry := range stepEntries {

		stepStatus := api_model.RunStatusEnum2String(stepEntry.Status)

		stepStatusText, err := api_model.GetRunStatusText(stepStatus, api_model.ReportLevelStep)
		if err != nil {
			return map[string]api_model.StepRunReport{}, err
		}

		if stepEntry.Error != nil {
			stepStatusText = stepStatusText + " - error: " + stepEntry.Error.Error()
		}

		parsedEntries[stepRunKey] = api_model.StepRunReport{
			RunId:        stepEntry.RunId.String(),
			StepId:       stepEntry.StepId,
			StepRunId:    stepEntry.StepRunId.String(),
			Name:         stepEntry.Name,
			Description:  stepEntry.Description,
			Started:      stepEntry.Started,
			Ended:        stepEntry.Ended,
			Status:       stepStatus,
			StatusText:   stepStatusText,
			ExecutedBy:   "soarca",
			CommandsB64:  stepEntry.CommandsB64,
			Variables:    stepEntry.Variables,
			AutomatedRun: stepEntry.IsAutomated,
		}
	}
	return parsedEntries, nil
}
