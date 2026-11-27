export function createApi() {
  return async function api(path, options = {}) {
    const response = await fetch(path, {
      credentials: 'same-origin',
      headers: { 'Content-Type': 'application/json', ...(options.headers || {}) },
      ...options,
    })
    const data = response.status === 204 ? {} : await response.json()
    if (!response.ok) throw new Error(data.error || 'No fue posible completar la operación')
    return data
  }
}
