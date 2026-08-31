# REST Binary Endpoint Contract

The shipped API is REST-based. The repository also contains a GraphQL SDL draft because GraphQL was preferred in the brief, but REST was explicitly accepted and is the implemented runtime boundary. REST is used for metadata/actions as well as multipart and binary operations.

| Method | Path | Auth | Purpose |
|---|---|---|---|
| `POST` | `/api/uploads` | Session | Multipart single/multi-file upload; returns per-file result objects. |
| `GET` | `/api/files/{id}/content` | Session/share | Authorized download stream; increments count after stream begins. |
| `GET` | `/api/files/{id}/preview` | Session/share | Safe preview for allowlisted image/PDF types. |
| `GET` | `/api/files/{id}` | Session | Owner-scoped detailed metadata. |
| `GET` | `/api/files` | Session | Owner-scoped list with filename, MIME, size, date, tag, uploader, folder, and cursor filters. |
| `GET` | `/api/folders` | Session | List the owner's child folders for a parent folder. |
| `POST` | `/api/folders` | Session | Create an owner-scoped folder. |
| `PATCH` | `/api/folders/{id}` | Session | Rename an owner-scoped folder. |
| `DELETE` | `/api/folders/{id}` | Session | Delete an empty owner-scoped folder; non-empty folders return `FOLDER_NOT_EMPTY`. |
| `PATCH` | `/api/files/{id}/folder` | Session | Move an owned file to an owned folder or root. |
| `GET` | `/api/admin/files` | Admin session | All-file inventory with uploader details and download counts. |
| `GET` | `/api/admin/stats` | Admin session | Aggregate users, files, logical/physical bytes, and downloads. |
| `GET` | `/public/{token}` | Public token | Public file preview/download landing page or shared-folder listing. |
| `GET` | `/public/{token}/download` | Public token | Public file download stream; `?preview=1` requests inline preview. Folder shares accept `?fileId=...` for an authorized child file. |
| `GET` | `/health/live` | None | Process liveness. |
| `GET` | `/health/ready` | None | Database/Redis/blob-store readiness. |

Upload responses preserve partial success:

```json
{
  "results": [
    { "filename": "report.pdf", "fileId": "...", "status": "created", "deduplicated": false },
    { "filename": "bad.docx", "status": "rejected", "error": { "code": "INVALID_MIME" } }
  ]
}
```
