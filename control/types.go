package control

const (
	CMD_NONE           = 0
	CMD_PC_READY       = 18 // [cmd, len, 1]
	CMD_NEW_PHONES     = 19 // [cmd, len, data]
	CMD_SEND_SMS       = 20 // [cmd, len, idx, type, phone, msg]
	CMD_REQ_MODEM_INFO = 21 // [cmd, len, idx]
	CMD_REQ_CONN_INFO  = 22 // [cmd, len, idx]
	CMD_REQ_PHONES     = 23 // [cmd, len, 0]
	CMD_REQ_REASON     = 24 // [cmd, len, 0]
	CMD_OUT_SHUTDOWN   = 25 // [cmd, len, 1]
	CMD_OUT_SAVE_STATE = 26 // [cmd, len, data]
	CMD_OUT_SIM_CHANGE = 27 // [cmd, len, data]
	CMD_OUT_SMS        = 28 // [cmd, len, idx, type, phone, msg]
	CMD_OUT_AT_CMD     = 29 // [cmd, len, ]

	IMEI_SIZE   = 15
	PHONE_SIZE  = 16
	ICCID_SIZE  = 18
	OPERID_SIZE = 5

	phonesFilePath = "phones.csv"
	configFilePath = "config.csv"
)

type ListElement struct {
	Phone      string `json:"phone,omitempty"`
	Surname    string `json:"surname,omitempty"`
	Name       string `json:"name,omitempty"`
	Patronymic string `json:"patronymic,omitempty"`
	Role       string `json:"role,omitempty"`
	AreaNum    string `json:"area_num,omitempty"`
}

type FilePhones struct {
	Elem []ListElement
}

type SmsMessage struct {
	Phone   string
	Message string
}

type ModemState struct {
	Status uint8  // Modem connection status
	Phone  string // Current phone number
	Iccid  string // ICCID of current sim-card
	Imei   string // IMEI of modem
}

type ErrorStates struct {
	connGsm  bool
	connM4   bool
	connBase bool
	Global   bool
}
