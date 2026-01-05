# Router Observability

## Overview

Kthena provides comprehensive observability features for monitoring and debugging the router component, which serves as the data plane entry point for inference traffic. This documentation implements the comprehensive observability framework outlined in the Router Observability Proposal, providing detailed metrics, access logs, and debug interfaces for effective AI workload management.

## Router Architecture & Observability

The Kthena router implements a multi-layered observability stack that provides real-time insights into:

- Request routing and load balancing decisions
- Model server health and performance
- Gateway API inference extension integration
- Prefill-decode disaggregation routing
- Token processing and AI-specific metrics

### Prometheus Metrics

Design Principles

**Essential Labels:** Include key dimensions for effective monitoring and debugging:

- `method`: HTTP method (GET, POST) - useful for differentiating request types
- `path`: API path (/v1/chat/completions, /v1/completions) - track different API usage
- `status_code`: HTTP response status code (200, 400, 500) - monitor success/failure patterns
- `model`: AI model name - essential for AI workload monitoring
- `error_type`: Specific error categories for detailed troubleshooting

**Label Cardinality Management:** Keep label values bounded to avoid high cardinality issues:

- Limited set of endpoints and methods
- Standard HTTP status codes
- Controlled model catalog
- Predefined error types

### Built-in Metrics Collection

Kthena router exposes Prometheus metrics on port `9090` at the `/metrics` endpoint, implementing the comprehensive metrics framework from the proposal.

### Request Processing Metrics

HTTP Request Metrics

```yaml
# Total number of HTTP requests processed by the router
kthena_router_requests_total{model="<model_name>",path="<path>",status_code="<code>",error_type="<error_type>"}

# End-to-end request processing latency distribution
kthena_router_request_duration_seconds{model="<model_name>",path="<path>",status_code="<code>"}
# Buckets: [0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10, 30, 60]

# Prefill phase processing latency for PD-disaggregated requests
kthena_router_request_prefill_duration_seconds{model="<model_name>",path="<path>",status_code="<code>"}
# Buckets: [0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10, 30, 60]

# Decode phase processing latency for PD-disaggregated requests
kthena_router_request_decode_duration_seconds{model="<model_name>",path="<path>",status_code="<code>"}
# Buckets: [0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10, 30, 60]

# Current active downstream requests (from clients to router)
kthena_router_active_downstream_requests{model="<model_name>"}

# Current active upstream requests (from router to backend pods)
kthena_router_active_upstream_requests{model_route="<route_name>",model_server="<server_name>"}
```

#### AI-Specific Token Metrics

```yaml
# Total tokens processed/generated
kthena_router_tokens_total{model="<model_name>",path="<path>",token_type="input|output"}
```

#### Scheduler Plugin Metrics

```yaml
# Processing time per scheduler plugin
kthena_router_scheduler_plugin_duration_seconds{model="<model_name>",plugin="<plugin_name>",type="filter|score"}
# Buckets: [0.001, 0.005, 0.01, 0.05, 0.1, 0.5]
```

#### Rate Limiting Metrics

```yaml
# Number of requests rejected due to rate limiting
kthena_router_rate_limit_exceeded_total{model="<model_name>",limit_type="input_tokens|output_tokens|requests",path="<path>"}
```

#### Fairness Queue Metrics

```yaml
# Current fairness queue size for pending requests
kthena_router_fairness_queue_size{model="<model_name>",user_id="<user_id>"}

# Time requests spend in fairness queue before processing
kthena_router_fairness_queue_duration_seconds{model="<model_name>",user_id="<user_id>"}
# Buckets: [0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5]
```

### Core System Metrics

#### Model Server Health

```yaml
kthena_model_server_health_status{model_server="<server_name>",namespace="<namespace>"}
kthena_model_server_queue_depth{model_server="<server_name>",model="<model>"}
kthena_model_server_gpu_utilization{model_server="<server_name>",gpu_id="<gpu_id>"}
kthena_model_server_memory_usage_bytes{model_server="<server_name>",type="gpu|system"}
```

#### Load Balancing Metrics

```yaml
kthena_lb_decisions_total{algorithm="<algorithm>",model="<model>",target="<target>"}
kthena_lb_target_health{target="<target>",model="<model>"}
kthena_lb_connection_pool_size{target="<target>"}
```

#### Prefill-Decode Disaggregation Metrics

