package orm

type InvalidResultTypeError struct{}

func (a InvalidResultTypeError) Error() string {
	return "invalid type of result, result must be a ptr which pointers to slice(slice of struct/pointers), or struct, or pointer to struct"
}

type InvalidDatasourceError struct {
	datasource string
}

func (a InvalidDatasourceError) Error() string {
	return "invalid datasource" + a.datasource
}
