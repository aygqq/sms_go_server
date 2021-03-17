package control

import (
	"encoding/csv"
	"errors"

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
	WhiteList = append(WhiteList, elem)
	err := WritePhonesFile()
	if err != nil {
		return err
	}

	return nil
}

func RemFromWhiteListIdx(idx int) error {
	if idx < 0 || idx >= len(WhiteList) {
		err := errors.New("No such element in the list")
		return err
	}
	WhiteList[idx] = WhiteList[len(WhiteList)-1]
	WhiteList[len(WhiteList)-1].Name = ""
	WhiteList[len(WhiteList)-1].Phone = ""
	WhiteList = WhiteList[:len(WhiteList)-1]

	err := WritePhonesFile()
	if err != nil {
		return err
	}

	return nil
}

func SearchWhiteList(elem ListElement) int {
	for i := 0; i < len(WhiteList); i++ {
		if WhiteList[i].Name == elem.Name && WhiteList[i].Phone == elem.Phone {
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
		if err != nil {
			break
		}

		elem.Phone = record[0]
		elem.Name = record[1]

		WhiteList = append(WhiteList, elem)
	}

	return err
}

func WritePhonesFile() error {
	var record [2]string

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
		record[1] = WhiteList[i].Name

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
