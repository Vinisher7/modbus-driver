package models

type DeviceRow struct {
	DeviceID   string
	DeviceName string
	ModbusDeviceID string
	Host           string
	Port           int
	UnitID         int

	FetchID        string
	InitialAddress int
	Quantity       int
	FunctionCode   int

	Position int

	TagID          string
	TagName        string
	DataType       string
	Operation      *string
	OperationValue *string
}

type Tag struct {
	ID             string
	Name           string
	Positions      []int
	DataType       string
	Operation      *string
	OperationValue *string
}

type Fetch struct {
	ID             string
	InitialAddress int
	Quantity       int
	FunctionCode   int
	Tags           []*Tag
}

type Device struct {
	DeviceID       string
	DeviceName     string
	ModbusDeviceID string
	Host           string
	Port           int
	UnitID         int
	Fetches        []*Fetch
}

type MQTTPayload struct {
	Timestamp string `json:"TS"`
	Val       string `json:"Val"`
}

type HealthPayload struct {
	Timestamp string `json:"TS"`
	Status    string `json:"Status"`
}
