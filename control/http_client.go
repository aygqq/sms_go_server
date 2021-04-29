package control

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

var login string = "mp1"
var pw string = "1234"
var pwHash string = fmt.Sprintf("%x", md5.Sum([]byte(pw)))

var addr string = "http://192.168.8.100:8080"
var auth string = "module=complete&login=" + login + "&password=" + pwHash

var singleGroupID string = ""
var carNewIdx int = 1
var singleGroupName string = "Открытие шлагбаума"
var ourExtID string = "mp2_with_macroscop"

func HttpTest() {
	dbCheckAndCreateGroup(singleGroupName)

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
}

func regularGroupClear() {
	dbRemoveCarsByExternalID()
}

func dbCheckAndCreateGroup(grName string) bool {
	singleGroupID = ""
	var groups []interface{} = getCarGroups()

	for _, group := range groups {
		gr := group.(map[string]interface{})
		if gr["name"] == grName && gr["open_barrier"] == true {
			singleGroupID = gr["id"].(string)
			log.Println("Group is already exists: " + gr["name"].(string))
			return true
		}
	}
	return false

	// singleGroupID = addCarGroup(true, "123", grName)

	// if singleGroupID == "" {
	// 	log.Println("Unable to add group")
	// 	return false
	// } else {
	// 	log.Println("Group successfully added")
	// 	return true
	// }
}

func dbSearchAndRemoveGroup(grName string) bool {
	singleGroupID = ""
	var groups []interface{} = getCarGroups()

	for _, group := range groups {
		gr := group.(map[string]interface{})
		if gr["name"] == grName {
			singleGroupID = gr["id"].(string)
			log.Println("Group is exists")
		}
	}

	if singleGroupID == "" {
		log.Println("Group is not exists")
		return true
	}

	if remCarGroup(singleGroupID) {
		log.Println("Group successfully removed")
		return true
	} else {
		log.Println("Unable to remove group")
		return false
	}
}

func dbSearchAndAddCar(nPlate string) bool {
	var plates []interface{} = getCarsByPlate(nPlate, singleGroupID)

	for _, plate := range plates {
		onePlate := plate.(map[string]interface{})
		if onePlate["external_id"] == ourExtID {
			log.Println("Car is already in database")
			return true
		}
	}

	if addCarToGroup(nPlate, ourExtID, singleGroupID) {
		log.Printf("Car %d successfully added\n", carNewIdx)
		carNewIdx++
		return true
	}
	log.Println("Unable to add car")
	return false
}

func dbRemoveAllCars() bool {
	cars := getAllCars()

	for _, car := range cars {
		oneCar := car.(map[string]interface{})
		remCar(oneCar["id"].(string))
	}
	return true
}

func dbRemoveCarsByExternalID() bool {
	cars, totalCount := getCarsByExtID(ourExtID, 0, 10)
	log.Println(totalCount)

	for _, car := range cars {
		oneCar := car.(map[string]interface{})
		remCar(oneCar["id"].(string))
	}

	for totalCount > 10 {
		cars, totalCount = getCarsByExtID(ourExtID, 0, 10)
		log.Println(totalCount)
		for _, car := range cars {
			oneCar := car.(map[string]interface{})
			remCar(oneCar["id"].(string))
		}
	}
	return true
}

func dbGetCarsFromGroup() bool {
	var offset int = 0
	var totalCount int = 10

	for totalCount > offset {
		_, totalCount = getCarsByGroup(singleGroupID, offset, 10)
		offset += 10
	}
	return true
}

func dbGetCarsByExtID() bool {
	var offset int = 0
	var totalCount int = 10

	for totalCount > offset {
		_, totalCount = getCarsByExtID(ourExtID, offset, 10)
		offset += 10
	}
	return true
}

