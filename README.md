# Saga Purchase
Purchase service of the [saga pattern implementation](https://github.com/minghsu0107/saga-example).

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
See [docker-compose example](https://github.com/minghsu0107/saga-example/blob/main/docker-compose.yaml) for details.
## Exported Metrics
| Metric                                                                                                                                   | Description                                                                                                 | Labels                                                           |
| ---------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------- |
| purchase_pubsub_subscriber_messages_received_total                                                                                                       | A Prometheus Counter. Counts the number of messages obtained by the subscriber.                             | `acked` ("acked" or "nacked"), `handler_name`, `subscriber_name` |
| purchase_pubsub_publish_time_seconds (purchase_pubsub_publish_time_seconds_count, purchase_pubsub_publish_time_seconds_bucket, purchase_pubsub_publish_time_seconds_sum)                                 | A Prometheus Histogram. Registers the time of execution of the Publish function of the decorated publisher. | `handler_name`, `success` ("true" or "false"), `publisher_name`  |
| purchase_http_request_duration_seconds (purchase_http_request_duration_seconds_count, purchase_http_request_duration_seconds_bucket, purchase_http_request_duration_sum)     | A Prometheus histogram. Records the latency of the HTTP requests.                                           | `code`, `handler`, `method`                                      |
| purchase_http_requests_inflight                                                                                                                   | A Prometheus gauge. Records the number of inflight requests being handled at the same time.                 | `code`, `handler`, `method`                                      |
| purchase_http_response_size_bytes (purchase_http_response_size_bytes_count, purchase_http_response_size_bytes_bucket, purchase_http_response_size_bytes_sum)                 | A Prometheus histogram. Records the size of the HTTP responses.                                             | `handler`                                                        |
