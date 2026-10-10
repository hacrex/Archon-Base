-- Archon Base project-scoped authorization, migration 004.
-- Organization owners/admins retain organization-wide access. Other users
-- require an explicit membership row for each project they can access.

CREATE TABLE project_memberships (
    project_id  uuid NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    user_id     uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role        text NOT NULL,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (project_id, user_id),
    CONSTRAINT project_memberships_role_check
        CHECK (role IN ('owner', 'admin', 'operator', 'developer', 'viewer'))
);

CREATE INDEX project_memberships_user_idx
    ON project_memberships (user_id, project_id);

CREATE INDEX project_memberships_project_role_idx
    ON project_memberships (project_id, role);

CREATE TRIGGER project_memberships_set_updated_at
    BEFORE UPDATE ON project_memberships
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

COMMENT ON TABLE project_memberships IS
    'Explicit user access grants for projects; organization owners and admins are implicit project administrators.';
COMMENT ON COLUMN project_memberships.role IS
    'Effective project role: owner, admin, operator, developer, or viewer.';
