package reporter

import (
	api_model "soarca/pkg/models/api"
	cache_model "soarca/pkg/models/cache"
)

const defaultRequestInterval int = 5

func parseCachePlaybookEntry(cacheEntry cache_model.ExecutionEntry) (api_model.PlaybookRunReport, error) {
	playbookStatus := api_model.CacheStatusEnum2String(cacheEntry.Status)

	playbookStatusText, err := api_model.GetCacheStatusText(playbookStatus, api_model.ReportLevelPlaybook)
	if err != nil {
		return api_model.PlaybookRunReport{}, err
	}
	if cacheEntry.Error != nil {
		playbookStatusText = playbookStatusText + " - error: " + cacheEntry.Error.Error()
	}

	stepResults, err := parseCacheStepEntries(cacheEntry.StepResults)
	if err != nil {
		return api_model.PlaybookRunReport{}, err
	}

	runReport := api_model.PlaybookRunReport{
		Type:            "run_status",
		Name:            cacheEntry.Name,
		Description:     cacheEntry.Description,
		RunId:           cacheEntry.ExecutionId.String(),
		PlaybookId:      cacheEntry.PlaybookId,
		Started:         cacheEntry.Started,
		Ended:           cacheEntry.Ended,
		Status:          playbookStatus,
		StatusText:      playbookStatusText,
		StepResults:     stepResults,
		RequestInterval: defaultRequestInterval,
	}
	return runReport, nil
}

func parseCacheStepEntries(cacheStepEntries map[string]cache_model.StepResult) (map[string]api_model.StepRunReport, error) {
	parsedEntries := map[string]api_model.StepRunReport{}
	for stepRunKey, stepEntry := range cacheStepEntries {

		stepStatus := api_model.CacheStatusEnum2String(stepEntry.Status)

		stepStatusText, err := api_model.GetCacheStatusText(stepStatus, api_model.ReportLevelStep)
		if err != nil {
			return map[string]api_model.StepRunReport{}, err
		}

		if stepEntry.Error != nil {
			stepStatusText = stepStatusText + " - error: " + stepEntry.Error.Error()
		}

		parsedEntries[stepRunKey] = api_model.StepRunReport{
			RunId:              stepEntry.ExecutionId.String(),
			StepId:             stepEntry.StepId,
			StepRunId:          stepEntry.StepExecutionId.String(),
			Name:               stepEntry.Name,
			Description:        stepEntry.Description,
			Started:            stepEntry.Started,
			Ended:              stepEntry.Ended,
			Status:             stepStatus,
			StatusText:         stepStatusText,
			ExecutedBy:         "soarca",
			CommandsB64:        stepEntry.CommandsB64,
			Variables:          stepEntry.Variables,
			AutomatedExecution: stepEntry.IsAutomated,
		}
	}
	return parsedEntries, nil
}
