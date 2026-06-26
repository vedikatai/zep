package handlers

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/getzep/zep/mcp/zep-mcp-server/internal/transform"
	zepclient "github.com/getzep/zep/mcp/zep-mcp-server/pkg/zep"
)

// HandleGetNodeEdges registers an MCP tool handler for get_node_edges.
// It lists edges connected to a node via client.Graph.Node.GetEdges and returns them as MCP text JSON.
//
// Parameters:
//   - client: authenticated Zep API client
//
// Returns:
//   - an mcp.ToolHandlerFor that accepts GetNodeEdgesInput and returns (*mcp.CallToolResult, any, error)
func HandleGetNodeEdges(client *zepclient.Client) mcp.ToolHandlerFor[GetNodeEdgesInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input GetNodeEdgesInput) (*mcp.CallToolResult, any, error) {
		edges, err := client.Graph.Node.GetEdges(ctx, input.NodeUUID)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get node edges: %w", err)
		}

		resultJSON, err := transform.FormatJSON(edges)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to format results: %w", err)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: resultJSON,
				},
			},
		}, edges, nil
	}
}