```yaml
kthena_pd_group_requests_total{group_key="<group>",role="prefill|decode",model="<model>"}
kthena_pd_group_routing_decisions{group_key="<group>",source_role="<role>",target_role="<role>"}
kthena_pd_group_kv_cache_hit_rate{group_key="<group>",model="<model>"}
kthena_kv_cache_operations_total{connector_type="<type>",operation="<op>",status="<status>"}
kthena_kv_cache_transfer_bytes_total{connector_type="<type>",direction="send|receive"}
kthena_kv_cache_latency_seconds{connector_type="<type>",operation="<op>",quantile="<q>"}
```

### Prometheus Configuration

Create a `prometheus.yml` configuration file to scrape Kthena router metrics:

```yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'kthena-router'
    static_configs:
      - targets: ['kthena-router-service:9090']
        labels:
          component: 'router'
          environment: 'production'
    
  - job_name: 'kthena-model-servers'
    kubernetes_sd_configs:
      - role: pod
        namespaces:
          names:
            - kthena-system
            - ai-inference
    relabel_configs:
      - source_labels: [__meta_kubernetes_pod_label_app]
        action: keep
        regex: 'kthena-model-server.*'
      - source_labels: [__meta_kubernetes_pod_name]
        target_label: pod
      - source_labels: [__meta_kubernetes_namespace]
        target_label: namespace
```

### Access Log Format

The router generates structured access logs for each request, following the comprehensive format defined in the proposal with AI-specific extensions to track model routing and processing stages.

### Log Format Configuration

Access logging can be configured through the AccessLoggerConfig:

```yaml
accessLogger:
  format: "json"  # Options: "json", "text"
  output: "stdout"  # Options: "stdout", "stderr", or file path
  enabled: true
```

#### JSON Format (Default)

```json
{
  "timestamp": "2024-01-15T10:30:45.123Z",
  "method": "POST",
  "path": "/v1/chat/completions",
  "protocol": "HTTP/1.1",
  "status_code": 200,
  
  "model_name": "llama2-7b",
  "model_route": "default/llama2-route-v1",
  "model_server": "default/llama2-server",
  "selected_pod": "llama2-deployment-5f7b8c9d-xk2p4",
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  
  "input_tokens": 150,
  "output_tokens": 75,
  
  "duration_total": 2350,
  "duration_request_processing": 45,
  "duration_upstream_processing": 2180,
  "duration_response_processing": 5,
  
  "error": {
    "type": "timeout",
    "message": "Model inference timeout after 30s"
  }
}
```

#### Text Format (Alternative)

For environments preferring text logs:
`[2024-01-15T10:30:45.123Z] "POST /v1/chat/completions HTTP/1.1" 200 model_name=llama2-7b model_route=default/llama2-route-v1 model_server=default/llama2-server selected_pod=llama2-deployment-5f7b8c9d-xk2p4 request_id=550e8400-e29b-41d4-a716-446655440000 tokens=150/75 timings=2350ms(45+2180+5)`

**Error Example:**
`[2024-01-15T10:30:45.123Z] "POST /v1/chat/completions HTTP/1.1" 500 error=timeout:"Model inference timeout after 30s" model_name=llama2-7b timings=30000ms(45+29950+5)`

Key Fields Explanation

#### Standard HTTP Fields

- `timestamp`: Request start time in ISO 8601 format with nanosecond precision
- `method`, `path`, `protocol`: Standard HTTP request information
- `status_code`: HTTP response status code

#### AI-Specific Routing Fields

- `model_name`: The AI model requested in the request body
- `model_route`: Which ModelRoute CR was matched (namespace/name format)
- `model_server`: Which ModelServer CR was selected (namespace/name format)
- `selected_pod`: The specific pod that processed the inference request
- `request_id`: Unique request identifier

#### Token Tracking

- `input_tokens`: Number of tokens in the input prompt
- `output_tokens`: Number of tokens in the response

#### Detailed Timing Breakdown (milliseconds)

- `duration_total`: End-to-end request processing time
- `duration_request_processing`: Time spent in router request processing
- `duration_upstream_processing`: Actual model inference time on backend pod
- `duration_response_processing`: Time spent processing and formatting response

#### Error Information

- `error`: Detailed error information for failed requests (type and message)

#### Debug Interface

The router exposes debug endpoints at /debug/config_dump to help operators inspect internal state and troubleshoot issues, providing access to the router's datastore information for ModelRoutes, ModelServers, and Pod details.

#### Debug Endpoints

#### List Resources

- `/debug/config_dump/modelroutes` - List all ModelRoute configurations
- `/debug/config_dump/modelservers` - List all ModelServer configurations
- `/debug/config_dump/pods` - List all Pod information

