package service

import (
	"encoding/json"

	"github.com/Wei-Shaw/sub2api/internal/pkg/apicompat"
)

func getOpenAIReasoningEffortOverride(account *Account) string {
	if account == nil {
		return ""
	}
	return normalizeOpenAIReasoningEffort(account.GetCredential("reasoning_effort_override"))
}

func applyOpenAIReasoningEffortOverrideToMap(reqBody map[string]any, account *Account) (string, bool) {
	override := getOpenAIReasoningEffortOverride(account)
	if override == "" || reqBody == nil {
		return "", false
	}

	reasoning, _ := reqBody["reasoning"].(map[string]any)
	if reasoning == nil {
		reasoning = map[string]any{}
		reqBody["reasoning"] = reasoning
	}
	reasoning["effort"] = override
	delete(reqBody, "reasoning_effort")
	return override, true
}

func applyOpenAIReasoningEffortOverrideToBody(body []byte, account *Account) ([]byte, *string, bool, error) {
	override := getOpenAIReasoningEffortOverride(account)
	if override == "" || len(body) == 0 {
		return body, nil, false, nil
	}

	var reqBody map[string]any
	if err := json.Unmarshal(body, &reqBody); err != nil {
		return nil, nil, false, err
	}
	if _, changed := applyOpenAIReasoningEffortOverrideToMap(reqBody, account); !changed {
		return body, nil, false, nil
	}

	updated, err := json.Marshal(reqBody)
	if err != nil {
		return nil, nil, false, err
	}
	return updated, &override, true, nil
}

func applyOpenAIReasoningEffortOverrideToResponsesRequest(req *apicompat.ResponsesRequest, account *Account) *string {
	override := getOpenAIReasoningEffortOverride(account)
	if override == "" || req == nil {
		return nil
	}
	if req.Reasoning == nil {
		req.Reasoning = &apicompat.ResponsesReasoning{}
	}
	req.Reasoning.Effort = override
	return &override
}

func applyOpenAIReasoningEffortOverrideToChatCompletionsRequest(req *apicompat.ChatCompletionsRequest, account *Account) *string {
	override := getOpenAIReasoningEffortOverride(account)
	if override == "" || req == nil {
		return nil
	}
	req.ReasoningEffort = override
	return &override
}
