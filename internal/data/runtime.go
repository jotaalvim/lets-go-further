package data

import (
	"fmt"
	"strconv"
)

type Runtime int

func (r Runtime) MarshalJSON() ([]byte, error) {

	jsonValue := fmt.Sprintf("%d mins", r)
	// `"Fran & Freddie's Diner	☺"`
	// "\"Fran & Freddie's Diner\t☺\""
	quotedJSONValue := strconv.Quote(jsonValue)

	return []byte(quotedJSONValue), nil
}
