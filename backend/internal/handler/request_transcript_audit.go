package handler

import (
	"bytes"
	"context"
	"strings"
	"sync"

	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
	"go.uber.org/zap"
)

const (
	requestTranscriptAuditComponent = "audit.request_transcript"
	requestTranscriptCaptureLimit   = 128 * 1024
	requestTranscriptMaxChars       = 4000
)

type requestTranscriptCaptureWriter struct {
	gin.ResponseWriter
	limit     int
	truncated bool
	buf       bytes.Buffer
}

var requestTranscriptCaptureWriterPool = sync.Pool{
	New: func() any {
		return &requestTranscriptCaptureWriter{limit: requestTranscriptCaptureLimit}
	},
}

func acquireRequestTranscriptCaptureWriter(rw gin.ResponseWriter) *requestTranscriptCaptureWriter {
	w, ok := requestTranscriptCaptureWriterPool.Get().(*requestTranscriptCaptureWriter)
	if !ok || w == nil {
		w = &requestTranscriptCaptureWriter{}
	}
	w.ResponseWriter = rw
	w.limit = requestTranscriptCaptureLimit
	w.truncated = false
	w.buf.Reset()
	return w
}

func releaseRequestTranscriptCaptureWriter(w *requestTranscriptCaptureWriter) {
	if w == nil {
		return
	}
	w.ResponseWriter = nil
	w.limit = requestTranscriptCaptureLimit
	w.truncated = false
	w.buf.Reset()
	requestTranscriptCaptureWriterPool.Put(w)
}

func (w *requestTranscriptCaptureWriter) captureBytes(b []byte) {
	if w == nil || w.limit <= 0 || len(b) == 0 || w.buf.Len() >= w.limit {
		if w != nil && w.buf.Len() >= w.limit {
			w.truncated = true
		}
		return
	}
	remaining := w.limit - w.buf.Len()
	if len(b) > remaining {
		_, _ = w.buf.Write(b[:remaining])
		w.truncated = true
		return
	}
	_, _ = w.buf.Write(b)
}

func (w *requestTranscriptCaptureWriter) Write(b []byte) (int, error) {
	w.captureBytes(b)
	return w.ResponseWriter.Write(b)
}

func (w *requestTranscriptCaptureWriter) WriteString(s string) (int, error) {
	w.captureBytes([]byte(s))
	return w.ResponseWriter.WriteString(s)
}

