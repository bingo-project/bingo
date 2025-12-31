# AI Reliability & Stability Fix Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Improve the reliability of the AI Chat module by implementing a retry mechanism for transient errors (P0) and fixing a potential goroutine leak in stream handling (P1).

**Architecture:**
- **Retry Logic:** A generic `retry` package in `pkg/ai/retry.go` that handles exponential backoff and error classification.
- **Provider Update:** Integrate the retry logic into `Chat` methods of all providers (OpenAI, Claude, Gemini, Qwen) and add context cancellation checks to their `ChatStream` goroutines.

**Tech Stack:** Go, standard library `context`, `time`.

### Task 1: Implement Retry Logic

**Files:**
- Create: `pkg/ai/retry.go`
- Test: `pkg/ai/retry_test.go`

**Step 1: Write the failing test for Retry**

Create `pkg/ai/retry_test.go` with a test case that simulates a transient error (e.g., 503) succeeds after retries, and a permanent error fails immediately.

**Step 2: Run test to verify it fails**

Run: `go test -v pkg/ai/retry_test.go`
Expected: FAIL (compilation error or undefined)

**Step 3: Write implementation**

Create `pkg/ai/retry.go` with:
- `RetryConfig` struct.
- `Do` function with exponential backoff.
- `isRetriable` helper function.

**Step 4: Run test to verify it passes**

Run: `go test -v pkg/ai/retry.go pkg/ai/retry_test.go`
Expected: PASS

**Step 5: Commit**

```bash
git add pkg/ai/retry.go pkg/ai/retry_test.go
git commit -m "feat(ai): implement retry mechanism for transient errors"
```

### Task 2: Fix OpenAI Provider (Retry + Context Leak)

**Files:**
- Modify: `pkg/ai/providers/openai/provider.go`

**Step 1: Apply changes**

- In `Chat`: Wrap `p.client.Generate` with `retry.Do` using default config.
- In `ChatStream`: Add `select { case <-ctx.Done(): ... }` inside the streaming goroutine.

**Step 2: Verify Compilation**

Run: `go build ./pkg/ai/providers/openai/...`

**Step 3: Commit**

```bash
git add pkg/ai/providers/openai/provider.go
git commit -m "fix(ai/openai): add retry mechanism and fix stream context leak"
```

### Task 3: Fix Claude Provider (Retry + Context Leak)

**Files:**
- Modify: `pkg/ai/providers/claude/provider.go`

**Step 1: Apply changes**

- In `Chat`: Wrap client call with `retry.Do`.
- In `ChatStream`: Add context cancellation check.

**Step 2: Verify Compilation**

Run: `go build ./pkg/ai/providers/claude/...`

**Step 3: Commit**

```bash
git add pkg/ai/providers/claude/provider.go
git commit -m "fix(ai/claude): add retry mechanism and fix stream context leak"
```

### Task 4: Fix Gemini Provider (Retry + Context Leak)

**Files:**
- Modify: `pkg/ai/providers/gemini/provider.go`

**Step 1: Apply changes**

- In `Chat`: Wrap client call with `retry.Do`.
- In `ChatStream`: Add context cancellation check.

**Step 2: Verify Compilation**

Run: `go build ./pkg/ai/providers/gemini/...`

**Step 3: Commit**

```bash
git add pkg/ai/providers/gemini/provider.go
git commit -m "fix(ai/gemini): add retry mechanism and fix stream context leak"
```

### Task 5: Fix Qwen Provider (Retry + Context Leak)

**Files:**
- Modify: `pkg/ai/providers/qwen/provider.go`

**Step 1: Apply changes**

- In `Chat`: Wrap client call with `retry.Do`.
- In `ChatStream`: Add context cancellation check.

**Step 2: Verify Compilation**

Run: `go build ./pkg/ai/providers/qwen/...`

**Step 3: Commit**

```bash
git add pkg/ai/providers/qwen/provider.go
git commit -m "fix(ai/qwen): add retry mechanism and fix stream context leak"
```

### Task 6: Final Verification

**Files:**
- None (Running suite)

**Step 1: Run all tests**

Run: `make test`

**Step 2: Lint check**

Run: `make lint`

