BEGIN;
CREATE TABLE IF NOT EXISTS modbus_tags(
    id UUID PRIMARY KEY,
    tag_name VARCHAR(30) NOT NULL,
    operation VARCHAR(30),
    operation_value VARCHAR(30),
    data_type data_type_enum NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP
);
COMMIT;