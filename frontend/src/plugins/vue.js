// Punto único de acceso a Vue para la versión sin dependencias locales.
// Cuando el frontend se migre a Vite, solo este archivo cambiará de importación.
export { createApp, computed, onMounted, reactive, ref } from 'https://unpkg.com/vue@3/dist/vue.esm-browser.prod.js'
