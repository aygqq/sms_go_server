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

var dbClearHour int = 4

var blinkMillis time.Duration = 50
var ErrorSt ErrorStates

func waitForResponce() error {
	var err error

	FlagControlWaitResp = true

	select {
	case read := <-ControlReqChan:
		if read == 0 {
			err = errors.New("Wrong response received")
		}
		ErrorSt.connM4 = false
		// log.Printf("Chanel recv %d\n", read)
	case <-time.After(2 * time.Second):
		log.Println("No response received")
		err = errors.New("No response received")
		ErrorSt.connM4 = true
	}

	FlagControlWaitResp = false

	return err
}

// Вызов переданной функции раз в сутки в указанное время.
func callAt(hour, min, sec int, f func()) error {
	loc, err := time.LoadLocation("Local")
	if err != nil {
		return err
	}

	// Вычисляем время первого запуска.
	now := time.Now().Local()
	firstCallTime := time.Date(
		now.Year(), now.Month(), now.Day(), hour, min, sec, 0, loc)
	if firstCallTime.Before(now) {
		// Если получилось время раньше текущего, прибавляем сутки.
		firstCallTime = firstCallTime.Add(time.Hour * 24)
	}

	// Вычисляем временной промежуток до запуска.
	duration := firstCallTime.Sub(time.Now().Local())

	go func() {
		time.Sleep(duration)
		for {
			f()
			// Следующий запуск через сутки.
			time.Sleep(time.Hour * 24)
		}
	}()

	return nil
}

func setLedRed(state bool) {
	if state {
		err := exec.Command("gpioset", "gpiochip0", "13=0").Run()
		if err != nil {
			log.Println(err)
		}
	} else {
		err := exec.Command("gpioset", "gpiochip0", "13=1").Run()
		if err != nil {
			log.Println(err)
		}
	}
}

func setLedGreen(state bool) {
	if state {
		err := exec.Command("gpioset", "gpiochip0", "14=0").Run()
		if err != nil {
			log.Println(err)
		}
	} else {
		err := exec.Command("gpioset", "gpiochip0", "14=1").Run()
		if err != nil {
			log.Println(err)
		}
	}
}

func blinkRedLedOnce() time.Duration {
	setLedRed(true)
	time.Sleep(time.Millisecond * blinkMillis)
	setLedRed(false)
	time.Sleep(time.Millisecond * blinkMillis * 2)
	return blinkMillis * 3
}

func blinkLedRed() {
	var delayMs time.Duration = 1000
	for {
		if !ErrorSt.connGsm && !ErrorSt.connM4 && !ErrorSt.connBase {
			setLedGreen(true)
			setLedRed(false)
			time.Sleep(time.Second * 5)
			continue
		}

		delayMs = 1000
		if ErrorSt.connGsm {
			setLedGreen(false)
			for i := 0; i < 1; i++ {
				delayMs -= blinkRedLedOnce()
			}
			log.Println("Error GSM")
		}
		time.Sleep(time.Millisecond * delayMs)

		delayMs = 1000
		if ErrorSt.connM4 {
			setLedGreen(false)
			for i := 0; i < 2; i++ {
				delayMs -= blinkRedLedOnce()
			}
			log.Println("Error M4")
		}
		time.Sleep(time.Millisecond * delayMs)

		delayMs = 1000
		if ErrorSt.connBase {
			setLedGreen(false)
			for i := 0; i < 3; i++ {
				delayMs -= blinkRedLedOnce()
			}
			log.Println("Error Database")
		}
		time.Sleep(time.Millisecond * delayMs)

		time.Sleep(time.Second * 2)
	}
}

func CheckModemState() {
	for {
		FlagControlWaitResp = true
		SendShort(CMD_REQ_MODEM_INFO, 0)
		waitForResponce()

		if ModemSt.Status != 1 && ModemSt.Status != 5 {
			ErrorSt.connGsm = true
		} else {
			ErrorSt.connGsm = false
		}
		time.Sleep(time.Second * 20)
	}
}

func procRecvSms(sms SmsMessage) {
	var answer SmsMessage
	answer.Phone = sms.Phone

	log.Printf("Message from %s, content: %s", sms.Phone, sms.Message)

	idx := SearchWhiteListByPhone(sms.Phone)
	if idx < 0 {
		log.Printf("Input number %s is not in white list\r\n", sms.Phone)
		// SmsList.PushBack(&sms)
		return
	}

	nPlate := sms.Message[0:len(sms.Message)]
	nPlate = strings.Trim(nPlate, " ")
	log.Println("Car number is: ", nPlate)

	nPlate, err := nPlateCheckAndFormat(nPlate)
	if err != nil {
		log.Println(err)
		answer.Message = "Ошибка. Неверный формат номера"
		SendSmsMessage(&answer)
		return
	}

	res := dbSearchAndAddCar(WhiteList[idx], nPlate)
	if res == 1 {
		answer.Message = nPlate + " - Въезд разрешен"
		SendSmsMessage(&answer)
	} else if res == 2 {
		answer.Message = nPlate + " - Автомобиль уже существует в базе данных"
		SendSmsMessage(&answer)
	}
}

// ProcStart function
func ProcStart() error {
	go blinkLedRed()

	go CheckModemState()
	time.Sleep(time.Second)

	err := readConfigFile()
	if err != nil {
		log.Printf("Failed to read config file: %q\n", err)
		FlagControlWaitResp = true
		SendCommand(CMD_PC_READY, true)
		waitForResponce()
		return err
	}

	for {
		if !dbCheckAndCreateGroup(ourGroupName) {
			err = errors.New("Unable to create group")
		}
		if err == nil {
			break
		}
		time.Sleep(time.Minute)
	}

	err = callAt(dbClearHour, 0, 0, regularGroupClear)
	if err != nil {
		FlagControlWaitResp = true
		SendCommand(CMD_PC_READY, true)
		waitForResponce()
		return err
	}

	err = readPhonesFile()
	if err != nil {
		log.Printf("Failed to read phones file: %q\n", err)
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
