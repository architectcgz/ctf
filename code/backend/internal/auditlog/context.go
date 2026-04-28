package auditlog

import "context"

type Control struct {
	Skip bool
}

type controlKey struct{}

func WithControl(ctx context.Context, control *Control) context.Context {
	if ctx == nil {
		return nil
	}
	return context.WithValue(ctx, controlKey{}, control)
}

func ControlFromContext(ctx context.Context) *Control {
	if ctx == nil {
		return nil
	}
	control, _ := ctx.Value(controlKey{}).(*Control)
	return control
}

func MarkSkip(ctx context.Context) {
	if control := ControlFromContext(ctx); control != nil {
		control.Skip = true
	}
}
