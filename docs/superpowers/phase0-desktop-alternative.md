# Phase 0 Implementation: Desktop Alternative Baseline

This document implements **Phase 0** from `ROADMAP.md` by defining:

1. User profiles
2. Docker Desktop parity baseline (top workflows + table-stakes)
3. Success metrics and targets
4. Capability matrix (current vs target)
5. UX quality rubric
6. Telemetry and logging plan

## 1) Primary User Profiles

### Profile A: Solo Developer
- Works mostly on a single machine.
- Needs fast daily control over containers/images/volumes/networks.
- Success criteria: complete common tasks without leaving the TUI.

### Profile B: Backend Team Developer
- Uses multiple services and Compose projects.
- Frequently checks logs, restarts services, inspects config/state.
- Success criteria: can diagnose and recover local stack issues quickly.

### Profile C: DevOps / Platform Engineer
- Works across multiple Docker contexts and environments.
- Needs reliability, auditability, and safe destructive operations.
- Success criteria: high-confidence operations with clear impact and recovery.

## 2) Top Daily Workflows (Parity Baseline)

The following 20 workflows define parity targets.

| # | Workflow | Current State | Target State |
|---|----------|---------------|--------------|
| 1 | List all containers with status | Supported | Keep + polish |
| 2 | Start/stop/restart a container | Supported | Keep + polish |
| 3 | Remove container(s) with confirmation | Supported | Safer preflight |
| 4 | View container inspect details | Supported | Keep + polish |
| 5 | View container logs in-app | Partial (shell out) | Native integrated logs |
| 6 | Exec into running container | Partial (shell out) | Native integrated terminal pane |
| 7 | Pull image from registry | Supported | Better auth/progress/cancel |
| 8 | Build image from Dockerfile | Supported | Better diagnostics/history |
| 9 | Tag image | Supported | Keep + polish |
| 10 | Remove/prune images safely | Supported | Better impact preview |
| 11 | Create container from image | Supported | Keep + polish |
| 12 | List volumes and inspect usage | Supported | Better dependency mapping |
| 13 | Create/remove/prune volumes | Supported | Better impact preview |
| 14 | List networks and inspect attachments | Supported | Keep + polish |
| 15 | Create/remove/prune networks | Supported | Better impact preview |
| 16 | List compose-derived services | Supported | Keep + polish |
| 17 | Start/stop/restart service containers | Supported | Keep + polish |
| 18 | Open compose file from service | Supported | Better project-level controls |
| 19 | Search local + remote images | Supported | Better filtering/discoverability |
| 20 | Complete operations with clear error recovery guidance | Partial | Actionable error UX |

## 3) Table-Stakes Expectations

These are the 10 baseline expectations users compare against Docker Desktop.

1. Consistent keyboard behavior across all tabs.
2. Predictable filtering and focus behavior.
3. Safe destructive operations with clear impact.
4. Integrated logs and shell/exec experience.
5. Reliable service and compose workflows.
6. Clear and actionable error messages.
7. Fast, responsive rendering under load.
8. Cross-platform reliability (Linux/macOS/Windows).
9. Good discoverability for common actions.
10. Stable behavior across releases.

## 4) Success Metrics and Targets

### Core Product Metrics

| Metric | Definition | Baseline Method | Phase 1 Target |
|--------|------------|-----------------|----------------|
| Task completion rate | % of users who complete key workflow without docs | Structured usability sessions | >= 85% on top 10 workflows |
| Median task time | Time to finish common operation | Timed scenario runs | -25% from initial baseline |
| Crash-free session rate | % sessions without panic/terminal breakage | Automated + manual session tracking | >= 99% |
| Error recovery success | % failed operations resolved without external docs | Scenario-based tests | >= 80% |
| UX consistency defects | Count of cross-tab inconsistency bugs | Weekly QA triage | -60% from baseline |

### Suggested Baseline Scenarios

- Stop and restart a failing service.
- Pull and run a new image.
- Remove a stopped container safely.
- Prune unused resources without surprise deletions.
- Inspect container/service configuration and copy output.

## 5) Capability Matrix (Current vs Target)

| Capability Area | Current | Gap Level | Target |
|-----------------|---------|----------|--------|
| Container lifecycle ops | Strong | Low | Stability + polish |
| Logs workflow | Medium | High | Native in-app logs experience |
| Exec workflow | Medium | High | Native terminal/exec pane |
| Compose project lifecycle | Medium | High | Project-level controls + health |
| Image lifecycle management | Medium | Medium | Rich metadata + safer cleanup |
| Volume/network ergonomics | Medium | Medium | Better dependency/impact UX |
| Discoverability | Medium | Medium | Command palette + contextual hints |
| Error recovery UX | Medium | High | Actionable, guided recovery |
| Cross-platform resilience | Low/Medium | High | Defined parity and tested support |
| Extensibility/integration | Low | Medium | Lightweight plugin/action hooks |

## 6) UX Quality Rubric

Score each category from 1 to 5. A release candidate should average >= 4.0 with no category below 3.

| Category | 1 (Poor) | 3 (Acceptable) | 5 (Excellent) |
|----------|----------|----------------|---------------|
| Consistency | Different behavior per tab | Mostly aligned, minor drift | Fully aligned interaction model |
| Discoverability | Hidden key actions | Help exists but fragmented | Fast, obvious action discovery |
| Feedback | Delayed/unclear operation status | Basic status and errors | Clear progress, outcomes, and next steps |
| Safety | Risky destructive flows | Confirmations present | Impact preview + safe defaults |
| Recovery | Raw errors only | Some actionable errors | Guided remediation and retries |
| Performance | Noticeable lag/stutter | Mostly responsive | Snappy and stable under realistic load |

## 7) Telemetry and Logging Plan (Privacy-Safe)

Telemetry is **opt-in** and local-first by default.

### Event Schema (No Sensitive Payloads)

| Event | Properties |
|-------|------------|
| `tab_switched` | `from_tab`, `to_tab`, `duration_ms` |
| `action_invoked` | `tab`, `action`, `selection_count` |
| `action_result` | `tab`, `action`, `status`, `duration_ms`, `error_category` |
| `filter_used` | `tab`, `filter_len`, `results_count` |
| `overlay_opened` | `overlay_type`, `tab` |
| `overlay_closed` | `overlay_type`, `tab`, `reason` |
| `panic_recovered` | `component`, `context` |

### Logging Rules

- Never log full command output containing secrets.
- Never log image pull credentials, env var values, or compose secrets.
- Use coarse error categories (`network`, `permission`, `daemon_unreachable`, `validation`, `unknown`).
- Redact identifiers when persisted outside debug mode.

### Storage and Transport

- Default: local session logs only.
- Optional: export anonymized metrics snapshot manually.
- No automatic remote upload in MVP.

### Initial Instrumentation Targets

1. Tab switches and action invocation results.
2. Error categories for failed operations.
3. Operation duration histograms for top workflows.

## 8) Phase 0 Exit Criteria Checklist

- [x] User profiles are defined.
- [x] Top 20 workflows and table-stakes expectations are documented.
- [x] Success metrics and targets are defined.
- [x] Capability matrix is documented.
- [x] UX quality rubric is documented.
- [x] Telemetry/logging policy and schema are documented.

Phase 0 is complete once this document is accepted as the baseline for Phase 1 planning and implementation.
