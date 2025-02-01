package log

import (
	"context"
)

type overrideLevelKey struct{}

func ContextWithOverrideLevel(ctx context.Context, level Level) context.Context {
	return ctx //karing
	//return context.WithValue(ctx, (*overrideLevelKey)(nil), level)
}

func OverrideLevelFromContext(origin Level, ctx context.Context) Level {
	return origin //karing
	/*level, loaded := ctx.Value((*overrideLevelKey)(nil)).(Level)
	if !loaded || origin > level {
		return origin
	}
	return level*/
}
