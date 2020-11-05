package helper

func IndexOfSliceNumber(arr interface{}, item interface{}) int {
	index := -1
	switch array := arr.(type) {
	case []int64:
		for idx, itemArr := range array {
			if itemArr == item.(int64) {
				index = idx
				break
			}
		}
	case []int32:
		for idx, itemArr := range array {
			if itemArr == item.(int32) {
				index = idx
				break
			}
		}
	case []uint32:
		for idx, itemArr := range array {
			if itemArr == item.(uint32) {
				index = idx
				break
			}
		}
	case []float64:
		item = item.(float64)
		for idx, itemArr := range array {
			if itemArr == item {
				index = idx
				break
			}
		}
	default:
		index = -1
	}
	return index
}
