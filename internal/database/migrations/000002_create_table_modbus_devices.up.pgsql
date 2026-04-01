BEGIN;
CREATE TABLE IF NOT EXISTS modbus_devices(
    id UUID PRIMARY KEY,
    id_device UUID NOT NULL,
    unit_id INT NOT NULL,
    host VARCHAR(15) NOT NULL,
    port INT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP
);
COMMIT;