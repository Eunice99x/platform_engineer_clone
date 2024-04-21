package strings

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"github.com/atotto/clipboard"
	"golang.org/x/crypto/bcrypt"
	"strings"
)

func Hash(str string) string {
	hash := sha1.New()
	hash.Write([]byte(str))
	return string(hash.Sum(nil))
}

// Encrypt as password using bcrypt
func Encrypt(text string) (string, error) {
	var encryptedPassword = ""

	data, err := bcrypt.GenerateFromPassword([]byte(text), 10)

	if err != nil {
		return encryptedPassword, err
	}

	encryptedPassword = string(data)

	return encryptedPassword, err
}

// EncloseCSVStr adds string wraps to a csv string, specified by the enclosure
func EncloseCSVStr(str, enclosure string) string {
	str = strings.TrimSpace(str)
	if str == "" {
		return `""`
	}
	strArr := strings.Split(str, ",")
	strArrJoined := strings.Join(strArr, fmt.Sprintf(`%v,%v`, enclosure, enclosure))
	return fmt.Sprintf(`"%v"`, strArrJoined)
}

// InterfaceToJSON returns a variable as type JSON... Or "error" (string) if invalid.
// This is used for "debugging" purposes only, do not use in prod.
func InterfaceToJSON(i interface{}) string {
	jsonBytes, err := json.Marshal(i)
	if err != nil {
		return "error"
	} else {
		return string(jsonBytes)
	}
}

func GetJSONAndCopyToClipboard(i ...interface{}) string {
	json := GetJSON(i)
	err := clipboard.WriteAll(json)
	if err != nil {
		fmt.Println(err, "error writing to clipboard")
	}
	return json
}

// GetJSON return an interface as type json
// Returns a generic error formatted string if fails
func GetJSON(i ...interface{}) string {
	if len(i) > 1 {
		d := make([]interface{}, 0)
		for _, v := range i {
			d = append(d, v)
		}
		j, err := json.Marshal(d)
		if err != nil {
			return "Error dumping to json"
		} else {
			return string(j)
		}
	} else {
		j, err := json.Marshal(i)
		if err != nil {
			return "Error dumping to json"
		} else {
			return string(j)
		}
	}
}

func MultiDimensionalStringArrayToCSVText(arr [][]string) string {
	csvTxt := ""
	for _, v := range arr {
		row := `"` + strings.Join(v[:], `","`)
		row = row[:len(row)-2]
		row = strings.Join(v[:], ",")
		row = TrimSuffix(row, ",")
		csvTxt = csvTxt + row + "\n"
	}
	return csvTxt
}

func TrimSuffix(s, suffix string) string {
	if strings.HasSuffix(s, suffix) {
		s = s[:len(s)-len(suffix)]
	}
	return s
}
