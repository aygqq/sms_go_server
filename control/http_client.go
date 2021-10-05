package control

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

type dbConfig struct {
	module    string
	login     string
	pw        string
	pwHash    string
	superuser string
	sudo_sms  bool

	ip   string
	port string
	addr string

	auth string
}

var dbCfg dbConfig

var ourGroupID string = ""
var carNewIdx int = 1
var ourGroupName string = "РАЗОВЫЙ"
var ourExtID string = "mp1_with_macroscop"

/*
func HttpTest() {
	dbCheckAndCreateGroup(ourGroupName)

	getAllCars()

	nPlate, _ := nPlateCheckAndFormat("cy783t198")
	dbSearchAndAddCar(nPlate)

	nPlate, _ = nPlateCheckAndFormat("tx756x198")
	dbSearchAndAddCar(nPlate)

	nPlate, _ = nPlateCheckAndFormat("py385c198")
	dbSearchAndAddCar(nPlate)

	nPlate, _ = nPlateCheckAndFormat("ka134e198")
	dbSearchAndAddCar(nPlate)

	nPlate, _ = nPlateCheckAndFormat("ta976k198")
	dbSearchAndAddCar(nPlate)

	nPlate, _ = nPlateCheckAndFormat("py123c198")
	dbSearchAndAddCar(nPlate)

	nPlate, _ = nPlateCheckAndFormat("tk797h198")
	dbSearchAndAddCar(nPlate)

	nPlate, _ = nPlateCheckAndFormat("ka348c198")
	dbSearchAndAddCar(nPlate)

	nPlate, _ = nPlateCheckAndFormat("be538y198")
	dbSearchAndAddCar(nPlate)

	nPlate, _ = nPlateCheckAndFormat("be538y198")
	dbSearchAndAddCar(nPlate)

	nPlate, _ = nPlateCheckAndFormat("to820t198")
	dbSearchAndAddCar(nPlate)

	nPlate, _ = nPlateCheckAndFormat("ck941x198")
	dbSearchAndAddCar(nPlate)

	dbGetCarsByExtID()
	getAllCars()

	// time.Sleep(time.Second)
	// getAllCars()

	dbRemoveCarsByExternalID()

	// time.Sleep(time.Second)
	dbGetCarsByExtID()
	getAllCars()

	time.Sleep(30 * time.Second)
	getAllCars()
}
*/

func regularGroupClear() {
	// dbRemoveCarsByExternalID()
	dbRemoveCarsByGroupID()
}

func dbCheckAndCreateGroup(grName string) bool {
	ourGroupID = ""
	var groups []interface{}

	log.Println("Trying to check or create group: " + grName)
	err, groups := getCarGroups()
	if err != nil {
		log.Printf("Failed to get groups, %s\r\n", err)
		return false
	}

	for _, group := range groups {
		gr := group.(map[string]interface{})
		log.Printf("\t%s:, barrier %t", gr["name"], gr["open_barrier"])
	}

	for _, group := range groups {
		gr := group.(map[string]interface{})
		if gr["name"] == grName && gr["open_barrier"] == true {
			ourGroupID = gr["id"].(string)
			log.Println("Group is already exists: " + gr["name"].(string))
			return true
		}
	}

	err, ourGroupID := addCarGroup(true, grName)
	if err != nil {
		log.Printf("Failed to add group, %s\r\n", err)
		return false
	} else {
		log.Printf("Group successfully added: %s, %s\r\n", grName, ourGroupID)
		return true
	}
}

func dbSearchAndRemoveGroup(grName string) bool {
	ourGroupID = ""
	var groups []interface{}

	log.Println("Trying to remove group: " + grName)
	err, groups := getCarGroups()
	if err != nil {
		log.Printf("Failed to remove group, %s\r\n", err)
		return false
	}

	for _, group := range groups {
		gr := group.(map[string]interface{})
		log.Printf("\t%s:, barrier %t", gr["name"], gr["open_barrier"])
		if gr["name"] == grName {
			ourGroupID = gr["id"].(string)
		}
	}

	if ourGroupID == "" {
		log.Println("Group is not exists")
		return true
	}

	err = remCarGroup(ourGroupID)
	if err == nil {
		log.Printf("Group removed: %s, %s\r\n", grName, ourGroupID)
		return true
	} else {
		log.Printf("Failed to remove group, %s\r\n", err)
		return false
	}
}

func dbSearchAndAddCar(user ListElement, nPlate string) int {
	var plates []interface{}

	log.Printf("Trying to add car %s to group %s\r\n", nPlate, ourGroupName)

	err, plates := getCarsByPlate(nPlate)
	if err != nil {
		log.Printf("Failed to get car by plate: %s\r\n", err)
		return 0
	}

	if len(plates) > 0 {
		log.Println("Car is already in database")
		return 2
	}

	// for _, plate := range plates {
	// 	onePlate := plate.(map[string]interface{})
	// 	if onePlate["external_id"] == ourExtID {
	// 		log.Println("Car is already in database")
	// 		return 2
	// 	}
	// }

	err = addCarToGroup(user, nPlate, ourExtID, ourGroupID)
	if err == nil {
		log.Printf("Car added: %s", nPlate)
		return 1
	} else {
		log.Printf("Failed to add car: %s\r\n", err)
		return 0
	}
}

