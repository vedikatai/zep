package handlers

import (
	"context"
	"fmt"

	zep "github.com/getzep/zep-go/v3"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/getzep/zep/mcp/zep-mcp-server/internal/transform"
	zepclient "github.com/getzep/zep/mcp/zep-mcp-server/pkg/zep"
)

// HandleGetUserContext registers an MCP tool handler for get_user_context.
// It loads thread user context via client.Thread.GetUserContext (optional template_id)
// and returns the context payload as MCP text JSON.
//
// Parameters:
//   - client: authenticated Zep API client
//
// Returns:
//   - an mcp.ToolHandlerFor that accepts GetUserContextInput and returns (*mcp.CallToolResult, any, error)
func HandleGetUserContext(client *zepclient.Client) mcp.ToolHandlerFor[GetUserContextInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input GetUserContextInput) (*mcp.CallToolResult, any, error) {
		// Build request
		contextReq := &zep.ThreadGetUserContextRequest{}

		if input.TemplateID != "" {
			contextReq.TemplateID = &input.TemplateID
		}

		// Get user context
		memory, err := client.Thread.GetUserContext(ctx, input.ThreadID, contextReq)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get user context: %w", err)
		}

		// Format results as JSON
		resultJSON, err := transform.FormatJSON(memory)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to format results: %w", err)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: resultJSON,
				},
			},
		}, memory, nil
	}
}
