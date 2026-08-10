/**
 * A diff is read when a task is opened, not carried on its report, so the mock
 * keeps the changes here rather than in the session fixtures. A task the mock
 * sessions name gets the diff written for it; every other task id — the ones a
 * simulated run invents — gets a deterministic slice of the same changes, the
 * way mockOutcome derives a run from a task id.
 */

import type {FileChange} from '@/features/sessions/types'

const CHANGES: Record<string, FileChange> = {
    proxyDeadline: {
        path: 'internal/implementation/mcp_proxy/v1.go',
        oldPath: '',
        changeType: 'modified',
        additions: 11,
        deletions: 4,
        unifiedDiff: `--- a/internal/implementation/mcp_proxy/v1.go
+++ b/internal/implementation/mcp_proxy/v1.go
@@ -38,10 +38,17 @@ func (p *proxyV1) Forward(req *output_itf.MCPRequest) (*output_itf.MCPResponse,
     if req == nil {
         return nil, custom_error.Critical("forward called with no request")
     }

-    resp, err := p.client.Do(req.HTTP())
-    if err != nil {
-        return nil, err
-    }
+    ctx, cancel := context.WithTimeout(p.ctx, p.deadline)
+    defer cancel()
+
+    resp, err := p.client.Do(req.HTTP().WithContext(ctx))
+    if err != nil {
+        if errors.Is(err, context.DeadlineExceeded) {
+            return nil, custom_error.TypedCritical(enums.ErrorTypeUpstream, err)
+        }
+        return nil, custom_error.Critical(err.Error())
+    }

     return output_itf.ReadMCPResponse(resp)
 }`,
    },
    proxyConfig: {
        path: 'internal/implementation/config/v1.go',
        oldPath: '',
        changeType: 'modified',
        additions: 6,
        deletions: 0,
        unifiedDiff: `--- a/internal/implementation/config/v1.go
+++ b/internal/implementation/config/v1.go
@@ -22,6 +22,12 @@ type Config struct {
     Agents   AgentsConfig   \`mapstructure:"agents"\`
     Database DatabaseConfig \`mapstructure:"database"\`
+    MCPProxy MCPProxyConfig \`mapstructure:"mcp_proxy"\`
 }
+
+type MCPProxyConfig struct {
+    ForwardDeadline time.Duration \`mapstructure:"forward_deadline"\`
+}`,
    },
    proxyYaml: {
        path: 'config.yaml',
        oldPath: '',
        changeType: 'modified',
        additions: 3,
        deletions: 0,
        unifiedDiff: `--- a/config.yaml
+++ b/config.yaml
@@ -9,3 +9,6 @@ database:
   path: ./harness.db

+mcp_proxy:
+  forward_deadline: 20s
+`,
    },
    proxyDocs: {
        path: 'README.md',
        oldPath: '',
        changeType: 'modified',
        additions: 2,
        deletions: 1,
        unifiedDiff: `--- a/README.md
+++ b/README.md
@@ -74,7 +74,8 @@ The proxy sits between the agent and every MCP server it may reach.
-Forwarded calls inherit the agent's lifetime, so a hung server hangs the node.
+Forwarded calls carry a deadline from \`mcp_proxy.forward_deadline\`, so a hung
+server fails one call instead of stalling the node that made it.`,
    },
    walReplay: {
        path: 'internal/implementation/wal/replay.go',
        oldPath: '',
        changeType: 'modified',
        additions: 9,
        deletions: 1,
        unifiedDiff: `--- a/internal/implementation/wal/replay.go
+++ b/internal/implementation/wal/replay.go
@@ -55,7 +55,15 @@ func (w *walV1) Replay(from uint64) ([]*output_itf.Record, error) {
     for offset < size {
         record, err := decode(chunk)
         if err != nil {
-            return nil, err
+            if errors.Is(err, errTornRecord) {
+                w.log.Warn("torn record, stopping replay", "offset", offset)
+                break
+            }
+
+            return nil, custom_error.TypedCritical(enums.ErrorTypeCorrupt, err)
         }

         records = append(records, record)`,
    },
    walProbe: {
        path: 'internal/implementation/wal/torn_record.go',
        oldPath: '',
        changeType: 'created',
        additions: 24,
        deletions: 0,
        unifiedDiff: `--- /dev/null
+++ b/internal/implementation/wal/torn_record.go
@@ -0,0 +1,24 @@
+package wal
+
+import "errors"
+
+var errTornRecord = errors.New("record ends before its declared length")
+
+// a torn record is the last write the process did not finish, so it can only
+// ever be the tail of the log
+func isTorn(chunk []byte, declared int) bool {
+    return len(chunk) < declared
+}`,
    },
    sessionRename: {
        path: 'internal/implementation/session_manager/v1.go',
        oldPath: 'internal/implementation/task_manager/v1.go',
        changeType: 'renamed',
        additions: 5,
        deletions: 5,
        unifiedDiff: `--- a/internal/implementation/task_manager/v1.go
+++ b/internal/implementation/session_manager/v1.go
@@ -1,10 +1,10 @@
-package task_manager
+package session_manager

 import (
     "hexago/internal/interface/core"
 )

-type taskManagerV1 struct {
+type sessionManagerV1 struct {
     retry RetryPolicy
 }`,
    },
    retryTable: {
        path: 'internal/implementation/session_manager/retry.go',
        oldPath: '',
        changeType: 'created',
        additions: 18,
        deletions: 0,
        unifiedDiff: `--- /dev/null
+++ b/internal/implementation/session_manager/retry.go
@@ -0,0 +1,18 @@
+package session_manager
+
+import "time"
+
+var backoff = []time.Duration{
+    2 * time.Second,
+    8 * time.Second,
+    30 * time.Second,
+}
+
+func (s *sessionManagerV1) waitBefore(attempt int) time.Duration {
+    if attempt >= len(backoff) {
+        return backoff[len(backoff)-1]
+    }
+
+    return backoff[attempt]
+}`,
    },
}

const DIFF_BY_TASK: Record<string, FileChange[]> = {
    'task-6': [],
    'task-7': [CHANGES.proxyDeadline, CHANGES.proxyConfig, CHANGES.proxyYaml],
    'task-8': [CHANGES.proxyDocs],
    'task-9': [CHANGES.walProbe, CHANGES.walReplay],
    'task-10': [CHANGES.walProbe],
    'task-12': [CHANGES.sessionRename],
    'task-16': [],
    'task-19': [CHANGES.proxyDeadline, CHANGES.proxyConfig],
}

const GENERATED = [
    CHANGES.proxyDeadline,
    CHANGES.retryTable,
    CHANGES.walReplay,
    CHANGES.proxyDocs,
    CHANGES.sessionRename,
]

function seedOf(id: string) {
    let seed = 0
    for (let index = 0; index < id.length; index += 1)
        seed = (seed * 31 + id.charCodeAt(index)) % 99991
    return seed
}

export function mockTaskDiff(taskId: string): FileChange[] {
    const named = DIFF_BY_TASK[taskId]
    if (named) return structuredClone(named)

    const seed = seedOf(taskId)
    const first = seed % GENERATED.length
    const count = 1 + (seed % 3)

    return Array.from({length: count}, (_, index) =>
        structuredClone(GENERATED[(first + index) % GENERATED.length]),
    )
}
