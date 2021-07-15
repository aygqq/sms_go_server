package swagger

import (
	"log"
	"net/http"
	"time"

	"../control"
)

func parseNumberState(r *http.Request) (uint8, bool, uint8) {
	var idx uint8
	var state bool
	var err uint8

	for k, v := range r.URL.Query() {
		if k == "number" {
			tmp := []byte(v[0])
			idx = tmp[0] - '0'
			if idx > 2 || idx < 1 {
				err = 1
			}
			idx--
		} else if k == "state" {
			if v[0] == "true" {
				state = true
			} else if v[0] == "false" {
				state = false
			} else {
				err = 1
			}
		}
	}

	return idx, state, err
}

func parseNumberSim(r *http.Request) (uint8, uint8, uint8) {
	var idx uint8
	var num uint8
	var err uint8

	for k, v := range r.URL.Query() {
		if k == "number" {
			tmp := []byte(v[0])
			idx = tmp[0] - '0'
			if idx > 2 || idx < 1 {
				err = 1
			}
			idx--
		} else if k == "sim_num" {
			tmp := []byte(v[0])
			num = tmp[0] - '0'
			if num > 4 || num < 1 {
				err = 1
			}
		}
	}

	return idx, num, err
}

func parsePhoneSms(r *http.Request) (string, string, uint8) {
	var phone string
	var sms string
	var err uint8

	for k, v := range r.URL.Query() {
		if k == "phone" {
			phone = v[0]
			if len(phone) > control.PHONE_SIZE {
				err = 1
			}
		} else if k == "message" {
			sms = v[0]
		}
	}

	return phone, sms, err
}

func parsePhoneName(r *http.Request) (control.ListElement, uint8) {
	var err uint8
	var elem control.ListElement

	for k, v := range r.URL.Query() {
		if k == "phone" {
			elem.Phone = v[0]
			if len(elem.Phone) > control.PHONE_SIZE {
				err = 1
			}
		} else if k == "name" {
			elem.Name = v[0]
		} else if k == "surname" {
			elem.Surname = v[0]
		} else if k == "patronymic" {
			elem.Patronymic = v[0]
		} else if k == "role" {
			elem.Role = v[0]
		} else if k == "area_num" {
			elem.AreaNum = v[0]
		}
	}

	return elem, err
}

func waitForResponce(secs int) (string, bool) {
	var ret bool
	var status string

	control.FlagHTTPWaitResp = true
	select {
	case read := <-control.HTTPReqChan:
		//! COM now in echo mode, so that "read" value doesn't matter
		if read == 1 {
			status = "OK"
			ret = true
		} else {
			status = "EXECUTE_ERROR"
			ret = false
		}
		log.Printf("Chanel recv %d\n", read)
		// control.ErrorSt.connM4 = false
		// status = "OK"
		// ret = true
	case <-time.After(time.Duration(secs) * time.Second):
		log.Println("No response received")
		status = "EXECUTE_ERROR"
		// control.ErrorSt.connM4 = true
		ret = false
	}
	control.FlagHTTPWaitResp = false
	return status, ret
}
