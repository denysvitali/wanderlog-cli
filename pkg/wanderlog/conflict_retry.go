package wanderlog

import (
	"context"
	"errors"
	"net/http"

	"github.com/sirupsen/logrus"
)

const maxJSON0MutationAttempts = 3

type json0OperationBuilder func(context.Context) ([]Operation, error)

// retryJSON0MutationContext applies freshly built JSON0 operations and retries
// conflicts only. The builder runs before every attempt so an operation is
// never replayed against a newer document with stale indices or old values.
func (c *Client) retryJSON0MutationContext(ctx context.Context, tripKey, operation string, build json0OperationBuilder) error {
	if ctx == nil {
		ctx = context.Background()
	}
	for attempt := 1; attempt <= maxJSON0MutationAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}

		ops, err := build(ctx)
		if err != nil {
			return err
		}
		if len(ops) == 0 {
			return nil
		}

		err = c.ApplyOperationsContext(ctx, tripKey, ops)
		if err == nil {
			return nil
		}
		if !isJSON0Conflict(err) || attempt == maxJSON0MutationAttempts {
			return err
		}

		c.logger.WithFields(logrus.Fields{
			"tripKey":   tripKey,
			"operation": operation,
			"attempt":   attempt,
		}).Debug("JSON0 conflict; refetching trip and rebuilding operations")
	}
	return nil
}

func isJSON0Conflict(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr) && apiErr.HTTPStatus == http.StatusConflict
}