// OpsRequestTranscriptAuditMiddleware captures successful gateway request/response
// text excerpts and writes them as audit events into ops_system_logs via the
// normal logger sink.
func OpsRequestTranscriptAuditMiddleware(ops *service.OpsService) gin.HandlerFunc {
	return func(c *gin.Context) {
		originalWriter := c.Writer
		w := acquireRequestTranscriptCaptureWriter(originalWriter)
		defer func() {
			if c.Writer == w {
				c.Writer = originalWriter
			}
			releaseRequestTranscriptCaptureWriter(w)
		}()
		c.Writer = w
		c.Next()

		if ops == nil || !ops.IsMonitoringEnabled(c.Request.Context()) {
			return
		}
		status := c.Writer.Status()
		if status < 200 || status >= 300 {
			return
		}
		if isCountTokensRequest(c) {
			return
		}

		requestBody := getTranscriptRequestBody(c)
		if len(requestBody) == 0 {
			return
		}

		requestText := extractRequestTranscriptText(requestBody)
		responseText := extractResponseTranscriptText(w.buf.Bytes(), c.Writer.Header().Get("Content-Type"))
		if requestText == "" && responseText == "" {
			return
		}

		requestText, requestTextTruncated := truncateTranscriptText(requestText, requestTranscriptMaxChars)
		responseText, responseTextTruncated := truncateTranscriptText(responseText, requestTranscriptMaxChars)

		var (
			requestID       string
			clientRequestID string
			userID          *int64
			accountID       *int64
			apiKeyID        *int64
			groupID         *int64
			platform        string
			model           string
			upstreamModel   string
			stream          bool
		)

		if c != nil && c.Request != nil {
			requestID, _ = c.Request.Context().Value(ctxkey.RequestID).(string)
			clientRequestID, _ = c.Request.Context().Value(ctxkey.ClientRequestID).(string)
			platform, _ = c.Request.Context().Value(ctxkey.Platform).(string)
		}
		if requestID == "" && c != nil && c.Writer != nil {
			requestID = strings.TrimSpace(c.Writer.Header().Get("X-Request-Id"))
		}
		if v, ok := c.Get(opsModelKey); ok {
			model, _ = v.(string)
		}
		if v, ok := c.Get(opsUpstreamModelKey); ok {
			upstreamModel, _ = v.(string)
		}
		if v, ok := c.Get(opsStreamKey); ok {
			stream, _ = v.(bool)
		}
		if v, ok := c.Get(opsAccountIDKey); ok {
			switch t := v.(type) {
			case int64:
				if t > 0 {
					accountID = &t
				}
			case int:
				if t > 0 {
					id := int64(t)
					accountID = &id
				}
			}
		}

		apiKey, _ := middleware2.GetAPIKeyFromContext(c)
		if apiKey != nil {
			apiKeyID = &apiKey.ID
			groupID = apiKey.GroupID
			if apiKey.User != nil {
				userID = &apiKey.User.ID
			}
			if platform == "" && apiKey.Group != nil {
				platform = apiKey.Group.Platform
			}
		}

		fields := []zap.Field{
			zap.String("component", requestTranscriptAuditComponent),
			zap.String("request_id", strings.TrimSpace(requestID)),
			zap.String("client_request_id", strings.TrimSpace(clientRequestID)),
			zap.String("platform", strings.TrimSpace(platform)),
			zap.String("model", strings.TrimSpace(model)),
			zap.String("request_text", requestText),
			zap.String("response_text", responseText),
			zap.Bool("request_text_truncated", requestTextTruncated),
			zap.Bool("response_text_truncated", responseTextTruncated || w.truncated),
			zap.Bool("stream", stream),
			zap.String("request_path", transcriptRequestPath(c)),
			zap.String("inbound_endpoint", GetInboundEndpoint(c)),
			zap.String("upstream_endpoint", GetUpstreamEndpoint(c, platform)),
			zap.String("upstream_model", strings.TrimSpace(upstreamModel)),
			zap.Int("request_body_bytes", len(requestBody)),
			zap.Int("response_body_bytes", w.buf.Len()),
		}
		if userID != nil {
			fields = append(fields, zap.Int64("user_id", *userID))
		}
		if accountID != nil {
			fields = append(fields, zap.Int64("account_id", *accountID))
		}
		if apiKeyID != nil {
			fields = append(fields, zap.Int64("api_key_id", *apiKeyID))
		}
		if groupID != nil {
			fields = append(fields, zap.Int64("group_id", *groupID))
		}

		log := logger.FromContext(context.Background())
		if c != nil && c.Request != nil {
			log = logger.FromContext(c.Request.Context())
		}
		log.With(fields...).Info("request transcript captured")
	}
}

func getTranscriptRequestBody(c *gin.Context) []byte {
	if c == nil {
		return nil
	}
	v, ok := c.Get(opsRequestBodyKey)
	if !ok {
		return nil
	}
	body, ok := v.([]byte)
	if !ok || len(body) == 0 {
		return nil
	}
	return body
}

func transcriptRequestPath(c *gin.Context) string {
	if c == nil || c.Request == nil || c.Request.URL == nil {
		return ""
	}
	return strings.TrimSpace(c.Request.URL.Path)
}

func truncateTranscriptText(text string, maxChars int) (string, bool) {
	text = strings.TrimSpace(text)
	if text == "" || maxChars <= 0 {
		return text, false
	}
	runes := []rune(text)
	if len(runes) <= maxChars {
		return text, false
	}
	return strings.TrimSpace(string(runes[:maxChars])) + " ...", true
}

func extractRequestTranscriptText(body []byte) string {
	body = bytes.TrimSpace(body)
	if len(body) == 0 {
		return ""
	}
	if !gjson.ValidBytes(body) {
		return strings.TrimSpace(string(body))
	}

	if input := gjson.GetBytes(body, "input"); input.Exists() {
		if text := extractResponsesInputRequestText(input); text != "" {
			return text
		}
	}
	if messages := gjson.GetBytes(body, "messages"); messages.Exists() {
		if text := extractMessagesRequestText(messages); text != "" {
			return text
		}
	}
	if contents := gjson.GetBytes(body, "contents"); contents.Exists() {
		if text := extractGeminiRequestText(contents); text != "" {
			return text
		}
	}

	for _, path := range []string{"prompt", "instructions", "system"} {
		if text := extractTextFromStructuredNode(gjson.GetBytes(body, path)); text != "" {
			return text
		}
	}
	return ""
}

