-- FTTH: datos de demostración (DML).
-- No contiene contraseñas ni información personal real.

BEGIN;

INSERT INTO ftth_requests
    (id, folio, type, priority, crew, description, status, created_at, created_by)
VALUES
    (1, 'FTTH-0001', 'Mantenimiento', 'Alta', 'Por asignar', 'Revisar potencia óptica del enlace troncal en la zona Centro.', 'Pendiente', '2026-11-13T03:48:43Z', 'May Torres'),
    (2, 'FTTH-0002', 'Instalación', 'Alta', 'Cuadrilla Centro', 'Activar servicio FTTH y configurar ONT en el edificio Aurora.', 'En proceso', '2026-11-13T02:10:00Z', 'Operador Demo'),
    (3, 'FTTH-0003', 'Incidencia de red', 'Alta', 'Soporte NOC', 'Validar potencia óptica en el splitter del sector Norte.', 'Pendiente', '2026-11-13T01:35:00Z', 'Operador Demo'),
    (4, 'FTTH-0004', 'Ampliación', 'Media', 'Cuadrilla Sur', 'Revisar factibilidad para extender cobertura en Colonia Jardines.', 'Pendiente', '2026-11-12T23:20:00Z', 'Operador Demo'),
    (5, 'FTTH-0005', 'Soporte técnico', 'Baja', 'Soporte NOC', 'Confirmar estabilidad del servicio después del cambio de ONT.', 'Resuelta', '2026-11-12T21:15:00Z', 'Operador Demo')
ON CONFLICT (id) DO NOTHING;

SELECT setval(
    pg_get_serial_sequence('ftth_requests', 'id'),
    (SELECT COALESCE(MAX(id), 1) FROM ftth_requests),
    true
);

INSERT INTO ftth_module_items
    (module, id, name, detail, status, map_type, map_zone, map_x, map_y, created_at, updated_at)
VALUES
    ('network', 1, 'OLT-QRO-CENTRO-01', 'Huawei MA5800-X7 · 612 ONT en línea · óptica estable en PON 0/1', 'Operativa', 'OLT', 'Centro', 20, 30, '2026-11-13T00:00:00Z', '2026-11-18T09:20:00Z'),
    ('network', 2, 'OLT-QRO-NORTE-01', 'ZTE C320 · 428 ONT en línea · potencia PON 0/2 en revisión', 'Atención', 'OLT', 'Norte', 77, 27, '2026-11-13T00:00:00Z', '2026-11-18T08:45:00Z'),
    ('network', 3, 'OLT-QRO-SUR-01', 'Huawei MA5800-X2 · 376 ONT en línea · mantenimiento programado', 'En revisión', 'OLT', 'Sur', 68, 74, '2026-11-13T00:00:00Z', '2026-11-18T07:55:00Z'),
    ('customers', 1, 'May Torres', 'may.torres@teracom.mx|Administrador|TERACOM|442 214 6101', 'Activo', NULL, NULL, NULL, NULL, '2026-10-04T00:00:00Z', '2026-11-18T09:42:00Z'),
    ('customers', 2, 'Roberto Méndez', 'roberto.mendez@teracom.mx|Líder|TERACOM|442 214 6102', 'Activo', NULL, NULL, NULL, NULL, '2026-10-02T00:00:00Z', '2026-11-18T08:15:00Z'),
    ('coverage', 1, 'Fraccionamiento El Refugio', 'Factibilidad para 186 viviendas · levantamiento de postes en curso', 'En análisis', NULL, NULL, NULL, NULL, '2026-11-13T00:00:00Z', '2026-11-18T08:30:00Z'),
    ('coverage', 2, 'Parque Industrial Benito Juárez', '22 empresas · troncal de 2.4 km autorizada', 'Aprobada', NULL, NULL, NULL, NULL, '2026-11-13T00:00:00Z', '2026-11-17T16:10:00Z'),
    ('inventory', 1, 'ONT GPON Wi-Fi 6', '84 disponibles · punto de reorden: 40 unidades', 'Disponible', NULL, NULL, NULL, NULL, '2026-11-13T00:00:00Z', '2026-11-18T08:05:00Z'),
    ('inventory', 2, 'Splitter PLC 1:8 SC/APC', '18 disponibles · punto de reorden: 25 unidades', 'Reabastecer', NULL, NULL, NULL, NULL, '2026-11-13T00:00:00Z', '2026-11-18T09:10:00Z'),
    ('commercial', 1, 'TELMEX', 'Cuenta de demostración · Centro · Enlace empresarial 500 Mbps', 'Activo', NULL, NULL, NULL, NULL, '2026-11-13T00:00:00Z', '2026-11-18T09:18:00Z'),
    ('commercial', 2, 'Totalplay', 'Cuenta de demostración · Norte · Enlace empresarial 300 Mbps', 'Activo', NULL, NULL, NULL, NULL, '2026-11-13T00:00:00Z', '2026-11-18T08:52:00Z'),
    ('commercial', 3, 'Megacable', 'Cuenta de demostración · Sur · Enlace empresarial 500 Mbps', 'Pendiente de pago', NULL, NULL, NULL, NULL, '2026-11-13T00:00:00Z', '2026-11-18T07:40:00Z')
ON CONFLICT (module, id) DO NOTHING;

INSERT INTO ftth_module_counters (module, last_id)
SELECT module, MAX(id) FROM ftth_module_items GROUP BY module
ON CONFLICT (module) DO UPDATE SET last_id = EXCLUDED.last_id;

COMMIT;
