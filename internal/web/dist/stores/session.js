import { ref } from '../plugins/vue.js'

// Estado compartido de la sesión y mensajes globales de la operación.
export function useSessionState() {
  return {
    user: ref(null),
    loading: ref(false),
    error: ref(''),
    success: ref(''),
  }
}
