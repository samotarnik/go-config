package url

import (
	"context"

	"github.com/samotarnik/go-config/source"
)

type urlKey struct{}

func WithURL(u string) source.Option {
	return func(o *source.Options) {
		if o.Context == nil {
			o.Context = context.Background()
		}
		o.Context = context.WithValue(o.Context, urlKey{}, u)
	}
}
