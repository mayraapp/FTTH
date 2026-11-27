// Serie demostrativa hasta que la telemetría e histórico de órdenes estén conectados.
export const weeklyOperationTrend = [
  { label: 'Lun', opened: 13, resolved: 12 },
  { label: 'Mar', opened: 16, resolved: 15 },
  { label: 'Mié', opened: 11, resolved: 11 },
  { label: 'Jue', opened: 18, resolved: 16 },
  { label: 'Vie', opened: 15, resolved: 14 },
  { label: 'Sáb', opened: 8, resolved: 8 },
  { label: 'Dom', opened: 10, resolved: 9 },
]

export const dashboardGrowthTrend = [
  { label: 'Ene', value: 2384 },
  { label: 'Feb', value: 2418 },
  { label: 'Mar', value: 2461 },
  { label: 'Abr', value: 2507 },
  { label: 'May', value: 2552 },
]

export const dashboardAvailabilityTrend = [
  { label: '00:00', value: 99.80 },
  { label: '04:00', value: 99.90 },
  { label: '08:00', value: 99.87 },
  { label: '12:00', value: 99.72 },
  { label: '16:00', value: 99.82 },
  { label: '20:00', value: 99.90 },
  { label: '24:00', value: 99.80 },
]

export const dashboardAlerts = [
  { tone: 'attention', title: 'Potencia óptica por revisar', detail: 'OLT-QRO-SUR-01 · PON 0/3 · ONT-QRO-SUR-018' },
  { tone: 'info', title: 'Mantenimiento programado', detail: 'OLT-QRO-NORTE-01 · 23:00–00:30 h' },
  { tone: 'healthy', title: 'Respaldo de configuración finalizado', detail: '4 OLT y 2,552 abonados sincronizados' },
]

export const dashboardStats = [
  { label: 'Uptime promedio', value: '99.82%' },
  { label: 'Ocupación PON', value: '52%' },
  { label: 'Abonados activos', value: '2,552' },
]
