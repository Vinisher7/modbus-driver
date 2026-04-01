DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_id_modbus_tag') THEN
        ALTER TABLE modbus_tag_fetchs 
        ADD CONSTRAINT fk_id_modbus_tag 
        FOREIGN KEY (id_modbus_tag) REFERENCES modbus_tags(id);
    END IF;
END;
$$;