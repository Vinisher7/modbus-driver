-- ==========================================================
-- 1. TABELA: devices
-- ==========================================================
INSERT INTO devices (id, device_name, attached_protocol, protocol, device_type, latitude, longitude, device_address, created_at) VALUES 
('550e8400-e29b-41d4-a716-446655440000', 'Gateway Industrial Sul', true, 'modbus', 'Gateway', '-23.55', '-46.63', 'Setor A1', NOW()),
('67e61d30-22bb-4201-9231-649035293672', 'CLP Tanque de Resfriamento', true, 'modbus', 'PLC', '-23.56', '-46.64', 'Setor B3', NOW());

-- ==========================================================
-- 2. TABELA: mqtt_tags
-- ==========================================================
-- INSERT INTO mqtt_tags (id, raw_topic, formatted_topic, id_device, tag_name, created_at) VALUES 
-- ('a1b2c3d4-e5f6-4a5b-8c9d-0e1f2a3b4c5d', 'telemetry/temp', 'south/g1/temp', '550e8400-e29b-41d4-a716-446655440000', 'Temperatura Externa', NOW()),
-- ('b2c3d4e5-f6a7-4b6c-9d0e-1f2a3b4c5d6e', 'telemetry/hum', 'south/g1/humidity', '550e8400-e29b-41d4-a716-446655440000', 'Umidade Relativa', NOW());

-- ==========================================================
-- 3. TABELA: modbus_devices
-- ==========================================================
INSERT INTO modbus_devices (id, host, port, unit_id, id_device, created_at) VALUES 
('f47ac10b-58cc-4372-a567-0e02b2c3d479', '10.0.0.50', 502, 1, '67e61d30-22bb-4201-9231-649035293672', NOW());

-- ==========================================================
-- 4. TABELA: modbus_tags
-- ==========================================================
INSERT INTO modbus_tags (id, tag_name, data_type, created_at) VALUES 
('7c9e663f-7482-4c2d-9051-561b36994191', 'Vazão de Entrada', 'int16', NOW()),
('8d0f774a-8593-5d3e-0162-672c470052a2', 'Temperatura Tanque', 'float32', NOW());

-- ==========================================================
-- 5. TABELA: modbus_fetchs
-- ==========================================================
INSERT INTO modbus_fetchs (id, id_modbus_device, initial_address, quantity, function_code, created_at) VALUES 
('d1421714-257a-4c28-9721-a3f789965a78', 'f47ac10b-58cc-4372-a567-0e02b2c3d479', 100, 5, 3, NOW()),
('e2532825-368b-5d39-0832-b40890076b89', 'f47ac10b-58cc-4372-a567-0e02b2c3d479', 0, 10, 1, NOW());

-- ==========================================================
-- 6. TABELA: modbus_tag_fetch
-- ==========================================================
INSERT INTO modbus_tag_fetchs (id, id_modbus_fetch, id_modbus_tag, position, created_at) VALUES 
('01a1b2c3-d4e5-4f6a-7b8c-9d0e1f2a3b4c', 'd1421714-257a-4c28-9721-a3f789965a78', '7c9e663f-7482-4c2d-9051-561b36994191', 0, NOW()),
('02b2c3d4-e5f6-5a7b-8c9d-0e1f2a3b4c5d', 'd1421714-257a-4c28-9721-a3f789965a78', '8d0f774a-8593-5d3e-0162-672c470052a2', 1, NOW());