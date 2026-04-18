package handler

import "testing"

func TestExtractRequestTranscriptText_OpenAIResponses(t *testing.T) {
	body := []byte(`{
		"model":"gpt-5.4",
		"input":[
			{"role":"user","content":[{"type":"input_text","text":"前一条"}]},
			{"role":"user","content":[{"type":"input_text","text":"今天的天气怎么样"}]}
		]
	}`)

	got := extractRequestTranscriptText(body)
	if got != "今天的天气怎么样" {
		t.Fatalf("extractRequestTranscriptText() = %q, want %q", got, "今天的天气怎么样")
	}
}

func TestExtractResponseTranscriptText_OpenAIJSON(t *testing.T) {
	body := []byte(`{
		"id":"resp_1",
		"output":[
			{
				"type":"message",
				"role":"assistant",
				"content":[
					{"type":"output_text","text":"旧回复"}
				]
			},
			{
				"type":"message",
				"role":"assistant",
				"content":[
					{"type":"output_text","text":"很不错"}
				]
			}
		]
	}`)

	got := extractResponseTranscriptText(body, "application/json")
	if got != "很不错" {
		t.Fatalf("extractResponseTranscriptText(JSON) = %q, want %q", got, "很不错")
	}
}

func TestExtractResponseTranscriptText_OpenAISSE(t *testing.T) {
	body := []byte("data: {\"type\":\"response.output_text.delta\",\"delta\":\"很\"}\n\n" +
		"data: {\"type\":\"response.output_text.delta\",\"delta\":\"不错\"}\n\n" +
		"data: [DONE]\n\n")

	got := extractResponseTranscriptText(body, "text/event-stream")
	if got != "很不错" {
		t.Fatalf("extractResponseTranscriptText(SSE) = %q, want %q", got, "很不错")
	}
}

func TestExtractResponseTranscriptText_OpenAIResponsesCompletedEvent(t *testing.T) {
	body := []byte("event: response.created\n" +
		"data: {\"type\":\"response.created\",\"response\":{\"id\":\"resp_1\",\"status\":\"in_progress\"}}\n\n" +
		"event: response.completed\n" +
		"data: {\"type\":\"response.completed\",\"response\":{\"id\":\"resp_1\",\"output\":[{\"type\":\"message\",\"role\":\"assistant\",\"content\":[{\"type\":\"output_text\",\"text\":\"最终答案\"}]}]}}\n\n")

	got := extractResponseTranscriptText(body, "text/event-stream")
	if got != "最终答案" {
		t.Fatalf("extractResponseTranscriptText(completed SSE) = %q, want %q", got, "最终答案")
	}
}
