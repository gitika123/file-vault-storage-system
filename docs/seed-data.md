# Local demo seed data

The seed command creates or updates three local-only accounts. Passwords are supplied through environment variables and are hashed with Argon2id at runtime; no credentials are stored in the repository.

From `backend/`, load the project `.env` values into the current shell, set three dedicated development passwords of at least 12 characters, and run:

```powershell
$env:SEED_ALICE_PASSWORD = "choose-a-local-password"
$env:SEED_BOB_PASSWORD = "choose-another-local-password"
$env:SEED_ADMIN_PASSWORD = "choose-an-admin-password"
go run ./cmd/seed
```

The command is idempotent for the three documented email addresses. It does not delete other users, files, sessions, or database data.
