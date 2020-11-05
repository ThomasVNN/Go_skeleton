package helpers

import (
	"gitlab.thovnn.vn/protobuf/internal-apis-go/product/es_service"
	"sort"
)

const SortTypeListingV2 = "listing_v2"
const SortTypeListingV3 = "listing_v3"

type SortMapV2 []*es_service.ProductData

type SortMapV3 []*es_service.ProductData

func (a SortMapV2) Len() int {
	return len(a)
}
func (a SortMapV2) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
func (a SortMapV2) Less(i, j int) bool {
	return listingScoreSort(a, i, j, "v2")
}

func (a SortMapV3) Len() int {
	return len(a)
}
func (a SortMapV3) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
func (a SortMapV3) Less(i, j int) bool {
	return listingScoreSort(a, i, j, "v3")
}

func listingScoreSort(a []*es_service.ProductData, i, j int, field string) bool {
	var p1, p2 float64
	var listingScore, listingScore2 = a[i].GetListingScore(), a[j].GetListingScore()
	switch field {
	case "v3":
		p1 = listingScore.GetScoreV3()
		p2 = listingScore2.GetScoreV3()
	default:
		p1 = listingScore.GetScoreV2()
		p2 = listingScore2.GetScoreV2()
	}

	return p1 > p2
}

func SortProducts(input []*es_service.ProductData, sortType string) []*es_service.ProductData {
	if sortType == SortTypeListingV2 {
		sort.Sort(SortMapV2(input))
	}
	if sortType == SortTypeListingV3 {
		sort.Sort(SortMapV3(input))
	}

	return input
}
