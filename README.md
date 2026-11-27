# FTTH

Proyecto Go y Vue para operar una red FTTH: GIS e inventario, órdenes de trabajo, aprovisionamiento, monitoreo NMS y gestión comercial BSS.

```text
FTTH/
├── backend/
│   ├── cmd/          # punto de entrada del servidor
│   ├── config/       # puerto y rutas de ejecución
│   ├── database/     # persistencia JSON local
│   ├── domain/       # modelos de negocio
│   ├── handler/      # autenticación y API HTTP
│   ├── middleware/   # cabeceras y protecciones HTTP
│   ├── repository/   # lectura y escritura de datos
│   └── router/       # registro centralizado de rutas
├── frontend/
│   ├── src/          # App, pages, components, router, stores y servicios
│   └── scripts/      # publicación de los estáticos
├── internal/web/dist/ # interfaz publicada por Go
└── data/             # solicitudes y módulos persistidos
```

## Ejecutar

```bash
cd /home/may/Documentos/ChatGPT/Yerry/FTTH
go run ./backend/cmd
```

Abre `http://127.0.0.1:8081`.

Credenciales iniciales:

- Empresa: `TERACOM`
- Usuario: `admin`
- Contraseña: `FTTH2026!`

Las órdenes de trabajo se conservan en `data/solicitudes.json`. GIS e inventario, aprovisionamiento, monitoreo NMS y gestión comercial permiten crear, editar y eliminar registros; sus datos se conservan en `data/modulos.json`.

## Publicar frontend

Después de editar la interfaz, actualiza el artefacto que sirve Go:

```bash
cd frontend
npm run build
```
