export const defaultView = 'home'

const metadata = {
  home: { title: 'Inicio', section: 'INICIO' },
  dashboard: { title: 'Resumen operativo', section: 'CENTRO DE CONTROL' },
  network: { title: 'Monitoreo de red (NMS)', section: 'MONITOREO' },
  customers: { title: 'Usuarios y Roles', section: 'ADMINISTRACIÓN' },
  coverage: { title: 'Cobertura y expansión', section: 'PLANEACIÓN' },
  inventory: { title: 'GIS e inventario de red', section: 'GIS' },
  commercial: { title: 'Clientes', section: 'GESTIÓN' },
  settings: { title: 'Configuración', section: 'CONFIGURACIÓN' },
}

export function viewMeta(view) {
  return metadata[view] || metadata.dashboard
}
