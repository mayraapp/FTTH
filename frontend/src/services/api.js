export function createApi() {
  return async function api(path, options = {}) {
    const response = await fetch(path, {
      credentials: 'same-origin',
      headers: { 'Content-Type': 'application/json', ...(options.headers || {}) },
      ...options,
    })
    const content = response.status === 204 ? '' : await response.text()
    let data = {}

    if (content) {
      try {
        data = JSON.parse(content)
      } catch (_) {
        // Un proxy o una capa externa puede responder texto plano. No se debe
        // exponer un error de análisis técnico a quien usa la aplicación.
      }
    }

    if (!response.ok) {
      const gatewayUnauthorized = response.status === 401 && !data.error
      const message = data.error
        || (gatewayUnauthorized
          ? 'El acceso externo requiere volver a verificar el correo. Recarga la página y completa el PIN si se solicita.'
          : 'No fue posible completar la operación')
      throw new Error(message)
    }
    return data
  }
}
