package requestid

import "context"

type key struct{}

func With(ctx context.Context, id string) context.Context {
	if id == "" {
		return ctx
	}
	return context.WithValue(ctx, key{}, id)
}

func From(ctx context.Context) (string, bool) {
	v := ctx.Value(key{})
	s, ok := v.(string)
	if !ok || s == "" {
		return "", false
	}
	return s, true
}
