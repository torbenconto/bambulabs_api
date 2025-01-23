package bambulabs_api

type PrinterConfig struct {
	Host         string
	AccessCode   string
	SerialNumber string
	MqttUser     string
	Mode         ConnectionMode
}
