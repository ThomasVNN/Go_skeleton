package helpers

import (
	"gitlab.thovnn.vn/protobuf/internal-apis-go/product/es_service"
	"math"
	"math/rand"
	"time"
)

func ShuffleProducts(chunkSize int64, list []*es_service.ProductData) []*es_service.ProductData {
	if chunkSize == 0 {
		return list
	}
	rand.Seed(time.Now().UnixNano())
	var total = int64(len(list))
	var chunkNum = math.Ceil(float64(total) / float64(chunkSize))
	var chunks [][]*es_service.ProductData
	var i int64
	for ; i < int64(chunkNum); i++ {
		var min = i * chunkSize
		var max = (i + 1) * chunkSize
		if max > total {
			max = total
		}
		var chunk []*es_service.ProductData
		chunk = append(chunk, list[min:max]...)
		chunks = append(chunks, chunk)
	}
	for _, e := range chunks {
		shuffle(len(e), func(j, k int) {
			e[j], e[k] = e[k], e[j]
		})
	}
	return mergeProductScoreArray(chunks...)
}

func mergeProductScoreArray(arr ...[]*es_service.ProductData) []*es_service.ProductData {
	var output []*es_service.ProductData
	for _, a := range arr {
		output = append(output, a...)
	}
	return output
}

func shuffle(n int, swap func(i, j int)) {
	if n < 0 {
		panic("invalid argument to Shuffle")
	}

	i := n - 1

	for ; i > 1<<31-1-1; i-- {
		j := int(rand.Int63n(int64(i + 1)))
		swap(i, j)
	}

	for ; i > 0; i-- {
		j := int(rand.Int31n(int32(i + 1)))
		swap(i, j)
	}
}
