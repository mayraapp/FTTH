# Reporte técnico — Plataforma FTTH

## Propósito del proyecto

FTTH es una plataforma web para supervisar una red de fibra óptica. Reúne un dashboard de indicadores, monitoreo NOC, mapa GIS, clientes, usuarios/roles, configuración y registro de solicitudes. El objetivo es que personal de una empresa de telecomunicaciones pueda consultar la operación, ubicar infraestructura y registrar eventos desde computadora o teléfono.

## Full stack utilizado

| Capa | Tecnología | Uso en FTTH |
| --- | --- | --- |
| Interfaz | HTML, CSS y JavaScript modular | SPA responsive publicada por el backend Go |
| Backend | Go 1.23, `net/http` | API REST, autenticación por cookie y reglas de negocio |
| Balanceador | NGINX | Punto público, distribución `least_conn` y ruta `/estado/` |
| Sesiones | Redis 7 | Sesiones compartidas entre réplicas del backend |
| Datos | PostgreSQL 16 | Solicitudes y registros de módulos compartidos y persistentes |
| Contenedores | Podman + Podman Compose | Ejecución reproducible de todos los servicios |
| Sistema propuesto | Rocky Linux 9 (VM KVM) | Servidor Linux compatible con el ecosistema RHEL |

## Arquitectura de publicación

```text
Usuario
  │ HTTP/HTTPS
  ▼
NGINX (único puerto público)
  ├── /         → FTTH Go réplica 1 ─┐
  ├── /         → FTTH Go réplica 2 ─┼── PostgreSQL
  └── /estado/  → sitio estático     └── Redis (sesiones)
```

La definición ejecutable está en `deploy/compose.yaml`. NGINX hace el balanceo de carga entre las dos aplicaciones Go. Ambas consultan el mismo PostgreSQL y Redis; por ello no dependen de memoria local ni de archivos JSON cuando se despliegan en contenedores.

## Scripts de base de datos

El proyecto versiona los scripts en `database/migrations/` para cumplir el
proceso de base de datos mostrado en clase:

- `001_schema_up.sql`: DDL que crea tablas, llaves primarias e índices.
- `001_schema_down.sql`: DDL de reversión para un entorno de pruebas.
- `002_seed_data.sql`: DML con datos de demostración, sin contraseñas ni datos
  personales reales.

El contenedor PostgreSQL monta y ejecuta los scripts `up` y de datos sólo si su
volumen aún está vacío. Esto evita que una reinicialización normal borre o
sobrescriba información persistida.

## Redes separadas

| Red | Miembros | Razón |
| --- | --- | --- |
| `public_net` | NGINX y sitio estático | Recibe únicamente tráfico publicado por la puerta de enlace. |
| `app_net` | NGINX y réplicas FTTH | NGINX se comunica con el backend sin exponerlo al host. |
| `session_net` | Réplicas FTTH y Redis | Limita las sesiones a la aplicación. |
| `database_net` | Réplicas FTTH y PostgreSQL | Limita la base de datos a la aplicación. |

Las tres redes internas no publican puertos. Sólo NGINX mapea `${HOST_PORT}:80` al servidor virtual.

## Manejo de sesión global

1. Al iniciar sesión, Go genera un identificador aleatorio y lo coloca en una cookie `HttpOnly`, `SameSite=Strict`.
2. El perfil y la expiración se almacenan en Redis durante ocho horas.
3. Cualquier réplica valida ese identificador contra Redis, de modo que el balanceador puede cambiar de réplica sin cerrar la sesión.
4. Al salir, la sesión se elimina de Redis y la cookie se vence en el navegador.

## Seguridad aplicada y pendiente antes de producción

- NGINX es el único servicio expuesto; Redis, PostgreSQL y las apps no tienen puertos de host.
- Las credenciales se entregan por `deploy/.env`, fuera del repositorio, y se restringen con `chmod 600`.
- El backend establece `HttpOnly`, `SameSite=Strict`, `X-Content-Type-Options`, `X-Frame-Options` y `Referrer-Policy`.
- La imagen ejecuta la app sin usuario root.
- Antes de publicar en Internet faltará colocar TLS, firewall, copias de seguridad verificadas, actualización de imágenes y secretos generados en el servidor.

## Funciones de administración del servidor

Según la estructura indicada en clase, se proponen responsabilidades separadas:

| Nivel | Responsabilidad |
| --- | --- |
| `sysadmin` | Administración del Linux/KVM, red, firewall, parches y cuentas. |
| `poweradmin` | Podman Compose, revisión de servicios, despliegue y recuperación controlada. |
| `admin` | Operación de la plataforma FTTH: usuarios, clientes e infraestructura lógica. |

`poweradmin` no debe usar la cuenta administrativa de la aplicación como sustituto de `sysadmin`; son ámbitos diferentes. Se recomienda cuentas personales con `sudo` limitado y nunca compartir contraseñas.

## Procedimiento de demostración

1. Crear una VM Rocky Linux 9 en KVM con recursos definidos por el docente.
2. Instalar Podman y el complemento Compose.
3. Copiar el proyecto y crear `deploy/.env` con secretos no públicos.
4. Ejecutar `npm --prefix frontend run build` y `podman compose --env-file .env up --build -d` desde `deploy/`.
5. Mostrar `podman compose ps`, abrir `/` y `/estado/`.
6. Iniciar sesión, actualizar la página varias veces y cerrar sesión para demostrar Redis.
7. Mostrar que sólo NGINX tiene un puerto publicado y que PostgreSQL/Redis pertenecen a redes internas.

## Publicación y administración de la VM

La topología de publicación sigue el esquema visto en clase: el DNS de un
subdominio apunta a la IP pública; el router reenvía los puertos 80 y 443 a la
IP privada de la VM Rocky; NGINX recibe esas peticiones y las distribuye a las
réplicas FTTH. SSH (22) se usa exclusivamente para administración.

FirewallD debe permitir HTTP/HTTPS y SELinux debe permanecer en modo
`Enforcing`. Cockpit usa 9090, pero no se publica al exterior: se limita a una
red de administración o VPN. Los procedimientos reproducibles están en
`deploy/rocky/`.

## Evidencia de validación local

- Pruebas Go: `go test ./...`.
- Validación de la interfaz: `node --check frontend/src/app.js` y `npm --prefix frontend run build`.
- Validación por contenedores pendiente de instalar Podman y disponer de permisos de libvirt en el equipo anfitrión.
