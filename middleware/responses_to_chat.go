package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/poixeai/proxify/infra/ctx"
	"github.com/poixeai/proxify/infra/logger"
)

// ResponsesAPIRequest represents the OpenAI Responses API request format
type ResponsesAPIRequest struct {
	Model             string                 `json:"model"`
	Input             interface{}            `json:"input"` // can be string or array
	Instructions      string                 `json:"instructions,omitempty"`
	Store             bool                   `json:"store,omitempty"`
	PreviousResponseID string                `json:"previous_response_id,omitempty"`
	Tools             []interface{}          `json:"tools,omitempty"`
	ToolChoice        interface{}            `json:"tool_choice,omitempty"`
	MaxOutputTokens   int                    `json:"max_output_tokens,omitempty"`
	Temperature       float64                `json:"temperature,omitempty"`
	TopP              float64                `json:"top_p,omitempty"`
	Truncation        string                 `json:"truncation,omitempty"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
	Stream            bool                   `json:"stream,omitempty"`
}

// ChatCompletionRequest represents the OpenAI Chat Completions API request format
type ChatCompletionRequest struct {
	Model            string                 `json:"model"`
	Messages         []ChatMessage          `json:"messages"`
	MaxTokens        int                    `json:"max_tokens,omitempty"`
	Temperature      float64                `json:"temperature,omitempty"`
	TopP             float64                `json:"top_p,omitempty"`
	Stream           bool                   `json:"stream,omitempty"`
	Tools            []interface{}          `json:"tools,omitempty"`
	ToolChoice       interface{}            `json:"tool_choice,omitempty"`
}

// ChatMessage represents a single message in chat completions
type ChatMessage struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"`
}

// ResponsesToChatMiddleware converts Responses API requests to Chat Completions API format
// and converts Chat Completions responses back to Responses API format
func ResponsesToChat() gin.HandlerFunc {
	return func(c *gin.Context) {
		route := ctx.GetRoute(c)
		if route == nil || route.Transform != "responses_to_chat" {
			c.Next()
			return
		}

		logger.Infof("ResponsesToChat: route=%s, path=%s, method=%s", route.Name, c.Request.URL.Path, c.Request.Method)

		// Only handle POST requests to /responses endpoint
		if c.Request.Method != "POST" || !strings.HasSuffix(c.Request.URL.Path, "/responses") {
			logger.Infof("ResponsesToChat: skipping (not POST /responses)")
			c.Next()
			return
		}

		// Read request body
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			logger.Warnf("ResponsesToChat: failed to read request body: %v", err)
			c.Next()
			return
		}

		logger.Infof("ResponsesToChat: raw request body: %s", string(bodyBytes))

		// Parse Responses API request
		var respReq ResponsesAPIRequest
		if err := json.Unmarshal(bodyBytes, &respReq); err != nil {
			logger.Warnf("ResponsesToChat: failed to parse Responses API request: %v", err)
			c.Next()
			return
		}

		// Convert to Chat Completions format
		chatReq := convertResponsesToChat(&respReq)

		// Serialize the converted request
		newBody, err := json.Marshal(chatReq)
		if err != nil {
			logger.Warnf("ResponsesToChat: failed to marshal Chat Completions request: %v", err)
			c.Next()
			return
		}

		// Update the request path from /responses to /chat/completions
		subPath := c.GetString(ctx.SubPath)
		newSubPath := strings.Replace(subPath, "/responses", "/chat/completions", 1)
		c.Set(ctx.SubPath, newSubPath)

		// Restore body for downstream handlers
		c.Request.Body = io.NopCloser(bytes.NewReader(newBody))
		c.Request.ContentLength = int64(len(newBody))

		logger.Infof("ResponsesToChat: converted request for model=%s", respReq.Model)

		c.Next()
	}
}