func dbRemoveCarsByExternalID() bool {
	var res bool = true
	var totalCount int = 11

	log.Printf("Trying to remove all cars by extID %s\r\n", ourExtID)

	for totalCount > 10 {
		err, cars, totalCount := getCarsByExtID(ourExtID, 0, 10)
		if err != nil {
			log.Printf("Failed to get cars by extID: %s\r\n", err)
			return false
		}

		log.Printf("There are %d cars to remove\r\n", totalCount)
		for _, car := range cars {
			oneCar := car.(map[string]interface{})
			err = remCar(oneCar["id"].(string))
			if err == nil {
				log.Printf("Car removed: %s\r\n", oneCar["license_plate_number"])
			} else {
				res = false
				log.Printf("Car %s not removed: %s\r\n", oneCar["license_plate_number"], err)
			}
		}
	}
	return res
}

func dbRemoveCarsByGroupID() bool {
	var res bool = true
	var totalCount int = 11

	log.Printf("Trying to remove all cars from group %s\r\n", ourGroupName)

	for totalCount > 10 {
		err, cars, totalCount := getCarsByGroup(ourGroupID, 0, 10)
		if err != nil {
			log.Printf("Failed to get cars by group: %s\r\n", err)
			return false
		}

		log.Printf("There are %d cars to remove\r\n", totalCount)
		for _, car := range cars {
			oneCar := car.(map[string]interface{})
			err = remCar(oneCar["id"].(string))
			if err == nil {
				log.Printf("Car removed: %s\r\n", oneCar["license_plate_number"])
			} else {
				res = false
				log.Printf("Car %s not removed: %s\r\n", oneCar["license_plate_number"], err)
			}
		}
	}
	return res
}

// Not in use yet
func getCarConfig() {
	resp, err := http.Get(dbCfg.addr + "/api/carconfig?" + dbCfg.auth)
	if err != nil {
		SetErrorState(&ErrorSt.connBase, true)
		log.Printf("Bad request, err: %s\r\n", err)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		SetErrorState(&ErrorSt.connBase, true)
		log.Printf("Bad request, err: %s\r\n", err)
		return
	}
	SetErrorState(&ErrorSt.connBase, false)

	log.Println(resp.Status)
	log.Println(string(body))
	log.Println()
}

func getCarGroups() (error, []interface{}) {
	resp, err := http.Get(dbCfg.addr + "/api/cars-groups?offset=0&portion=10&" + dbCfg.auth)
	if err != nil {
		SetErrorState(&ErrorSt.connBase, true)
		return err, nil
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		SetErrorState(&ErrorSt.connBase, true)
		return err, nil
	}
	SetErrorState(&ErrorSt.connBase, false)

	var result map[string]interface{}

	json.Unmarshal(body, &result)

	if resp.Status == "200 OK" {
		groups := result["groups"].([]interface{})
		return nil, groups
	} else {
		err = errors.New(result["ErrorMessage"].(string))
		return err, nil
	}
}

func addCarGroup(openBar bool, name string) (error, string) {
	message := map[string]interface{}{
		"external_id":  ourExtID,
		"name":         name,
		"intercept":    false,
		"open_barrier": openBar,
		"color":        "008b00ff",
	}

	bytesRepresentation, err := json.Marshal(message)
	if err != nil {
		SetErrorState(&ErrorSt.connBase, true)
		return err, ""
	}

	resp, err := http.Post(dbCfg.addr+"/api/cars-groups?"+dbCfg.auth, "application/json", bytes.NewBuffer(bytesRepresentation))
	if err != nil {
		SetErrorState(&ErrorSt.connBase, true)
		return err, ""
	}
	SetErrorState(&ErrorSt.connBase, false)

	var result map[string]interface{}

	json.NewDecoder(resp.Body).Decode(&result)

	if resp.Status == "200 OK" {
		return nil, result["id"].(string)
	} else {
		err = errors.New(result["ErrorMessage"].(string))
		return err, ""
	}
}

func remCarGroup(groupID string) error {
	req, err := http.NewRequest(http.MethodDelete, dbCfg.addr+"/api/cars-groups/"+groupID+"?"+dbCfg.auth, nil)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		SetErrorState(&ErrorSt.connBase, true)
		return err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		SetErrorState(&ErrorSt.connBase, true)
		return err
	}
	SetErrorState(&ErrorSt.connBase, false)

	var result map[string]interface{}

	json.Unmarshal(body, &result)

	if resp.Status == "200 OK" {
		return nil
	} else {
		err = errors.New(result["ErrorMessage"].(string))
		return err
	}
}

