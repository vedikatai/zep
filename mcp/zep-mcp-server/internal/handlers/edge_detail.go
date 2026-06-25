package handlers

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/getzep/zep/mcp/zep-mcp-server/internal/transform"
	zepclient "github.com/getzep/zep/mcp/zep-mcp-server/pkg/zep"
)

// HandleGetEdge registers an MCP tool handler for get_edge.
// It fetches a single graph edge by UUID via client.Graph.Edge.Get and returns it as MCP text JSON.
//
// Parameters:
//   - client: authenticated Zep API client
//
// Returns:
//   - an mcp.ToolHandlerFor that accepts GetEdgeInput and returns (*mcp.CallToolResult, any, error)
func HandleGetEdge(client *zepclient.Client) mcp.ToolHandlerFor[GetEdgeInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input GetEdgeInput) (*mcp.CallToolResult, any, error) {
		edge, err := client.Graph.Edge.Get(ctx, input.UUID)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get edge: %w", err)
		}

		resultJSON, err := transform.FormatJSON(edge)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to format results: %w", err)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: resultJSON,
				},
			},
		}, edge, nil
	}
}
