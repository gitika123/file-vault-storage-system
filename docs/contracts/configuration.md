# Configuration Contract

All runtime configuration is supplied through environment variables. Local development uses `.env.example` as the documented template; secrets must never be committed.

| Variable | Required | Default | Meaning |
|---|---:|---:|---|
| `APP_ENV` | no | `development` | Runtime environment. |
| `HTTP_ADDR` | no | `:8080` | Go HTTP bind address. |
| `DATABASE_URL` | yes | - | PostgreSQL connection string. |
| `REDIS_URL` | yes in production | - | Redis connection string for distributed limits. |
| `FRONTEND_ORIGIN` | yes | `http://localhost:5173` | Allowed browser origin. |
| `SESSION_SECRET` | yes | - | Secret for signing/encrypting session material; minimum 32 bytes. |
| `CSRF_SECRET` | yes | - | Secret for CSRF token generation; minimum 32 bytes. |
| `COOKIE_SECURE` | no | `false` | Must be `true` behind HTTPS. |
| `API_RATE_PER_SECOND` | no | `2` | Per-user steady-state API rate. |
| `API_RATE_BURST` | no | `4` | Per-user burst allowance. |
| `USER_QUOTA_BYTES` | no | `10485760` | Default logical quota: 10 MiB. |
| `MAX_FILE_BYTES` | no | `10485760` | Maximum individual file size. |
| `MAX_UPLOAD_BYTES` | no | `52428800` | Maximum aggregate multipart request size. |
| `BLOB_STORE_DRIVER` | no | `local` | `local` for Docker volume; `s3` is a future adapter. |
| `BLOB_STORE_PATH` | no | `/var/lib/balkanid/blobs` | Local blob root. |
| `PUBLIC_SHARE_BASE_URL` | no | `http://localhost:5173/share` | Browser base URL for generated public links. |
| `LOG_LEVEL` | no | `info` | Structured log level. |

The application must fail fast on missing required production configuration and must redact secret values from logs and diagnostics.
