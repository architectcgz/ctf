# 2026-06-02 CSP Security Headers Review

- Review target:
  - Repository: `ctf`
  - Branch: `main`
  - Diff source: local working tree (`code/frontend/nginx/default.conf`, `code/backend/internal/app/router.go`, `code/backend/internal/middleware/security_headers.go`)
  - Files reviewed:
    - `code/frontend/nginx/default.conf`
    - `code/backend/internal/app/router.go`
    - `code/backend/internal/middleware/security_headers.go`
    - `code/frontend/index.html`
    - `code/frontend/src/shared/model/realtime/useWebSocket.ts`
    - `code/frontend/src/pages/utility/UILabRoutePage.vue`
- Classification check: agree with non-trivial security review
- Gate verdict: blocked
- Dominant scope: frontend CSP containment quality

## Findings

### 1. Blocker: CSP still leaves arbitrary outbound channels open, so the new policy does not deliver the intended XSS containment

- Location:
  - `code/frontend/nginx/default.conf:10`
  - `code/frontend/src/shared/model/realtime/useWebSocket.ts:31-46`
  - `code/frontend/src/pages/utility/UILabRoutePage.vue:160-163`
- Risk:
  - The new CSP blocks inline script execution, but it still allows an injected script to open arbitrary `ws:` / `wss:` connections and arbitrary `https:` image beacons.
  - That means the policy does not meaningfully constrain outbound exfiltration once script execution is obtained, which is the main containment value CSP is supposed to add after an XSS bug slips through.
- Evidence:
  - Runtime WebSocket code only uses same-origin `/ws/...` and derives the final URL from `window.location.host`.
  - The only reviewed external image origin is `https://api.dicebear.com/...` in the UI lab page.
  - The current policy grants much broader permissions than the application behavior requires:
    - `connect-src 'self' ws: wss:`
    - `img-src 'self' data: https:`
- Required fix:
  - Tighten `connect-src` to same-origin plus any explicitly required websocket origin, instead of all `ws:` / `wss:`.
  - Tighten `img-src` to the concrete external host that is actually needed, or remove the external dependency.
- Required re-validation:
  - Rebuild frontend assets.
  - Open the app in a browser and verify:
    - main shell loads without CSP violations
    - notification / scoreboard websocket connections still succeed
    - the UI lab avatar still renders if that page remains supported

## Material findings

- Finding 1 must be fixed before considering this CSP hardening merge-ready.

## Senior implementation assessment

- Adding CSP at the Nginx entrypoint is the right ownership choice for document responses.
- Adding minimal defensive headers on API responses is acceptable as a secondary layer.
- The current problem is not placement but policy width: the implementation is close, but the allowlist is broader than the application behavior needs, so the security value is weaker than intended.

## Required re-validation

- `cd code/frontend && pnpm build`
- Browser validation against the served app, including websocket paths

## Residual risk

- I did not run a browser session against the served Nginx container, so final CSP enforcement was inferred from the built artifact and current source usage.
- I did not audit every possible user-provided URL that may open in a new tab; this review only covered the new CSP/security-header diff and its immediate runtime dependencies.

## Touched known-debt status

- This diff closes the previously tracked “missing CSP header” todo only partially.
- The touched surface now has a CSP policy, but the containment policy is still too broad, so the security debt should remain open until the allowlist is tightened and re-validated.
