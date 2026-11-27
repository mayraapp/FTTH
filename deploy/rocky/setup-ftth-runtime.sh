#!/usr/bin/env bash
# Prepara una VM Rocky Linux para ejecutar FTTH con Podman.
# Debe ejecutarlo sysadmin como root dentro de la VM, no en una laptop Mint.
set -euo pipefail

if [[ "${EUID}" -ne 0 ]]; then
  echo "Ejecuta este script con sudo dentro de Rocky Linux." >&2
  exit 1
fi

# Al invocarlo con sudo, conserva como dueño al operador que lanzó el script.
# El valor explícito DEPLOY_USER sigue teniendo prioridad para otros entornos.
DEPLOY_USER="${DEPLOY_USER:-${SUDO_USER:-poweradmin}}"
INSTALL_ROOT="${INSTALL_ROOT:-/opt/ftth/production-system}"

# Rocky 9 distribuye Podman en AppStream. podman-compose vive en EPEL y
# algunas de sus dependencias necesitan CRB; ambos son repositorios gratuitos.
dnf -y install dnf-plugins-core
dnf config-manager --set-enabled crb
dnf -y install epel-release
dnf -y install podman podman-compose firewalld policycoreutils
systemctl enable --now firewalld

# Sólo NGINX se publica. SSH se mantiene para administración; restringirlo a
# una red administrativa se configura después, cuando se conozca esa red.
firewall-cmd --permanent --add-service=http
firewall-cmd --permanent --add-service=https
firewall-cmd --permanent --add-service=ssh
firewall-cmd --reload

install -d -m 0750 "${INSTALL_ROOT}"
if id "${DEPLOY_USER}" >/dev/null 2>&1; then
  chown "${DEPLOY_USER}:${DEPLOY_USER}" "${INSTALL_ROOT}"
else
  echo "Aviso: el usuario ${DEPLOY_USER} no existe; /opt queda bajo control de sysadmin." >&2
fi

echo "Runtime FTTH preparado en ${INSTALL_ROOT}."
echo "SELinux debe permanecer en Enforcing: $(getenforce)."
echo "No se abrió Cockpit (9090); úsalo únicamente desde red privada o VPN."
