package helper

type ParamStruct struct {
	Limit  int64
	Page   int64
	Offset int64
	Last   int64
	Since  int64
	Until  int64
}

func GetPagination(page int64, limit int64) ParamStruct {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit
	last := offset + limit
	return ParamStruct{
		Page:   page,
		Offset: offset,
		Limit:  limit,
		Last:   last - 1,
	}
}
