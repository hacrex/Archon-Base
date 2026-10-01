BEGIN;

DROP TRIGGER IF EXISTS database_instances_set_updated_at ON database_instances;
DROP TRIGGER IF EXISTS projects_set_updated_at ON projects;
DROP FUNCTION IF EXISTS set_updated_at();
DROP TABLE IF EXISTS outbox_events;
DROP TABLE IF EXISTS database_instances;
DROP TABLE IF EXISTS projects;

COMMIT;
