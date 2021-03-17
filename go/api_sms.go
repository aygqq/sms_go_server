/*
 * API для взаимодействия с STM32MP1
 *
 * Данное API чото гдето зочемто нужно, не очень понятно. Но пусть будет, что мешает штоли?
 *
 * API version: 1.0
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */
package swagger

import (
	"encoding/json"
	"fmt"
	"net/http"

	"../control"
)

func GetSmsModemSt(w http.ResponseWriter, r *http.Request) {
	var res RespSmsmodemstateResults
	var resp RespSmsmodemstate

	control.FlagHTTPWaitResp = true
	control.SendShort(control.CMD_REQ_MODEM_INFO, 0)
	status, ret := waitForResponce(5)

	if ret == true {
		res.Status = control.ModemSt.Status
		res.Phone = control.ModemSt.Phone
		res.Imei = control.ModemSt.Imei
		res.Iccid = control.ModemSt.Iccid
		resp.Results = &res
	}
	resp.Status = status

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	str, _ := json.Marshal(resp)
	fmt.Fprintf(w, string(str))
}

func GetSmsMessage(w http.ResponseWriter, r *http.Request) {
	var res RespSmsResults
	var resp RespSms

	resp.Status = "OK"

	if control.SmsList.Len() > 0 {
		e := control.SmsList.Front()
		sms, ok := e.Value.(*control.SmsMessage)
		if !ok {
			resp.Status = "EXECUTE_ERROR"
		} else {
			res.Phone = sms.Phone
			res.Message = sms.Message
			resp.Results = &res

			control.SmsList.Remove(e)
		}
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	str, _ := json.Marshal(resp)
	fmt.Fprintf(w, string(str))
}

func SetSendSms(w http.ResponseWriter, r *http.Request) {
	var res RespSmsResults
	var resp RespSms

	phone, sms, err := parsePhoneSms(r)

	var smsMes control.SmsMessage

	if err == 0 {
		smsMes.Phone = phone
		smsMes.Message = sms
		control.FlagHTTPWaitResp = true
		control.SendSmsMessage(&smsMes)
		status, ret := waitForResponce(21)
		if ret == true {
			res.Phone = phone
			res.Message = sms
			resp.Results = &res
		}
		resp.Status = status
	} else {
		resp.Status = "INVALID_REQUEST"
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	str, _ := json.Marshal(resp)
	fmt.Fprintf(w, string(str))
}
