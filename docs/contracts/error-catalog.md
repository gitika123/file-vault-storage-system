# API Error Contract

Errors are stable domain errors, not database or filesystem messages. GraphQL responses use `errors[].extensions.code`; REST responses use the same `code` in JSON.

| Code | HTTP | Meaning |
|---|---:|---|
| `UNAUTHENTICATED` | 401 | No valid session. |
| `FORBIDDEN` | 403 | Session exists but lacks permission. |
| `NOT_FOUND` | 404 | Resource is missing or intentionally hidden by authorization policy. |
| `INVALID_INPUT` | 400 | Input fails structural validation. |
| `INVALID_MIME` | 415 | Content does not match its declared type/policy. |
| `FILE_TOO_LARGE` | 413 | File/request exceeds configured size limit. |
| `QUOTA_EXCEEDED` | 413 | Logical user quota would be exceeded. |
| `RATE_LIMITED` | 429 | User or endpoint rate limit exceeded; include `Retry-After`. |
| `CONFLICT` | 409 | Operation conflicts with current state. |
| `DEPENDENCY_UNAVAILABLE` | 503 | Database, Redis, or blob store is unavailable. |
| `INTERNAL` | 500 | Unexpected server failure; expose only a request ID. |

Example REST error:

```json
{
  "error": {
    "code": "QUOTA_EXCEEDED",
    "message": "Upload would exceed your 10 MiB logical quota.",
    "requestId": "req_01...",
    "details": { "currentBytes": 8388608, "requestedBytes": 3145728, "limitBytes": 10485760 }
  }
}
```

Messages are safe for users. Internal causes are logged with a request ID and are never returned directly.
