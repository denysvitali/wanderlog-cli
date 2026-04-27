package wanderlog

import (
	"context"
	"net/http"

	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog/openapi"
)

// OpenAPIClient returns a spec-generated Wanderlog API client wired to the same
// HTTP transport, user agent, and optional authentication as Client.
func (c *Client) OpenAPIClient() (*openapi.ClientWithResponses, error) {
	return c.openAPI()
}

func (c *Client) openAPIRequestEditor(_ context.Context, req *http.Request) error {
	req.Header.Set("User-Agent", c.userAgent)
	if c.auth == nil {
		return nil
	}
	return c.addAuthHeaders(req)
}
