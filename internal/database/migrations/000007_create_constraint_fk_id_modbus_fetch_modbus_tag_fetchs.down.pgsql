BEGIN;
ALTER TABLE modbus_tag_fetchs DROP CONSTRAINT IF EXISTS fk_id_modbus_fetch;
COMMIT;