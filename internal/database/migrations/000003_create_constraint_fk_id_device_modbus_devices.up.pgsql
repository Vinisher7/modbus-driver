DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_id_device') THEN
        ALTER TABLE modbus_devices 
        ADD CONSTRAINT fk_id_device 
        FOREIGN KEY (id_device) REFERENCES devices(id);
    END IF;
END;
$$;