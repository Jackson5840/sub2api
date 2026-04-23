package service

import (
	"encoding/json"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/apicompat"
	"github.com/stretchr/testify/require"
)

func TestApplyOpenAIReasoningEffortOverrideToMap(t *testing.T) {
	account := &Account{
		Credentials: map[string]any{
			"reasoning_effort_override": "low",
		},
	}
	reqBody := map[string]any{
		"model":             "gpt-5.4",
		"reasoning_effort":  "high",
		"reasoning":         map[string]any{"summary": "auto"},
		"service_tier":      "priority",
		"max_output_tokens": 256,
	}

	effort, changed := applyOpenAIReasoningEffortOverrideToMap(reqBody, account)
	require.True(t, changed)
	require.Equal(t, "low", effort)

	reasoning, ok := reqBody["reasoning"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "low", reasoning["effort"])
	require.Equal(t, "auto", reasoning["summary"])
	_, hasFlat := reqBody["reasoning_effort"]
	require.False(t, hasFlat)
}

func TestApplyOpenAIReasoningEffortOverrideToBody(t *testing.T) {
	account := &Account{
		Credentials: map[string]any{
			"reasoning_effort_override": "medium",
		},
	}

	body, effort, changed, err := applyOpenAIReasoningEffortOverrideToBody([]byte(`{"model":"gpt-5.4","reasoning_effort":"xhigh"}`), account)
	require.NoError(t, err)
	require.True(t, changed)
	require.NotNil(t, effort)
	require.Equal(t, "medium", *effort)

	var reqBody map[string]any
	require.NoError(t, json.Unmarshal(body, &reqBody))
	reasoning, ok := reqBody["reasoning"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "medium", reasoning["effort"])
	_, hasFlat := reqBody["reasoning_effort"]
	require.False(t, hasFlat)
}

func TestApplyOpenAIReasoningEffortOverrideToResponsesRequest(t *testing.T) {
	account := &Account{
		Credentials: map[string]any{
			"reasoning_effort_override": "high",
		},
	}
	req := &apicompat.ResponsesRequest{
		Model: "claude-sonnet-4.5",
		Reasoning: &apicompat.ResponsesReasoning{
			Effort:  "low",
			Summary: "auto",
		},
	}

	effort := applyOpenAIReasoningEffortOverrideToResponsesRequest(req, account)
	require.NotNil(t, effort)
	require.Equal(t, "high", *effort)
	require.NotNil(t, req.Reasoning)
	require.Equal(t, "high", req.Reasoning.Effort)
	require.Equal(t, "auto", req.Reasoning.Summary)
}

func TestApplyOpenAIReasoningEffortOverrideToChatCompletionsRequest(t *testing.T) {
	account := &Account{
		Credentials: map[string]any{
			"reasoning_effort_override": "xhigh",
		},
	}
	req := &apicompat.ChatCompletionsRequest{
		Model:           "gpt-5.4",
		ReasoningEffort: "low",
	}

	effort := applyOpenAIReasoningEffortOverrideToChatCompletionsRequest(req, account)
	require.NotNil(t, effort)
	require.Equal(t, "xhigh", *effort)
	require.Equal(t, "xhigh", req.ReasoningEffort)
}
