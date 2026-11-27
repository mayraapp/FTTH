export const homeModules = [
  { view: 'dashboard', icon: '▦', title: 'Dashboard', description: 'Resumen general y métricas de la operación.', accent: 'cyan' },
  { view: 'inventory', icon: '⌖', title: 'Mapa GIS', description: 'Visualización y gestión geográfica de red.', accent: 'green' },
  { view: 'network', icon: '▣', title: 'Monitoreo NOC', description: 'Supervisión en tiempo real de OLT y PON.', accent: 'purple' },
  { view: 'customers', icon: '◉', title: 'Usuarios y Roles', description: 'Gestión de usuarios, roles y permisos de acceso.', accent: 'orange' },
  { view: 'commercial', icon: '◈', title: 'Clientes', description: 'Altas y seguimiento de servicios FTTH.', accent: 'yellow' },
  { view: 'settings', icon: '⚙', title: 'Configuración', description: 'Perfil, seguridad y parámetros de sistema.', accent: 'slate' },
]

export const nmsPorts = [
  { id: 'P1', name: 'Centro · PON 0/1', letter: 'A', subscribers: 108, capacity: 128, state: 'healthy', stateLabel: 'Operativo' },
  { id: 'P2', name: 'Norte · PON 0/2', letter: 'B', subscribers: 86, capacity: 128, state: 'healthy', stateLabel: 'Operativo' },
  { id: 'P3', name: 'Sur · PON 0/3', letter: 'C', subscribers: 71, capacity: 128, state: 'warning', stateLabel: 'Potencia por revisar' },
  { id: 'P4', name: 'Juriquilla · PON 0/4', letter: 'D', subscribers: 0, capacity: 128, state: 'idle', stateLabel: 'Capacidad disponible' },
]

// Corte de referencia estático. Sus valores usan unidades y rangos propios de
// una red GPON, pero no proceden de equipos conectados en tiempo real.
export const nmsReferenceSnapshot = Object.freeze({
  capturedAt: '27 nov 2026 · 14:00 CST',
  source: 'Muestra estática de laboratorio',
  scope: '4 PON · 265 ONT provisionadas · 512 puertos disponibles',
})

export const nmsPortTelemetry = Object.freeze({
  P1: {
    trafficGbps: 1.84, averagePowerDbm: -20.1, latencyMs: 2.4, packetLoss: '0.00', uplink: 'XGE 0/1/0 · 10 GbE', availability: '99.98',
    ont: { name: 'ONT-QRO-CENTRO-014', subscriber: 'Abonado residencial de referencia', serial: 'ZTEG-CNTR-0014', model: 'ZTE F6600P', service: 'Internet 300 / 100 Mbps', rxDbm: -19.4, txDbm: 2.3, status: 'Sesión simulada activa', lastSeen: 'Corte estático: 14:00 CST', connected: true },
  },
  P2: {
    trafficGbps: 1.42, averagePowerDbm: -21.8, latencyMs: 2.8, packetLoss: '0.01', uplink: 'XGE 0/1/0 · 10 GbE', availability: '99.97',
    ont: { name: 'ONT-QRO-NORTE-032', subscriber: 'Abonado comercial de referencia', serial: 'HWTC-NORT-0032', model: 'Huawei EG8145V5', service: 'Internet 500 / 200 Mbps', rxDbm: -21.8, txDbm: 2.1, status: 'Sesión simulada activa', lastSeen: 'Corte estático: 14:00 CST', connected: true },
  },
  P3: {
    trafficGbps: 1.09, averagePowerDbm: -26.8, latencyMs: 3.6, packetLoss: '0.08', uplink: 'XGE 0/1/0 · 10 GbE', availability: '99.91',
    ont: { name: 'ONT-QRO-SUR-018', subscriber: 'Abonado residencial de referencia', serial: 'ZTEG-SUR-0018', model: 'ZTE F670L', service: 'Internet 200 / 80 Mbps', rxDbm: -26.8, txDbm: 1.9, status: 'Sesión simulada · señal en atención', lastSeen: 'Corte estático: 14:00 CST', connected: true },
  },
  P4: {
    trafficGbps: 0, averagePowerDbm: null, latencyMs: null, packetLoss: '0.00', uplink: 'Disponible para expansión', availability: '—',
    ont: { name: 'Sin ONT asociada', subscriber: 'Puerto disponible para alta', serial: '—', model: 'Sin equipo registrado', service: 'Sin perfil asignado', rxDbm: null, txDbm: null, status: 'Sin sesión simulada', lastSeen: 'Sin lectura de referencia', connected: false },
  },
})

