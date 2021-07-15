package control

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

type dbConfig struct {
	module string
	login  string
	pw     string
	pwHash string

	ip   string
	port string
	addr string

	auth string
}

var dbCfg dbConfig

var ourGroupID string = ""
var carNewIdx int = 1
var ourGroupName string = "Временная"
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
	dbRemoveCarsByExternalID()
}

func dbCheckAndCreateGroup(grName string) bool {
	ourGroupID = ""
	var groups []interface{} = getCarGroups()

	for _, group := range groups {
		gr := group.(map[string]interface{})
		if gr["name"] == grName && gr["open_barrier"] == true {
			ourGroupID = gr["id"].(string)
			log.Println("Group is already exists: " + gr["name"].(string))
			return true
		}
	}

	// ourGroupID = addCarGroup(true, grName)

	// if ourGroupID == "" {
	// 	log.Println("Unable to add group")
	// 	return false
	// } else {
	// 	log.Println("Group successfully added")
	// 	return true
	// }

	return false
}

func dbSearchAndRemoveGroup(grName string) bool {
	ourGroupID = ""
	var groups []interface{} = getCarGroups()

	for _, group := range groups {
		gr := group.(map[string]interface{})
		if gr["name"] == grName {
			ourGroupID = gr["id"].(string)
			log.Println("Group is exists")
		}
	}

	if ourGroupID == "" {
		log.Println("Group is not exists")
		return true
	}

	if remCarGroup(ourGroupID) {
		log.Println("Group successfully removed")
		return true
	} else {
		log.Println("Unable to remove group")
		return false
	}
}

func dbSearchAndAddCar(user ListElement, nPlate string) int {
	var plates []interface{} = getCarsByPlate(nPlate)

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

	if addCarToGroup(user, nPlate, ourExtID, ourGroupID) {
		log.Printf("Car %d successfully added\n", carNewIdx)
		carNewIdx++
		return 1
	}
	log.Println("Unable to add car")
	return 0
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
		_, totalCount = getCarsByGroup(ourGroupID, offset, 10)
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
	resp, err := http.Get(dbCfg.addr + "/api/carconfig?" + dbCfg.auth)
	if err != nil {
		ErrorSt.connBase = true
		log.Println(err)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		ErrorSt.connBase = true
		log.Println(err)
		return
	}
	ErrorSt.connBase = false

	log.Println(resp.Status)
	log.Println(string(body))
	log.Println()
}

func getCarGroups() []interface{} {
	resp, err := http.Get(dbCfg.addr + "/api/cars-groups?offset=0&portion=10&" + dbCfg.auth)
	if err != nil {
		ErrorSt.connBase = true
		log.Println(err)
		return nil
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		ErrorSt.connBase = true
		log.Println(err)
		return nil
	}
	ErrorSt.connBase = false

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

func addCarGroup(openBar bool, name string) string {
	message := map[string]interface{}{
		"external_id":  ourExtID,
		"name":         name,
		"intercept":    openBar,
		"open_barrier": openBar,
		"color":        "00ffff00",
	}

	bytesRepresentation, err := json.Marshal(message)
	if err != nil {
		ErrorSt.connBase = true
		log.Println(err)
		return ""
	}

	resp, err := http.Post(dbCfg.addr+"/api/cars-groups?"+dbCfg.auth, "application/json", bytes.NewBuffer(bytesRepresentation))
	if err != nil {
		ErrorSt.connBase = true
		log.Println(err)
		return ""
	}
	ErrorSt.connBase = false

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
	req, err := http.NewRequest(http.MethodDelete, dbCfg.addr+"/api/cars-groups/"+groupID+"?"+dbCfg.auth, nil)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		ErrorSt.connBase = true
		log.Println(err)
		return false
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		ErrorSt.connBase = true
		log.Println(err)
		return false
	}
	ErrorSt.connBase = false

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
	resp, err := http.Get(dbCfg.addr + "/api/cars?offset=0&portion=50&" + dbCfg.auth)
	if err != nil {
		ErrorSt.connBase = true
		log.Println(err)
		return nil
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		ErrorSt.connBase = true
		log.Println(err)
		return nil
	}
	ErrorSt.connBase = false

	var result map[string]interface{}

	json.Unmarshal(body, &result)
	// log.Println(result)

	if resp.Status == "200 OK" {
		log.Println("All cars")
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
	resp, err := http.Get(dbCfg.addr + "/api/cars?" + count + filter + dbCfg.auth)
	if err != nil {
		ErrorSt.connBase = true
		log.Println(err)
		return nil, 0
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		ErrorSt.connBase = true
		log.Println(err)
		return nil, 0
	}
	ErrorSt.connBase = false

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
	resp, err := http.Get(dbCfg.addr + "/api/cars?" + count + filter + dbCfg.auth)
	if err != nil {
		ErrorSt.connBase = true
		log.Println(err)
		return nil, 0
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		ErrorSt.connBase = true
		log.Println(err)
		return nil, 0
	}
	ErrorSt.connBase = false

	var result map[string]interface{}

	// json.NewDecoder(resp.Body).Decode(&result)
	json.Unmarshal(body, &result)

	if resp.Status == "200 OK" {
		totalCount := result["total_count"].(float64)
		plates := result["plates"].([]interface{})
		log.Printf("Cars by external id: %s, total_count: %f\r\n", extID, totalCount)
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

func getCarsByPlate(nPlate string) []interface{} {
	filter := "filter=license_plate_number='" + url.QueryEscape(nPlate) + "'&"
	req, err := http.NewRequest(http.MethodGet, dbCfg.addr+"/api/cars?offset=0&portion=50&"+filter+dbCfg.auth, nil)
	req.Header.Add("Accept-Charset", "utf-8")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		ErrorSt.connBase = true
		log.Println(err)
		return nil
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		ErrorSt.connBase = true
		log.Println(err)
		return nil
	}
	ErrorSt.connBase = false

	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if resp.Status == "200 OK" {
		// log.Println("Cars by plate " + nPlate)
		plates := result["plates"].([]interface{})
		// for _, plate := range plates {
		// 	pl := plate.(map[string]interface{})
		// 	log.Printf("%s, %s, %s\r\n", pl["license_plate_number"], pl["id"], pl["external_id"])
		// }
		return plates
	} else {
		log.Println(result["ErrorMessage"])
		return nil
	}
}

func addCarToGroup(user ListElement, nPlate string, extID string, groupID string) bool {
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
		ErrorSt.connBase = true
		log.Println(err)
		return false
	}

	resp, err := http.Post(dbCfg.addr+"/api/cars?"+dbCfg.auth, "application/json", bytes.NewBuffer(bytesRepresentation))
	if err != nil {
		ErrorSt.connBase = true
		log.Println(err)
		return false
	}
	ErrorSt.connBase = false

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
	req, err := http.NewRequest(http.MethodDelete, dbCfg.addr+"/api/cars/"+carID+"?"+dbCfg.auth, nil)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		ErrorSt.connBase = true
		log.Println(err)
		return false
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		ErrorSt.connBase = true
		log.Println(err)
		return false
	}
	ErrorSt.connBase = false

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