func extractResponsesInputRequestText(input gjson.Result) string {
	if !input.Exists() {
		return ""
	}
	if input.Type == gjson.String {
		return strings.TrimSpace(input.String())
	}

	last := ""
	for _, item := range input.Array() {
		role := strings.ToLower(strings.TrimSpace(item.Get("role").String()))
		typ := strings.ToLower(strings.TrimSpace(item.Get("type").String()))
		if role == "assistant" {
			continue
		}
		switch {
		case role == "user":
			if text := extractTextFromStructuredNode(item.Get("content")); text != "" {
				last = text
			}
		case typ == "input_text" || typ == "text":
			if text := extractTextFromStructuredNode(item); text != "" {
				last = text
			}
		case typ == "message":
			if text := extractTextFromStructuredNode(item.Get("content")); text != "" {
				last = text
			}
		}
	}
	return strings.TrimSpace(last)
}

func extractMessagesRequestText(messages gjson.Result) string {
	if !messages.Exists() {
		return ""
	}
	last := ""
	for _, msg := range messages.Array() {
		if strings.ToLower(strings.TrimSpace(msg.Get("role").String())) != "user" {
			continue
		}
		if text := extractTextFromStructuredNode(msg.Get("content")); text != "" {
			last = text
		}
	}
	return strings.TrimSpace(last)
}

func extractGeminiRequestText(contents gjson.Result) string {
	if !contents.Exists() {
		return ""
	}
	last := ""
	for _, content := range contents.Array() {
		role := strings.ToLower(strings.TrimSpace(content.Get("role").String()))
		if role != "" && role != "user" {
			continue
		}
		if text := extractTextFromStructuredNode(content.Get("parts")); text != "" {
			last = text
		}
	}
	return strings.TrimSpace(last)
}

func extractResponseTranscriptText(body []byte, contentType string) string {
	body = bytes.TrimSpace(body)
	if len(body) == 0 {
		return ""
	}
	lowerContentType := strings.ToLower(strings.TrimSpace(contentType))
	if strings.Contains(lowerContentType, "image/") ||
		strings.Contains(lowerContentType, "audio/") ||
		strings.Contains(lowerContentType, "application/octet-stream") {
		return ""
	}
	if strings.Contains(lowerContentType, "text/event-stream") || bytes.Contains(body, []byte("data:")) {
		if text := extractResponseTranscriptFromSSE(body); text != "" {
			return text
		}
	}
	if !gjson.ValidBytes(body) {
		return strings.TrimSpace(string(body))
	}
	return extractResponseTranscriptFromJSON(body)
}

func extractResponseTranscriptFromSSE(body []byte) string {
	lines := strings.Split(string(body), "\n")
	var builder strings.Builder
	lastComplete := ""

	appendChunk := func(chunk string) {
		if chunk == "" {
			return
		}
		_, _ = builder.WriteString(chunk)
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if payload == "" || payload == "[DONE]" || !gjson.Valid(payload) {
			continue
		}

		data := []byte(payload)
		eventType := strings.TrimSpace(gjson.GetBytes(data, "type").String())
		switch eventType {
		case "response.output_text.delta":
			appendChunk(gjson.GetBytes(data, "delta").String())
			continue
		case "response.output_item.added", "response.output_item.done":
			if item := gjson.GetBytes(data, "item"); item.Exists() && item.Raw != "" {
				if text := extractResponseTranscriptFromJSON([]byte(item.Raw)); text != "" {
					lastComplete = text
				}
			}
			continue
		case "response.output_text.done":
			if text := strings.TrimSpace(gjson.GetBytes(data, "text").String()); text != "" {
				lastComplete = text
			}
			continue
		case "content_block_start":
			if strings.TrimSpace(gjson.GetBytes(data, "content_block.type").String()) == "text" {
				appendChunk(gjson.GetBytes(data, "content_block.text").String())
			}
			continue
		case "content_block_delta":
			if strings.TrimSpace(gjson.GetBytes(data, "delta.type").String()) == "text_delta" {
				appendChunk(gjson.GetBytes(data, "delta.text").String())
			}
			continue
		case "response.completed", "response.done", "response.incomplete", "response.failed":
			response := gjson.GetBytes(data, "response")
			if response.Exists() && response.Raw != "" {
				if text := extractResponseTranscriptFromJSON([]byte(response.Raw)); text != "" {
					lastComplete = text
				}
			}
			continue
		}

		if delta := extractChatCompletionsDeltaText(data); delta != "" {
			appendChunk(delta)
			continue
		}
		if text := extractResponseTranscriptFromJSON(data); text != "" {
			lastComplete = text
		}
	}

	if text := strings.TrimSpace(builder.String()); text != "" {
		return text
	}
	return strings.TrimSpace(lastComplete)
}

