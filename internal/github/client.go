package github

import (
	"fmt"
	"log/slog"

	"github.com/cli/go-gh/v2/pkg/api"
)

// Client wraps the GitHub GraphQL API client
type Client struct {
	gqlClient *api.GraphQLClient
}

// NewClient creates a new GitHub API client using gh CLI authentication
func NewClient() (*Client, error) {
	// Use the gh CLI's authentication
	client, err := api.DefaultGraphQLClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create GitHub client: %w", err)
	}

	return &Client{
		gqlClient: client,
	}, nil
}

// ExecuteQuery executes a GraphQL query and returns the response
func (c *Client) ExecuteQuery(query string, variables interface{}, result interface{}) error {
	slog.Debug("Executing GraphQL query", "query", query, "variables", variables)

	// Convert variables to map[string]interface{} if needed
	var varMap map[string]interface{}
	if variables != nil {
		if vm, ok := variables.(map[string]interface{}); ok {
			varMap = vm
		} else {
			// If it's not already a map, we need to convert it
			// This is a simple approach, might need more sophisticated handling in production
			varMap = make(map[string]interface{})
			// For now, assume QueryVariables struct format
			if qv, ok := variables.(QueryVariables); ok {
				varMap["query"] = qv.Query
				if qv.Cursor != nil {
					varMap["cursor"] = *qv.Cursor
				}
			}
		}
	}

	err := c.gqlClient.Do(query, varMap, result)
	if err != nil {
		return fmt.Errorf("GraphQL query failed: %w", err)
	}

	return nil
}
