BEGIN;
CREATE TABLE IF NOT EXISTS modbus_tag_fetchs(
    id UUID PRIMARY KEY,
    id_modbus_fetch UUID NOT NULL,
    id_modbus_tag UUID NOT NULL,
    position int,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP
);
COMMIT;