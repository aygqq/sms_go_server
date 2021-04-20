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

var addr string = "http://192.168.8.123:8080"
var auth string = "module=complete&login=" + login + "&password=" + pwHash

var singleGroupID string = ""
var carNewIdx int = 1
var singleGroupName string = "Разовый въезд"

func HttpTest() {
	// dbRemoveAllCars()
	// time.Sleep(time.Second * 5)
	// dbCheckAndCreateGroup(singleGroupName)
	// time.Sleep(time.Second * 5)

	// dbSearchAndAddCar("sr123t78")
	// dbSearchAndAddCar("vy123a78")
	// dbSearchAndAddCar("sr782t52")
	// dbSearchAndAddCar("bn123d34")
	// dbSearchAndAddCar("kg123i02")
	// dbSearchAndAddCar("kd123n02")
	// dbSearchAndAddCar("fg173i06")
	// dbSearchAndAddCar("eg128b55")
	// dbSearchAndAddCar("ks176a45")
	// dbSearchAndAddCar("sf103i92")
	// dbSearchAndAddCar("ft123k54")
	// dbSearchAndAddCar("pl123c64")
	// dbSearchAndAddCar("ok233t74")
	// dbSearchAndAddCar("op913i84")
	// dbSearchAndAddCar("sf105i92")
	// dbSearchAndAddCar("fg128k54")
	// dbSearchAndAddCar("zl121c64")
	// dbSearchAndAddCar("oz233t74")
	// dbSearchAndAddCar("op910z84")
	// dbSearchAndAddCar("oh567z84")
	// dbSearchAndAddCar("oh567z84")

	// dbSearchAndAddCar("аб567д84")

	// getAllCars()

	nPlateCheckAndFormat("ay123c047")
	nPlateCheckAndFormat("bm123h047")
	nPlateCheckAndFormat("pC123T047")
	nPlateCheckAndFormat("мХ123К047")
}

func regularGroupClear() {
	dbRemoveCarsFromGroup()
}

func dbCheckAndCreateGroup(grName string) bool {
	singleGroupID = ""
	var groups []interface{} = getCarGroups()

	for _, group := range groups {
		gr := group.(map[string]interface{})
		if gr["name"] == grName {
			singleGroupID = gr["id"].(string)
			log.Println("Group is already exists")
			return true
		}
	}

	singleGroupID = addCarGroup(true, "123", grName)

	if singleGroupID == "" {
		log.Println("Unable to add group")
		return false
	} else {
		log.Println("Group successfully added")
		return true
	}
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

	if len(plates) > 0 {
		log.Println("Car is already in database")
		return true
	}

	if addCarToGroup(nPlate, fmt.Sprintf("%d", carNewIdx), singleGroupID) {
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

func dbRemoveCarsFromGroup() bool {
	cars, totalCount := getCarsByGroup(singleGroupID, 0, 10)

	for _, car := range cars {
		oneCar := car.(map[string]interface{})
		remCar(oneCar["id"].(string))
	}

	for totalCount > 10 {
		log.Println(totalCount)
		cars, totalCount = getCarsByGroup(singleGroupID, 0, 10)
		for _, car := range cars {
			oneCar := car.(map[string]interface{})
			remCar(oneCar["id"].(string))
		}
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
			log.Printf("%s: id %s, barrier %t", gr["name"], gr["id"], gr["open_barrier"])
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

	if resp.Status == "200 OK" {
		log.Println("Cars")
		plates := result["plates"].([]interface{})
		for _, plate := range plates {
			pl := plate.(map[string]interface{})
			log.Printf("%s, %s\r\n", pl["license_plate_number"], pl["id"])
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
		log.Println(result["total_count"])
		totalCount := result["total_count"].(float64)
		plates := result["plates"].([]interface{})
		for _, plate := range plates {
			pl := plate.(map[string]interface{})
			log.Printf("%s, %s\r\n", pl["license_plate_number"], pl["id"])
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
			log.Printf("%s, %s\r\n", pl["license_plate_number"], pl["id"])
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
