BEGIN;
CREATE TABLE IF NOT EXISTS modbus_fetchs(
    id UUID PRIMARY KEY,
    id_modbus_device UUID NOT NULL,
    initial_address INT NOT NULL,
    quantity INT NOT NULL,
    function_code INT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP
);
COMMIT;