export const topologyONTs = [
  { id: 101, name: 'ONT-QRO-CENTRO-014', detail: 'Residencial · señal de recepción -19.4 dBm', status: 'Operativa' },
  { id: 102, name: 'ONT-QRO-NORTE-032', detail: 'Comercial · enlace estable a 300 Mbps', status: 'Operativa' },
  { id: 103, name: 'ONT-QRO-SUR-018', detail: 'Residencial · señal de recepción -26.8 dBm', status: 'Atención' },
  { id: 104, name: 'ONT-QRO-JARDINES-041', detail: 'Residencial · señal de recepción -20.1 dBm', status: 'Operativa' },
  { id: 105, name: 'ONT-QRO-PARQUE-009', detail: 'Empresarial · enlace simétrico 500 Mbps', status: 'Operativa' },
  { id: 106, name: 'ONT-QRO-CENTRO-067', detail: 'Mantenimiento preventivo programado', status: 'En revisión' },
]

export const roleCatalog = [
  { id: 'Administrador', icon: '♛', accent: 'cyan', description: 'Control total de la plataforma' },
  { id: 'Líder', icon: '✦', accent: 'violet', description: 'Coordina equipos y operación' },
  { id: 'Técnico', icon: '⚒', accent: 'orange', description: 'Atiende red y servicios' },
  { id: 'Sub-técnico', icon: '⌁', accent: 'yellow', description: 'Soporte de campo' },
  { id: 'Capturista', icon: '▣', accent: 'emerald', description: 'Registra y valida información' },
  { id: 'Cliente', icon: '◉', accent: 'slate', description: 'Consulta de servicio' },
]

export const moduleConfigs = {
  network: {
    panelTitle: 'Monitoreo de red (NMS)', description: 'Supervisa OLT, enlaces y eventos operativos de la red óptica.', nameLabel: 'Equipo, OLT o alerta', detailLabel: 'Detalle operativo', statuses: ['Operativa', 'Atención', 'En revisión'],
    metrics: [{ label: 'Elementos monitoreados', detail: 'En este módulo' }, { label: 'Con atención', status: 'Atención', detail: 'Requieren seguimiento' }, { label: 'Operativos', status: 'Operativa', detail: 'Estado saludable' }],
  },
  customers: {
    panelTitle: 'Usuarios y Roles', description: 'Administra usuarios, roles y accesos de la operación.', nameLabel: 'Usuario', detailLabel: 'Correo, rol y contacto', statuses: ['Activo', 'Inactivo'],
    metrics: [{ label: 'Total usuarios', detail: 'En este módulo' }, { label: 'Activos hoy', status: 'Activo', detail: 'Con acceso vigente' }, { label: 'Usuarios inactivos', status: 'Inactivo', detail: 'Sin acceso operativo' }],
  },
  coverage: {
    panelTitle: 'Cobertura y expansión', description: 'Controla zonas, factibilidades y proyectos de crecimiento de la red.', nameLabel: 'Zona o proyecto', detailLabel: 'Detalle de cobertura', statuses: ['En análisis', 'Aprobada', 'Pendiente'],
    metrics: [{ label: 'Zonas y proyectos', detail: 'En este módulo' }, { label: 'En análisis', status: 'En análisis', detail: 'Por validar' }, { label: 'Aprobadas', status: 'Aprobada', detail: 'Listas para ejecutar' }],
  },
  inventory: {
    panelTitle: 'GIS e inventario de red', description: 'Ubica infraestructura y controla materiales, cajas, puertos y equipos.', nameLabel: 'Material, caja o equipo', detailLabel: 'Ubicación o detalle de existencias', statuses: ['Disponible', 'Reabastecer', 'En tránsito'],
    metrics: [{ label: 'Activos registrados', detail: 'En este módulo' }, { label: 'Por reabastecer', status: 'Reabastecer', detail: 'Requieren compra' }, { label: 'Disponibles', status: 'Disponible', detail: 'Para operación' }],
  },
  commercial: {
    panelTitle: 'Clientes', description: 'Administra clientes, datos de contacto y servicios de fibra.', nameLabel: 'Nombre completo', detailLabel: 'Contacto y servicio', statuses: ['Activo', 'Pendiente de pago', 'Suspendido'],
    metrics: [{ label: 'Clientes registrados', detail: 'En este módulo' }, { label: 'Pagos pendientes', status: 'Pendiente de pago', detail: 'Para seguimiento' }, { label: 'Servicios activos', status: 'Activo', detail: 'Con servicio vigente' }],
  },
}