// convertResponsesToChat converts Responses API request to Chat Completions format
func convertResponsesToChat(respReq *ResponsesAPIRequest) *ChatCompletionRequest {
	chatReq := &ChatCompletionRequest{
		Model:       respReq.Model,
		Temperature: respReq.Temperature,
		TopP:        respReq.TopP,
		Stream:      respReq.Stream,
		Tools:       convertTools(respReq.Tools), // Convert tools format
		ToolChoice:  respReq.ToolChoice,
	}

	if respReq.MaxOutputTokens > 0 {
		chatReq.MaxTokens = respReq.MaxOutputTokens
	}

	// Build messages array
	var messages []ChatMessage

	// Add system instructions if present
	if respReq.Instructions != "" {
		messages = append(messages, ChatMessage{
			Role:    "system",
			Content: respReq.Instructions,
		})
	}

	// Handle input - can be string or array of messages
	switch v := respReq.Input.(type) {
	case string:
		// Simple string input
		messages = append(messages, ChatMessage{
			Role:    "user",
			Content: v,
		})
	case []interface{}:
		// Array format - each item should be a message object
		for _, item := range v {
			if msg, ok := item.(map[string]interface{}); ok {
				role, _ := msg["role"].(string)
				content := msg["content"]

				// Convert unsupported roles to supported ones
				// "developer" -> "system" (Codex uses developer role)
				if role == "developer" {
					role = "system"
				}

				// Convert content to string if it's an array
				// Codex sends: [{"type": "input_text", "text": "actual content"}]
				// Zhipu expects: "actual content"
				contentStr := convertContentToString(content)

				if role != "" && contentStr != "" {
					messages = append(messages, ChatMessage{
						Role:    role,
						Content: contentStr,
					})
				}
			}
		}
	default:
		// Fallback: treat as user message
		if respReq.Input != nil {
			messages = append(messages, ChatMessage{
				Role:    "user",
				Content: respReq.Input,
			})
		}
	}

	chatReq.Messages = messages

	return chatReq
}

// convertTools converts tools from Responses API format to Chat Completions format
// Responses API: {"type": "function", "name": "xxx", "parameters": {...}}
// Chat Completions: {"type": "function", "function": {"name": "xxx", "parameters": {...}}}
// Also filters out unsupported tool types like web_search
func convertTools(tools []interface{}) []interface{} {
	if tools == nil {
		return nil
	}

	var converted []interface{}
	for _, tool := range tools {
		t, ok := tool.(map[string]interface{})
		if !ok {
			continue // Skip invalid tools
		}

		toolType, _ := t["type"].(string)

		// Skip unsupported tool types (OpenAI built-in tools)
		// Zhipu only supports "function" type
		if toolType != "function" {
			logger.Infof("ResponsesToChat: skipping unsupported tool type: %s", toolType)
			continue
		}

		// Check if already in Chat Completions format (has "function" key)
		if _, hasFunction := t["function"]; hasFunction {
			converted = append(converted, tool)
			continue
		}

		// Convert from Responses API format to Chat Completions format
		// Extract name, description, parameters from top level
		name, _ := t["name"].(string)
		description, _ := t["description"].(string)
		parameters := t["parameters"]
		strict, _ := t["strict"].(bool)

		// Build new function object
		functionObj := map[string]interface{}{
			"name": name,
		}
		if description != "" {
			functionObj["description"] = description
		}
		if parameters != nil {
			functionObj["parameters"] = parameters
		}
		if strict {
			functionObj["strict"] = strict
		}

		// Create new tool in Chat Completions format
		newTool := map[string]interface{}{
			"type":     "function",
			"function": functionObj,
		}

		converted = append(converted, newTool)
	}

	return converted
}

// convertContentToString converts content to string format
// Handles both string content and array content with type/text fields
func convertContentToString(content interface{}) string {
	switch c := content.(type) {
	case string:
		return c
	case []interface{}:
		// Array of content parts - extract text from each
		var texts []string
		for _, part := range c {
			if p, ok := part.(map[string]interface{}); ok {
				// Handle {"type": "input_text", "text": "content"}
				if text, ok := p["text"].(string); ok {
					texts = append(texts, text)
				}
			}
		}
		if len(texts) > 0 {
			return texts[0] // Return first text for simplicity
		}
	}
	return ""
}
