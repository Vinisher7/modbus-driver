DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'data_type_enum') THEN
        CREATE TYPE data_type_enum AS ENUM (
            'uint16', 'int16', 'float32', 'float32_swapped', 'int32', 'uint32'
        );
    END IF;
END $$;