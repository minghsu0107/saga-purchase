package repo

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/trace"
)

var (
	ServiceNameHeader   string = "Service-Name"
	W3CSupportedVersion        = 0
)

func encodeGRPCRequest(_ context.Context, request interface{}) (interface{}, error) {
	return request, nil
}

func decodeGRPCResponse(_ context.Context, grpcReply interface{}) (interface{}, error) {
	return grpcReply, nil
}

func spanContextToW3C(ctx context.Context) string {
	sc := trace.SpanContextFromContext(ctx)
	if !sc.IsValid() {
		return ""
	}
	// Clear all flags other than the trace-context supported sampling bit.
	flags := sc.TraceFlags() & trace.FlagsSampled
	return fmt.Sprintf("%.2x-%s-%s-%s",
		W3CSupportedVersion,
		sc.TraceID(),
		sc.SpanID(),
		flags)
}
