package control

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"strings"
	"unicode"
)

var enRu = map[string]string{
	"A": "А",
	"B": "В",
	"E": "Е",
	"K": "К",
	"M": "М",
	"H": "Н",
	"O": "О",
	"P": "Р",
	"C": "С",
	"T": "Т",
	"Y": "У",
	"X": "Х",
}

var ruPlateChar = [12]string{"А", "В", "Е", "К", "М", "Н", "О", "Р", "С", "Т", "У", "Х"}
var enPlateChar = [12]string{"A", "B", "E", "K", "M", "H", "O", "P", "C", "T", "Y", "X"}

func nPlateCheckAndFormat(nPlate string) (string, error) {
	var err error = nil

	nPlate = strings.ToUpper(nPlate)

	if !checkPlateChars(nPlate) {
		err = errors.New("Wrong nPlate characters")
		return nPlate, err
	}

	nPlate = transliteEnRu(nPlate)

	if len(nPlate) != 12 {
		errStr := fmt.Sprintf("Wrong nPlate length %d", len(nPlate))
		err = errors.New(errStr)
		return nPlate, err
	}

	bytes := []byte(nPlate)
	letters := string(bytes[0:4]) + string(bytes[7:9])
	numbers := string(bytes[4:7])
	region := string(bytes[9:12])
	log.Printf("Analyse %s, %s, %s\n", letters, numbers, region)
	if !checkIfNumberString(letters, true) {
		err = errors.New("Wrong nPlate letters")
		return nPlate, err
	}
	if !checkIfNumberString(numbers, false) {
		err = errors.New("Wrong nPlate numbers")
		return nPlate, err
	}
	if !checkIfNumberString(region, false) {
		err = errors.New("Wrong nPlate region")
		return nPlate, err
	}

	return nPlate, err
}

func transliteEnRu(text string) string {
	if text == "" {
		return ""
	}

	var input = bytes.NewBufferString(text)
	var output = bytes.NewBuffer(nil)

	// Previous, next letter for special processor
	// var p, n rune
	var rr string
	var ok bool

	log.Printf("Translite %s ->\n", input.String())
	for {
		r, _, err := input.ReadRune()
		if err != nil {
			break
		}

		if isRussianChar(r) || unicode.IsNumber(r) {
			output.WriteRune(r)
			continue
		}

		rr, ok = enRu[string(r)]
		if ok {
			log.Println(rr)
			output.WriteString(rr)
		}
	}

	log.Printf("Translite -> %s\n", output.String())
	return output.String()
}

func checkIfNumberString(str string, isNumbers bool) bool {
	var input = bytes.NewBufferString(str)

	for {
		r, _, err := input.ReadRune()
		if err != nil {
			break
		}

		if unicode.IsNumber(r) == isNumbers {
			return false
		}
	}
	return true
}

func checkPlateChars(nPlate string) bool {
	var input = bytes.NewBufferString(nPlate)

	for {
		r, _, err := input.ReadRune()
		if err != nil {
			break
		}

		if !isPlateChar(r) {
			return false
		}
	}
	return true
}

func isPlateChar(r rune) bool {
	if unicode.IsNumber(r) {
		return true
	}
	for i := 0; i < 12; i++ {
		if string(r) == ruPlateChar[i] || string(r) == enPlateChar[i] {
			return true
		}
	}
	log.Printf("Wrong char %x\n", r)
	return false
}

func isRussianChar(r rune) bool {
	switch {
	case r >= 1040 && r <= 1103,
		r == 1105, r == 1025:
		return true
	}

	return false
}
