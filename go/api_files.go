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
	"io/ioutil"
	"log"
	"net/http"

	"../control"
)

func FileAddElem(w http.ResponseWriter, r *http.Request) {
	var res RespFileElemResults
	var resp RespFileElem

	phone, name, err := parsePhoneName(r)

	var elem control.ListElement

	if err == 0 {
		elem.Phone = phone
		elem.Name = name
		log.Printf("Name %s, Phone %s", name, phone)
		ret := control.AddToWhiteList(elem)
		if ret == nil {
			res.Phone = phone
			res.Name = name
			resp.Results = &res
			resp.Status = "OK"
		} else {
			log.Println(ret)
			resp.Status = "EXECUTE_ERROR"
		}
	} else {
		resp.Status = "INVALID_REQUEST"
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	str, _ := json.Marshal(resp)
	fmt.Fprintf(w, string(str))
}

func FileRemoveElem(w http.ResponseWriter, r *http.Request) {
	var res RespFileElemResults
	var resp RespFileElem

	phone, name, err := parsePhoneName(r)

	var elem control.ListElement

	if err == 0 {
		elem.Phone = phone
		elem.Name = name
		idx := control.SearchWhiteList(elem)
		ret := control.RemFromWhiteListIdx(idx)
		if ret == nil {
			res.Phone = phone
			res.Name = name
			resp.Results = &res
			resp.Status = "OK"
		} else {
			log.Println(ret)
			resp.Status = "EXECUTE_ERROR"
		}
	} else {
		resp.Status = "INVALID_REQUEST"
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	str, _ := json.Marshal(resp)
	fmt.Fprintf(w, string(str))
}

func GetFilePhones(w http.ResponseWriter, r *http.Request) {
	var resp RespFilephones

	resp.Results = make([][2]string, len(control.WhiteList))
	for i := 0; i < len(control.WhiteList); i++ {
		resp.Results[i][0] = control.WhiteList[i].Phone
		resp.Results[i][1] = control.WhiteList[i].Name
	}

	resp.Status = "OK"

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	str, _ := json.Marshal(resp)
	fmt.Fprintf(w, string(str))
}

func SetFilePhones(w http.ResponseWriter, r *http.Request) {
	var resp RespFilephones

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
	}

	err = json.Unmarshal(body, &resp.Results)
	for i := 0; i < len(resp.Results); i++ {
		control.WhiteList[i].Phone = resp.Results[i][0]
		control.WhiteList[i].Name = resp.Results[i][1]
	}
	control.WritePhonesFile()

	resp.Status = "OK"

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	str, _ := json.Marshal(resp)
	fmt.Fprintf(w, string(str))
}