package handlers

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/getzep/zep/mcp/zep-mcp-server/internal/transform"
	zepclient "github.com/getzep/zep/mcp/zep-mcp-server/pkg/zep"
)

// HandleGetEpisodeMentions registers an MCP tool handler for get_episode_mentions.
// It loads nodes and edges mentioned by an episode via client.Graph.Episode.GetNodesAndEdges
// and returns the mentions payload as MCP text JSON.
//
// Parameters:
//   - client: authenticated Zep API client
//
// Returns:
//   - an mcp.ToolHandlerFor that accepts GetEpisodeMentionsInput and returns (*mcp.CallToolResult, any, error)
func HandleGetEpisodeMentions(client *zepclient.Client) mcp.ToolHandlerFor[GetEpisodeMentionsInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input GetEpisodeMentionsInput) (*mcp.CallToolResult, any, error) {
		mentions, err := client.Graph.Episode.GetNodesAndEdges(ctx, input.UUID)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get episode mentions: %w", err)
		}

		resultJSON, err := transform.FormatJSON(mentions)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to format results: %w", err)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: resultJSON,
				},
			},
		}, mentions, nil
	}
}
