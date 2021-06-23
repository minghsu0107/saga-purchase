# Saga Purchase
Purchase microservice of the [saga pattern implementation](https://github.com/minghsu0107/saga-example).

Features:
- Realtime event-driven subscription using [NATS Streaming](https://docs.nats.io/nats-streaming-concepts/intro) and [server-sent events (SSE)](https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events/Using_server-sent_events)
- Prometheus metrics
- Distributed tracing exporter
  - HTTP server 
  - gPRC client
- Comprehensive application struture with domain-driven design (DDD), decoupling service implementations from configurations and transports
- Compile-time dependecy injection using [wire](https://github.com/google/wire)
- Graceful shutdown
- Unit testing and continuous integration using [Drone CI](https://www.drone.io)
## Usage
Build from source:
```bash
go mod tidy
make build
```
Start the service:
```bash
NATS_URL=nats://nats-streaming:4222 \
NATS_CLUSTER_ID=test-cluster \
RPC_AUTH_SVC_HOST=saga-account:8000 \
RPC_PRODUCT_SVC_HOST=saga-product:8000 \
OC_AGENT_HOST=oc-collector:55678 \
./server
```
- `NATS_URL`: NATS Streaming server URL.
- `NATS_CLUSTER_ID`: NATS Cluster ID
- `RPC_AUTH_SVC_HOST`: gRPC account service host
- `RPC_PRODUCT_SVC_HOST`: gRPC product service host
## Running in Docker
See [docker-compose example](https://github.com/minghsu0107/saga-example/blob/a47f998fee6112941133a08ad4dd75e5a342b0bf/docker-compose.yaml#L23) for details.
## Exported Metrics
| Metric                                                                                                                                   | Description                                                                                                 | Labels                                                           |
| ---------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------- |
| subscriber_messages_received_total                                                                                                       | A Prometheus Counter. Counts the number of messages obtained by the subscriber.                             | `acked` ("acked" or "nacked"), `handler_name`, `subscriber_name` |
| handler_execution_time_seconds (handler_execution_time_seconds_count, handler_execution_time_bucket, handler_execution_time_seconds_sum) | A Prometheus Histogram. Records the execution time of the handler function wrapped by the middleware.       | `handler_name`, `success` ("true" or "false")                    |
| publish_time_seconds (publish_time_seconds_count, publish_time_seconds_bucket, publish_time_seconds_sum)                                 | A Prometheus Histogram. Registers the time of execution of the Publish function of the decorated publisher. | `handler_name`, `success` ("true" or "false"), `publisher_name`  |
| http_request_duration_seconds (http_request_duration_seconds_count, http_request_duration_seconds_bucket, http_request_duration_sum)     | A Prometheus histogram. Records the latency of the HTTP requests.                                           | `code`, `handler`, `method`                                      |
| http_requests_inflight                                                                                                                   | A Prometheus gauge. Records the number of inflight requests being handled at the same time.                 | `code`, `handler`, `method`                                      |
| http_response_size_bytes (http_response_size_bytes_count, http_response_size_bytes_bucket, http_response_size_bytes_sum)                 | A Prometheus histogram. Records the size of the HTTP responses.                                             | `handler`                                                        |