func getCarsByGroup(groupID string, offset int, portion int) (error, []interface{}, int) {
	filter := "filter=group_id='" + groupID + "'&"
	count := fmt.Sprintf("offset=%d&portion=%d&", offset, portion)
	resp, err := http.Get(dbCfg.addr + "/api/cars?" + count + filter + dbCfg.auth)
	if err != nil {
		SetErrorState(&ErrorSt.connBase, true)
		return err, nil, 0
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		SetErrorState(&ErrorSt.connBase, true)
		return err, nil, 0
	}
	SetErrorState(&ErrorSt.connBase, false)

	var result map[string]interface{}

	json.Unmarshal(body, &result)

	if resp.Status == "200 OK" {
		totalCount := result["total_count"].(float64)
		plates := result["plates"].([]interface{})
		return nil, plates, int(totalCount)
	} else {
		err = errors.New(result["ErrorMessage"].(string))
		return err, nil, 0
	}
}

// Not in use yet
func getCarsByExtID(extID string, offset int, portion int) (error, []interface{}, int) {
	filter := "filter=external_id='" + extID + "'&"
	count := fmt.Sprintf("offset=%d&portion=%d&", offset, portion)
	resp, err := http.Get(dbCfg.addr + "/api/cars?" + count + filter + dbCfg.auth)
	if err != nil {
		SetErrorState(&ErrorSt.connBase, true)
		return err, nil, 0
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		SetErrorState(&ErrorSt.connBase, true)
		return err, nil, 0
	}
	SetErrorState(&ErrorSt.connBase, false)

	var result map[string]interface{}

	json.Unmarshal(body, &result)

	if resp.Status == "200 OK" {
		totalCount := result["total_count"].(float64)
		plates := result["plates"].([]interface{})
		return nil, plates, int(totalCount)
	} else {
		err = errors.New(result["ErrorMessage"].(string))
		return err, nil, 0
	}
}

func getCarsByPlate(nPlate string) (error, []interface{}) {
	filter := "filter=license_plate_number='" + url.QueryEscape(nPlate) + "'&"
	req, err := http.NewRequest(http.MethodGet, dbCfg.addr+"/api/cars?offset=0&portion=50&"+filter+dbCfg.auth, nil)
	req.Header.Add("Accept-Charset", "utf-8")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		SetErrorState(&ErrorSt.connBase, true)
		return err, nil
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		SetErrorState(&ErrorSt.connBase, true)
		return err, nil
	}
	SetErrorState(&ErrorSt.connBase, false)

	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if resp.Status == "200 OK" {
		plates := result["plates"].([]interface{})
		return nil, plates
	} else {
		err = errors.New(result["ErrorMessage"].(string))
		return err, nil
	}
}

func addCarToGroup(user ListElement, nPlate string, extID string, groupID string) error {
	message := map[string]interface{}{
		"owner": map[string]string{
			"first_name":  user.Name,
			"second_name": user.Surname,
			"third_name":  user.Patronymic,
		},
		"external_id":          extID,
		"license_plate_number": nPlate,
		"additional_info":      user.AreaNum,
		"model":                user.Role,
		"color":                "",
		"groups": []map[string]string{
			0: {"id": groupID},
		},
	}

	bytesRepresentation, err := json.Marshal(message)
	if err != nil {
		SetErrorState(&ErrorSt.connBase, true)
		return err
	}

	resp, err := http.Post(dbCfg.addr+"/api/cars?"+dbCfg.auth, "application/json", bytes.NewBuffer(bytesRepresentation))
	if err != nil {
		SetErrorState(&ErrorSt.connBase, true)
		return err
	}
	SetErrorState(&ErrorSt.connBase, false)

	var result map[string]interface{}

	json.NewDecoder(resp.Body).Decode(&result)

	if resp.Status == "200 OK" {
		return nil
	} else {
		err = errors.New(result["ErrorMessage"].(string))
		return err
	}
}

func remCar(carID string) error {
	req, err := http.NewRequest(http.MethodDelete, dbCfg.addr+"/api/cars/"+carID+"?"+dbCfg.auth, nil)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		SetErrorState(&ErrorSt.connBase, true)
		return err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		SetErrorState(&ErrorSt.connBase, true)
		return err
	}
	SetErrorState(&ErrorSt.connBase, false)

	var result map[string]interface{}

	json.Unmarshal(body, &result)

	if resp.Status == "200 OK" {
		return nil
	} else {
		err = errors.New(result["ErrorMessage"].(string))
		return err
	}
}

func getMacroscopTime() (error, string) {
	// return nil, "22.09.2018 3:33:06"
	resp, err := http.Get(dbCfg.addr + "/command?type=gettime&" + dbCfg.auth)
	// resp, err := http.Get(dbCfg.addr + "/command?type=gettime&" + dbCfg.auth + "&responsetype=json")
	if err != nil {
		SetErrorState(&ErrorSt.connBase, true)
		return err, ""
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		SetErrorState(&ErrorSt.connBase, true)
		return err, ""
	}
	SetErrorState(&ErrorSt.connBase, false)

	return nil, string(body)

	// 	var result map[string]interface{}
	// 	json.Unmarshal(body, &result)

	// 	log.Println(result)
	// 	return nil, result["time"].(string)
}
