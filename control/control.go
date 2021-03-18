package control

import (
	"container/list"
	"errors"

	// "flag"
	"log"
	"os/exec"

	// "reflect"
	"strings"
	"time"
)

// ModemSt - Modem params
var ModemSt ModemState

// SmsList - List of recieved sms messages
var SmsList *list.List

// WhiteList - Phones white list
var WhiteList []ListElement = make([]ListElement, 0, 50)

// HTTPReqChan - Chanel to proceed reply to API
var HTTPReqChan chan uint8 = make(chan uint8)

// ControlReqChan - Chanel to proceed reply to control
var ControlReqChan chan uint8 = make(chan uint8)

// FlagHTTPWaitResp - What chanel is in use
var FlagHTTPWaitResp bool = false

// FlagControlWaitResp - What chanel is in use
var FlagControlWaitResp bool = false

func waitForResponce() error {
	var err error

	FlagControlWaitResp = true

	select {
	case read := <-ControlReqChan:
		if read == 0 {
			err = errors.New("Wrong response received")
		}
		// log.Printf("Chanel recv %d\n", read)
	case <-time.After(2 * time.Second):
		log.Println("No response received")
		err = errors.New("No response received")
	}

	FlagControlWaitResp = false

	return err
}

func carIDCheckFormat(carID string) error {
	var err error = nil
	if len(carID) < 10 || len(carID) > 13 {
		log.Printf("Wrong carID length %d", len(carID))
		err = errors.New("Wrong carID length")
	}
	return err
}

func procRecvSms(sms SmsMessage) {
	var answer SmsMessage
	answer.Phone = sms.Phone

	log.Printf("Message from %s, content: %s", sms.Phone, sms.Message)

	idx := SearchWhiteListByPhone(sms.Phone)
	if idx < 0 {
		log.Printf("Input number %s is not in white list\r\n", sms.Phone)
		answer.Message = "Ошибка. Вашего номера нет в белом списке."
		SendSmsMessage(&answer)
		SmsList.PushBack(&sms)
		return
	}

	if strings.HasPrefix(sms.Message, "Добавить: ") {
		carID := sms.Message[len("Добавить: "):len(sms.Message)]
		carID = strings.Trim(carID, " ")
		log.Println("Car number is: ", carID)

		err := carIDCheckFormat(carID)
		if err != nil {
			log.Println(err)
			answer.Message = "Ошибка. Неверный формат номера"
			SendSmsMessage(&answer)
			return
		}

		answer.Message = "Номер добавлен в базу на пол часа"
		SendSmsMessage(&answer)
	} else if strings.HasPrefix(sms.Message, "Удалить: ") {
		carID := sms.Message[len("Удалить: "):len(sms.Message)]
		carID = strings.Trim(carID, " ")
		log.Println("Car number is: ", carID)

		err := carIDCheckFormat(carID)
		if err != nil {
			log.Println(err)
			answer.Message = "Ошибка. Неверный формат номера"
			SendSmsMessage(&answer)
			return
		}

		answer.Message = "Номер удален из базы"
		SendSmsMessage(&answer)
	} else {
		answer.Message = "Ошибка. Неверный формат сообщения"
		SendSmsMessage(&answer)
		SmsList.PushBack(&sms)
	}
}

// ProcStart function
func ProcStart() error {
	err := readPhonesFile()
	if err != nil {
		log.Printf("Failed to read file: %q\n", err)
		FlagControlWaitResp = true
		SendCommand(CMD_PC_READY, true)
		waitForResponce()
		return err
	}

	err = checkPhonesFile()
	if err != nil {
		log.Printf("Failed to read file: %q\n", err)
		FlagControlWaitResp = true
		SendCommand(CMD_PC_READY, true)
		waitForResponce()
		return err
	}

	WritePhonesFile()

	FlagControlWaitResp = true
	SendCommand(CMD_PC_READY, true)
	if err = waitForResponce(); err != nil {
		return err
	}

	return nil
}

func procShutdown() {
	err := exec.Command("/bin/sh", "shutdown.sh").Run()
	if err != nil {
		log.Println(err)
	}
}
