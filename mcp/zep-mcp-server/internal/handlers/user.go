package handlers

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/getzep/zep/mcp/zep-mcp-server/internal/transform"
	zepclient "github.com/getzep/zep/mcp/zep-mcp-server/pkg/zep"
)

// HandleGetUser registers an MCP tool handler for get_user.
// It fetches a user by ID via client.User.Get and returns the user as MCP text JSON.
//
// Parameters:
//   - client: authenticated Zep API client
//
// Returns:
//   - an mcp.ToolHandlerFor that accepts GetUserInput and returns (*mcp.CallToolResult, any, error)
func HandleGetUser(client *zepclient.Client) mcp.ToolHandlerFor[GetUserInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input GetUserInput) (*mcp.CallToolResult, any, error) {
		// Get user
		user, err := client.User.Get(ctx, input.UserID)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get user: %w", err)
		}

		// Format results as JSON
		resultJSON, err := transform.FormatJSON(user)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to format results: %w", err)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: resultJSON,
				},
			},
		}, user, nil
	}
}
