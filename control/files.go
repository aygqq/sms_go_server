package control

import (
	"crypto/md5"
	"encoding/csv"
	"errors"
	"fmt"

	// "io/ioutil"
	"log"
	"os"
)

var phFile FilePhones

func checkPhone(str string) error {
	data := []byte(str)
	err := errors.New("Failed to parse Phone")
	isWrong := false

	if len(data) > PHONE_SIZE {
		return err
	}

	for i := 0; i < len(data); i++ {
		if (data[i] < '0' || data[i] > '9') && data[i] != '+' {
			isWrong = true
		}
	}

	if isWrong == true {
		return err
	}

	return nil
}

func checkPhonesFile() error {
	for i := 0; i < len(WhiteList); i++ {
		err := checkPhone(WhiteList[i].Phone)
		if err != nil {
			return err
		}
	}

	return nil
}

func AddToWhiteList(elem ListElement) error {
	// TODO: Check if exists
	idx := SearchWhiteListByPhone(elem.Phone)
	if idx != -1 {
		err := errors.New("Element with yhis phone is already exists")
		return err
	}
	WhiteList = append(WhiteList, elem)
	err := WritePhonesFile()
	if err != nil {
		return err
	}

	return nil
}

func RemFromWhiteListIdx(idx int) (ListElement, error) {
	var elem ListElement
	if idx < 0 || idx >= len(WhiteList) {
		err := errors.New("No such element in the list")
		return elem, err
	}
	elem = WhiteList[idx]
	WhiteList[idx] = WhiteList[len(WhiteList)-1]
	WhiteList[len(WhiteList)-1].Name = ""
	WhiteList[len(WhiteList)-1].Phone = ""
	WhiteList[len(WhiteList)-1].Surname = ""
	WhiteList[len(WhiteList)-1].Patronymic = ""
	WhiteList[len(WhiteList)-1].Role = ""
	WhiteList[len(WhiteList)-1].AreaNum = ""
	WhiteList = WhiteList[:len(WhiteList)-1]

	err := WritePhonesFile()
	if err != nil {
		return elem, err
	}

	return elem, nil
}

func SearchWhiteList(elem ListElement) int {
	for i := 0; i < len(WhiteList); i++ {
		if WhiteList[i] == elem {
			return i
		}
	}

	return -1
}

func SearchWhiteListByPhone(phone string) int {
	for i := 0; i < len(WhiteList); i++ {
		if WhiteList[i].Phone == phone {
			return i
		}
	}

	return -1
}

// func GetPhonesFile(records *[]ListElement) error {
// 	log.Println("GetPhonesFile")
// 	for i := 0; i < len(WhiteList); i++ {
// 		records[i].Name = WhiteList[i].Name
// 		records[i].Name = WhiteList[i].Name
// 	}
// 	return nil
// }

func readConfigFile() error {
	var cfg dbConfig

	log.Println("readConfigFile")

	csvfile, err := os.Open(configFilePath)
	if err != nil {
		return err
	}
	defer csvfile.Close()

	// Parse the file
	r := csv.NewReader(csvfile)

	// Iterate through the records
	for {
		// Read each record from csv
		record, err := r.Read()
		log.Println(record)
		if err != nil {
			break
		}

		if record[0] == "addr" {
			cfg.ip = record[1]
		} else if record[0] == "port" {
			cfg.port = record[1]
		} else if record[0] == "module" {
			cfg.module = record[1]
		} else if record[0] == "login" {
			cfg.login = record[1]
		} else if record[0] == "password" {
			cfg.pw = record[1]
		}
	}

	cfg.addr = "http://" + cfg.ip + ":" + cfg.port

	cfg.pwHash = fmt.Sprintf("%x", md5.Sum([]byte(cfg.pw)))
	// cfg.auth = "module=" + cfg.module + "&login=" + cfg.login + "&password=" + cfg.pwHash
	cfg.auth = "login=" + cfg.login + "&password=" + cfg.pwHash

	log.Println(cfg.addr)
	log.Println(cfg.auth)

	dbCfg = cfg

	return err
}

func readPhonesFile() error {
	log.Println("readPhonesFile")

	var elem ListElement

	csvfile, err := os.Open(phonesFilePath)
	if err != nil {
		return err
	}
	defer csvfile.Close()

	// Parse the file
	r := csv.NewReader(csvfile)

	// Iterate through the records
	for {
		// Read each record from csv
		record, err := r.Read()
		log.Println(record)
		if err != nil || len(record) != 6 {
			break
		}

		elem.Phone = record[0]
		elem.Surname = record[1]
		elem.Name = record[2]
		elem.Patronymic = record[3]
		elem.Role = record[4]
		elem.AreaNum = record[5]

		WhiteList = append(WhiteList, elem)
	}

	return err
}

func WritePhonesFile() error {
	var record [6]string

	err := checkPhonesFile()
	if err != nil {
		return err
	}

	file, err := os.Create(phonesFilePath)
	if err != nil {
		return err
	}
	defer file.Close()
	w := csv.NewWriter(file)
	defer w.Flush()

	for i := 0; i < len(WhiteList); i++ {
		record[0] = WhiteList[i].Phone
		record[1] = WhiteList[i].Surname
		record[2] = WhiteList[i].Name
		record[3] = WhiteList[i].Patronymic
		record[4] = WhiteList[i].Role
		record[5] = WhiteList[i].AreaNum

		err := w.Write(record[:])
		if err != nil {
			return err
		}
	}

	return nil
}

func deleteFile(path string) error {
	log.Println("Deleting config file")

	return os.Remove(path)
}
