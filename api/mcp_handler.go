package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/yt-mcp-server/service"
	
)

// MCPHandler handles all MCP protocol requests.
type MCPHandler struct {
	youtubeService *service.YouTubeService
}

// NewMCPHandler creates a new MCPHandler.
func NewMCPHandler(youtubeService *service.YouTubeService) *MCPHandler {
	return &MCPHandler{
		youtubeService: youtubeService,
	}
}

// ... (rest of the MCP struct definitions are the same)

type MCPRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

type MCPResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *MCPError   `json:"error,omitempty"`
}

type MCPError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type InitializeResult struct {
	ProtocolVersion string                 `json:"protocolVersion"`
	Capabilities    map[string]interface{} `json:"capabilities"`
	ServerInfo      map[string]string      `json:"serverInfo"`
}

type ToolsListResult struct {
	Tools []Tool `json:"tools"`
}

type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

type ToolsCallParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

type ToolsCallResult struct {
	Content []ContentItem `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}

type ContentItem struct {
	Type string      `json:"type"`
	Text string      `json:"text,omitempty"`
	JSON interface{} `json:"json,omitempty"`
}

// HandleMCP handles all MCP protocol requests.
func (h *MCPHandler) HandleMCP(w http.ResponseWriter, r *http.Request) {
	var req MCPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendErrorResponse(w, req.ID, -32700, "Parse error", nil)
		return
	}

	switch req.Method {
	case "initialize":
		h.handleInitialize(w, &req)
	case "tools/list":
		h.handleToolsList(w, &req)
	case "tools/call":
		h.handleToolsCall(w, &req)
	default:
		h.sendErrorResponse(w, req.ID, -32601, "Method not found", nil)
	}
}

func (h *MCPHandler) handleInitialize(w http.ResponseWriter, req *MCPRequest) {
	result := InitializeResult{
		ProtocolVersion: "2025-06-18",
		Capabilities:    map[string]interface{}{"tools": map[string]interface{}{}},
		ServerInfo:      map[string]string{"name": "youtube-toolkit-server", "version": "2.0.0"},
	}
	h.sendSuccessResponse(w, req.ID, result)
}

func (h *MCPHandler) handleToolsList(w http.ResponseWriter, req *MCPRequest) {
	tools := []Tool{
		{
			Name:        "get_video_metadata",
			Description: "Gets detailed information for a specific video.",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{"video_id": map[string]interface{}{"type": "string", "description": "The ID of the YouTube video."}},
				"required":   []string{"video_id"},
			},
		},
		{
			Name:        "search_videos",
			Description: "Searches for YouTube videos.",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{
					"query":      map[string]interface{}{"type": "string", "description": "The search term."},
					"channel_id": map[string]interface{}{"type": "string", "description": "Optional: Restricts search to a specific channel."},
					"limit":      map[string]interface{}{"type": "integer", "description": "Optional: Max number of results (default: 10)."},
				},
				"required":   []string{"query"},
			},
		},
		{
			Name:        "get_video_comments",
			Description: "Fetches top-level comment threads for a video.",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{
					"video_id": map[string]interface{}{"type": "string", "description": "The ID of the YouTube video."},
					"sort_by":  map[string]interface{}{"type": "string", "description": "Optional: Sort order (default: top)."},
					"limit":    map[string]interface{}{"type": "integer", "description": "Optional: Max number of results (default: 20)."},
				},
				"required":   []string{"video_id"},
			},
		},
	}
	result := ToolsListResult{Tools: tools}
	h.sendSuccessResponse(w, req.ID, result)
}

func (h *MCPHandler) handleToolsCall(w http.ResponseWriter, req *MCPRequest) {
	var toolParams ToolsCallParams
	if err := h.decodeParams(req.Params, &toolParams); err != nil {
		h.sendErrorResponse(w, req.ID, -32602, "Invalid params", err.Error())
		return
	}

	switch toolParams.Name {
	case "get_video_metadata":
		h.handleGetVideoMetadata(w, req.ID, &toolParams)
	case "search_videos":
		h.handleSearchVideos(w, req.ID, &toolParams)
	case "get_video_comments":
		h.handleGetVideoComments(w, req.ID, &toolParams)
	default:
		h.sendErrorResponse(w, req.ID, -32601, "Unknown tool", nil)
	}
}

func (h *MCPHandler) handleGetVideoMetadata(w http.ResponseWriter, id interface{}, params *ToolsCallParams) {
	videoID, _ := params.Arguments["video_id"].(string)
	if videoID == "" {
		h.sendToolError(w, id, "video_id is required")
		return
	}

	metadata, err := h.youtubeService.GetVideoMetadata(context.Background(), videoID)
	if err != nil {
		h.sendToolError(w, id, fmt.Sprintf("API Error: %v", err))
		return
	}

	h.sendToolResult(w, id, metadata)
}

func (h *MCPHandler) handleSearchVideos(w http.ResponseWriter, id interface{}, params *ToolsCallParams) {
	query, _ := params.Arguments["query"].(string)
	if query == "" {
		h.sendToolError(w, id, "query is required")
		return
	}
	channelID, _ := params.Arguments["channel_id"].(string)
	limit, ok := params.Arguments["limit"].(float64) // JSON numbers are float64
	if !ok || limit == 0 {
		limit = 10
	}

	results, err := h.youtubeService.SearchVideos(context.Background(), query, channelID, int64(limit))
	if err != nil {
		h.sendToolError(w, id, fmt.Sprintf("API Error: %v", err))
		return
	}

	h.sendToolResult(w, id, results)
}

func (h *MCPHandler) handleGetVideoComments(w http.ResponseWriter, id interface{}, params *ToolsCallParams) {
	videoID, _ := params.Arguments["video_id"].(string)
	if videoID == "" {
		h.sendToolError(w, id, "video_id is required")
		return
	}
	sortBy, _ := params.Arguments["sort_by"].(string)
	if sortBy == "" {
		sortBy = "top"
	}
	limit, ok := params.Arguments["limit"].(float64)
	if !ok || limit == 0 {
		limit = 20
	}

	comments, err := h.youtubeService.GetVideoComments(context.Background(), videoID, sortBy, int64(limit))
	if err != nil {
		h.sendToolError(w, id, fmt.Sprintf("API Error: %v", err))
		return
	}

	h.sendToolResult(w, id, comments)
}

func (h *MCPHandler) decodeParams(source interface{}, dest interface{}) error {
	bytes, err := json.Marshal(source)
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, dest)
}

func (h *MCPHandler) sendSuccessResponse(w http.ResponseWriter, id interface{}, result interface{}) {
	response := MCPResponse{JSONRPC: "2.0", ID: id, Result: result}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *MCPHandler) sendErrorResponse(w http.ResponseWriter, id interface{}, code int, message string, data interface{}) {
	response := MCPResponse{JSONRPC: "2.0", ID: id, Error: &MCPError{Code: code, Message: message, Data: data}}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *MCPHandler) sendToolError(w http.ResponseWriter, id interface{}, message string) {
	result := ToolsCallResult{
		Content: []ContentItem{{Type: "text", Text: message}},
		IsError: true,
	}
	h.sendSuccessResponse(w, id, result)
}

func (h *MCPHandler) sendToolResult(w http.ResponseWriter, id interface{}, result interface{}) {
	// The YouTube API responses are complex structs. We send them as JSON.
	// An LLM can then parse this JSON to extract the information it needs.
	finalResult := ToolsCallResult{
		Content: []ContentItem{{Type: "json", JSON: result}},
	}
	h.sendSuccessResponse(w, id, finalResult)
}
