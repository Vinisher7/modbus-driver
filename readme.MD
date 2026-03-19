# modbus-driver

Driver de polling Modbus TCP escrito em Go. Lê configurações de dispositivos e tags diretamente de um banco de dados **SQL Server**, converte os registros e publica os valores como JSON em um broker **MQTT**.

---

## Visão Geral

```
SQL Server (metadados)
        │
        ▼
  modbus-driver
  ┌─────────────────────────────┐
  │  Carrega Devices + Tags     │
  │  Poll Loop (ticker)         │
  │    └── Grupos paralelos     │
  │          └── Por device     │
  │               └── Fetches   │
  │                    └── Tags │
  └─────────────────────────────┘
        │
        ▼
  MQTT Broker
  /devices/formatted_data/erd/{deviceUUID}/{tagID}
```

O driver realiza o ciclo de polling de forma concorrente: os dispositivos são divididos em grupos (`DEVICE_GROUP_SIZE`) e cada grupo é processado em paralelo usando goroutines.

---

## Estrutura de Pastas

```
modbus-driver/
├── cmd/
│   ├── main.go          # Ponto de entrada da aplicação
│   └── .env             # Variáveis de ambiente (não versionado em produção)
├── internal/
│   ├── config/
│   │   └── config.go    # Carregamento e tipagem das configurações
│   ├── database/
│   │   └── database.go  # Conexão com SQL Server e query de metadados
│   ├── modbus/
│   │   ├── client.go    # Cliente TCP Modbus (wraps goburrow/modbus)
│   │   └── driver.go    # Loop de polling, conversão de registros e publicação
│   ├── models/
│   │   └── models.go    # Structs de domínio (Device, Fetch, Tag, MQTTPayload)
│   └── mqtt/
│       └── publisher.go # Publisher MQTT (wraps paho.mqtt.golang)
├── Dockerfile            # Build multi-stage para imagem Alpine
├── go.mod
└── go.sum
```

---

## Detalhamento dos Pacotes

### `cmd/`

Ponto de entrada da aplicação. `main.go` orquestra toda a inicialização na sequência:

1. Carrega as configurações via `config.Load()`.
2. Conecta-se ao banco de dados SQL Server.
3. Carrega a lista de dispositivos e tags do banco (`database.LoadDevices`).
4. Cria o publisher MQTT.
5. Inicia o `Driver.RunPollLoop` em uma goroutine separada.
6. Aguarda sinal de shutdown (`SIGINT` / `SIGTERM`) para encerrar graciosamente.

O arquivo `.env` (localizado em `cmd/.env`) é carregado automaticamente ao iniciar a aplicação.

---

### `internal/config/`

| Arquivo | Responsabilidade |
|---|---|
| `config.go` | Define a struct `Config` e a função `Load()` que lê as variáveis de ambiente (com suporte a `.env` via `godotenv`). |

#### Variáveis de Ambiente

| Variável | Padrão | Descrição |
|---|---|---|
| `DB_SERVER` | `localhost` | Host do SQL Server |
| `DB_PORT` | `1433` | Porta do SQL Server |
| `DB_USER` | `sa` | Usuário do banco |
| `DB_PASSWORD` | _(vazio)_ | Senha do banco |
| `DB_NAME` | _(vazio)_ | Nome do banco de dados |
| `MQTT_BROKER` | `tcp://localhost:1883` | URL do broker MQTT |
| `MQTT_CLIENT_ID` | `modbus-driver` | Client ID MQTT |
| `MQTT_USERNAME` | _(vazio)_ | Usuário MQTT (opcional) |
| `MQTT_PASSWORD` | _(vazio)_ | Senha MQTT (opcional) |
| `MODBUS_TIMEOUT` | `5s` | Timeout de conexão Modbus TCP |
| `POLL_INTERVAL` | `10s` | Intervalo entre ciclos de polling |
| `DEVICE_GROUP_SIZE` | `5` | Número de devices por grupo paralelo |

---

### `internal/database/`

| Arquivo | Responsabilidade |
|---|---|
| `database.go` | Conecta ao SQL Server e executa a query de startup que retorna todos os dispositivos, fetches e tags em um resultado "flat" (JOIN de 5 tabelas). |

#### Função `Connect`

Abre e valida a conexão com o SQL Server usando a string de conexão no formato:

```
sqlserver://<user>:<password>@<host>:<port>?database=<dbname>
```

#### Função `LoadDevices`

Executa uma query JOIN entre as tabelas:

| Tabela | Papel |
|---|---|
| `devices` | Dispositivo lógico (filtrado por `protocol = 'modbus'`) |
| `modbus_devices` | Configuração de rede Modbus (host, port, unit_id) |
| `modbus_fetchs` | Grupo de leitura (initial_address, quantity, function_code) |
| `modbus_tag_fetch` | Vínculo entre fetch e tag com posição do registrador |
| `modbus_tags` | Definição da tag (nome, data_type, operação) |

O resultado flat é reconstituído em uma hierarquia `Device → Fetch → Tag`. UUIDs retornados pelo SQL Server em formato `uniqueidentifier` (little-endian) são convertidos corretamente para string.

---

### `internal/modbus/`

#### `client.go`

