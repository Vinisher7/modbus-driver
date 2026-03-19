package models

// DeviceRow é a linha "flat" retornada pelo JOIN da query de startup
type DeviceRow struct {
	DeviceID       string
	DeviceName     string
	ModbusDeviceID string
	Host           string
	Port           int
	UnitID         int
	FetchID        string
	InitialAddress int
	Quantity       int
	FunctionCode   int
	Position       int
	TagID          string
	TagName        string
	DataType       string
	Operation      *string
	OperationValue *string
}

// Tag representa uma tag Modbus mapeada a uma posição no batch de registros
type Tag struct {
	ID             string
	Name           string
	Positions      []int
	DataType       string
	Operation      *string
	OperationValue *string
}

// Fetch representa um grupo de leitura Modbus
type Fetch struct {
	ID             string
	InitialAddress int
	Quantity       int
	FunctionCode   int
	Tags           []Tag
}

// Device agrupa todos os dados necessários para polling de um dispositivo
type Device struct {
	DeviceID       string // UUID de devices.id — usado no tópico MQTT
	DeviceName     string
	ModbusDeviceID string
	Host           string
	Port           int
	UnitID         int
	Fetches        []Fetch
}

// MQTTPayload é o formato JSON padrão de publicação
type MQTTPayload struct {
	Timestamp string `json:"TS"`
	Val       string `json:"Val"`
}
