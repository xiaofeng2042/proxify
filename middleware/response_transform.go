package middleware

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/poixeai/proxify/infra/ctx"
	"github.com/poixeai/proxify/infra/logger"
)

// ChatCompletionStreamChunk represents a streaming chunk from Chat Completions API
type ChatCompletionStreamChunk struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index int `json:"index"`
		Delta struct {
			Role             string `json:"role,omitempty"`
			Content          string `json:"content,omitempty"`
			ReasoningContent string `json:"reasoning_content,omitempty"` // Zhipu GLM reasoning
		} `json:"delta"`
		FinishReason *string `json:"finish_reason"`
	} `json:"choices"`
}

// ResponseTransform transforms Chat Completions responses to Responses API format
func ResponseTransform() gin.HandlerFunc {
	return func(c *gin.Context) {
		route := ctx.GetRoute(c)
		if route == nil || route.Transform != "responses_to_chat" {
			c.Next()
			return
		}

		// Only handle responses endpoint
		if !strings.HasSuffix(c.Request.URL.Path, "/responses") {
			c.Next()
			return
		}

		logger.Infof("ResponseTransform: activating transform for path=%s", c.Request.URL.Path)

		// Wrap the response writer
		proxy := &responseProxy{
			ResponseWriter: c.Writer,
			transform:      true,
		}
		c.Writer = proxy

		c.Next()
	}
}

type responseProxy struct {
	gin.ResponseWriter
	transform        bool
	initialized      bool   // whether we've sent response.created event
	responseID       string
	itemID           string // message item ID for delta events
	sequenceNumber   int    // sequence counter for all events
	contentPartAdded bool   // whether we've added content_part for main output
	accumulatedText  string // accumulated text content for response.output_text.done
	completed        bool   // whether we've already sent response.completed event
}

func (p *responseProxy) Write(data []byte) (int, error) {
	if !p.transform {
		return p.ResponseWriter.Write(data)
	}

	// Check if this is a streaming response
	contentType := p.Header().Get("Content-Type")
	logger.Infof("ResponseTransform.Write: contentType=%s, dataLen=%d, dataPreview=%s", contentType, len(data), string(data)[:min(100, len(data))])

	if !strings.Contains(contentType, "text/event-stream") {
		// Non-streaming response, pass through
		return p.ResponseWriter.Write(data)
	}

	// Transform SSE chunks
	return p.transformSSE(data)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (p *responseProxy) transformSSE(data []byte) (int, error) {
	logger.Infof("transformSSE called with data: %s", string(data))

	// Parse SSE format: "data: {...}\n\n"
	lines := strings.Split(string(data), "\n")
	var totalWritten int

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		logger.Infof("transformSSE processing line: %s", line)

		if !strings.HasPrefix(line, "data: ") {
			// Pass through non-data lines
			n, err := p.ResponseWriter.Write([]byte(line + "\n"))
			totalWritten += n
			if err != nil {
				return totalWritten, err
			}
			continue
		}

		// Extract JSON payload
		jsonStr := strings.TrimPrefix(line, "data: ")
		logger.Infof("transformSSE jsonStr: %s", jsonStr)

		// Handle [DONE] marker - only emit if we haven't already sent response.completed
		if jsonStr == "[DONE]" {
			// If we already sent response.completed via finish_reason, skip this
			if p.completed {
				logger.Infof("transformSSE: skipping [DONE] marker, already sent response.completed")
				continue
			}
			// Convert to Responses API done format
			// SSE format with event type
			doneEvent := "event: response.completed\ndata: {\"type\":\"response.completed\"}\n\n"
			n, err := p.ResponseWriter.Write([]byte(doneEvent))
			totalWritten += n
			if err != nil {
				return totalWritten, err
			}
			continue
		}

		// Parse Chat Completions chunk
		var ccChunk ChatCompletionStreamChunk
		if err := json.Unmarshal([]byte(jsonStr), &ccChunk); err != nil {
			logger.Warnf("ResponseTransform: failed to parse chunk: %v", err)
			// Pass through on error
			n, err := p.ResponseWriter.Write([]byte(line + "\n"))
			totalWritten += n
			if err != nil {
				return totalWritten, err
			}
			continue
		}

		// Convert to Responses API format
		events := p.convertChunkToResponsesEvents(&ccChunk, !p.initialized)
		if !p.initialized && len(events) > 0 {
			p.initialized = true
			p.responseID = ccChunk.ID
		}
		logger.Infof("transformSSE: converted to %d events for chunk id=%s, initialized=%v", len(events), ccChunk.ID, p.initialized)
		if len(events) == 0 {
			continue
		}

		// Write each event
		for _, event := range events {
			eventBytes, err := json.Marshal(event.Data)
			if err != nil {
				logger.Warnf("ResponseTransform: failed to marshal response: %v", err)
				continue
			}

			// SSE format: event line + data line + blank line
			output := fmt.Sprintf("event: %s\ndata: %s\n\n", event.Type, string(eventBytes))
			logger.Infof("transformSSE: writing output: %s", output)
			n, err := p.ResponseWriter.Write([]byte(output))
			totalWritten += n
			if err != nil {
				return totalWritten, err
			}
		}
	}

	return totalWritten, nil
}

func (p *responseProxy) WriteHeader(code int) {
	p.ResponseWriter.WriteHeader(code)
}

func (p *responseProxy) Flush() {
	if flusher, ok := p.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (p *responseProxy) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := p.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, fmt.Errorf("response writer does not support hijacking")
}

