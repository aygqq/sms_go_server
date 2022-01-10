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

var blinkMillis time.Duration = 75
var ErrorSt ErrorStates
var prevErr ErrorStates
var errCount int = 0
var sysInited bool = false

func waitForResponce() error {
	var err error

	FlagControlWaitResp = true

	select {
	case read := <-ControlReqChan:
		if read == 0 {
			err = errors.New("M4: Wrong response received")
		}
		SetErrorState(&ErrorSt.connM4, false)
		// log.Printf("Chanel recv %d\n", read)
	case <-time.After(2 * time.Second):
		err = errors.New("M4: No response received")
		SetErrorState(&ErrorSt.connM4, true)
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

func SetErrorState(curErr *bool, state bool) {
	*curErr = state
	if ErrorSt.connGsm && !prevErr.connGsm {
		log.Println("Set Error connection GSM")
	} else if !ErrorSt.connGsm && prevErr.connGsm {
		log.Println("Unset Error connection GSM")
	}
	prevErr.connGsm = ErrorSt.connGsm

	if ErrorSt.connM4 && !prevErr.connM4 {
		log.Println("Set Error connection M4")
	} else if !ErrorSt.connM4 && prevErr.connM4 {
		log.Println("Unset Error connection M4")
	}
	prevErr.connM4 = ErrorSt.connM4

	if ErrorSt.connBase && !prevErr.connBase {
		log.Println("Set Error Database")
	} else if !ErrorSt.connBase && prevErr.connBase {
		log.Println("Unset Error Database")
	}
	prevErr.connBase = ErrorSt.connBase
}

func setLedRed(state bool) {
	if state {
		exec.Command("gpioset", "gpiochip0", "13=0").Run()
	} else {
		exec.Command("gpioset", "gpiochip0", "13=1").Run()
	}
}

func setLedGreen(state bool) {
	if state {
		exec.Command("gpioset", "gpiochip0", "14=0").Run()
	} else {
		exec.Command("gpioset", "gpiochip0", "14=1").Run()
	}
}

func blinkRedLedOnce() time.Duration {
	setLedRed(true)
	time.Sleep(time.Millisecond * blinkMillis)
	setLedRed(false)
	time.Sleep(time.Millisecond * blinkMillis * 3)
	return blinkMillis * 4
}

// period is 5s
func blinkLedRed() {
	var delayMs time.Duration = 1000
	for {
		if ErrorSt.Global {
			setLedGreen(false)
			setLedRed(true)
			time.Sleep(time.Second * 5)
			continue
		}
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
		}
		time.Sleep(time.Millisecond * delayMs)

		delayMs = 1000
		if ErrorSt.connM4 {
			setLedGreen(false)
			for i := 0; i < 2; i++ {
				delayMs -= blinkRedLedOnce()
			}
		}
		time.Sleep(time.Millisecond * delayMs)

		delayMs = 1000
		if ErrorSt.connBase {
			setLedGreen(false)
			for i := 0; i < 3; i++ {
				delayMs -= blinkRedLedOnce()
			}
		}
		time.Sleep(time.Millisecond * delayMs)

		time.Sleep(time.Second * 2)

		if !ErrorSt.connGsm && !ErrorSt.connM4 {
			errCount = 0
		} else {
			errCount++
		}

		if sysInited && errCount > 72 {
			log.Printf("Some error is active about 6 minutes (%t, %t, %t)", ErrorSt.connGsm, ErrorSt.connM4, ErrorSt.connBase)
			procRestart()
		}
	}
}

func CheckModemState() {
	for {
		FlagControlWaitResp = true
		SendShort(CMD_REQ_MODEM_INFO, 0)
		waitForResponce()

		if ModemSt.Status != 1 && ModemSt.Status != 5 {
			SetErrorState(&ErrorSt.connGsm, true)
		} else {
			SetErrorState(&ErrorSt.connGsm, false)
		}
		time.Sleep(time.Second * 30)
	}
}

func SuperuserInform(text string) {
	var msg SmsMessage
	var cnt int = 0

	for cnt < 12 {
		if ModemSt.Status != 1 && ModemSt.Status != 5 {
			time.Sleep(10 * time.Second)
			cnt++
			continue
		} else {
			break
		}
	}
	time.Sleep(5 * time.Second)

	if dbCfg.sudo_sms {
		msg.Phone = dbCfg.superuser
		msg.Message = text
		SendSmsMessage(&msg)
	}
}

func UpdateTime() {
	err, macroscopTime := getMacroscopTime()
	if err != nil {
		log.Println("Failed to get Macroscop time")
		return
	}
	log.Println("Recieved DateTime " + macroscopTime)

	pcDateTime := strings.Split(macroscopTime, " ")
	if len(pcDateTime) < 2 {
		return
	}
	pcDate := strings.Split(pcDateTime[0], ".")
	if len(pcDate) < 3 {
		return
	}

	mpDate := pcDate[2] + "-" + pcDate[1] + "-" + pcDate[0]
	mpTime := pcDateTime[1]

	mpDateTime := mpDate + " " + mpTime
	log.Println("Converted DateTime " + mpDateTime)

	procSetTime(mpDateTime)
}

func procRecvSms(sms SmsMessage) {
	var answer SmsMessage
	answer.Phone = sms.Phone

	log.Printf("SMS message from %s, content: %s", sms.Phone, sms.Message)
	if ErrorSt.Global || !sysInited {
		log.Println("Can't process message. Init not finished or finished with error.")
		return
	}

	idx := SearchWhiteListByPhone(sms.Phone)
	if idx < 0 {
		log.Printf("Input number %s is not in white list\r\n", sms.Phone)
		// SmsList.PushBack(&sms)
		return
	}

	nPlate := sms.Message[0:len(sms.Message)]
	nPlate = strings.Trim(nPlate, " ")
	nPlate = strings.ReplaceAll(nPlate, " ", "")

	nPlate, err := nPlateCheckAndFormat(nPlate)
	if err != nil {
		log.Printf("Failed to parse car plate: %s", err)
		answer.Message = "Ошибка. Неверный формат номера"
		SendSmsMessage(&answer)
		return
	}

	// UpdateTime()
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
	time.Sleep(3 * time.Second)

	err := readConfigFile()
	if err != nil {
		log.Printf("Failed to read config file: %q\n", err)
		return err
	}
	UpdateTime()

	err = ReadPhonesFile()
	if err != nil {
		log.Printf("Failed to read phones file: %q\n", err)
		var elem ListElement
		elem.Phone = ""
		elem.Surname = ""
		elem.Name = ""
		elem.Patronymic = ""
		elem.Role = ""
		elem.AreaNum = ""
		WhiteList = append(WhiteList, elem)
	}

	checkPhonesFile(&WhiteList)
	WritePhonesFile(&WhiteList)

	err = errors.New("Failed to check or create group")
	for i := 0; i < 5; i++ {
		if dbCheckAndCreateGroup(ourGroupName) {
			err = nil
			break
		}
		time.Sleep(time.Minute)
	}
	if err != nil {
		return err
	}

	err = callAt(dbClearHour, 0, 0, regularGroupClear)
	if err != nil {
		return err
	}

	FlagControlWaitResp = true
	SendCommand(CMD_PC_READY, true)
	if err = waitForResponce(); err != nil {
		return err
	}

	sysInited = true
	return nil
}

func procSetTime(time string) error {
	cmd := exec.Command("timedatectl", "set-time", time)
	output, err := cmd.Output()

	if err != nil {
		log.Printf("Failed to set time: %s\r\n", output)
		return err
	} else {
		log.Printf("Time set successfull: %s\r\n", time)
	}

	return nil
}

func procRestart() {
	log.Printf("Restart\r\n\r\n")
	err := exec.Command("shutdown", "-r", "now").Run()
	if err != nil {
		log.Println("Failed to send restart command", err)
	}
}

func procShutdown() {
	log.Printf("Shutdown\r\n\r\n")
	err := exec.Command("shutdown", "now").Run()
	if err != nil {
		log.Println("Failed to send shutdown command", err)
	}
}
