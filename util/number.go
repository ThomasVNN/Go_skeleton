package util

import (
	"fmt"
	"math"
	"strconv"
)

func ProcessPrice(price float64) float64 {
	n := math.Trunc(price)
	priceString := fmt.Sprintf("%.0f", n)
	x := len(priceString)
	a := math.Pow10(x - 1)
	b := a / 2
	c := math.Trunc(math.Mod(n, a))
	d := math.Trunc(n / a)
	finalPrice := 0.0
	if c < b/2 {
		finalPrice = d * a
	} else if c > b {
		finalPrice = (d + 1) * a
	} else {
		finalPrice = (d * a) + b
	}
	return finalPrice
}

func InterfaceToUint(n interface{}) (num uint32, ok bool) {
	ok = true
	switch t := n.(type) {
	case float64:
		num = uint32(t)
	case float32:
		num = uint32(t)
	case int:
		num = uint32(t)
	case int32:
		num = uint32(t)
	case int64:
		num = uint32(t)
	case uint32:
		num = t
	default:
		ok = false
	}

	return
}

func InterfaceToInt(n interface{}) (num int, ok bool) {
	ok = true
	var err error
	switch t := n.(type) {
	case float64:
		num = int(t)
	case float32:
		num = int(t)
	case int32:
		num = int(t)
	case int64:
		num = int(t)
	case string:
		num, err = strconv.Atoi(t)
		ok = err == nil
	default:
		ok = false
	}

	return
}