DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_id_modbus_device') THEN
        ALTER TABLE modbus_fetchs 
        ADD CONSTRAINT fk_id_modbus_device 
        FOREIGN KEY (id_modbus_device) REFERENCES modbus_devices(id);
    END IF;
END;
$$;