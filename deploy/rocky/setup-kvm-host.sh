#!/usr/bin/env bash
# Prepara un servidor físico Rocky/RHEL como host KVM de laboratorio.
# No ejecutar dentro de una VM Rocky: ahí no hace falta virtualización anidada.
set -euo pipefail

if [[ "${EUID}" -ne 0 ]]; then
  echo "Ejecuta este script con sudo en el host Rocky/RHEL." >&2
  exit 1
fi

dnf -y install @virtualization firewalld
systemctl enable --now libvirtd firewalld

echo "Host KVM preparado. Verifica con: virsh list --all"
echo "Cockpit no se instala ni publica por defecto."
