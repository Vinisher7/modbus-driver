DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_id_modbus_fetch') THEN
        ALTER TABLE modbus_tag_fetchs 
        ADD CONSTRAINT fk_id_modbus_fetch 
        FOREIGN KEY (id_modbus_fetch) REFERENCES modbus_fetchs(id);
    END IF;
END;
$$;