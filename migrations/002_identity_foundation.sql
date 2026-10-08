-- Archon Base identity foundation, migration 002.
-- This stores users and memberships but does not enable unauthenticated API
-- mutations. Authentication and authorization must gate future endpoints.

CREATE TABLE users (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email        text NOT NULL,
    display_name text NOT NULL,
    status       text NOT NULL DEFAULT 'invited',
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT users_email_check CHECK (position('@' in email) > 1),
    CONSTRAINT users_status_check CHECK (status IN ('active', 'invited', 'suspended'))
);

CREATE UNIQUE INDEX users_email_lower_idx ON users (lower(email));

CREATE TABLE organization_memberships (
    organization_id uuid NOT NULL,
    user_id         uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role            text NOT NULL,
    created_at      timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (organization_id, user_id),
    CONSTRAINT memberships_role_check CHECK (role IN ('owner', 'admin', 'operator', 'developer', 'viewer'))
);

CREATE INDEX memberships_user_idx ON organization_memberships (user_id);

CREATE TRIGGER users_set_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
