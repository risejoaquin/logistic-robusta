# 🍣 SUSHI LOSPLEBES | POS & WhatsApp Automation Engine

[![Go Version](https://img.shields.io/badge/Go-1.22%2B-00ADD8?style=for-the-badge&logo=go)](https://golang.org)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16%2B-336791?style=for-the-badge&logo=postgresql)](https://postgresql.org)
[![Architecture](https://img.shields.io/badge/Architecture-Zero_Overhead-black?style=for-the-badge)](#)
[![Deployment](https://img.shields.io/badge/Ready-Railway%20%7C%20Cloud_Run-6A00FF?style=for-the-badge)](#)

Un sistema monolítico hiper-eficiente diseñado para la orquestación logística y punto de venta (POS) del restaurante **Sushi Los Plebes**. 

Desarrollado bajo una arquitectura estricta de **Clean Code**, cero sobreingeniería, y orientado puramente a eventos. Combina un bot conversacional por WhatsApp con un Monitor de Cocina en Tiempo Real (SSE) y automatización de impresión térmica de tickets directos al mostrador.

---

## ⚡ Características Principales (MVP)

- **Cero Frameworks Bloated:** Construido 100% sobre la librería estándar `net/http` de Go.
- **Motor de Estados WhatsApp:** Máquina de estados embebida en base de datos para manejar el embudo de ventas asíncrono sin depender de NLP pesado.
- **Receptor Webhook Ultrarrápido:** Persistencia directa y asíncrona usando `JSONB` en PostgreSQL.
- **Transmisión de Eventos Sever-Sent (SSE):** Monitor de cocina pasivo que se actualiza al instante sin recargas.
- **Impresión Térmica Autónoma:** Script embebido nativo que formatea strings (32 caracteres / línea estilo ESC/POS) para tickets y llama a la API de pre-impresión `window.print()` vía IFrame invisible.

---

## 🏗️ Schema de la Base de Datos (PostgreSQL)

El sistema opera sobre un modelo relacional estricto optimizado para lecturas y transacciones síncronas rápidas (no ORM, puro driver nativo `pgx/v5`).

### 1. `mensajes_raw` (Data Lake Conversacional)
Bitácora estricta de auditoría y respaldo de webhooks para garantizar que ningún payload de Meta se pierda.
| Columna     | Tipo      | Atributos                   | Descripción                                 |
| ----------- | --------- | --------------------------- | ------------------------------------------- |
| `id`        | SERIAL    | PRIMARY KEY                 | Identificador único.                        |
| `telefono`  | VARCHAR   | (20)                        | Identificador del cliente.                  |
| `payload`   | JSONB     |                             | Dumping binario del body JSON enviado por Meta. |
| `creado_en` | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP   | Marca de tiempo exacta del ingreso.         |

### 2. `pedidos` (Entity transaccional y Motor de Estados)
Controla el ciclo de vida del "Embudo de Ventas".
| Columna             | Tipo      | Atributos                   | Descripción                                        |
| ------------------- | --------- | --------------------------- | -------------------------------------------------- |
| `id`                | SERIAL    | PRIMARY KEY                 | Ticket ID.                                         |
| `telefono`          | VARCHAR   | (20)                        | Ref al cliente.                                    |
| `nombre`            | VARCHAR   | (100)                       | Nombre extraído del perfil de WhatsApp.             |
| `detalles_orden`    | TEXT      |                             | Texto crudo con el pedido del cliente.              |
| `direccion_entrega` | TEXT      |                             | Modalidad y destino. (`PICKUP` o Domicilio).        |
| `metodo_pago`       | VARCHAR   | (50)                        | Identificador del medio (`EFECTIVO` / `TRANSF.`).   |
| `total`             | DECIMAL   | (10, 2)                     | (MVP) Hardcodeado, a expandir luego al carrito.     |
| `estado`            | INT       | DEFAULT 0                   | Estado del bot: 0=Nuevo, 1=Leyendo, 2=Pickup, 4=Fin|
| `creado_en`         | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP   | Registro del evento comercial.                     |

### 3. `ganancias` (Impacto Directo POS)
Reflejo atómico de ingresos para cuadres de caja y estados financieros.
| Columna     | Tipo      | Atributos                   | Descripción                                 |
| ----------- | --------- | --------------------------- | ------------------------------------------- |
| `id`        | SERIAL    | PRIMARY KEY                 | Identificador de Transacción.               |
| `pedido_id` | INT       | FOREIGN KEY -> pedidos(id)  | Relación al ticket de origen.               |
| `monto`     | DECIMAL   | (10, 2)                     | Monto real inyectado a la economía.         |
| `creado_en` | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP   | Fecha de cierre monetario.                  |

### 4. `inventario` (Gestión Lógica de Insumos)
Almacén para el log de egresos lógicos atados a ventas.
| Columna     | Tipo      | Atributos                   | Descripción                                 |
| ----------- | --------- | --------------------------- | ------------------------------------------- |
| `id`        | SERIAL    | PRIMARY KEY                 | Identificador de insumo.                    |
| `insumo`    | VARCHAR   | (100)                       | Nombre del insumo (`Arroz`, `Salmón`).      |
| `cantidad`  | INT       | DEFAULT 100                 | Inventario activo.                          |

---

## 🛠️ Mapa de Arquitectura

```text
📦 ROOT
 ┣ 📂 templates/
 ┃ ┗ 📜 monitor.html   # Interfaz UI SSE y generador de tickets ESC/POS nativo
 ┣ 📜 main.go          # Entrypoint de Go y Servidor HTTP
 ┣ 📜 db.go            # Connection Pooling y Auto-Migraciones PostgreSQL (DDL)
 ┣ 📜 webhook.go       # Listener del API WhatsApp Cloud de Meta
 ┣ 📜 bot.go           # Cerebro (Árbol de decisión/Estados) y emisor de mensajes
 ┗ 📜 monitor.go       # Endpoints Render HTML e inyector Server-Sent Events (SSE)
```

---

## 🚀 Despliegue y Ejecución

### Prerrequisitos de Entorno

1. **Golang 1.22+** instalado y en el PATH del sistema (`go version`).
2. **PostgreSQL 16+** corriendo en local o cloud.
3. Crear un proyecto en **Meta for Developers** (WhatsApp Cloud API) para obtener tokens.

### Variables de Entorno `.env` necesarias

```env
WHATSAPP_VERIFY_TOKEN="MiSuperTokenDeVerificacion123"
WHATSAPP_ACCESS_TOKEN="EAALvxxxxxxxxxxxxx"
WHATSAPP_PHONE_ID="1234567890"
DATABASE_URL="postgres://usuario:password@localhost:5432/sushipos"
PORT="8080"
```

### Ejecutar en Desarrollo (Local)

1. Inicializa el módulo e instala la dependencia de PostgreSQL:
   ```bash
   go mod init sushipos
   go get github.com/jackc/pgx/v5/pgxpool
   ```
2. Ejecuta el servidor (importante compilar todo el directorio):
   ```bash
   go run .
   ```
   *(Alternativa: `go build -o server.exe && ./server.exe`)*

### Ejecutar en Producción (ej. Railway, Render, Cloud Run)

El sistema es 100% _Container-Ready_. Configura el build de la plataforma apuntando al directorio raíz. En Railway:
1. Agrega el repo.
2. Añade un plugin de PostgreSQL.
3. Mapea las variables de entorno.
4. El comando Start será inferido como `go build` o `go run`.

---

## 🔒 Privacidad y Rendimiento

Este MVP fue pensado primariamente para mitigar costos operativos y saturación de hardware.
* Se evitó intencionalmente el uso de frameworks frontend grandes para el monitor, limitando la carga de red y priorizando el flujo constante SSE.
* Se estructuró un cierre automático (`r.Context().Done()`) para gestionar múltiples terminales de cocina sin fugas de memoria.

*Desarrollado en 2026 para redefinir el estándar local de Los Plebes.*
