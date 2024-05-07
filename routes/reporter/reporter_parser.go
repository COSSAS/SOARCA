package reporter

import (
	api_model "soarca/models/api"
	cache_model "soarca/models/report"
)

func parseCachePlaybookEntry(cacheEntry cache_model.ExecutionEntry) (api_model.PlaybookExecutionReport, error) {
	playbookStatus, err := api_model.CacheStatusEnum2String(cacheEntry.Status)
	if err != nil {
		return api_model.PlaybookExecutionReport{}, err
	}

	stepResults, err := parseCacheStepEntries(cacheEntry.StepResults)
	if err != nil {
		return api_model.PlaybookExecutionReport{}, err
	}

	executionReport := api_model.PlaybookExecutionReport{
		Type:            "execution_status",
		ExecutionId:     cacheEntry.ExecutionId.String(),
		PlaybookId:      cacheEntry.PlaybookId,
		Started:         cacheEntry.Started.String(),
		Ended:           cacheEntry.Ended.String(),
		Status:          playbookStatus,
		StatusText:      cacheEntry.PlaybookResult.Error(),
		Error:           cacheEntry.PlaybookResult.Error(),
		StepResults:     stepResults,
		RequestInterval: 5,
	}
	return executionReport, nil
}

func parseCacheStepEntries(cacheStepEntries map[string]cache_model.StepResult) (map[string]api_model.StepExecutionReport, error) {
	parsedEntries := map[string]api_model.StepExecutionReport{}
	for stepId, stepEntry := range cacheStepEntries {

		stepStatus, err := api_model.CacheStatusEnum2String(stepEntry.Status)
		if err != nil {
			return map[string]api_model.StepExecutionReport{}, err
		}

		parsedEntries[stepId] = api_model.StepExecutionReport{
			ExecutionId: stepEntry.ExecutionId.String(),
			StepId:      stepEntry.StepId,
			Started:     stepEntry.Started.String(),
			Ended:       stepEntry.Ended.String(),
			Status:      stepStatus,
			StatusText:  stepEntry.Error.Error(),
			Error:       stepEntry.Error.Error(),
			Variables:   stepEntry.Variables,
		}
	}
	return parsedEntries, nil
}
