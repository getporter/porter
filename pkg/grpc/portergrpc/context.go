package portergrpc

import (
	"context"
	"errors"

	"get.porter.sh/porter/pkg/porter"
)

type ctxKey int

const porterConnCtxKey ctxKey = 0

// GetPorterConnectionFromContext returns the porter connection from the give context.
// Use AddPorterConnectionToContext to add one.
func GetPorterConnectionFromContext(ctx context.Context) (*porter.Porter, error) {
	p, ok := ctx.Value(porterConnCtxKey).(*porter.Porter)
	if !ok {
		return nil, errors.New("Unable to find porter connection in context")
	}
	return p, nil
}

// AddPorterConnectionToContext adds the porter connection to the given context
// use GetPorterConnectionFromContext to read it back out
func AddPorterConnectionToContext(p *porter.Porter, ctx context.Context) context.Context {
	return context.WithValue(ctx, porterConnCtxKey, p)
}
