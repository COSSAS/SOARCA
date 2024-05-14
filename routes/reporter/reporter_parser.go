package reporter

import (
	api_model "soarca/models/api"
	cache_model "soarca/models/cache"
)

const defaultRequestInterval int = 5

func parseCachePlaybookEntry(cacheEntry cache_model.ExecutionEntry) (api_model.PlaybookExecutionReport, error) {
	playbookStatus, err := api_model.CacheStatusEnum2String(cacheEntry.Status)
	if err != nil {
		return api_model.PlaybookExecutionReport{}, err
	}
	playbookStatusText, err := api_model.GetCacheStatusText(playbookStatus, api_model.ReportLevelPlaybook)
	playbookErrorStr := ""
	if cacheEntry.PlaybookResult != nil {
		playbookErrorStr = cacheEntry.PlaybookResult.Error()
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
		StatusText:      playbookStatusText,
		Error:           playbookErrorStr,
		StepResults:     stepResults,
		RequestInterval: defaultRequestInterval,
	}
	return executionReport, nil
}

func parseCacheStepEntries(cacheStepEntries map[string]cache_model.StepResult) (map[string]api_model.StepExecutionReport, error) {
	parsedEntries := map[string]api_model.StepExecutionReport{}
	for stepId, stepEntry := range cacheStepEntries {

		stepStatus, err := api_model.CacheStatusEnum2String(stepEntry.Status)
		stepStatusText, err := api_model.GetCacheStatusText(stepStatus, api_model.ReportLevelStep)
		if err != nil {
			return map[string]api_model.StepExecutionReport{}, err
		}

		stepError := stepEntry.Error
		stepErrorStr := ""
		if stepError != nil {
			stepErrorStr = stepError.Error()
		}

		automatedExecution := "true"
		if !stepEntry.IsAutomated {
			automatedExecution = "false"
		}

		parsedEntries[stepId] = api_model.StepExecutionReport{
			ExecutionId:        stepEntry.ExecutionId.String(),
			StepId:             stepEntry.StepId,
			Started:            stepEntry.Started.String(),
			Ended:              stepEntry.Ended.String(),
			Status:             stepStatus,
			StatusText:         stepStatusText,
			ExecutedBy:         "soarca",
			CommandsB64:        stepEntry.CommandsB64,
			Error:              stepErrorStr,
			Variables:          stepEntry.Variables,
			AutomatedExecution: automatedExecution,
		}
	}
	return parsedEntries, nil
}
