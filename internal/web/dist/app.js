import { createApp, computed, onMounted, reactive, ref } from './plugins/vue.js'
import { createApi } from './services/api.js'
import { homeModules, moduleConfigs, nmsPorts, nmsPortTelemetry, nmsReferenceSnapshot, roleCatalog, topologyONTs } from './models/operation.js?v=28'
import { weeklyOperationTrend, dashboardAlerts, dashboardAvailabilityTrend, dashboardGrowthTrend, dashboardStats } from './models/dashboard.js?v=3'
import { defaultView, viewMeta } from './router/navigation.js?v=2'
import { useSessionState } from './stores/session.js'

export function mountOperationsApp(selector) {
createApp({
  setup() {
    const { user, loading, error, success } = useSessionState()
    const view = ref(defaultView)
    const mobileMenuOpen = ref(false)
    const credentials = reactive({ company: 'TERACOM', username: 'admin', password: '' })
    const globalSearch = ref('')
    const theme = ref(readPreferredTheme())
    const settingsTab = ref('profile')
    const profileForm = reactive({ name: '', email: '', phone: '', company: '', position: 'Administrador de red' })
    const securityForm = reactive({ multiFactor: true, sessionDuration: '8 horas', accessAlerts: true, recoveryEmail: true })
    const systemForm = reactive({ networkName: 'TERACOM FTTH', timezone: 'America/Mexico_City', dateFormat: 'DD/MM/AAAA', compactMode: false, automaticRefresh: true })
    const selectedNmsPort = ref('P1')
    const selectedTerminal = ref(null)
    const selectedNmsOLTId = ref('')
    const nmsOltEditorOpen = ref(false)
    const nmsOltEditor = reactive({ name: '', zone: 'Centro', status: 'Operativa', detail: '' })

    const moduleKeys = Object.keys(moduleConfigs)
    const moduleItems = reactive(Object.fromEntries(moduleKeys.map(module => [module, []])))
    const moduleForms = reactive(Object.fromEntries(moduleKeys.map(module => [module, { name: '', detail: '', status: moduleConfigs[module].statuses[0] }])))
    const moduleEdit = reactive({ module: '', id: null, name: '', detail: '', status: '' })
    const userSearch = ref('')
    const roleFilter = ref('')
    const userEditorOpen = ref(false)
    const userEditor = reactive({ id: null, name: '', email: '', role: 'Técnico', company: 'TERACOM', phone: '', status: 'Activo' })
    const clientEditorOpen = ref(false)
    const clientEditor = reactive({ id: null, name: '', phone: '', email: '', address: '', coordinates: '', plan: '100 Mbps', status: 'Activo' })

    const title = computed(() => viewMeta(view.value).title)
    const sectionLabel = computed(() => viewMeta(view.value).section)
    const currentModule = computed(() => moduleConfigs[view.value] || null)
    const userInitial = computed(() => (user.value?.name || 'A').trim().charAt(0).toUpperCase())
    const currentDate = computed(() => new Intl.DateTimeFormat('es-MX', { day: 'numeric', month: 'short', year: 'numeric' }).format(new Date()))
    const currentDateLong = computed(() => new Intl.DateTimeFormat('es-MX', { weekday: 'long', day: 'numeric', month: 'long', year: 'numeric' }).format(new Date()))
    const activeNmsPort = computed(() => nmsPorts.find(port => port.id === selectedNmsPort.value) || nmsPorts[0])
    const activeNmsTelemetry = computed(() => nmsPortTelemetry[activeNmsPort.value.id] || nmsPortTelemetry.P4)
    const nmsTotalTraffic = computed(() => Object.values(nmsPortTelemetry).reduce((total, telemetry) => total + telemetry.trafficGbps, 0).toFixed(2))
    const nmsTerminals = computed(() => Array.from({ length: 10 }, (_, index) => `${activeNmsPort.value.letter}${index + 1}`))
    const nmsMatrixRows = computed(() => {
      const liveReadings = {
        P1: ['-18.42', '-20.06', '-21.18'],
        P2: ['-19.74', '-23.86'],
        P3: ['-26.80'],
      }
      return Array.from({ length: 8 }, (_, rowIndex) => {
        const port = `P${rowIndex + 1}`
        return {
          port,
          cells: nmsTerminals.value.map((id, columnIndex) => {
            const value = (liveReadings[port] || [])[columnIndex] || ''
            const status = value === '-N.C-' ? 'offline' : value && Number(value) <= -25 ? 'danger' : value ? 'ok' : 'empty-state'
            return { id, value, status }
          }),
        }
      })
    })
    const nmsAlertCount = computed(() => nmsMatrixRows.value.flatMap(row => row.cells).filter(cell => cell.status === 'danger').length)
    const nmsOccupancy = computed(() => {
      const occupied = nmsPorts.reduce((total, port) => total + port.subscribers, 0)
      const capacity = nmsPorts.reduce((total, port) => total + port.capacity, 0)
      return capacity ? Math.round((occupied / capacity) * 100) : 0
    })
    const selectedNetworkNodeKey = ref('')
    const networkLayer = ref('all')
    const mapTool = ref('manage-plans')
    const mapZoom = ref(1)
    const mapEditorOpen = ref(false)
    const mapEditor = reactive({ type: 'OLT', name: '', detail: '', status: 'Operativa', zone: 'Centro', x: 50, y: 50 })
    const mapTopologyZones = [
      { key: 'centro', label: 'Centro', x: 11, y: 38, width: 26, height: 18, oltX: 20, oltY: 30 },
      { key: 'norte', label: 'Norte', x: 62, y: 12, width: 25, height: 18, oltX: 77, oltY: 27 },
      { key: 'juriquilla', label: 'Juriquilla', x: 24, y: 8, width: 23, height: 17, oltX: 37, oltY: 18 },
      { key: 'sur', label: 'Sur', x: 56, y: 60, width: 29, height: 21, oltX: 68, oltY: 74 },
    ]
    const ontTopologyPositions = {
      101: { zoneKey: 'centro', x: 12, y: 67 },
      102: { zoneKey: 'norte', x: 78, y: 48 },
      103: { zoneKey: 'sur', x: 62, y: 58 },
      104: { zoneKey: 'juriquilla', x: 35, y: 31 },
      105: { zoneKey: 'norte', x: 89, y: 15 },
      106: { zoneKey: 'centro', x: 34, y: 79 },
    }
    function mapZoneKey(value = '') {
      const text = String(value).toLocaleLowerCase('es-MX')
      if (text.includes('juriquilla') || text.includes('jardines')) return 'juriquilla'
      if (text.includes('norte') || text.includes('parque')) return 'norte'
      if (text.includes('sur')) return 'sur'
      return 'centro'
    }
    function mapNodeZone(record) {
      return mapZoneKey(`${record?.map?.zone || ''} ${record?.name || ''}`)
    }
    const networkMapNodes = computed(() => {
      const oltNodes = moduleItems.network.map((record, index) => {
        const map = record.map || {}
        const type = map.type || 'OLT'
        const zoneKey = mapNodeZone(record)
        const zone = mapTopologyZones.find(item => item.key === zoneKey)
        const duplicate = moduleItems.network.slice(0, index).filter(item => (item.map?.type || 'OLT') === type && mapNodeZone(item) === zoneKey).length
        return {
          key: `network-${record.id}`,
          module: 'network',
          record,
          type,
          zoneKey,
          attention: ['Atención', 'En revisión'].includes(record.status),
          x: Number.isFinite(map.x) ? map.x : Math.min(91, (zone?.oltX || 50) + (duplicate * 4)),
          y: Number.isFinite(map.y) ? map.y : Math.min(89, (zone?.oltY || 50) + (duplicate * 4)),
        }
      })
      const ontNodes = topologyONTs.map(record => {
        const position = ontTopologyPositions[record.id] || { zoneKey: mapNodeZone(record), x: 50, y: 50 }
        return { key: `ont-${record.id}`, module: 'inventory', record, type: 'ONT', readOnly: true, attention: ['Atención', 'En revisión'].includes(record.status), ...position }
      })
      return [...oltNodes, ...ontNodes]
    })
    const nmsOLTs = computed(() => moduleItems.network.filter(item => !item.map || item.map.type === 'OLT'))
    const activeNmsOLT = computed(() => nmsOLTs.value.find(item => String(item.id) === selectedNmsOLTId.value) || nmsOLTs.value[0] || null)
    const selectedNetworkNode = computed(() => networkMapNodes.value.find(node => node.key === selectedNetworkNodeKey.value) || null)
    const visibleNetworkMapNodes = computed(() => networkMapNodes.value.filter(node => {
      const term = globalSearch.value.trim().toLocaleLowerCase('es-MX')
      if (term && !`${node.type} ${node.record.name} ${node.record.detail} ${node.record.status}`.toLocaleLowerCase('es-MX').includes(term)) return false
      if (networkLayer.value === 'olt') return node.type === 'OLT'
      if (networkLayer.value === 'ont') return node.type === 'ONT'
      if (networkLayer.value === 'attention') return ['Atención', 'En revisión'].includes(node.record.status)
      return true
    }))
    const networkOLTCount = computed(() => networkMapNodes.value.filter(node => node.type === 'OLT').length)
    const networkONTCount = computed(() => networkMapNodes.value.filter(node => node.type === 'ONT').length)
    const networkAttentionNodes = computed(() => networkMapNodes.value.filter(node => ['Atención', 'En revisión'].includes(node.record.status)))
    const networkTopologyZones = computed(() => mapTopologyZones.map(zone => ({
      ...zone,
      oltCount: networkMapNodes.value.filter(node => node.type === 'OLT' && node.zoneKey === zone.key).length,
      ontCount: networkMapNodes.value.filter(node => node.type === 'ONT' && node.zoneKey === zone.key).length,
    })))
    const networkBackboneLinks = computed(() => {
      const oltByZone = Object.fromEntries(networkMapNodes.value.filter(node => node.type === 'OLT').map(node => [node.zoneKey, node]))
      return [['centro', 'norte'], ['norte', 'sur'], ['sur', 'centro'], ['centro', 'juriquilla']]
        .map(([from, to]) => ({ from: oltByZone[from], to: oltByZone[to] }))
        .filter(link => link.from && link.to && link.from.key !== link.to.key)
        .map(link => ({ key: `backbone-${link.from.key}-${link.to.key}`, x1: link.from.x, y1: link.from.y, x2: link.to.x, y2: link.to.y }))
    })
    const networkAccessLinks = computed(() => networkMapNodes.value
      .filter(node => node.type !== 'OLT')
      .map(node => {
        const parent = networkMapNodes.value.find(candidate => candidate.type === 'OLT' && candidate.zoneKey === node.zoneKey) || networkMapNodes.value.find(candidate => candidate.type === 'OLT')
        return parent ? { key: `access-${parent.key}-${node.key}`, x1: parent.x, y1: parent.y, x2: node.x, y2: node.y, attention: node.attention || parent.attention } : null
      })
      .filter(Boolean))
    const networkMapLinks = computed(() => [...networkBackboneLinks.value, ...networkAccessLinks.value])
    const operationTrendMax = computed(() => Math.max(...weeklyOperationTrend.flatMap(point => [point.opened, point.resolved]), 1))
    const operationChartPoints = computed(() => weeklyOperationTrend.map((point, index) => {
      const x = 42 + (index * 90)
      const scale = 132 / operationTrendMax.value
      return { ...point, x, openedY: 168 - (point.opened * scale), resolvedY: 168 - (point.resolved * scale) }
    }))
    const operationOpenedPolyline = computed(() => operationChartPoints.value.map(point => `${point.x},${point.openedY}`).join(' '))
    const operationResolvedPolyline = computed(() => operationChartPoints.value.map(point => `${point.x},${point.resolvedY}`).join(' '))
    const operationOpenedArea = computed(() => `42,168 ${operationOpenedPolyline.value} 582,168`)
    const weeklyOpened = computed(() => weeklyOperationTrend.reduce((total, point) => total + point.opened, 0))
    const weeklyResolved = computed(() => weeklyOperationTrend.reduce((total, point) => total + point.resolved, 0))
    const resolutionRate = computed(() => Math.round((weeklyResolved.value / weeklyOpened.value) * 100))
    const dashboardGrowthMax = computed(() => Math.max(...dashboardGrowthTrend.map(point => point.value), 1))
    const dashboardGrowthPoints = computed(() => dashboardGrowthTrend.map((point, index) => {
      const x = 46 + (index * 134)
      const y = 188 - ((point.value / dashboardGrowthMax.value) * 142)
      return { ...point, x, y }
    }))
    const dashboardGrowthPolyline = computed(() => dashboardGrowthPoints.value.map(point => `${point.x},${point.y}`).join(' '))
    const dashboardGrowthArea = computed(() => `46,188 ${dashboardGrowthPolyline.value} 582,188`)
    const dashboardAvailabilityPoints = computed(() => dashboardAvailabilityTrend.map((point, index) => {
      const x = 42 + (index * 90)
      const y = 188 - (((point.value - 99) / 1) * 142)
      return { ...point, x, y }
    }))
    const dashboardAvailabilityPolyline = computed(() => dashboardAvailabilityPoints.value.map(point => `${point.x},${point.y}`).join(' '))
    const dashboardAvailabilityArea = computed(() => `42,188 ${dashboardAvailabilityPolyline.value} 582,188`)
    const directoryUsers = computed(() => moduleItems.customers.map((item, index) => {
      const [email = '', role = '', company = '', phone = ''] = String(item.detail || '').split('|').map(value => value.trim())
      const fallbackRole = roleCatalog[index % roleCatalog.length].id
      const initial = item.name.trim().split(/\s+/).slice(0, 2).map(word => word[0]).join('').toUpperCase()
      return {
        ...item,
        email: email || `${item.name.toLowerCase().replace(/\s+/g, '.')}@teracom.mx`,
        role: role || fallbackRole,
        company: company || 'TERACOM',
        phone: phone || 'Sin teléfono',
        initials: initial || 'U',
        lastAccess: item.status === 'Activo' ? ['Hoy, 09:42', 'Hoy, 08:15', 'Ayer, 18:36', 'Hoy, 07:50'][index % 4] : '—',
      }
    }))
    const filteredDirectoryUsers = computed(() => {
      const term = userSearch.value.trim().toLowerCase()
      return directoryUsers.value.filter(entry => {
        const matchesRole = !roleFilter.value || entry.role === roleFilter.value
        const matchesTerm = !term || [entry.name, entry.email, entry.company, entry.phone, entry.role].some(value => String(value).toLowerCase().includes(term))
        return matchesRole && matchesTerm
      })
    })
    const userRoleCards = computed(() => roleCatalog.map(role => ({
      ...role,
      count: directoryUsers.value.filter(entry => entry.role === role.id).length,
    })))
    const activeDirectoryUsers = computed(() => directoryUsers.value.filter(entry => entry.status === 'Activo').length)
    const clientDirectory = computed(() => moduleItems.commercial.map(item => {
      const values = String(item.detail || '').split('|').map(value => value.trim())
      const structuredRecord = values.length >= 5
      const [phone = '', email = '', address = '', coordinates = '', plan = ''] = values
      const inferredPlan = String(item.detail || '').match(/\b\d+(?:\s*)Mbps\b/i)?.[0] || '100 Mbps'
      const initials = item.name.trim().split(/\s+/).slice(0, 2).map(word => word[0]).join('').toUpperCase()
      return {
        ...item,
        phone: structuredRecord ? phone || 'Sin teléfono' : 'Sin teléfono',
        email: structuredRecord ? email : '',
        address: structuredRecord ? address || 'Sin dirección registrada' : 'Sin dirección registrada',
        coordinates: structuredRecord ? coordinates : '',
        plan: structuredRecord ? plan || '100 Mbps' : inferredPlan,
        initials: initials || 'C',
      }
    }))
    const activeClients = computed(() => clientDirectory.value.filter(entry => entry.status === 'Activo').length)
    const pendingClients = computed(() => clientDirectory.value.filter(entry => entry.status === 'Pendiente de pago').length)
    const dashboardMetrics = computed(() => [
      { label: 'Clientes activos', value: activeClients.value, detail: 'Servicios vigentes', icon: '◉', tone: 'clients' },
      { label: 'Fallas activas', value: nmsAlertCount.value, detail: 'Requieren atención', icon: '!', tone: 'alerts' },
      { label: 'Eventos abiertos', value: Math.max(weeklyOpened.value - weeklyResolved.value, 0), detail: 'Durante la semana', icon: '▤', tone: 'events' },
      { label: 'OLT monitoreadas', value: networkOLTCount.value, detail: 'Red supervisada', icon: '⌖', tone: 'network' },
    ])

    const api = createApi()
    let sessionGeneration = 0

    async function restoreSession() {
      const generation = sessionGeneration
      try {
        const restoredUser = (await api('/api/me')).user
        if (generation !== sessionGeneration) return
        user.value = restoredUser
        hydrateProfileForm()
        await loadModules()
      } catch (_) {
        if (generation === sessionGeneration) user.value = null
      }
    }
    async function login() {
      const generation = ++sessionGeneration
      loading.value = true; error.value = ''
      try {
        const authenticatedUser = (await api('/api/login', { method: 'POST', body: JSON.stringify(credentials) })).user
        if (generation !== sessionGeneration) return
        user.value = authenticatedUser
        hydrateProfileForm()
        credentials.password = ''
        await loadModules()
      } catch (err) {
        error.value = err.message
      } finally {
        loading.value = false
      }
    }
    async function logout() {
      const generation = ++sessionGeneration
      loading.value = true; error.value = ''
      try {
        await api('/api/logout', { method: 'POST' })
        if (generation !== sessionGeneration) return
        user.value = null; view.value = defaultView
        mobileMenuOpen.value = false
        moduleKeys.forEach(module => { moduleItems[module] = [] })
        cancelModuleEdit()
        closeUserEditor()
        closeClientEditor()
      } catch (err) {
        if (generation === sessionGeneration) error.value = err.message || 'No fue posible cerrar sesión'
      } finally {
        if (generation === sessionGeneration) loading.value = false
      }
    }
    async function loadModuleItems(module) {
      const result = await api(`/api/modulos/${module}`)
      moduleItems[module] = Array.isArray(result.items) ? result.items : []
    }
    async function loadModules() {
      try {
        await Promise.all(moduleKeys.map(loadModuleItems))
      } catch (err) {
        if (user.value) error.value = err.message
      }
    }
    async function refreshModule(module) {
      loading.value = true; error.value = ''
      try {
        await loadModuleItems(module)
      } catch (err) {
        error.value = err.message
      } finally {
        loading.value = false
      }
    }
    async function createModuleItem(module) {
      loading.value = true; error.value = ''; success.value = ''
      try {
        const result = await api(`/api/modulos/${module}`, { method: 'POST', body: JSON.stringify(moduleForms[module]) })
        moduleItems[module].unshift(result.item)
        Object.assign(moduleForms[module], { name: '', detail: '', status: moduleConfigs[module].statuses[0] })
        success.value = `${result.item.name} se agregó correctamente.`
      } catch (err) {
        error.value = err.message
      } finally {
        loading.value = false
      }
    }
    function startModuleEdit(module, item) {
      error.value = ''; success.value = ''
      Object.assign(moduleEdit, { module, id: item.id, name: item.name, detail: item.detail, status: item.status })
    }
    function cancelModuleEdit() {
      Object.assign(moduleEdit, { module: '', id: null, name: '', detail: '', status: '' })
    }
    async function updateModuleItem() {
      loading.value = true; error.value = ''; success.value = ''
      try {
        const result = await api(`/api/modulos/${moduleEdit.module}/${moduleEdit.id}`, { method: 'PUT', body: JSON.stringify({ name: moduleEdit.name, detail: moduleEdit.detail, status: moduleEdit.status }) })
        const index = moduleItems[moduleEdit.module].findIndex(item => item.id === result.item.id)
        if (index !== -1) moduleItems[moduleEdit.module].splice(index, 1, result.item)
        success.value = `${result.item.name} se actualizó correctamente.`
        cancelModuleEdit()
      } catch (err) {
        error.value = err.message
      } finally {
        loading.value = false
      }
    }
    async function deleteModuleItem(module, item) {
      if (!window.confirm(`¿Eliminar ${item.name}? Esta acción no se puede deshacer.`)) return
      loading.value = true; error.value = ''; success.value = ''
      try {
        await api(`/api/modulos/${module}/${item.id}`, { method: 'DELETE' })
        moduleItems[module] = moduleItems[module].filter(entry => entry.id !== item.id)
        success.value = `${item.name} se eliminó correctamente.`
      } catch (err) {
        error.value = err.message
      } finally {
        loading.value = false
      }
    }
    function hydrateProfileForm() {
      if (!user.value) return
      Object.assign(profileForm, {
        name: user.value.name,
        email: `operaciones@${user.value.company.toLowerCase()}.mx`,
        phone: '+52 442 100 0001',
        company: user.value.company,
        position: user.value.role === 'Administrador' ? 'Administrador de red' : user.value.role,
      })
    }
    function saveProfileSettings() {
      const name = profileForm.name.trim()
      const company = profileForm.company.trim()
      if (name.length < 3 || company.length < 2) {
        error.value = 'Completa el nombre y la empresa para guardar los cambios.'
        return
      }
      user.value = { ...user.value, name, company }
      error.value = ''
      success.value = 'Cambios de perfil aplicados a esta sesión.'
    }
    function saveSecuritySettings() {
      error.value = ''
      success.value = 'Preferencias de seguridad aplicadas a esta sesión.'
    }
    function saveSystemSettings() {
      error.value = ''
      success.value = 'Preferencias del sistema aplicadas a esta sesión.'
    }
    function showSettingsNotice(message) {
      error.value = ''
      success.value = message
    }
    function readPreferredTheme() {
      try {
        return window.localStorage.getItem('ftth-ui-theme') === 'light' ? 'light' : 'dark'
      } catch (_) {
        return 'dark'
      }
    }
    function applyTheme(nextTheme) {
      const normalizedTheme = nextTheme === 'light' ? 'light' : 'dark'
      theme.value = normalizedTheme
      document.documentElement.style.colorScheme = normalizedTheme
      document.querySelector('#app')?.classList.toggle('theme-light', normalizedTheme === 'light')
      const themeColor = document.querySelector('meta[name="theme-color"]')
      if (themeColor) themeColor.setAttribute('content', normalizedTheme === 'light' ? '#f4f7fb' : '#090a18')
      try {
        window.localStorage.setItem('ftth-ui-theme', normalizedTheme)
      } catch (_) {
        // La apariencia sigue funcionando aunque el navegador no permita persistirla.
      }
    }
    function toggleTheme() {
      applyTheme(theme.value === 'dark' ? 'light' : 'dark')
    }
    function userDetailFrom(editor) {
      return [editor.email, editor.role, editor.company, editor.phone].map(value => String(value || '').trim()).join('|')
    }
    function resetUserEditor() {
      Object.assign(userEditor, { id: null, name: '', email: '', role: 'Técnico', company: user.value?.company || 'TERACOM', phone: '', status: 'Activo' })
    }
    function openNewUserEditor() {
      error.value = ''; success.value = ''
      resetUserEditor()
      userEditorOpen.value = true
    }
    function openUserEditor(entry) {
      error.value = ''; success.value = ''
      Object.assign(userEditor, { id: entry.id, name: entry.name, email: entry.email, role: entry.role, company: entry.company, phone: entry.phone, status: entry.status })
      userEditorOpen.value = true
    }
    function closeUserEditor() {
      userEditorOpen.value = false
      resetUserEditor()
    }
    async function saveUser() {
      loading.value = true; error.value = ''; success.value = ''
      const payload = { name: userEditor.name.trim(), detail: userDetailFrom(userEditor), status: userEditor.status }
      try {
        if (userEditor.id) {
          const result = await api(`/api/modulos/customers/${userEditor.id}`, { method: 'PUT', body: JSON.stringify(payload) })
          const index = moduleItems.customers.findIndex(item => item.id === result.item.id)
          if (index !== -1) moduleItems.customers.splice(index, 1, result.item)
          success.value = `${result.item.name} se actualizó correctamente.`
        } else {
          const result = await api('/api/modulos/customers', { method: 'POST', body: JSON.stringify(payload) })
          moduleItems.customers.unshift(result.item)
          success.value = `${result.item.name} se agregó correctamente.`
        }
        closeUserEditor()
      } catch (err) {
        error.value = err.message
      } finally {
        loading.value = false
      }
    }
    async function toggleUserStatus(entry) {
      loading.value = true; error.value = ''; success.value = ''
      try {
        const nextStatus = entry.status === 'Activo' ? 'Inactivo' : 'Activo'
        const result = await api(`/api/modulos/customers/${entry.id}`, { method: 'PUT', body: JSON.stringify({ name: entry.name, detail: userDetailFrom(entry), status: nextStatus }) })
        const index = moduleItems.customers.findIndex(item => item.id === result.item.id)
        if (index !== -1) moduleItems.customers.splice(index, 1, result.item)
        success.value = `${entry.name} ahora está ${nextStatus.toLowerCase()}.`
      } catch (err) {
        error.value = err.message
      } finally {
        loading.value = false
      }
    }
    function clientDetailFrom(editor) {
      return [editor.phone, editor.email, editor.address, editor.coordinates, editor.plan]
        .map(value => String(value || '').trim().replace(/\|/g, ' '))
        .join('|')
    }
    function resetClientEditor() {
      Object.assign(clientEditor, { id: null, name: '', phone: '', email: '', address: '', coordinates: '', plan: '100 Mbps', status: 'Activo' })
    }
    function openNewClientEditor() {
      error.value = ''; success.value = ''
      resetClientEditor()
      clientEditorOpen.value = true
    }
    function openClientEditor(entry) {
      error.value = ''; success.value = ''
      Object.assign(clientEditor, {
        id: entry.id,
        name: entry.name,
        phone: entry.phone === 'Sin teléfono' ? '' : entry.phone,
        email: entry.email,
        address: entry.address === 'Sin dirección registrada' ? '' : entry.address,
        coordinates: entry.coordinates,
        plan: entry.plan,
        status: entry.status,
      })
      clientEditorOpen.value = true
    }
    function closeClientEditor() {
      clientEditorOpen.value = false
      resetClientEditor()
    }
    async function saveClient() {
      loading.value = true; error.value = ''; success.value = ''
      const payload = { name: clientEditor.name.trim(), detail: clientDetailFrom(clientEditor), status: clientEditor.status }
      try {
        if (clientEditor.id) {
          const result = await api(`/api/modulos/commercial/${clientEditor.id}`, { method: 'PUT', body: JSON.stringify(payload) })
          const index = moduleItems.commercial.findIndex(item => item.id === result.item.id)
          if (index !== -1) moduleItems.commercial.splice(index, 1, result.item)
          success.value = `${result.item.name} se actualizó correctamente.`
        } else {
          const result = await api('/api/modulos/commercial', { method: 'POST', body: JSON.stringify(payload) })
          moduleItems.commercial.unshift(result.item)
          success.value = `${result.item.name} se agregó correctamente.`
        }
        closeClientEditor()
      } catch (err) {
        error.value = err.message
      } finally {
        loading.value = false
      }
    }
    function filterUsersByRole(role) {
      roleFilter.value = roleFilter.value === role ? '' : role
    }
    function moduleMetrics(module) {
      return moduleConfigs[module].metrics.map(metric => ({
        ...metric,
        value: metric.status ? moduleItems[module].filter(item => item.status === metric.status).length : moduleItems[module].length,
      }))
    }
    function selectNetworkNode(node) { selectedNetworkNodeKey.value = node.key }
    function clearNetworkNodeSelection() { selectedNetworkNodeKey.value = '' }
    function openMapEditor(type) {
      mapTool.value = type
      Object.assign(mapEditor, { type, name: '', detail: '', status: 'Operativa', zone: 'Centro', x: 50, y: 50 })
      globalSearch.value = ''
      error.value = ''
      success.value = ''
      mapEditorOpen.value = true
    }
    function closeMapEditor() {
      mapEditorOpen.value = false
      mapTool.value = 'manage-plans'
    }
    function placeMapElement(event) {
      if (!mapEditorOpen.value) return
      const bounds = event.currentTarget.getBoundingClientRect()
      mapEditor.x = Math.min(96, Math.max(4, Math.round(((event.clientX - bounds.left) / bounds.width) * 100)))
      mapEditor.y = Math.min(94, Math.max(6, Math.round(((event.clientY - bounds.top) / bounds.height) * 100)))
    }
    async function saveMapElement() {
      const name = mapEditor.name.trim()
      const zone = mapEditor.zone.trim()
      if (name.length < 3 || zone.length < 2) {
        error.value = 'Indica un nombre y una zona para el elemento.'
        return
      }
      loading.value = true
      error.value = ''
      success.value = ''
      try {
        const detail = mapEditor.detail.trim() || `Elemento FTTH ubicado en ${zone}.`
        const result = await api('/api/modulos/network', { method: 'POST', body: JSON.stringify({ name, detail, status: mapEditor.status, map: { type: mapEditor.type, zone, x: mapEditor.x, y: mapEditor.y } }) })
        moduleItems.network.unshift(result.item)
        selectedNetworkNodeKey.value = `network-${result.item.id}`
        success.value = `${result.item.name} se agregó al mapa.`
        closeMapEditor()
      } catch (err) {
        error.value = err.message
      } finally {
        loading.value = false
      }
    }
    function resetNmsOltEditor() {
      Object.assign(nmsOltEditor, { name: '', zone: 'Centro', status: 'Operativa', detail: '' })
    }
    function openNmsOltEditor() {
      error.value = ''
      success.value = ''
      resetNmsOltEditor()
      nmsOltEditorOpen.value = true
    }
    function closeNmsOltEditor() {
      nmsOltEditorOpen.value = false
      resetNmsOltEditor()
    }
    async function saveNmsOLT() {
      const name = nmsOltEditor.name.trim()
      if (name.length < 3) {
        error.value = 'Escribe un nombre de al menos 3 caracteres para la OLT.'
        return
      }
      const mapPositionByZone = { Centro: [20, 30], Norte: [77, 27], Sur: [68, 74], Jardines: [37, 18], Parque: [88, 60] }
      const [x, y] = mapPositionByZone[nmsOltEditor.zone] || [50, 50]
      loading.value = true
      error.value = ''
      success.value = ''
      try {
        const detail = nmsOltEditor.detail.trim() || `OLT asignada a la zona ${nmsOltEditor.zone}.`
        const result = await api('/api/modulos/network', { method: 'POST', body: JSON.stringify({ name, detail, status: nmsOltEditor.status, map: { type: 'OLT', zone: nmsOltEditor.zone, x, y } }) })
        moduleItems.network.unshift(result.item)
        selectedNmsOLTId.value = String(result.item.id)
        selectedNmsPort.value = 'P1'
        selectedTerminal.value = null
        success.value = `${result.item.name} se agregó al monitoreo y al mapa.`
        closeNmsOltEditor()
      } catch (err) {
        error.value = err.message
      } finally {
        loading.value = false
      }
    }
    function zoomMap(delta) { mapZoom.value = Math.min(1.22, Math.max(.9, Number((mapZoom.value + delta).toFixed(2)))) }
    function resetMapView() { mapZoom.value = 1; clearNetworkNodeSelection(); networkLayer.value = 'all'; globalSearch.value = '' }
    function selectNmsPort(portID) {
      selectedNmsPort.value = portID
      selectedTerminal.value = null
    }
    function selectNmsOLT(oltID) {
      selectedNmsOLTId.value = String(oltID)
      selectedNmsPort.value = 'P1'
      selectedTerminal.value = null
    }
    function selectTerminal(terminal) { selectedTerminal.value = terminal }
    function terminalStatusLabel(status) {
      return ({ ok: 'Dentro de parámetros', warning: 'En el límite', danger: 'Fuera de parámetro', offline: 'No conectado', 'empty-state': 'No asignado' })[status] || 'Sin lectura'
    }
    function navigateTo(nextView) {
      view.value = nextView
      mobileMenuOpen.value = false
    }
    function editSelectedNetworkNode() {
      const node = selectedNetworkNode.value
      if (!node) return
      if (node.readOnly) {
        navigateTo('inventory')
        return
      }
      navigateTo(node.module)
      startModuleEdit(node.module, node.record)
    }
    const moduleStatusClass = status => ({ Atención: 'high', Reabastecer: 'high', Pendiente: 'high', 'Pendiente de pago': 'high', Suspendido: 'high', 'En revisión': 'medium', Programadas: 'medium', 'En análisis': 'medium', 'En tránsito': 'medium', Inactivo: 'inactive', Operativa: 'low', Disponible: 'low', Aprobada: 'low', Activo: 'low' }[status] || '')
    const formatDate = value => new Intl.DateTimeFormat('es-MX', { dateStyle: 'medium', timeStyle: 'short' }).format(new Date(value))
    onMounted(() => {
      applyTheme(theme.value)
      restoreSession()
    })
    return { user, view, mobileMenuOpen, loading, error, success, credentials, globalSearch, theme, settingsTab, profileForm, securityForm, systemForm, homeModules, moduleConfigs, moduleItems, moduleForms, moduleEdit, currentModule, userInitial, currentDate, currentDateLong, nmsPorts, nmsReferenceSnapshot, selectedNmsPort, activeNmsPort, activeNmsTelemetry, nmsTotalTraffic, nmsOLTs, selectedNmsOLTId, activeNmsOLT, nmsOltEditorOpen, nmsOltEditor, nmsTerminals, nmsOccupancy, nmsAlertCount, selectedTerminal, networkLayer, mapTool, mapZoom, mapEditorOpen, mapEditor, selectedNetworkNode, networkMapNodes, visibleNetworkMapNodes, networkTopologyZones, networkBackboneLinks, networkAccessLinks, networkMapLinks, networkOLTCount, networkONTCount, networkAttentionNodes, title, sectionLabel, weeklyOperationTrend, operationChartPoints, operationOpenedPolyline, operationResolvedPolyline, operationOpenedArea, weeklyOpened, weeklyResolved, resolutionRate, dashboardMetrics, dashboardAlerts, dashboardStats, dashboardGrowthTrend, dashboardGrowthPoints, dashboardGrowthPolyline, dashboardGrowthArea, dashboardAvailabilityTrend, dashboardAvailabilityPoints, dashboardAvailabilityPolyline, dashboardAvailabilityArea, roleCatalog, userSearch, roleFilter, userEditorOpen, userEditor, clientEditorOpen, clientEditor, directoryUsers, filteredDirectoryUsers, userRoleCards, activeDirectoryUsers, clientDirectory, activeClients, pendingClients, login, logout, navigateTo, loadModuleItems, refreshModule, createModuleItem, startModuleEdit, cancelModuleEdit, updateModuleItem, deleteModuleItem, moduleMetrics, hydrateProfileForm, saveProfileSettings, saveSecuritySettings, saveSystemSettings, showSettingsNotice, toggleTheme, openNewUserEditor, openUserEditor, closeUserEditor, saveUser, toggleUserStatus, openNewClientEditor, openClientEditor, closeClientEditor, saveClient, filterUsersByRole, selectNetworkNode, clearNetworkNodeSelection, openMapEditor, closeMapEditor, placeMapElement, saveMapElement, openNmsOltEditor, closeNmsOltEditor, saveNmsOLT, zoomMap, resetMapView, selectNmsPort, selectNmsOLT, selectTerminal, terminalStatusLabel, editSelectedNetworkNode, moduleStatusClass, formatDate }
  },
}).mount(selector)
}
