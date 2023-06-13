package portergrpc

import (
	"context"
	"testing"

	"get.porter.sh/porter/pkg/porter"
	"github.com/stretchr/testify/assert"
)

func TestGetPorterConnectionFromContextReturnsErrIfNoConnectionInContext(t *testing.T) {
	ctx := context.Background()
	p, err := GetPorterConnectionFromContext(ctx)
	assert.Nil(t, p)
	assert.EqualError(t, err, "Unable to find porter connection in context")
}

func TestGetPorterConnectionFromContextReturnsPorterConnection(t *testing.T) {
	p := porter.New()
	ctx := context.Background()
	ctx = context.WithValue(ctx, porterConnCtxKey, p)
	newP, err := GetPorterConnectionFromContext(ctx)
	assert.Nil(t, err)
	assert.Equal(t, p, newP)
}
func TestAddPorterConnectionToContextReturnsContextUpdatedWithPorterConnection(t *testing.T) {
	p := porter.New()
	ctx := context.Background()
	ctx = AddPorterConnectionToContext(p, ctx)
	newP, ok := ctx.Value(porterConnCtxKey).(*porter.Porter)
	assert.True(t, ok)
	assert.Equal(t, p, newP)
}
