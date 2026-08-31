-- BalkanID Secure File Vault
-- Migration 000001: extensions, enums, core metadata, sharing, analytics, and indexes.

BEGIN;

CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS citext;
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TYPE user_role AS ENUM ('user', 'admin');
CREATE TYPE blob_state AS ENUM ('pending', 'ready', 'pending_delete');
CREATE TYPE share_visibility AS ENUM ('private', 'public', 'direct');
CREATE TYPE share_permission AS ENUM ('view', 'download');

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email CITEXT NOT NULL UNIQUE,
    display_name TEXT NOT NULL CHECK (length(trim(display_name)) BETWEEN 1 AND 120),
    password_hash TEXT NOT NULL,
    role user_role NOT NULL DEFAULT 'user',
    quota_bytes BIGINT NOT NULL DEFAULT 10485760 CHECK (quota_bytes >= 0),
    disabled_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE blobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sha256 BYTEA NOT NULL CHECK (octet_length(sha256) = 32),
    size_bytes BIGINT NOT NULL CHECK (size_bytes >= 0),
    detected_mime TEXT NOT NULL CHECK (length(trim(detected_mime)) > 0),
    storage_key TEXT NOT NULL UNIQUE CHECK (length(trim(storage_key)) > 0),
    reference_count BIGINT NOT NULL DEFAULT 0 CHECK (reference_count >= 0),
    state blob_state NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (sha256)
);

CREATE TABLE folders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    parent_id UUID REFERENCES folders(id) ON DELETE RESTRICT,
    name TEXT NOT NULL CHECK (length(trim(name)) BETWEEN 1 AND 255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE files (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    uploaded_by UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    blob_id UUID NOT NULL REFERENCES blobs(id) ON DELETE RESTRICT,
    folder_id UUID REFERENCES folders(id) ON DELETE RESTRICT,
    name TEXT NOT NULL CHECK (length(trim(name)) BETWEEN 1 AND 255),
    declared_mime TEXT NOT NULL CHECK (length(trim(declared_mime)) > 0),
    detected_mime TEXT NOT NULL CHECK (length(trim(detected_mime)) > 0),
    size_bytes BIGINT NOT NULL CHECK (size_bytes >= 0),
    was_deduplicated BOOLEAN NOT NULL DEFAULT false,
    visibility share_visibility NOT NULL DEFAULT 'private',
    download_count BIGINT NOT NULL DEFAULT 0 CHECK (download_count >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE tags (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL CHECK (length(trim(name)) BETWEEN 1 AND 64),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (owner_id, name)
);

CREATE TABLE file_tags (
    file_id UUID NOT NULL REFERENCES files(id) ON DELETE CASCADE,
    tag_id UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (file_id, tag_id)
);

CREATE TABLE public_shares (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    file_id UUID REFERENCES files(id) ON DELETE CASCADE,
    folder_id UUID REFERENCES folders(id) ON DELETE CASCADE,
    token_hash BYTEA NOT NULL UNIQUE CHECK (octet_length(token_hash) >= 32),
    created_by UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    expires_at TIMESTAMPTZ,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK ((file_id IS NOT NULL) <> (folder_id IS NOT NULL))
);

CREATE TABLE user_shares (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    file_id UUID REFERENCES files(id) ON DELETE CASCADE,
    folder_id UUID REFERENCES folders(id) ON DELETE CASCADE,
    recipient_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    granted_by UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    permission share_permission NOT NULL DEFAULT 'download',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK ((file_id IS NOT NULL) <> (folder_id IS NOT NULL)),
    CHECK (recipient_id <> granted_by)
);

CREATE TABLE download_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    file_id UUID NOT NULL REFERENCES files(id) ON DELETE RESTRICT,
    blob_id UUID NOT NULL REFERENCES blobs(id) ON DELETE RESTRICT,
    actor_id UUID REFERENCES users(id) ON DELETE SET NULL,
    public_share_id UUID REFERENCES public_shares(id) ON DELETE SET NULL,
    started_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE audit_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_id UUID REFERENCES users(id) ON DELETE SET NULL,
    action TEXT NOT NULL CHECK (length(trim(action)) BETWEEN 1 AND 80),
    resource_type TEXT NOT NULL CHECK (length(trim(resource_type)) BETWEEN 1 AND 80),
    resource_id UUID,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    request_id TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX folders_owner_parent_name_ci_idx
    ON folders (owner_id, COALESCE(parent_id, '00000000-0000-0000-0000-000000000000'::uuid), lower(name));
CREATE INDEX folders_owner_parent_idx ON folders (owner_id, parent_id, created_at DESC);

CREATE INDEX files_owner_created_idx ON files (owner_id, created_at DESC, id DESC);
CREATE INDEX files_owner_size_idx ON files (owner_id, size_bytes, id);
CREATE INDEX files_owner_mime_idx ON files (owner_id, detected_mime, id);
CREATE INDEX files_folder_idx ON files (folder_id, created_at DESC, id DESC);
CREATE INDEX files_name_trgm_idx ON files USING GIN (lower(name) gin_trgm_ops);
CREATE INDEX files_uploader_idx ON files (uploaded_by, created_at DESC);

CREATE INDEX file_tags_tag_file_idx ON file_tags (tag_id, file_id);
CREATE INDEX public_shares_file_idx ON public_shares (file_id) WHERE revoked_at IS NULL;
CREATE INDEX public_shares_folder_idx ON public_shares (folder_id) WHERE revoked_at IS NULL;
CREATE INDEX user_shares_recipient_idx ON user_shares (recipient_id, created_at DESC);
CREATE INDEX user_shares_file_idx ON user_shares (file_id, recipient_id);
CREATE INDEX user_shares_folder_idx ON user_shares (folder_id, recipient_id);
CREATE INDEX download_events_file_time_idx ON download_events (file_id, started_at DESC);
CREATE INDEX download_events_time_idx ON download_events (started_at DESC);
CREATE INDEX audit_events_resource_idx ON audit_events (resource_type, resource_id, created_at DESC);
CREATE INDEX audit_events_actor_idx ON audit_events (actor_id, created_at DESC);

COMMIT;
