package handlers

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/getzep/zep/mcp/zep-mcp-server/internal/transform"
	zepclient "github.com/getzep/zep/mcp/zep-mcp-server/pkg/zep"
)

// HandleGetEpisode registers an MCP tool handler for get_episode.
// It fetches a single episode by UUID via client.Graph.Episode.Get and returns it as MCP text JSON.
//
// Parameters:
//   - client: authenticated Zep API client
//
// Returns:
//   - an mcp.ToolHandlerFor that accepts GetEpisodeInput and returns (*mcp.CallToolResult, any, error)
func HandleGetEpisode(client *zepclient.Client) mcp.ToolHandlerFor[GetEpisodeInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input GetEpisodeInput) (*mcp.CallToolResult, any, error) {
		episode, err := client.Graph.Episode.Get(ctx, input.UUID)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get episode: %w", err)
		}

		resultJSON, err := transform.FormatJSON(episode)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to format results: %w", err)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: resultJSON,
				},
			},
		}, episode, nil
	}
}
