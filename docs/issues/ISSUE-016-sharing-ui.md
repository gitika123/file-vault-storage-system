# Issue 016 - Sharing, public-link, preview, and delete UX

**Status:** CLOSED

The dashboard now exposes public-link creation, preview, download, and delete actions on each file row, plus a details/share-link panel. The API capabilities are consumed through the authenticated frontend client.

## Verification evidence

`npm run build` passes strict TypeScript and Vite compilation. The corresponding backend public-share, private-download, preview, revoke, and authorization flows were verified live in Issues 011 and 012.
