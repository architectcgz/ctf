package commands

import "context"

type awdReadinessGateTraceKey struct{}

type AWDReadinessGateTrace struct {
	executed bool
	allowed  bool
}

func WithAWDReadinessGateTrace(ctx context.Context) (context.Context, *AWDReadinessGateTrace) {
	if ctx == nil {
		ctx = context.Background()
	}
	if trace := AWDReadinessGateTraceFromContext(ctx); trace != nil {
		return ctx, trace
	}

	trace := &AWDReadinessGateTrace{}
	return context.WithValue(ctx, awdReadinessGateTraceKey{}, trace), trace
}

func AWDReadinessGateTraceFromContext(ctx context.Context) *AWDReadinessGateTrace {
	if ctx == nil {
		return nil
	}
	trace, _ := ctx.Value(awdReadinessGateTraceKey{}).(*AWDReadinessGateTrace)
	return trace
}

func (t *AWDReadinessGateTrace) RecordDecision(allowed bool) {
	if t == nil {
		return
	}
	t.executed = true
	t.allowed = allowed
}

func (t *AWDReadinessGateTrace) Executed() bool {
	return t != nil && t.executed
}

func (t *AWDReadinessGateTrace) Allowed() bool {
	return t != nil && t.executed && t.allowed
}
