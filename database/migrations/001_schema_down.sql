-- FTTH: reversión del esquema (DDL).
-- No se ejecuta automáticamente. Usarlo sólo en una base de datos de pruebas.

BEGIN;
DROP TABLE IF EXISTS ftth_module_counters;
DROP TABLE IF EXISTS ftth_module_items;
DROP TABLE IF EXISTS ftth_requests;
COMMIT;
