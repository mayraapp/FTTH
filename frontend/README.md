# Frontend FTTH

`src/` contiene los archivos fuente de la interfaz Vue. Está organizado por responsabilidad: `components/`, `models/`, `pages/`, `plugins/`, `router/`, `services/`, `stores/` y `styles/`. El servidor Go no publica esta carpeta directamente: publica el resultado de `npm run build` desde `../internal/web/dist`.

La compilación actual no requiere paquetes adicionales porque usa Vue como módulo CDN; el script conserva la separación entre fuente y artefacto publicado.