func extractChatCompletionsDeltaText(data []byte) string {
	choices := gjson.GetBytes(data, "choices")
	if !choices.Exists() {
		return ""
	}
	out := make([]string, 0, 2)
	for _, choice := range choices.Array() {
		if text := choice.Get("delta.content").String(); text != "" {
			out = append(out, text)
		}
		if text := extractTextFromStructuredNode(choice.Get("message.content")); text != "" {
			out = append(out, text)
		}
	}
	return strings.TrimSpace(strings.Join(out, ""))
}

func extractResponseTranscriptFromJSON(body []byte) string {
	for _, path := range []string{"response", "message"} {
		if nested := gjson.GetBytes(body, path); nested.Exists() && nested.Raw != "" {
			if text := extractResponseTranscriptFromJSON([]byte(nested.Raw)); text != "" {
				return text
			}
		}
	}
	if output := gjson.GetBytes(body, "output"); output.Exists() {
		if text := extractResponsesOutputText(output); text != "" {
			return text
		}
	}
	if choices := gjson.GetBytes(body, "choices"); choices.Exists() {
		if text := extractChatChoicesText(choices); text != "" {
			return text
		}
	}
	if content := gjson.GetBytes(body, "content"); content.Exists() {
		if text := extractTextFromStructuredNode(content); text != "" {
			return text
		}
	}
	if candidates := gjson.GetBytes(body, "candidates"); candidates.Exists() {
		if text := extractGeminiCandidatesText(candidates); text != "" {
			return text
		}
	}
	return ""
}

func extractResponsesOutputText(output gjson.Result) string {
	if !output.Exists() {
		return ""
	}
	last := ""
	for _, item := range output.Array() {
		typ := strings.ToLower(strings.TrimSpace(item.Get("type").String()))
		role := strings.ToLower(strings.TrimSpace(item.Get("role").String()))
		if typ != "message" {
			continue
		}
		if role != "" && role != "assistant" {
			continue
		}
		if text := extractTextFromStructuredNode(item.Get("content")); text != "" {
			last = text
		}
	}
	return strings.TrimSpace(last)
}

func extractChatChoicesText(choices gjson.Result) string {
	if !choices.Exists() {
		return ""
	}
	last := ""
	for _, choice := range choices.Array() {
		if text := extractTextFromStructuredNode(choice.Get("message.content")); text != "" {
			last = text
			continue
		}
		if text := choice.Get("delta.content").String(); text != "" {
			last = text
		}
	}
	return strings.TrimSpace(last)
}

func extractGeminiCandidatesText(candidates gjson.Result) string {
	if !candidates.Exists() {
		return ""
	}
	last := ""
	for _, candidate := range candidates.Array() {
		if text := extractTextFromStructuredNode(candidate.Get("content.parts")); text != "" {
			last = text
		}
	}
	return strings.TrimSpace(last)
}

func extractTextFromStructuredNode(node gjson.Result) string {
	if !node.Exists() {
		return ""
	}

	switch {
	case node.Type == gjson.String:
		return strings.TrimSpace(node.String())
	case node.IsArray():
		out := make([]string, 0, len(node.Array()))
		for _, item := range node.Array() {
			if text := extractTextFromStructuredNode(item); text != "" {
				out = append(out, text)
			}
		}
		return strings.TrimSpace(strings.Join(out, "\n"))
	}

	if text := strings.TrimSpace(node.Get("text").String()); text != "" {
		return text
	}
	if text := strings.TrimSpace(node.Get("delta.text").String()); text != "" {
		return text
	}
	if text := strings.TrimSpace(node.Get("delta").String()); text != "" && !node.Get("role").Exists() {
		return text
	}
	for _, path := range []string{"content", "parts"} {
		if child := node.Get(path); child.Exists() {
			if text := extractTextFromStructuredNode(child); text != "" {
				return text
			}
		}
	}
	return ""
}
