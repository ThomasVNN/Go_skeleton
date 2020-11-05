package helper

import (
	"encoding/json"
	"strconv"
	"strings"
)

func CheckStringDuplicate(capitals []string, haystack string) (result bool) {
	for _, item := range capitals {
		if item == haystack {
			result = true
		}
	}
	return
}

func ToUpperSpecial(str string) string {
	if str == "" {
		return str
	}
	return strings.Replace(str, str[0:1], strings.ToUpper(str[0:1]), 1)
}

func IndexOf(arr []string, item string) int {
	index := -1
	for idx, itemArr := range arr {
		if itemArr == item {
			index = idx
			break
		}
	}
	return index
}

func RemoveLastString(str, sep string) string {
	strReturn := ""
	arrString := strings.Split(str, sep)
	countString := len(arrString)
	if countString <= 0 {
		return strReturn
	}
	arrString = arrString[:countString-1]
	countString = len(arrString)
	for i := 0; i < countString; i++ {
		if i == countString-1 {
			strReturn += arrString[i]
			break
		}
		strReturn += arrString[i] + "_"

	}
	return strReturn
}

func GetLastCategoryId(path string) (int32, error) {
	sCategory := strings.Split(path, "/")
	lastCategoryId := sCategory[len(sCategory)-1]
	intCategoryId, err := strconv.Atoi(lastCategoryId)
	if err != nil {
		return 0, err
	}
	return int32(intCategoryId), nil
}

func JsonToString(value interface{}) string {
	resJson, _ := json.Marshal(value)
	return string(resJson)
}
