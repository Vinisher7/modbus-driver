BEGIN;
ALTER TABLE modbus_fetchs DROP CONSTRAINT IF EXISTS fk_id_modbus_device;
COMMIT;