#### Get Specific Resource

- `/debug/config_dump/namespaces/{namespace}/modelroutes/{name}` - Get specific ModelRoute
- `/debug/config_dump/namespaces/{namespace}/modelservers/{name}` - Get specific ModelServer
- `/debug/config_dump/namespaces/{namespace}/pods/{name}` - Get specific Pod

#### Enhanced Debug Endpoints

#### Router State Debug

```
GET /debug/router/state
GET /debug/routing/table
GET /debug/routing/decisions
GET /debug/routing/backends
```

#### Model Server Debug

```
GET /debug/model-servers
GET /debug/model-servers/{name}/status
GET /debug/model-servers/{name}/metrics
```

#### Gateway API Inference Extension Debug

```
GET /debug/inference-extension/pools
GET /debug/inference-extension/pools/{pool}/endpoints
GET /debug/inference-extension/decisions
```

#### Prefill-Decode Disaggregation Debug

```
GET /debug/pd-groups
GET /debug/pd-groups/{group_key}/status
GET /debug/kv-cache/connectors
```

#### Example Debug Responses

```json
{
  "modelroutes": [
    {
      "name": "llama2-route",
      "namespace": "default",
      "spec": {
        "modelName": "llama2-7b",
        "loraAdapters": ["lora-adapter-1", "lora-adapter-2"],
        "rules": [
          {
            "name": "default-rule",
            "modelMatch": {
              "body": {
                "model": "llama2-7b"
              }
            },
            "targetModels": [
              {
                "modelServer": {
                  "name": "llama2-server",
                  "namespace": "default"
                },
                "weight": 100
              }
            ]
          }
        ],
        "rateLimit": {
          "local": {
            "inputTokensPerSecond": 1000,
            "outputTokensPerSecond": 500
          }
        }
      }
    }
  ]
}
```

GET /debug/config_dump/namespaces/default/pods/llama2-deployment-5f7b8c9d-xk2p4

```json
{
  "name": "llama2-deployment-5f7b8c9d-xk2p4",
  "namespace": "default",
  "podInfo": {
    "podIP": "10.244.2.20",
    "nodeName": "worker-node-1",
    "phase": "Running",
    "startTime": "2024-01-15T10:00:00Z",
    "labels": {
      "app": "llama2",
      "version": "v1",
      "role": "inference"
    }
  },
  "engine": "vLLM",
  "metrics": {
    "gpuCacheUsage": 0.75,
    "requestWaitingNum": 3,
    "requestRunningNum": 2,
    "tpot": 0.045,
    "ttft": 1.2
  },
  "models": ["llama2-7b", "lora-adapter-1", "lora-adapter-2"],
  "modelServers": ["default/llama2-server"]
}
```

### Grafana Dashboard

#### Pre-configured Dashboard

Kthena provides a comprehensive Grafana dashboard that visualizes all metrics from the proposal:

**Dashboard Configuration:**

```json
{
  "dashboard": {
    "title": "Kthena Router Observability",
    "uid": "kthena-router-observability",
    "tags": ["kthena", "router", "ai-inference"],
    "panels": [
      {
        "title": "Request Rate by Model",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(kthena_router_requests_total[5m])",
            "legendFormat": "{{model}}"
          }
        ]
      },
      {
        "title": "Token Processing Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(kthena_router_tokens_total[5m])",
            "legendFormat": "{{model}} - {{token_type}}"
          }
        ]
      },
      {
        "title": "Request Latency Distribution",
        "type": "heatmap",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, kthena_router_request_duration_seconds)",
            "legendFormat": "95th percentile"
          }
        ]
      },
      {
        "title": "Rate Limit Violations",
        "type": "stat",
        "targets": [
          {
            "expr": "rate(kthena_router_rate_limit_exceeded_total[5m])",
            "legendFormat": "{{model}} - {{limit_type}}"
          }
        ]
      }
    ]
  }
}
```

Key Dashboard Sections

1. HTTP Request Overview

- Request rate by model, path, and status code
- Error rate tracking with error type breakdown
- End-to-end latency distribution with percentiles

1. AI-Specific Metrics

- Token processing rates (input vs output)
- Model-specific performance metrics
- Token cost analysis and trends

1. Advanced Features

- Prefill/decode phase latency breakdown
- Scheduler plugin performance
- Rate limiting effectiveness
- Fairness queue metrics

#### Alerting Configuration

**Prometheus Alert Rules**

Create alert rules based on the proposal's metrics:

