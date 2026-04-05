# ContainerTUI Desktop Alternative Roadmap

**Goal:** Turn ContainerTUI from a strong terminal utility into a trustworthy daily-driver alternative for Docker Desktop users by closing critical UX, workflow, and reliability gaps.

**Architecture:** Deliver in staged tracks: (1) trust/stability baseline, (2) core parity workflows, (3) advanced ecosystem capabilities, (4) differentiation. Each stage has explicit exit criteria and measurable user outcomes.

**Tech Stack:** Go, Bubble Tea/Bubbles/Lipgloss, Docker/Moby API, current tabbed TUI architecture.

---

## Phase 0: Product Definition + Success Metrics (2-3 weeks)

**Outcome:** Clear "what we are replacing" scope and measurable targets.

- Define user profiles:
  - solo developer
  - backend team member
  - DevOps/platform engineer
- Define parity baseline with Docker Desktop:
  - top 20 daily workflows
  - top 10 "table-stakes" expectations
- Set measurable goals:
  - task completion time for common flows
  - success rate without docs
  - crash-free session rate
  - top complaint categories trend

**Deliverables**
- Capability matrix: current vs target parity
- UX quality rubric (consistency, discoverability, feedback, error recovery)
- Telemetry/logging plan (privacy-safe)

---

## Phase 1: Trust and UX Consistency Foundation (4-6 weeks)

**Why first:** Users abandon alternatives that feel inconsistent or risky, regardless of feature count.

### 1.1 Interaction consistency across tabs
- Normalize:
  - filtering behavior
  - focus behavior
  - help ordering/wording
  - list/detail spacing and margins
  - overlay behavior and escape semantics
- Add consistency tests for all tabs (shared behavior contracts)

### 1.2 Error handling and recovery UX
- Replace raw operation errors with actionable messages:
  - "what failed"
  - "why likely"
  - "what to do next"
- Add "retry" patterns for transient failures
- Add context-aware daemon/socket diagnostics

### 1.3 Safety rails for destructive actions
- Preflight impact previews for prune/delete
- Better confirmations (resource count + sample names)
- Optional "safe mode" (no destructive actions without explicit opt-in)

### 1.4 Keyboard discoverability
- "Command palette" (`:`) with fuzzy search for actions
- In-context shortcut hints for selected item

**Exit criteria**
- No known cross-tab interaction drift
- destructive action confidence complaints materially reduced
- new-user task completion rate materially improved

---

## Phase 2: Core Desktop Workflow Parity (6-10 weeks)

### 2.1 Integrated logs + exec (no shelling out UX)
- Bring logs and terminal-like exec into native panes
- Add controls:
  - follow/pause
  - clear
  - copy/export
  - simple search in logs

### 2.2 Compose project lifecycle parity
- Project-level actions:
  - up/down/restart
  - rebuild/pull
  - scale service
- Show project topology and container health at service/project level
- Better compose file linkage and diagnostics

### 2.3 Image workflow parity
- Better pull/build feedback and cancellation
- Registry auth/session handling UX
- Image metadata richness:
  - size/layers/history
  - age + usage context
- Safe cleanup recommendations ("this prune will affect...")

### 2.4 Networks/Volumes operational parity
- richer inspect views
- attach/detach and connectivity context
- mount usage and "what depends on this" traces

**Exit criteria**
- top 10 Docker Desktop daily workflows supported end-to-end
- users no longer need frequent shell escape for routine operations

---

## Phase 3: Platform + Ecosystem Capability (8-12 weeks)

### 3.1 Contexts and multi-environment management
- Docker context switching UI
- clear target/environment indicators
- remote engine ergonomics

### 3.2 Cross-platform robustness
- Linux/macOS/Windows parity pass
- socket/path/clipboard/shell behavior hardening
- rootless Docker support maturity

### 3.3 Optional Kubernetes bridge (scoped)
- Decide explicit position:
  - either "not a k8s product"
  - or minimal k8s context visibility/workflows
- Avoid half-baked k8s UX

### 3.4 Plugin/extension architecture (lightweight)
- external commands/actions registry
- safe integration hooks for org-specific workflows

**Exit criteria**
- "works reliably on my platform/context" becomes common user sentiment
- external teams can integrate without forking core

---

## Phase 4: Differentiation (6-8 weeks, ongoing)

**Win where desktop GUIs are weaker:**

- Power-user command graph:
  - chain actions (inspect -> logs -> exec)
  - repeatable macro-like workflows
- Keyboard-first speed:
  - jump actions
  - saved views/queries
  - multi-resource batch ops with previews
- Low-latency remote workflows:
  - optimized for SSH/devbox/cloud shells
- Scriptability:
  - export operation plan
  - reproducible action recipes

**Exit criteria**
- users choose it for speed/flow, not only because it is "free/open"

---

## Cross-Cutting Tracks (all phases)

- **Testing and quality**
  - contract tests for tab consistency
  - integration tests for lifecycle operations
  - golden/snapshot tests for key panels
- **Documentation and onboarding**
  - "first 10 minutes" guide
  - migration guide from Docker Desktop
  - troubleshooting cookbook
- **Performance**
  - responsiveness SLAs
  - large-list rendering and update throttling
- **Accessibility**
  - no-nerd-font complete parity
  - high-contrast and reduced-visual-noise modes

---

## Potential user criticisms to proactively address

- "I still need shell commands for normal tasks."
- "It feels inconsistent depending on tab."
- "I don't trust destructive actions."
- "Errors tell me what failed, not how to recover."
- "Compose support is basic compared to what I expect."
- "Works on Linux best, rough elsewhere."

---

## Recommended sequencing (pragmatic)

1. **Phase 1 first** (trust/consistency)
2. **Phase 2.1 + 2.2 next** (logs/exec + compose lifecycle)
3. **Phase 2.3** (image maturity)
4. **Phase 3 platform hardening**
5. **Phase 4 differentiation**

This order maximizes perceived product quality early while building parity in the workflows users judge most.
