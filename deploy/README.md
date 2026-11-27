# Despliegue con Podman Compose

Este directorio publica FTTH con dos réplicas Go, NGINX, Redis, PostgreSQL y un sitio estático independiente.

## Preparación

1. Compila el frontend desde la raíz del proyecto:

   ```bash
   npm --prefix frontend run build
   ```

2. Crea el archivo de secretos local:

   ```bash
   cp deploy/.env.example deploy/.env
   chmod 600 deploy/.env
   ```

   Sustituye los dos valores de ejemplo. No uses la contraseña de la aplicación ni subas `deploy/.env` al repositorio.

3. Desde `deploy/`, inicia el entorno:

   ```bash
   podman compose --env-file .env up --build -d
   ```

   En un volumen PostgreSQL nuevo, el contenedor ejecuta automáticamente
   `database/migrations/001_schema_up.sql` y `002_seed_data.sql`. No ejecutes
   el archivo `001_schema_down.sql` salvo que estés en una base de pruebas y
   quieras eliminar todo el esquema FTTH.

4. Comprueba sólo el acceso público:

   ```bash
   curl -I http://127.0.0.1:8080/
   curl -I http://127.0.0.1:8080/estado/
   podman compose ps
   ```

En la VM final cambia `HOST_PORT=80` si el responsable del sistema usa Podman con permisos para publicar el puerto 80. Mientras se prueba sin privilegios, usa `8080`.

## Servicios y aislamiento

| Servicio | Función | Red(es) | Puerto publicado |
| --- | --- | --- | --- |
| `nginx-gateway` | Entrada, balanceo y ruteo | `public_net`, `app_net` | Sólo éste: `${HOST_PORT}:80` |
| `ftth-app-1`, `ftth-app-2` | Dos réplicas Go | `app_net`, `session_net`, `database_net` | Ninguno |
| `static-site` | Segundo proyecto estático en `/estado/` | `public_net` | Ninguno |
| `redis` | Sesiones globales | `session_net` | Ninguno |
| `postgres` | Datos persistentes | `database_net` | Ninguno |

Las redes `app_net`, `session_net` y `database_net` son internas. Redis y PostgreSQL no se pueden alcanzar desde el host ni desde el navegador.

## Scripts SQL versionados

| Archivo | Tipo | Propósito |
| --- | --- | --- |
| `database/migrations/001_schema_up.sql` | DDL | Crea tablas, claves e índices. |
| `database/migrations/001_schema_down.sql` | DDL | Revierte el esquema **sólo en pruebas**. |
| `database/migrations/002_seed_data.sql` | DML | Inserta datos de demostración sin secretos. |

Los scripts de inicialización de PostgreSQL se ejecutan únicamente la primera
vez que se crea `postgres_data`. Para repetirlos en local hay que eliminar ese
volumen de prueba de forma intencional; nunca hacerlo sobre datos reales.

## Validación de la sesión global

1. Inicia sesión en `/`.
2. Recarga varias veces; NGINX puede atender cada petición desde una réplica distinta.
3. La sesión continúa válida porque el identificador de cookie consulta Redis.
4. Cierra sesión y vuelve a cargar: Redis elimina la sesión y `/api/me` responde `401`.

## Operación segura

- Consulta registros: `podman compose logs --tail=100 nginx-gateway ftth-app-1 ftth-app-2`.
- Detén sin borrar datos: `podman compose down`.
- Para borrar intencionalmente todos los datos de prueba: `podman compose down -v`. Este último comando elimina los volúmenes de Redis y PostgreSQL.
- Antes de exponer Internet, configura un dominio, TLS con certificados válidos, firewall y actualizaciones del sistema.
