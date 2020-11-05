package util

import (
	"encoding/json"
	"errors"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

func GetCategory(path string) ([]int64, error) {
	if path == "" {
		return nil, errors.New("invalid argument")
	}

	ss := strings.Split(path, "/")
	if len(ss) < 0 {
		return nil, errors.New("invalid argument")
	}

	var rs []int64
	for _, s := range ss {
		val, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return nil, err
		}
		rs = append(rs, val)
	}

	return rs, nil
}

func CheckIsAccentMarks(strKeyword string) bool {
	keywordFinal := MakeAliasFromString(strKeyword)
	strKeyword = strings.Replace(strKeyword, " ", "-", -1)

	if strKeyword == keywordFinal {
		return false
	}
	return true

}

func MakeAliasFromString(s string) string {
	if s == "0" {
		return s
	}

	// convert to lower case
	s = strings.ToLower(s)

	// create a conversion table
	tableConvertCases := []struct{ in, out string }{
		{"ấ", "a"},
		{"ầ", "a"},
		{"ẩ", "a"},
		{"ẫ", "a"},
		{"ậ", "a"},
		{"Ấ", "a"},
		{"Ầ", "a"},
		{"Ẩ", "a"},
		{"Ẫ", "a"},
		{"Ậ", "a"},
		{"ắ", "a"},
		{"ằ", "a"},
		{"ẳ", "a"},
		{"ẵ", "a"},
		{"ặ", "a"},
		{"Ắ", "a"},
		{"Ằ", "a"},
		{"Ẳ", "a"},
		{"Ẵ", "a"},
		{"Ặ", "a"},
		{"á", "a"},
		{"à", "a"},
		{"ả", "a"},
		{"ã", "a"},
		{"ạ", "a"},
		{"â", "a"},
		{"ă", "a"},
		{"Á", "a"},
		{"À", "a"},
		{"Ả", "a"},
		{"Ã", "a"},
		{"Ạ", "a"},
		{"Â", "a"},
		{"Ă", "a"},
		{"ế", "e"},
		{"ề", "e"},
		{"ể", "e"},
		{"ễ", "e"},
		{"ệ", "e"},
		{"Ế", "e"},
		{"Ề", "e"},
		{"Ể", "e"},
		{"Ễ", "e"},
		{"Ệ", "e"},
		{"é", "e"},
		{"è", "e"},
		{"ẻ", "e"},
		{"ẽ", "e"},
		{"ẹ", "e"},
		{"ê", "e"},
		{"É", "e"},
		{"È", "e"},
		{"Ẻ", "e"},
		{"Ẽ", "e"},
		{"Ẹ", "e"},
		{"Ê", "e"},
		{"í", "i"},
		{"ì", "i"},
		{"ỉ", "i"},
		{"ĩ", "i"},
		{"ị", "i"},
		{"Í", "i"},
		{"Ì", "i"},
		{"Ỉ", "i"},
		{"Ĩ", "i"},
		{"Ị", "i"},
		{"ố", "o"},
		{"ồ", "o"},
		{"ổ", "o"},
		{"ỗ", "o"},
		{"ộ", "o"},
		{"Ố", "o"},
		{"Ồ", "o"},
		{"Ổ", "o"},
		{"Ô", "o"},
		{"Ộ", "o"},
		{"ớ", "o"},
		{"ờ", "o"},
		{"ở", "o"},
		{"ỡ", "o"},
		{"ợ", "o"},
		{"Ớ", "o"},
		{"Ờ", "o"},
		{"Ở", "o"},
		{"Ỡ", "o"},
		{"Ợ", "o"},
		{"ó", "o"},
		{"ò", "o"},
		{"ỏ", "o"},
		{"õ", "o"},
		{"ọ", "o"},
		{"ô", "o"},
		{"ơ", "o"},
		{"Ó", "o"},
		{"Ò", "o"},
		{"Ỏ", "o"},
		{"Õ", "o"},
		{"Ọ", "o"},
		{"Ô", "o"},
		{"Ơ", "o"},
		{"ứ", "u"},
		{"ừ", "u"},
		{"ử", "u"},
		{"ữ", "u"},
		{"ự", "u"},
		{"Ứ", "u"},
		{"Ừ", "u"},
		{"Ử", "u"},
		{"Ữ", "u"},
		{"Ự", "u"},
		{"ú", "u"},
		{"ù", "u"},
		{"ủ", "u"},
		{"ũ", "u"},
		{"ụ", "u"},
		{"ư", "u"},
		{"Ú", "u"},
		{"Ù", "u"},
		{"Ủ", "u"},
		{"Ũ", "u"},
		{"Ụ", "u"},
		{"Ư", "u"},
		{"ý", "y"},
		{"ỳ", "y"},
		{"ỷ", "y"},
		{"ỹ", "y"},
		{"ỵ", "y"},
		{"Ý", "y"},
		{"Ỳ", "y"},
		{"Ỷ", "y"},
		{"Ỹ", "y"},
		{"Ỵ", "y"},
		{"đ", "d"},
		{"Đ", "d"},
		{" ", "-"},
		{"&", "-"},
		{"<", "-"},
		{">", "-"},
		{`"`, ""},
		{"'", ""},
		{"?", ""},
		{"!", ""},
		{".", ""},
		{":", ""},
		{"@", ""},
		{"=", ""},
		{"#", ""},
		{"(", ""},
		{")", ""},
		{"[", ""},
		{"]", ""},
		{"_", "-"},
		{"ç", "c"},
		{"~", "-"},
		{"€", ""},
		{"‚", ""},
		{"ƒ", ""},
		{"„", ""},
		{"…", ""},
		{"†", ""},
		{"‡", ""},
		{"‰", ""},
		{"Š", ""},
		{"—", ""},
		{"™", ""},
		{"š", ""},
		{"›", ""},
		{"œ", "oe"},
		{"¡", ""},
		{"¢", ""},
		{"£", ""},
		{"¤", ""},
		{"¥", "y"},
		{"¦", ""},
		{"|", ""},
		{"§", ""},
		{"¨", ""},
		{"©", "c"},
		{"ª", ""},
		{"«", ""},
		{"¬", ""},
		{"®", "r"},
		{"¯", ""},
		{"°", ""},
		{"±", ""},
		{"²", ""},
		{"³", ""},
		{"µ", ""},
		{"¶", ""},
		{"·", ""},
		{"¸", ""},
		{"¹", ""},
		{"º", ""},
		{"»", ""},
		{"¼", ""},
		{"½", ""},
		{"¾", ""},
		{"¿", ""},
		{"×", ""},
		{"Þ", ""},
		{"ß", ""},
		{"÷", ""},
		{"ø", ""},
		{"þ", ""},
		{";", ""},
		{"/", ""},
		{"\\", ""},
		{" ", "-"},
	}
	for _, tc := range tableConvertCases {
		s = strings.Replace(s, tc.in, tc.out, -1)
	}
	s = regexp.MustCompile("[^a-z0-9-_]").ReplaceAllString(s, "-")
	s = regexp.MustCompile("-+").ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	s = url.QueryEscape(s)
	return s
}

func JsonToString(value interface{}) string {
	resJson, _ := json.Marshal(value)
	return string(resJson)
}
