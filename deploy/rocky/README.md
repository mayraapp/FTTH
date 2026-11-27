# Preparación de Rocky Linux

Estos scripts representan la capa del pizarrón: KVM, una VM Rocky y luego
Podman con NGINX, las dos apps, Redis, PostgreSQL y el sitio estático.

## Elegir el lugar correcto para KVM

| Situación | Dónde instalar KVM | Script o comando |
| --- | --- | --- |
| Servidor físico con Rocky/RHEL | En ese servidor físico | `sudo ./setup-kvm-host.sh` |
| Laptop Linux Mint que aloja una VM Rocky | En Linux Mint con `apt` | No ejecutar `setup-kvm-host.sh` en la VM |
| VM Rocky que ejecutará FTTH | No instala KVM | `sudo ./setup-ftth-runtime.sh` |

No es necesario ni recomendable instalar KVM dentro de la VM Rocky salvo que
se requiera virtualización anidada para otra práctica.

## Preparar la VM Rocky para FTTH

1. Crea una VM Rocky Linux 9 con una IP privada fija o reservada, por ejemplo
   `192.168.1.30`. Para esta práctica se recomiendan al menos 2 vCPU, 4 GB RAM
   y 30 GB de disco libre.
2. Copia este proyecto a la VM y ejecuta como cuenta administrativa:

   ```bash
   cd deploy/rocky
   sudo DEPLOY_USER=poweradmin ./setup-ftth-runtime.sh
   ```

3. Copia el proyecto a `/opt/ftth/production-system` y asegura que
   `poweradmin` sea propietario si esa cuenta existe.
4. Como `poweradmin`, crea `deploy/.env` a partir de `.env.example`, genera
   secretos hexadecimales y compila/inicia la plataforma:

   ```bash
   openssl rand -hex 24
   npm --prefix frontend run build
   cd deploy
   podman compose --env-file .env up --build -d
   ```

5. Para exposición final, configura `HOST_PORT=80`, un dominio y TLS. Durante
   pruebas locales se conserva `HOST_PORT=8080`.

## Red, dominio y puertos

```text
DNS: ftth.tudominio.example → IP pública
Router: 80/443 → 192.168.1.30 (VM Rocky)
VM: NGINX → app 1 / app 2 → Redis y PostgreSQL internos
```

Si la VM usa NAT de KVM, el reenvío debe llegar primero al host físico y luego
a la VM. Una interfaz en puente simplifica la práctica porque la VM recibe su
propia IP de la red local.

## Controles de seguridad

- Mantener SELinux en `Enforcing`: `getenforce`.
- FirewallD abre únicamente HTTP, HTTPS y SSH. Restringe SSH a la red de
  administración antes de publicar el servidor.
- No se abre Cockpit (`9090`) en el firewall. Si se utiliza, debe ser por VPN o
  red privada.
- Redis y PostgreSQL no exponen puertos al host; los aísla Podman Compose.
- Los secretos viven sólo en `deploy/.env`, con permisos `600`.
