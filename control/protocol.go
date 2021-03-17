package control

import (
	"container/list"
	"log"
	"strings"

	// "strings"

	// "unicode"

	"../com"
	"../crc16"
)

var table *crc16.Table

// InitProtocol - Init function
func InitProtocol() {
	log.Printf("Init protocol\n")

	com.Init(recieveHandler)

	SmsList = list.New()

	//! TODO: Table must be simmilar with PCB's table
	table = crc16.MakeMyTable(crc16.CRC16_MY)
}

func SendCommand(cmdType uint8, state bool) {
	// log.Printf("SendCommand\n")
	var buf [6]byte

	buf[0] = cmdType
	buf[1] = 1
	if state {
		buf[2] = 1
	}

	crc := crc16.Checksum(buf[:3], table)
	buf[3] = uint8(crc & 0xff)
	buf[4] = uint8(crc >> 8)
	buf[5] = byte('\n')

	com.Send(buf[:])
}

func SendShort(cmdType uint8, data byte) {
	// log.Printf("SendShort\n")
	var buf [6]byte

	buf[0] = cmdType
	buf[1] = 1
	buf[2] = data

	crc := crc16.Checksum(buf[:3], table)
	buf[3] = uint8(crc & 0xff)
	buf[4] = uint8(crc >> 8)
	buf[5] = byte('\n')

	com.Send(buf[:])
}

func SendData(cmdType uint8, data []byte) {
	// log.Printf("SendData\n")
	var dataLen = len(data)

	var buf = make([]byte, dataLen+5)

	buf[0] = cmdType
	buf[1] = uint8(dataLen)
	for i := 0; i < dataLen; i++ {
		buf[2+i] = data[i]
	}

	crc := crc16.Checksum(buf[0:len(buf)-3], table)
	buf[2+dataLen] = uint8(crc & 0xff)
	buf[3+dataLen] = uint8(crc >> 8)
	buf[4+dataLen] = byte('\n')

	com.Send(buf[:])
}

func SendDoubleByte(cmdType uint8, byte1 uint8, byte2 uint8) {
	var buf [2]byte

	buf[0] = byte1
	buf[1] = byte2

	SendData(cmdType, buf[:])
}

func SendSmsMessage(sms *SmsMessage) {
	len := 2 + PHONE_SIZE + len(sms.Message)
	var buf = make([]byte, len)

	var ptr int = 0

	// Modem num
	buf[ptr] = 0
	ptr++

	// Message type (now empty)
	buf[ptr] = 0
	ptr++

	// Phone number
	copy(buf[ptr:], sms.Phone)
	ptr += PHONE_SIZE

	// Message
	copy(buf[ptr:], sms.Message)

	SendData(CMD_SEND_SMS, buf[:])
}

func recieveHandler(data []byte) {
	var crcIn uint16
	var crc [2]uint8
	if int(data[1]) != (len(data) - 5) {
		log.Printf("Wrong length %d (real %d)\n", data[1], (len(data) - 5))
		return
	}

	crcPkt := crc16.Checksum(data[:len(data)-3], table)

	crc[0] = uint8(crcPkt)
	crc[1] = uint8(crcPkt >> 8)
	if crc[0] == 0xFE {
		crc[0] = 0xFD
	}
	if crc[1] == 0xFE {
		crc[1] = 0xFD
	}
	crcPkt = uint16(crc[1]) << 8
	crcPkt += uint16(crc[0])

	crcIn = uint16(data[len(data)-2]) << 8
	crcIn += uint16(data[len(data)-3])

	if crcPkt != crcIn {
		log.Printf("Bad crc16 0x%X 0x%X\n", crcPkt, crcIn)
		return
	}
	// log.Printf("recv: ")
	// for i := 0; i < len(data)-1; i++ {
	// 	log.Printf("%02X ", data[i])
	// }
	// log.Printf("  \n")
	// // ! Return here bacause of there are blocking by channel below
	// return

	switch data[0] {
	case CMD_PC_READY:
		// log.Printf("CMD_PC_READY\n")

		if FlagHTTPWaitResp == true {
			HTTPReqChan <- data[2]
			FlagHTTPWaitResp = false
		}
		if FlagControlWaitResp == true {
			ControlReqChan <- data[2]
		}
	case CMD_SEND_SMS:
		// log.Printf("CMD_SEND_SMS\n")

		if FlagHTTPWaitResp == true {
			HTTPReqChan <- data[2]
			FlagHTTPWaitResp = false
		}
		if FlagControlWaitResp == true {
			ControlReqChan <- data[2]
		}
	case CMD_REQ_MODEM_INFO:
		// log.Printf("CMD_REQ_MODEM_INFO\n")
		var ptr int = 2

		ptr++
		ModemSt.Status = uint8(data[ptr])
		ptr++

		// var iccid = make([]byte, ICCID_SIZE)
		// copy(iccid, data[ptr:ptr+ICCID_SIZE])
		// ModemSt.Iccid = string(iccid)
		// ptr += ICCID_SIZE

		var phone = make([]byte, PHONE_SIZE)
		copy(phone, data[ptr:ptr+PHONE_SIZE])
		ModemSt.Phone = string(phone)
		ModemSt.Phone = strings.Trim(ModemSt.Phone, "\u0000")
		ptr += PHONE_SIZE

		// var imei = make([]byte, IMEI_SIZE)
		// copy(imei, data[ptr:ptr+IMEI_SIZE])
		// ModemSt.Imei = string(imei)
		// ptr += IMEI_SIZE

		if FlagHTTPWaitResp == true {
			HTTPReqChan <- 1
			FlagHTTPWaitResp = false
		}
		if FlagControlWaitResp == true {
			ControlReqChan <- 1
		}
	case CMD_OUT_SHUTDOWN:
		// log.Printf("CMD_OUT_SHUTDOWN\n")

		go procShutdown()
	case CMD_OUT_SMS:
		// log.Printf("CMD_OUT_SMS\n")

		var ptr uint8 = 2
		var sms SmsMessage

		sms.Phone = string(data[ptr : ptr+PHONE_SIZE])
		sms.Phone = strings.Trim(sms.Phone, "\u0000")
		ptr = ptr + PHONE_SIZE
		msgLen := data[1] - PHONE_SIZE - 2
		sms.Message = string(data[ptr : ptr+msgLen])
		sms.Message = strings.Trim(sms.Message, "\r\n")

		procRecvSms(sms)

		//! sms may be cleared after end of function (make(sms, 1))
		// SmsList.PushBack(&sms)
	default:
		log.Println("Unknown command")
	}
}