Envolve a biblioteca `goburrow/modbus` para conexão TCP. Suporta os 4 function codes:

| Function Code | Operação |
|---|---|
| `1` | Read Coils |
| `2` | Read Discrete Inputs |
| `3` | Read Holding Registers |
| `4` | Read Input Registers |

O método `ReadRegisters` retorna os bytes do response como um slice de `uint16` (big-endian).

#### `driver.go`

Contém toda a lógica de orquestração de polling:

- **`RunPollLoop`** — Tick periódico baseado em `PollInterval`. A cada tick, chama `pollAllGroups`.
- **`pollAllGroups`** — Divide os devices em chunks de `DeviceGroupSize` e processa cada grupo em paralelo.
- **`processGroup`** — Processa cada device do grupo também em paralelo (goroutines).
- **`pollDevice`** — Abre uma conexão Modbus para o device, executa cada `Fetch`, converte os valores e publica no MQTT.

#### Conversão de Tipos (`combineRegisters`)

Converte um slice de `uint16` em `float64` conforme o `data_type` configurado na tag:

| `data_type` | Words | Descrição |
|---|---|---|
| `uint16` | 1 | Inteiro sem sinal 16 bits |
| `int16` | 1 | Inteiro com sinal 16 bits |
| `float32` | 2 | Float IEEE 754, words em ordem direta (big-endian) |
| `float32_swapped` | 2 | Float IEEE 754, words trocadas (little-endian) |
| `int32` | 2 | Inteiro com sinal 32 bits |
| `uint32` | 2 | Inteiro sem sinal 32 bits |
| `auto` | 1 ou 2 | Infere `uint16` (1 word) ou `float32` (2 words) |

#### Operações (`applyOperation`)

Após a conversão, é possível aplicar uma operação matemática ao valor:

| Operação | Efeito |
|---|---|
| `*` | Multiplica pelo `operation_value` |
| `/` | Divide pelo `operation_value` |
| `+` | Soma o `operation_value` |
| `-` | Subtrai o `operation_value` |

---

### `internal/models/`

| Struct | Descrição |
|---|---|
| `DeviceRow` | Linha "flat" retornada pelo JOIN da query de startup |
| `Device` | Dispositivo com lista de `Fetch` |
| `Fetch` | Grupo de leitura com lista de `Tag` |
| `Tag` | Tag com `Positions` (posições no batch de registros), `DataType`, `Operation` e `OperationValue` |
| `MQTTPayload` | Payload JSON publicado: `{ "TS": "<RFC3339>", "Val": "<valor>" }` |

---

### `internal/mqtt/`

| Arquivo | Responsabilidade |
|---|---|
| `publisher.go` | Cria e gerencia a conexão com o broker MQTT. Expõe o método `Publish` que envia o payload para o tópico correto. |

#### Tópico MQTT

```
/devices/formatted_data/erd/{deviceUUID}/{tagID}
```

- `deviceUUID` — UUID do dispositivo lógico (tabela `devices.id`)
- `tagID` — UUID da tag (tabela `modbus_tags.id`)

As mensagens são publicadas com QoS 0 e retain flag ativado.

---

## Como Executar

### Pré-requisitos

- Go 1.21+
- SQL Server acessível com as tabelas de metadados configuradas
- Broker MQTT acessível

### Execução Local

1. Configure o arquivo `cmd/.env`:

```env
DB_SERVER=192.168.0.12
DB_PORT=1433
DB_USER=sa
DB_PASSWORD=SuaSenha
DB_NAME=hydra

MQTT_BROKER=tcp://seu-broker:1883
MQTT_CLIENT_ID=modbus-driver
MQTT_USERNAME=
MQTT_PASSWORD=

MODBUS_TIMEOUT=3s
POLL_INTERVAL=4s
DEVICE_GROUP_SIZE=5
```

2. Execute:

```bash
go run ./cmd/main.go
```

### Build

```bash
go build -o modbus-driver ./cmd/main.go
./modbus-driver
```

---

## Docker

O projeto inclui um `Dockerfile` com build multi-stage que gera uma imagem mínima baseada em **Alpine Linux**.

```bash
# Build da imagem
docker build -t modbus-driver .

# Execução
docker run --rm modbus-driver
```

> **Atenção:** O `.env` é copiado para dentro da imagem durante o build (`COPY ./cmd/.env .`). Para ambientes de produção, prefira passar as variáveis via `--env-file` ou variáveis de ambiente do orquestrador.

---

## Dependências Principais

| Pacote | Uso |
|---|---|
| [`github.com/goburrow/modbus`](https://github.com/goburrow/modbus) | Cliente Modbus TCP |
| [`github.com/eclipse/paho.mqtt.golang`](https://github.com/eclipse/paho.mqtt.golang) | Cliente MQTT |
| [`github.com/microsoft/go-mssqldb`](https://github.com/microsoft/go-mssqldb) | Driver SQL Server |
| [`github.com/joho/godotenv`](https://github.com/joho/godotenv) | Leitura de arquivo `.env` |
| [`github.com/google/uuid`](https://github.com/google/uuid) | Manipulação de UUIDs |
| [`github.com/shopspring/decimal`](https://github.com/shopspring/decimal) | Precisão decimal |
