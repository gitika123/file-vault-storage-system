# Issue 021 - Continuous integration

**Status:** CLOSED

The GitHub Actions workflow runs Go formatting/tests, strict frontend build, Chromium installation, and the Playwright UAT suite on push and pull request. The authenticated UAT case remains environment-gated and is skipped when credentials are not present.
