# Release v0.3.0

Released: 2026-01-31

## Summary

This release introduces major networking, observability, and workload management enhancements to Kthena. Key highlights include **Gateway API support** for advanced traffic management, a comprehensive **Observability framework**, **LeaderWorkerSet integration**, and **HPA support** for ModelServing. Additionally, it brings a new **Binpack Scale Down** strategy and significant improvements to the CLI and E2E testing framework.

*Note: This release includes a breaking change to the PodGroup API.*

## What's New

### Gateway API Support

**Background and Motivation**:
Previously, `ModelRoute` resources shared a global `modelName` namespace, leading to conflicts when multiple users attempted to route the same model name (e.g., `deepseek-r1`). Gateway API support resolves this by allowing independent routing spaces bound to different Gateways. It also lays the foundation for supporting the Kubernetes Gateway API Inference Extension.

**Key Capabilities**:

- **Independent Routing**: Bind `ModelRoute` to specific Gateways to isolate traffic.
- **Conflict Resolution**: Multiple `ModelRoute` resources can use the same `modelName` if bound to different Gateways.
- **Flexible Listeners**: Support for multiple listeners and protocols via Gateway API.

**Configuration**:

```bash
helm install kthena ... --set networking.kthenaRouter.gatewayAPI.enabled=true
```

**Related**:

- User Guide: [Gateway API Support](docs/kthena/docs/user-guide/gateway-api-support.md)
- Issues: #546
- PRs: `e2e for gateway inference extension`, `doc for gateway api support`

### Observability Framework

**Background and Motivation**:
Diagnosing performance issues in LLM inference (e.g., slow time-to-first-token, 5xx errors) requires deep visibility. The new observability framework provides production-grade monitoring capabilities directly from the router.

**Key Capabilities**:

- **Prometheus Metrics**: Rich metrics for requests, latency (prefill/decode), tokens, and scheduler fairness exposed on port `8080`.
- **Structured Access Logs**: Detailed per-request JSON logs for forensic analysis.
- **Debug Endpoints**: Real-time inspection of routing tables and upstream health on a separate port `15000` (Issue #545).

**Configuration**:
Enable via Helm or environment variables (`ACCESS_LOG_ENABLED`, `ACCESS_LOG_FORMAT`).

**Related**:

- User Guide: [Router Observability](docs/kthena/docs/user-guide/router-observability.md)
- Issues: #545, #561

### ModelServing Scale Subresource (HPA Support)

**Background and Motivation**:
To enable autoscaling of inference workloads based on metrics like QPS or GPU utilization, `ModelServing` resources now expose a standard Kubernetes scale subresource. This allows integration with Horizontal Pod Autoscalers (HPA) and use of `kubectl scale` (#564).

**Configuration**:

```bash
kubectl autoscale modelserving my-model --min=1 --max=5 --cpu-percent=80
```

### Workload Management Enhancements

**LeaderWorkerSet Integration**:
Added support for **LeaderWorkerSet** in ModelServing Roles, enabling more complex distributed inference topologies where leader-worker coordination is required (#407, #683).

**ModelServing Version Control & Partitioning**:
Introduced revision tracking and partition protection for ModelServing. This allows for controlled rollouts and safer scaling operations by protecting specific partitions during updates (#584, #661, #653).

**vLLM Parallel Deployment**:
Enhanced support for vLLM with Data Parallel and Expert Parallel deployment modes, allowing for more efficient scaling of large models (#586).

### Binpack Scale Down

**Background and Motivation**:
Standard scale-down removes pods based on ID. Binpack scale-down utilizes `controller.kubernetes.io/pod-deletion-cost` to selectively remove the group or role with the "lowest cost" (least important workload) first, maximizing available contiguous node capacity for large upcoming jobs.

**Related**:

- User Guide: [Binpack Scale Down](docs/kthena/docs/user-guide/binpack-scale-down.md)

## Breaking Changes

### PodGroup API Update

The `PodGroup` API (imported from Volcano) has changed. The field `minTaskMember` has been replaced by `subGroupSize` in Gang scheduling policies (Issue #532).

**Action Required**:
Update your `PodGroup` manifests to use `subGroupSize`:

```yaml
spec:
  subGroupSize: 1  # Replaces minTaskMember
```

## Other Notable Changes

### Improvements

- **Client Performance**: Added ability to customize client QPS and Burst settings (#686).
- **CLI Templates**: Added templates for PD disaggregation use cases (#571).
- **One-Click Deploy**: Enhanced `hack/local-up-kthena.sh` for easy source deployment.
- **Controller Flags**: Added `--controllers` flag to selectively start controllers.
- **Service Naming**: Renamed services to use `kthena-` prefix for consistency.
- **Webhooks**: Enabled `ModelServing` webhooks by default in Helm charts (#694).

### Bug Fixes

- **Stability**: Fixed multiple panic scenarios in the scheduler and controller (#714, #703, #688).
- **Recovery**: Fixed issues where failed pods were not correctly recovered after controller restart (#697) and Role status transition issues (#706).
- **Controller Crash**: Fixed an issue where `kthena-controller-manager` failed to list ModelRoute CRDs when networking was disabled (#567).
- **Headless Service**: Fixed an issue where headless services would not recover after deletion (#169).
- **Scheduling**: Fixed a divide-by-zero error in `LeastRequest` scoring algorithm (#723).

### Documentation & Quality

- **Helm Docs**: Added automated generation for Helm chart documentation (#583).
- **New Guides**: Partition Revision Control (#653), LeaderWorkerSet Integration (#683).
- **E2E Tests**: Extensive new tests for LoRA (#649), Rate Limiting (#669), and Disaggregation (#593).
- **Linting**: Added CI workflows for Python linting (#547).

## Contributors

Thank you to all contributors who made this release possible:

@YaoZengzeng, @LiZhenCheng9527, @Lu Ma, @zhoujinyu, @katara-Jayprakash, @Zhonghu Xu, @liuhaiyu, @Yogesh Kumar