```yaml
groups:
- name: kthena_router_alerts
  rules:
  - alert: HighRouterErrorRate
    expr: rate(kthena_router_requests_total{status_code=~"5.."}[5m]) > 0.1
    for: 2m
    labels:
      severity: critical
    annotations:
      summary: "High error rate in Kthena router"
      description: "Router error rate is {{ $value | humanizePercentage }} for model {{ $labels.model }}"

  - alert: HighRequestLatency
    expr: histogram_quantile(0.95, kthena_router_request_duration_seconds) > 5
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "High request latency detected"
      description: "95th percentile latency is {{ $value }}s for model {{ $labels.model }}"

  - alert: HighTokenProcessingRate
    expr: rate(kthena_router_tokens_total[1m]) > 10000
    for: 2m
    labels:
      severity: info
    annotations:
      summary: "High token processing rate"
      description: "Token processing rate is {{ $value }}/s for model {{ $labels.model }}"

  - alert: RateLimitExceeded
    expr: rate(kthena_router_rate_limit_exceeded_total[5m]) > 0.05
    for: 2m
    labels:
      severity: warning
    annotations:
      summary: "Rate limit being exceeded"
      description: "Rate limit exceeded {{ $value }}/s for {{ $labels.limit_type }} on model {{ $labels.model }}"
```

#### Troubleshooting Guide

##### Using Metrics for Troubleshooting

1. High Error Rate Investigation

```bash
# Check error types and patterns
curl http://localhost:9090/metrics | grep kthena_router_requests_total

# Analyze latency breakdown
curl http://localhost:9090/metrics | grep kthena_router_request_duration_seconds
```

1. Performance Issues

```bash
# Check token processing rates
curl http://localhost:9090/metrics | grep kthena_router_tokens_total

# Monitor scheduler plugin performance
curl http://localhost:9090/metrics | grep kthena_router_scheduler_plugin_duration_seconds
```

1. Rate Limiting Analysis

```bash
# Check rate limit violations
curl http://localhost:9090/metrics | grep kthena_router_rate_limit_exceeded_total

# Monitor fairness queue metrics
curl http://localhost:9090/metrics | grep kthena_router_fairness_queue
```

##### Using Access Logs for Debugging

1. Request Tracing

```bash
# Find specific request by ID
grep "request_id=550e8400-e29b-41d4-a716-446655440000" router.log

# Analyze timing breakdown
grep "model_name=llama2-7b" router.log | jq '.duration_upstream_processing'
```

1. Error Analysis

```bash
# Find all error requests
grep '"error":' router.log

# Analyze error patterns by type
grep '"type": "timeout"' router.log | wc -l
```

Using Debug Interface

1. Configuration Verification

```bash
# Verify ModelRoute configuration
curl http://localhost:9090/debug/config_dump/modelroutes

# Check ModelServer details
curl http://localhost:9090/debug/config_dump/modelservers
```

1. Runtime State Inspection

```bash
# Check current routing table
curl http://localhost:9090/debug/routing/table

# Monitor active requests
curl http://localhost:9090/debug/router/state
```

#### Integration and Configuration

**Router Configuration**

```yaml
# Router observability configuration
observability:
  metrics:
    enabled: true
    port: 9090
    path: /metrics
    
  accessLog:
    enabled: true
    format: "json"  # or "text"
    output: "stdout"  # or "stderr", "/path/to/file"
    
  debug:
    enabled: true
    port: 9090
```

**Prometheus ServiceMonitor**

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: kthena-router-monitor
  namespace: kthena-system
spec:
  selector:
    matchLabels:
      app: kthena-router
  endpoints:
  - port: metrics
    interval: 15s
    path: /metrics
```

#### Performance Considerations

**Metrics Cardinality Management**

- `Bounded Labels`: All label values are constrained to prevent exponential cardinality growth
- `Controlled Model Catalog`: Model names come from a controlled set of deployed models
- `Standard Status Codes`: HTTP status codes are limited to standard values
- `Predefined Error Types`: Error types are predefined and bounded

**Access Log Performance**

- `Asynchronous Logging`: Log writing is asynchronous to avoid blocking request processing
- `Configurable Sampling`: Support for sampling high-traffic scenarios
- `Output Flexibility`: Multiple output options (stdout, files, external collectors)

**Debug Interface Security**

- `Configurable Access`: Debug endpoints can be restricted or disabled in production
- `Read-Only Operations`: All debug endpoints are read-only for safety
- `Sensitive Data Filtering`: No sensitive configuration data is exposed
