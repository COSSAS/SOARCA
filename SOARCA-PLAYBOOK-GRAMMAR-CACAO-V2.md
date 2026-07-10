# SOARCA Playbook Grammar (CACAO 2.0)

> Use this when writing CACAO 2.0 playbooks that must execute on SOARCA. It encodes runtime behavior of the SOARCA decomposer/executor (verified against the `development` branch) — not just spec compliance. A playbook that validates against the CACAO schema can still silently misbehave on SOARCA. Follow this grammar to avoid that.

## 1. Mental model

CACAO 2.0 is the format. SOARCA is the executor — and it diverges from the spec in important ways:

- **No parallelism.** `parallel` steps aren't implemented. Build everything sequentially.
- **Abort on failure.** A step that errors aborts the whole branch immediately. There is no continue-on-error.
- **`on_failure` is not followed** by the current decomposer. Define it for forward-compatibility, but don't expect runtime branching on it.
- **No loop, limited transform.** No iteration construct — a loop must happen outside SOARCA or in a shell command on the target. Transform is available for JSON responses via the `soarca-assignment` / jq step extension (§5.2), *if* your build has it (PR #373 and later on `development`). Without the extension, no JSON extraction and no variable-to-variable copy. With it, everything §5.2 covers.
- **Variables are set only by module outputs.** Each capability writes to a fixed variable name (see §5). `out_args` declares scope exit; it does not rename.
- **Sub-playbook variables are passed by name-merging the entire parent scope.** `in_args`/`out_args` on a `playbook-action` step are ignored by the executor today.

Internalize those five points before writing anything.

## 2. Object grammar

Every playbook needs `type: "playbook"`, `spec_version: "cacao-2.0"`, `id`, `name`, `created_by`, `created`, `modified`, `workflow_start`, and `workflow`. Anything that calls APIs/targets also needs `agent_definitions` and `target_definitions`; anything authenticated needs `authentication_info_definitions`.

### 2.1 IDs

Every CACAO object id is `<object-type>--<UUID>`. Step UUIDs **MUST** be RFC 4122 v4. Other objects **SHOULD** be v4. Don't hand-write placeholders like `aaaaaaaa-aaaa-...` — generate them with a real v4 generator. Every cross-reference (`workflow_start`, `on_completion`, `on_true`, `on_false`, `agent`, `targets[]`, `authentication_info`, `playbook_id`) must resolve to an id that exists in the document.

### 2.2 Step types in use

**`on_completion` is required on every non-end step.** It is the normal-flow exit — where execution proceeds after this step is done. Even on conditional steps, `on_completion` must be present and point at a real next step (typically an end). `on_true` and `on_false` are *not* substitutes for `on_completion`; they exist for the steps that are genuinely conditional.

| Type | Required fields | Notes |
|---|---|---|
| `start` | `on_completion` | One per playbook, referenced by `workflow_start`. |
| `end` | — | Multiple are fine. Give each a `name` if their roles differ. |
| `action` | `on_completion`, `commands[]`, `agent`, `targets[]` | The workhorse. Add `out_args` to capture the module's result variable. |
| `if-condition` | `on_completion`, `condition`, `on_true`, `on_false` | All four required. See §7 for what goes where. |
| `playbook-action` | `on_completion`, `playbook_id` | Invokes another stored playbook. See §6. |

`parallel`, `while-condition`, `switch-condition` are spec-defined but not in practical use here — avoid unless you've verified the version of SOARCA you target supports them.

## 3. Variable token syntax

Variables are referenced as `__name__:value` — double underscores either side, then `:value`.

### 3.1 Where interpolation happens — and where it does not

Substitution is a literal string replace performed **only at the step's `commands[]` level**:

- `commands[].command` — interpolated
- `commands[].content` — interpolated
- `commands[].headers` (every header value) — interpolated

**Not interpolated** (the values are taken verbatim and sent as-is):

- `target_definitions.*` — including `address.ipv4[]`, `address.url[]`, `port`, etc.
- `agent_definitions.*`
- `authentication_info_definitions.*` — including `user_id`, `username`, `password`, `token`, `private_key`, etc.

This matters because the decomposer/executor binds the target and auth info into the request object before the per-step interpolation pass runs. A `__var__:value` token sitting in a target IP, a port, a username, a password, or an oauth2 `token` field is sent **literally** to the remote endpoint — which means failed auth, an unreachable host, or a 401 with no clear cause.

Consequences for design:

- **Host IPs in targets must be hardcoded** in the target definition. If you need the host to vary, the playbook itself has to vary — either store one playbook per host, or carry the IP only inside the `command`/`content`/header and target something that proxies to the right host.
- **Credentials in `authentication_info_definitions` must be hardcoded** in the playbook. There is no way to inject them via variables. Practical handling: store secrets in the database playbook record (acceptable for lab/internal SOARCA), use a secrets manager that injects on playbook load, or accept that each credential rotation requires a playbook update.
- **Tokens for `oauth2` cannot be variabilized** the way you'd expect. The token field is read verbatim. For short-lived tokens (Graph, etc.), the playbook in storage either has a fresh token written in before invocation, or the call has to be modeled differently — e.g. pass the token in a header (which *is* interpolated) and use `oauth2` with an empty/dummy token, or use no auth-info block and assemble the `Authorization: Bearer …` header yourself from a variable.

### 3.2 Interpolation semantics (within the supported sites)

Substitution is purely a literal string replace — **no JSON awareness**. A variable interpolated into a JSON body is inserted raw:

- `"priority":__pri__:value` produces `"priority":7` when `__pri__` is `7` (correct for numbers/booleans).
- `"priority":"__pri__:value"` produces `"priority":"7"` (a string — often wrong; e.g. Gotify, Graph `accountEnabled`).
- A string variable carrying a `"` or newline lands unescaped and breaks the JSON. Escape upstream or use `command_b64`/`content_b64`.

Variable resolution: a declared-but-unbound `external: true` variable interpolates to the empty string. An undeclared token is left as the literal `__name__:value`. Always declare what you reference.

## 4. Variable declaration

```json
"__name__": {
  "type": "string",
  "description": "What this is for; required for handoff.",
  "value": "",
  "constant": false,
  "external": true
}
```

`external: true` means "the caller binds the value" (trigger payload, or parent playbook via name-merge). `external: false` means "internal to this playbook" — use it for module-output sinks like `__soarca_http_api_result__`. Don't set `external: true` on a module-output variable; that is the inverse of its meaning.

`constant: true` is documentation; it doesn't lock the value in SOARCA, but it tells humans not to mutate.

### 4.1 Types that matter for conditions

The condition comparator dispatches on the variable's declared `type`:

- `string` → **lexicographic** compare. `"10" < "3"` is true. Never use for numeric thresholds.
- `integer`, `long` → `strconv.Atoi` then numeric compare.
- `float` → numeric.
- `bool`, `ipv4-addr`, `mac-addr`, etc. → their own paths.

If you need `> 5` to mean what it says, the variable must be typed `integer`. The most common bug in field playbooks is doing severity comparisons on a string-typed variable; it works for single digits and silently inverts at two.

## 5. Module → output variable mapping

Each SOARCA capability writes its result into a fixed variable name. Declare that variable (typically `external: false`) and list it in the step's `out_args` to expose it to later steps.

| Module (agent `name`) | Result variable(s) | Target type | Auth types accepted |
|---|---|---|---|
| `soarca-http-api` | `__soarca_http_api_result__` | `http-api` | `http-basic` (user_id, password), `oauth2` (static `token`) |
| `soarca-ssh` | `__soarca_ssh_result__` | `ssh` | `user-auth` (username, password), `private-key` (with passphrase) |
| `soarca-powershell` | `__soarca_powershell_result__`, `__soarca_powershell_error__` | `net-address` (port 5985 = WinRM) | `user-auth` |
| `soarca-manual` | name in step's `out_args` | `individual` | — |
| `soarca-openc2` | `__soarca_openc2_result__` | `http-api` | as http-api |
| `soarca-fin` | dependent on Fin module | varies | — |

The agent block is always `type: "soarca"`, `name: "soarca-<module>"`.

### 5.1 `oauth2` is static — and not interpolated

SOARCA's `oauth2` auth type does not perform a token exchange. It attaches the literal `token` field as `Authorization: Bearer <token>`. Worse for automation: the `token` field is part of `authentication_info_definitions` (see §3.1), so it is **not interpolated** — a `__var__:value` token written there is sent literally as the bearer.

For APIs that need OAuth2 client-credentials (Microsoft Graph, Azure, Google), this means:

- A token cannot be passed in via a `__graph_token__:value` placeholder in the `token` field.
- The token has to either be written into the stored playbook (then rewritten on rotation), or — better — handled by skipping the `oauth2` auth-info entirely and assembling the `Authorization` header yourself: `"Authorization": ["Bearer __token__:value"]`. Headers *are* interpolated, so the token can be a real variable bound at trigger time.

SOARCA cannot fetch a token by itself (no HTTP-response transform without §5.2). With §5.2, a token *can* be extracted from a login response and interpolated into the next request's `Authorization` header — because header values are interpolated. That workaround unblocks real client-credentials flows.

### 5.2 Transforming step outputs (`soarca-assignment` / jq step extension)

CACAO's `step_extensions` map on each step can carry `soarca-assignment` extensions. After the step's commands complete, each assignment reads a named step-result variable, optionally runs it through an expression engine, and inserts/replaces a variable in scope with the result. The new variable merges into the step's outputs and propagates like any other module output.

**Availability.** Introduced by PR #373 on the `development` branch; not necessarily in every build. If your SOARCA doesn't have the extension code compiled in, `step_extensions` on a step are silently ignored — no error, no assignment, destination variables stay at their default. See §11 for the diagnostic (passthrough assignment) that tells you which side is failing.

**Shape:**

```json
"action--…": {
  ...
  "commands": [ ... ],
  "step_extensions": {
    "extension-definition--<uuid>": {
      "type": "soarca-assignment",
      "step-result": "__soarca_http_api_result__",
      "variable": "__extracted_field__",
      "expression": {"type": "jq", "expression": ".data.someField"}
    }
  }
}
```

Multiple assignments per step are fine — each map entry produces one variable, evaluated independently against the same step-result.

**Passthrough (no expression) copies raw.** Omit the `expression` block entirely and the assignment stores the step-result value as-is into the destination variable. Primary use: diagnostics.

**Constraints:**

- **jq is the only engine currently.** No regex, no XPath, no other transformers.
- **jq requires valid JSON input.** The engine `json.Valid`-checks the source and rejects otherwise. Plain-text or XML responses aren't extractable this way. For non-JSON, do the transform on the target (shell, PowerShell) and consume the plain result directly.
- **jq input must be a JSON *object* at the top level.** The engine unmarshals into `map[string]any{}`. A response that's a top-level array (`[...]`) fails; wrap the query at the source (`?fields=data` style) or work around by having jq run against an object-wrapped response.
- **Results are always string-typed.** The assignment hardcodes `VariableTypeString` regardless of what jq returned. A path that yields the JSON number `100` lands as the string `"100"`. Downstream numeric comparison on that variable is lexicographic — see §4.1. If you need numeric behavior, extract, then re-declare a fresh integer-typed variable in `playbook_variables` and set its value manually, or accept the string comparison.
- **Only the response body is exposed** as `__soarca_http_api_result__`. HTTP headers and status code aren't separately available for jq to read. (This is called out in issue #370 as a future improvement; PR #373 does not restructure HTTP outputs.)
- **100ms timeout per jq evaluation.** Path extractions are microseconds; complex filters (recursive descent, mapping over large arrays) can trip this.
- **Silent failure on jq errors.** Invalid query, invalid JSON input, or timeout: the extension logs a warning and skips the assignment. The destination variable retains whatever it had before (usually its declared default `""`).
- **jq `null` → the literal string `"null"`.** A missing path returns null, which the assignment stringifies. If you want cleaner handling, use jq's default operator at the source: `.data.field // "unknown"`.

**Chaining pattern — extract in one step, interpolate in the next.** Since assigned variables propagate through scope like any other, they're available in the *next* step's URL, body, or headers:

```json
"step_extensions": {
  "extension-definition--…": {
    "type": "soarca-assignment",
    "step-result": "__soarca_http_api_result__",
    "variable": "__lat__",
    "expression": {"type": "jq", "expression": ".results[0].latitude"}
  }
}
```

Then in the following step: `"command": "GET /v1/forecast?latitude=__lat__:value&… HTTP/1.1"`. This is how a geocoding-then-weather flow, a token-then-API flow, or a search-then-delete-by-uuid flow gets wired inside SOARCA — no external glue.

## 6. Sub-playbook calls (`playbook-action`)

When step A is a `playbook-action` pointing at child playbook B:

1. SOARCA takes A's parent scope (all `playbook_variables`) and **merges by name** into B's scope. Same-name variables in B are **overwritten** by the parent's value.
2. `in_args` and `out_args` on the `playbook-action` step are not read by the executor. Include them as documentation, not mechanism.
3. After B runs, B's final variable values are merged back into the parent scope, again by name.

Consequences:

- **No rename.** If two children both declare `__user__`, the parent has one `__user__` value that goes to both. To pass different values you must rename the variable in one of the children.
- **Accidental override.** A parent variable with the same name as a child's hardcoded value will silently replace it. Name parent variables specifically when this would be surprising.
- **"Without arguments" still merges by name.** A child with no overlapping variable names is what makes a call truly argument-less. Check variable names against the child before assuming.

## 7. Conditions

Condition syntax is strict. Expression must split into **exactly three space-separated parts**:

```
__var__:value <op> <literal>
```

- Variable on the **left**. SOARCA only interpolates the left side. `5 < __var__:value` makes the engine try to resolve `5` as a variable, finds nothing, errors.
- Literal on the right with **no quoting needed** (and no spaces — multi-word literals break the 3-part split).
- Operators: `=`, `!=`, `>`, `>=`, `<`, `<=`.
- Don't omit spaces: `__var__:value>=3` fails parse.

### 7.1 Modeling `on_completion` vs `on_true` / `on_false`

An `if-condition` step has all three pointers, with distinct roles:

- **`on_completion`** is the normal-flow exit. It is where execution continues after the conditional region is done — the follow-up that happens *regardless* of which branch was taken.
- **`on_true` / `on_false`** hold steps that are *conditional*: they run only when the condition evaluates to true (resp. false), and never otherwise.

**Two invariants that follow, and are worth enforcing in your generator:**

1. **Every branch terminates in its own end step.** Do not share an end between the two branches, and do not share an end with `on_completion`. Distinct ends per branch keep the flow traceable and stop conditions from accidentally reusing steps that shouldn't apply.
2. **`on_false` and `on_completion` must not point at the same step.** Same rule for `on_true` and `on_completion`. If a step should run regardless of the outcome, put it on `on_completion` and leave the branch pointing at its own end. If a step should run only on one branch, put it on that branch and never on `on_completion`. Never both.

**Two shapes cover most real conditionals:**

*Shape A — one branch does extra work, then flow continues:*

```
if-condition --on_true--> conditional step -> shared continuation
             --on_false--> end
             --on_completion--> shared continuation
```

Example: "Signs of exploitation? On true, Report Incident; either way, continue to Investigation." Report Incident is the conditional step on `on_true`, its own `on_completion` rejoins Investigation; `on_false` terminates in its own end; the `if-condition`'s `on_completion` also points at Investigation. Investigation has two incoming pointers — allowed and idiomatic.

*Shape B — each branch does different work, nothing shared afterwards:*

```
if-condition --on_true--> conditional step A -> end A
             --on_false--> conditional step B -> end B
             --on_completion--> end C   (formally required; unreachable at runtime)
```

Example: "Patch available? On true, ask about temporary mitigation; on false, run mitigation directly." Each branch does its own work and terminates. `on_completion` gets its own third end that never fires at runtime under the current decomposer, but is required structurally.

The `on_completion`-on-a-condition being sometimes unreachable is a known quirk. Include it anyway: the field is required, and its target documents intent.

## 8. Remote trigger via REST

To invoke a playbook on a different SOARCA instance, do not use `playbook-action` (it only resolves locally stored playbooks). Use http-api against the remote SOARCA's trigger endpoint:

```
POST /trigger/playbook/{playbook-id}
Content-Type: application/json

{
  "__var_name__": {"type": "...", "value": "..."},
  ...
}
```

The handler enforces three rules on the body. Any failure rejects the trigger:

1. Each variable name **must already exist** in the remote playbook.
2. Each variable's `type` must **match** the remote playbook's declared type exactly.
3. Each corresponding remote variable must be `external: true`.

The trigger response contains an execution id, useful for follow-up reporter GETs. Polling for completion is itself a separate http-api step with the same no-transform limitation on parsing the response body.

## 9. Authentication info field names

These differ between auth types and are a common silent-failure spot. Get them right:

- `http-basic`: `user_id`, `password`
- `user-auth`: `username`, `password`
- `oauth2`: `token` (static; see §5.1)
- `private-key`: `private_key`, `passphrase`, `username`

**None of these fields are interpolated** (see §3.1). Values must be hardcoded into the playbook at the auth-info level. Practical patterns:

- For lab/internal use, accept the credentials sitting in the playbook record.
- For rotating secrets, manage the playbook as code and redeploy on rotation.
- To get variable-driven auth, bypass `authentication_info_definitions` for that step and put the credential in a header instead — `"Authorization": ["Basic __b64creds__:value"]` or `"Authorization": ["Bearer __token__:value"]`. Header values *are* interpolated. This works for http-api; ssh/powershell have no equivalent escape.

Do not write `__var__:value` tokens into `username`, `password`, `user_id`, `token`, or any other auth-info field expecting them to resolve.

## 10. Output discipline

- **Generate v4 UUIDs.** Don't pattern them.
- **Validate JSON** before delivering.
- **Verify every cross-reference resolves.** Walk `on_completion`/`on_true`/`on_false`/`agent`/`targets[]`/`authentication_info`/`playbook_id` and confirm each id is defined.
- **Declare every variable you reference.** Including module-output sinks (set `external: false` on those).
- **Type variables used in conditions correctly.** Numeric thresholds → `integer` or `float`.
- **Numeric/boolean JSON body values are interpolated unquoted.** String values are quoted.
- **PowerShell paths:** `-LiteralPath "..."` to defeat wildcard parsing and tolerate spaces.
- **No secrets in committed playbooks.** Use variables, bind at trigger time or via KMS.

## 11. Anti-patterns observed in the wild

These are real bugs from field playbooks; do not reproduce them.

1. **Condition reads `__impact__` while the lookup writes `__soarca_http_api_result__`.** The condition silently evaluates a static default. Always read the variable the module actually writes.
2. **Severity threshold compared on a `string`-typed variable.** Works for single digits, inverts at two. Type the variable `integer`.
3. **Putting unconditional follow-up steps on `on_true` or `on_false`.** Anything that should run regardless of the condition's outcome belongs after `on_completion`, not duplicated into each branch. `on_true`/`on_false` are for the steps that are genuinely conditional.
4. **OAuth2 client-credentials assumed.** SOARCA's `oauth2` only attaches a static bearer. Build for token-injection at trigger time.
5. **Capturing `__soarca_powershell_error__` then defining only `on_completion`.** On failure the run aborts and the captured error never propagates — module outputs from failed steps are not merged back.
6. **Add-then-delete pairs that capture no UUID.** OPNsense rule deletes need the UUID, but the add step's response goes into `__soarca_http_api_result__` and is not parsed. Either capture and pass it through the same execution, or do find-and-delete by description over SSH (shell loop on the target).
7. **`__user__` name collision across sub-playbooks.** A parent calling two children that both declare `__user__` drives both with the same value. Rename one variable in the child.
8. **Plaintext credentials inline.** Bearer tokens, API keys, AD admin passwords sitting in the JSON. For values you must inline (auth-info fields, target IPs — see §3.1), accept the inlining and manage at the playbook-record level; for values you can move into a header, do so.
9. **Variable tokens in target/agent/auth-info fields.** `__token__:value` written into an `oauth2` `token`, `__ip__:value` in a target's `ipv4[]`, or `__admin_pw__:value` in a `user-auth` `password` are **sent verbatim** — SOARCA does not interpolate these sites (see §3.1). The request hits the wire with the literal token string, fails auth or hits the wrong host, and the cause looks like a credential or network problem.
10. **`on_false` and `on_completion` on the same if-condition point at the same step.** The two roles are mutually exclusive: `on_completion` is "runs regardless," branch pointers are "runs only in that case." Same-target across the two is a modeling error; even if it happens to work, it's ambiguous about intent and hides bugs. Same rule for `on_true`/`on_completion` and for the two branches sharing an end. Every branch and the on_completion each get their own end step. (§7.1.)
11. **Comparing a jq-extracted value numerically without redeclaring.** The assignment engine hardcodes result type to `string` (§5.2). A jq path returning JSON `100` lands as the string `"100"`. If the next `if-condition` says `__abuse_score__:value > 50`, that's a lexicographic compare — `"100" > "50"` is *false* because `'1' < '5'`. Extract to a string, then declare a fresh integer-typed variable and re-assign; or gate on it via a second jq expression that returns a comparable string.
12. **Assuming step_extensions ran because the playbook didn't error.** If your SOARCA build predates PR #373's executor wiring, `step_extensions` are silently skipped — no warning, no assignment, downstream interpolation shows empty variables. Diagnose with a *passthrough* assignment (a `soarca-assignment` block with no `expression` field): it copies `step-result` into the destination variable verbatim. If the passthrough works and jq doesn't, jq is failing silently (query wrong, response shape unexpected, or non-JSON input). If neither works, the extension isn't in the build. Also check SOARCA logs for `evaluating assignment extension` traces and `jq expression failed` warnings.
13. **jq query written against the success-response shape only.** APIs return different shapes on error — an AbuseIPDB `{"errors":[...]}`, an ip-api `{"status":"fail","message":"..."}`, a Graph `{"error":{"code":"..."}}`. `.data.field` on any of those returns null, which the assignment stringifies to literal `"null"`. Wrap with jq's default operator at the query — `.data.field // "unknown"` — or branch on a preflight check first.

## 12. Minimal templates

### http-api action

```json
"action--…": {
  "type": "action",
  "name": "…",
  "on_completion": "end--…",
  "commands": [{
    "type": "http-api",
    "command": "POST /path HTTP/1.1",
    "content": "{\"key\":\"__var__:value\"}",
    "headers": {"Content-Type": ["application/json"]}
  }],
  "agent": "soarca--…",
  "targets": ["http-api--…"],
  "out_args": ["__soarca_http_api_result__"]
}
```

### ssh action

```json
{
  "type": "action",
  "commands": [{"type": "ssh", "command": "…shell pipeline…"}],
  "agent": "soarca--…",        // {"type":"soarca","name":"soarca-ssh"}
  "targets": ["ssh--…"],       // ssh target with address, port:"22", authentication_info
  "out_args": ["__soarca_ssh_result__"]
}
```

### powershell action

```json
{
  "type": "action",
  "commands": [{"type":"powershell","command":"Disable-LocalUser -Name \"__user__:value\""}],
  "agent": "soarca--…",         // {"type":"soarca","name":"soarca-powershell"}
  "targets": ["net-address--…"],// net-address with ipv4, port:"5985", user-auth
  "out_args": ["__soarca_powershell_result__"]
}
```

### manual approval gate

```json
{
  "type": "action",
  "timeout": 120000,
  "commands": [{"type":"manual","command":"Approve? Respond: confirm"}],
  "agent": "soarca--…",         // {"type":"soarca","name":"soarca-manual"}
  "targets": ["individual--…"], // {"type":"individual","name":"SOC"}
  "out_args": ["__approval__"]  // operator response lands here
}
```

Then condition on `__approval__:value = confirm`, with the variable declared `type: "string"`.

### if-condition

```json
"if-condition--…": {
  "type": "if-condition",
  "name": "…",
  "condition": "__var__:value > 5",
  "on_true": "…",         // conditional-only step(s)
  "on_false": "…",        // conditional-only step(s)
  "on_completion": "end--…" // normal-flow exit, required
}
```

### playbook-action (sub-playbook call)

```json
"playbook-action--…": {
  "type": "playbook-action",
  "name": "…",
  "playbook_id": "playbook--…",
  "on_completion": "…"
}
```

— `in_args` ignored by executor; rely on name-merging instead.

### action with jq step extensions

```json
"action--…": {
  "type": "action",
  "name": "…",
  "on_completion": "…",
  "commands": [{
    "type": "http-api",
    "command": "GET /some/json/endpoint HTTP/1.1",
    "headers": {"Accept": ["application/json"]}
  }],
  "agent": "soarca--…",
  "targets": ["http-api--…"],
  "step_extensions": {
    "extension-definition--<uuid-a>": {
      "type": "soarca-assignment",
      "step-result": "__soarca_http_api_result__",
      "variable": "__field_a__",
      "expression": {"type": "jq", "expression": ".data.fieldA // \"unknown\""}
    },
    "extension-definition--<uuid-b>": {
      "type": "soarca-assignment",
      "step-result": "__soarca_http_api_result__",
      "variable": "__field_b__",
      "expression": {"type": "jq", "expression": ".data.fieldB"}
    }
  }
}
```

Both extracted variables declared in `playbook_variables` with `type: "string"` and `external: false` — they're populated at runtime, not by the caller. Passthrough form omits `expression` entirely.

## 13. Before delivering, check

1. JSON is valid.
2. Every step UUID is v4.
3. Every non-end step has `on_completion`. Including conditional steps.
4. Every cross-reference resolves.
5. Every variable referenced is declared.
6. **No `__var__:value` tokens in target, agent, or auth-info fields.** Only `commands[].command`, `commands[].content`, and header values are interpolated (§3.1).
7. Variables used in conditions are typed for the comparison.
8. JSON bodies interpolate numeric/boolean variables unquoted.
9. **`on_true`, `on_false`, and `on_completion` on every `if-condition` point at three *different* steps.** Each branch has its own end. No shared targets across the three pointers.
10. **If any step uses `step_extensions`:** the SOARCA target build has PR #373's assignment executor. Every destination variable is declared with `type: "string"` and `external: false`. jq queries are wrapped with `// "fallback"` where a missing field would produce misleading `"null"` output. Variables consumed by numeric conditions downstream are re-declared with the right numeric type, not conditioned on the jq output directly.
11. No secrets inline that don't have to be inline.
12. Anti-patterns from §11 are absent.

If any check fails, fix before output.
