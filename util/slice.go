package util

func UInt32SliceDifference(a, b []uint32) (diff []uint32) {
	m := make(map[uint32]bool)

	for _, item := range b {
		m[item] = true
	}

	for _, item := range a {
		if _, ok := m[item]; !ok {
			diff = append(diff, item)
		}
	}
	return
}