// ResponsesAPIEvent represents an SSE event for Responses API
type ResponsesAPIEvent struct {
	Type string
	Data interface{}
}

// nextSequenceNumber increments and returns the sequence number for events
func (p *responseProxy) nextSequenceNumber() int {
	p.sequenceNumber++
	return p.sequenceNumber
}

// convertChunkToResponsesEvents converts a Chat Completions chunk to Responses API events
// isFirstChunk indicates if this is the first chunk (to emit response.created)
func (p *responseProxy) convertChunkToResponsesEvents(ccChunk *ChatCompletionStreamChunk, isFirstChunk bool) []ResponsesAPIEvent {
	var events []ResponsesAPIEvent

	// CRITICAL: Always emit setup events on first chunk, even if no choices!
	// This fixes the "OutputTextDelta without active item" error from Codex client
	if isFirstChunk {
		// Generate message item ID
		p.itemID = "msg_" + ccChunk.ID

		// 1. Emit response.created event
		events = append(events, ResponsesAPIEvent{
			Type: "response.created",
			Data: map[string]interface{}{
				"type":           "response.created",
				"sequence_number": p.nextSequenceNumber(),
				"response": map[string]interface{}{
					"id":         ccChunk.ID,
					"object":     "response",
					"created_at": ccChunk.Created,
					"status":     "in_progress",
				},
			},
		})

		// 2. Emit response.in_progress event
		events = append(events, ResponsesAPIEvent{
			Type: "response.in_progress",
			Data: map[string]interface{}{
				"type":           "response.in_progress",
				"sequence_number": p.nextSequenceNumber(),
				"response": map[string]interface{}{
					"id":         ccChunk.ID,
					"object":     "response",
					"created_at": ccChunk.Created,
					"status":     "in_progress",
				},
			},
		})

		// 3. Emit response.output_item.added
		events = append(events, ResponsesAPIEvent{
			Type: "response.output_item.added",
			Data: map[string]interface{}{
				"type":           "response.output_item.added",
				"sequence_number": p.nextSequenceNumber(),
				"output_index":   0,
				"item": map[string]interface{}{
					"id":      p.itemID,
					"type":    "message",
					"role":    "assistant",
					"status":  "in_progress",
					"content": []interface{}{},
				},
			},
		})

		// 4. Emit response.content_part.added
		events = append(events, ResponsesAPIEvent{
			Type: "response.content_part.added",
			Data: map[string]interface{}{
				"type":           "response.content_part.added",
				"sequence_number": p.nextSequenceNumber(),
				"item_id":        p.itemID,
				"output_index":   0,
				"content_index":  0,
				"part": map[string]interface{}{
					"type":        "output_text",
					"text":        "",
					"annotations": []interface{}{},
				},
			},
		})

		p.contentPartAdded = true
	}

	// If no choices in this chunk, return just the setup events
	if len(ccChunk.Choices) == 0 {
		return events
	}

	choice := ccChunk.Choices[0]

	// Handle content delta - skip reasoning_content, only process actual content
	if choice.Delta.Content != "" {
		// Accumulate text for the done event
		p.accumulatedText += choice.Delta.Content

		events = append(events, ResponsesAPIEvent{
			Type: "response.output_text.delta",
			Data: map[string]interface{}{
				"type":           "response.output_text.delta",
				"sequence_number": p.nextSequenceNumber(),
				"item_id":        p.itemID,
				"output_index":   0,
				"content_index":  0,
				"delta":          choice.Delta.Content,
			},
		})
	}

	// Handle finish reason
	if choice.FinishReason != nil {
		// Emit response.content_part.done with full text
		events = append(events, ResponsesAPIEvent{
			Type: "response.content_part.done",
			Data: map[string]interface{}{
				"type":           "response.content_part.done",
				"sequence_number": p.nextSequenceNumber(),
				"item_id":        p.itemID,
				"output_index":   0,
				"content_index":  0,
				"part": map[string]interface{}{
					"type":        "output_text",
					"text":        p.accumulatedText,
					"annotations": []interface{}{},
				},
			},
		})

		// Emit response.output_text.done with full text
		events = append(events, ResponsesAPIEvent{
			Type: "response.output_text.done",
			Data: map[string]interface{}{
				"type":           "response.output_text.done",
				"sequence_number": p.nextSequenceNumber(),
				"item_id":        p.itemID,
				"output_index":   0,
				"content_index":  0,
				"text":           p.accumulatedText,
			},
		})

		// Emit response.output_item.done with full content
		events = append(events, ResponsesAPIEvent{
			Type: "response.output_item.done",
			Data: map[string]interface{}{
				"type":           "response.output_item.done",
				"sequence_number": p.nextSequenceNumber(),
				"output_index":   0,
				"item": map[string]interface{}{
					"id":     p.itemID,
					"type":   "message",
					"role":   "assistant",
					"status": "completed",
					"content": []interface{}{
						map[string]interface{}{
							"type":        "output_text",
							"text":        p.accumulatedText,
							"annotations": []interface{}{},
						},
					},
				},
			},
		})

		// Finally emit response.completed
		events = append(events, ResponsesAPIEvent{
			Type: "response.completed",
			Data: map[string]interface{}{
				"type":           "response.completed",
				"sequence_number": p.nextSequenceNumber(),
				"response": map[string]interface{}{
					"id":         ccChunk.ID,
					"object":     "response",
					"created_at": ccChunk.Created,
					"status":     "completed",
				},
			},
		})

		// Mark as completed so we don't send duplicate response.completed on [DONE]
		p.completed = true
	}

	return events
}
