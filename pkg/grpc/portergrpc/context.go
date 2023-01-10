package portergrpc

import (
	"context"
	"errors"

	"get.porter.sh/porter/pkg/porter"
)

type ctxKey int

const porterConnCtxKey ctxKey = 0

func GetPorterConnectionFromContext(ctx context.Context) (*porter.Porter, error) {
	p, ok := ctx.Value(porterConnCtxKey).(*porter.Porter)
	if !ok {
		return nil, errors.New("Unable to find porter connection in context")
	}
	return p, nil
}

func AddPorterConnectionToContext(p *porter.Porter, ctx context.Context) context.Context {
	return context.WithValue(ctx, porterConnCtxKey, p)
}
