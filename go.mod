module tracer-demo

go 1.23.3

replace github.com/alibaba/opentelemetry-go-auto-instrumentation/test/verifier => github.com/alibaba/opentelemetry-go-auto-instrumentation/test/verifier v0.0.0-20250217013025-877aae4c1b49

require (
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	go.opentelemetry.io/auto/sdk v1.2.0 // indirect
	go.opentelemetry.io/otel v1.38.0 // indirect
	go.opentelemetry.io/otel/metric v1.38.0 // indirect
	go.opentelemetry.io/otel/trace v1.38.0 // indirect
)