func getCarConfig() {
	resp, err := http.Get(addr + "/api/carconfig?" + auth)
	if err != nil {
		log.Fatalln(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println(resp.Status)
	log.Println(string(body))
	log.Println()
}

func getCarGroups() []interface{} {
	resp, err := http.Get(addr + "/api/cars-groups?offset=0&portion=10&" + auth)
	if err != nil {
		log.Fatalln(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	var result map[string]interface{}

	json.Unmarshal(body, &result)

	if resp.Status == "200 OK" {
		groups := result["groups"].([]interface{})
		for _, group := range groups {
			gr := group.(map[string]interface{})
			log.Printf("\t%s: id %s, barrier %t", gr["name"], gr["id"], gr["open_barrier"])
		}
		return groups
	} else {
		log.Println(result["ErrorMessage"])
		return nil
	}
}

func addCarGroup(openBar bool, extID string, name string) string {
	message := map[string]interface{}{
		"external_id":  extID,
		"name":         name,
		"intercept":    openBar,
		"open_barrier": openBar,
		"color":        "00ffff00",
	}

	bytesRepresentation, err := json.Marshal(message)
	if err != nil {
		log.Fatalln(err)
	}

	resp, err := http.Post(addr+"/api/cars-groups?"+auth, "application/json", bytes.NewBuffer(bytesRepresentation))
	if err != nil {
		log.Fatalln(err)
	}

	var result map[string]interface{}

	json.NewDecoder(resp.Body).Decode(&result)

	if resp.Status == "200 OK" {
		log.Printf("Group added: %s, id %s\r\n", result["name"], result["id"])
		return result["id"].(string)
	} else {
		log.Println(result["ErrorMessage"])
		return ""
	}
}

func remCarGroup(groupID string) bool {
	req, err := http.NewRequest(http.MethodDelete, addr+"/api/cars-groups/"+groupID+"?"+auth, nil)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	var result map[string]interface{}

	json.Unmarshal(body, &result)

	if resp.Status == "200 OK" {
		log.Printf("Group removed: id %s\r\n", groupID)
		return true
	} else {
		log.Println(result["ErrorMessage"])
		return false
	}
}

func getAllCars() []interface{} {
	resp, err := http.Get(addr + "/api/cars?offset=0&portion=50&" + auth)
	if err != nil {
		log.Fatalln(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	var result map[string]interface{}

	json.Unmarshal(body, &result)
	// log.Println(result)

	if resp.Status == "200 OK" {
		log.Println("Cars")
		plates := result["plates"].([]interface{})
		for _, plate := range plates {
			pl := plate.(map[string]interface{})
			log.Printf("\t%s, %s, %s\r\n", pl["license_plate_number"], pl["id"], pl["external_id"])
		}
		return plates
	} else {
		log.Println(result["ErrorMessage"])
		return nil
	}
}

func getCarsByGroup(groupID string, offset int, portion int) ([]interface{}, int) {
	filter := "filter=group_id='" + groupID + "'&"
	count := fmt.Sprintf("offset=%d&portion=%d&", offset, portion)
	resp, err := http.Get(addr + "/api/cars?" + count + filter + auth)
	if err != nil {
		log.Fatalln(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	var result map[string]interface{}

	// json.NewDecoder(resp.Body).Decode(&result)
	json.Unmarshal(body, &result)

	if resp.Status == "200 OK" {
		log.Println("Cars in group " + groupID)
		totalCount := result["total_count"].(float64)
		plates := result["plates"].([]interface{})
		for _, plate := range plates {
			pl := plate.(map[string]interface{})
			log.Printf("\t%s, %s\r\n", pl["license_plate_number"], pl["id"])
		}
		return plates, int(totalCount)
	} else {
		log.Println(result["ErrorMessage"])
		return nil, 0
	}
}

func getCarsByExtID(extID string, offset int, portion int) ([]interface{}, int) {
	filter := "filter=external_id='" + extID + "'&"
	count := fmt.Sprintf("offset=%d&portion=%d&", offset, portion)
	resp, err := http.Get(addr + "/api/cars?" + count + filter + auth)
	if err != nil {
		log.Fatalln(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	var result map[string]interface{}

	// json.NewDecoder(resp.Body).Decode(&result)
	json.Unmarshal(body, &result)

	if resp.Status == "200 OK" {
		log.Println("Cars by external id " + extID)
		totalCount := result["total_count"].(float64)
		plates := result["plates"].([]interface{})
		for _, plate := range plates {
			pl := plate.(map[string]interface{})
			log.Printf("\t%s, %s\r\n", pl["license_plate_number"], pl["id"])
		}
		return plates, int(totalCount)
	} else {
		log.Println(result["ErrorMessage"])
		return nil, 0
	}
}

func getCarsByPlate(nPlate string, groupID string) []interface{} {
	filter := "filter=license_plate_number='" + url.QueryEscape(nPlate) + "'&"
	req, err := http.NewRequest(http.MethodGet, addr+"/api/cars?offset=0&portion=50&"+filter+auth, nil)
	req.Header.Add("Accept-Charset", "utf-8")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if resp.Status == "200 OK" {
		log.Println("Cars by plate " + nPlate)
		plates := result["plates"].([]interface{})
		for _, plate := range plates {
			pl := plate.(map[string]interface{})
			log.Printf("%s, %s, %s\r\n", pl["license_plate_number"], pl["id"], pl["external_id"])
		}
		return plates
	} else {
		log.Println(result["ErrorMessage"])
		return nil
	}
}

func addCarToGroup(nPlate string, extID string, groupID string) bool {
	message := map[string]interface{}{
		"owner": map[string]string{
			"first_name":  "",
			"second_name": "",
			"third_name":  "",
		},
		"external_id":          extID,
		"license_plate_number": nPlate,
		"additional_info":      "",
		"model":                "",
		"color":                "",
		"groups": []map[string]string{
			0: {"id": groupID},
		},
	}

	bytesRepresentation, err := json.Marshal(message)
	if err != nil {
		log.Fatalln(err)
	}

	resp, err := http.Post(addr+"/api/cars?"+auth, "application/json", bytes.NewBuffer(bytesRepresentation))
	if err != nil {
		log.Fatalln(err)
	}

	var result map[string]interface{}

	json.NewDecoder(resp.Body).Decode(&result)

	if resp.Status == "200 OK" {
		log.Printf("Car added: %s, %s", result["license_plate_number"], result["id"])
		return true
	} else {
		log.Println(result["ErrorMessage"])
		return false
	}
}

func remCar(carID string) bool {
	req, err := http.NewRequest(http.MethodDelete, addr+"/api/cars/"+carID+"?"+auth, nil)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	var result map[string]interface{}

	json.Unmarshal(body, &result)

	if resp.Status == "200 OK" {
		log.Printf("Car removed: id %s\r\n", carID)
		return true
	} else {
		log.Println(result["ErrorMessage"])
		return false
	}
}
