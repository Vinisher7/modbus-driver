-- ╔═══════════════════════════════════════════════════════════════╗
-- ║  INSERTs de teste — IHM_TORTUGA (baseado na imagem)         ║
-- ╚═══════════════════════════════════════════════════════════════╝

-- 1) Device
INSERT INTO devices (id, device_name, attached_protocol, protocol, device_type, latitude, longitude, device_address, created_at)
VALUES (
    'a1b2c3d4-e5f6-7890-abcd-ef1234567890',
    'IHM_TORTUGA',
    true,
    'modbus',
    'ihm',
    '-23.550520',
    '-46.633308',
    '192.168.1.100',
    NOW()
);

-- 2) Modbus Device (unit_id = 1, host e porta do simulador/CLP)
INSERT INTO modbus_devices (id, id_device, unit_id, host, port, created_at)
VALUES (
    'b2c3d4e5-f6a7-8901-bcde-f12345678901',
    'a1b2c3d4-e5f6-7890-abcd-ef1234567890',
    1,
    '127.0.0.1',
    502,
    NOW()
);

-- 3) Modbus Fetchs
--    Fetch 1: registros 0-1 (40001-40002 na notação clássica), quantity=2, function_code=3 → Holding Register
--    Fetch 2: registro 4   (40005 na notação clássica),         quantity=1, function_code=3
INSERT INTO modbus_fetchs (id, id_modbus_device, initial_address, quantity, function_code, created_at)
VALUES
    ('c3d4e5f6-a7b8-9012-cdef-123456789012', 'b2c3d4e5-f6a7-8901-bcde-f12345678901', 0, 2, 3, NOW()),
    ('d4e5f6a7-b8c9-0123-defa-234567890123', 'b2c3d4e5-f6a7-8901-bcde-f12345678901', 4, 1, 3, NOW());

-- 4) Modbus Tags
INSERT INTO modbus_tags (id, tag_name, data_type, operation, operation_value, created_at)
VALUES
    ('e5f6a7b8-c9d0-1234-efab-345678901234', 'PRESS_JUSANTE',   'uint16', NULL, NULL, NOW()),
    ('f6a7b8c9-d0e1-2345-fabc-456789012345', 'PRESS_MONTANTE',  'uint16', NULL, NULL, NOW()),
    ('a7b8c9d0-e1f2-3456-abcd-567890123456', 'LEITURA_VAZAO',   'uint16', NULL, NULL, NOW());

-- 5) Modbus Tag ↔ Fetch (posição do registro dentro do fetch)
--    PRESS_JUSANTE  → Fetch 1, position 0 (offset 0 a partir do 40001)
--    PRESS_MONTANTE → Fetch 1, position 1 (offset 1 a partir do 40001)
--    LEITURA_VAZAO  → Fetch 2, position 0 (único registro no fetch)
INSERT INTO modbus_tag_fetchs (id, id_modbus_fetch, id_modbus_tag, position, created_at)
VALUES
    ('b8c9d0e1-f2a3-4567-bcde-678901234567', 'c3d4e5f6-a7b8-9012-cdef-123456789012', 'e5f6a7b8-c9d0-1234-efab-345678901234', 0, NOW()),
    ('c9d0e1f2-a3b4-5678-cdef-789012345678', 'c3d4e5f6-a7b8-9012-cdef-123456789012', 'f6a7b8c9-d0e1-2345-fabc-456789012345', 1, NOW()),
    ('d0e1f2a3-b4c5-6789-defa-890123456789', 'd4e5f6a7-b8c9-0123-defa-234567890123', 'a7b8c9d0-e1f2-3456-abcd-567890123456', 0, NOW());

-- ═══════════════════════════════════════════════════════════════
-- Verificação
-- ═══════════════════════════════════════════════════════════════
SELECT
    d.device_name,
    md.host,
    md.port,
    md.unit_id,
    mf.initial_address,
    mf.quantity,
    mf.function_code,
    mt.tag_name,
    mt.data_type,
    mtf.position
FROM devices d
JOIN modbus_devices md ON md.id_device = d.id
JOIN modbus_fetchs mf ON mf.id_modbus_device = md.id
JOIN modbus_tag_fetchs mtf ON mtf.id_modbus_fetch = mf.id
JOIN modbus_tags mt ON mt.id = mtf.id_modbus_tag
WHERE d.protocol = 'modbus'
ORDER BY d.id, mf.initial_address, mtf.position;