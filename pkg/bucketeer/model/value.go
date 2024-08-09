package model

type BKTValue struct {
	Value interface{}
}

func NewValue(value interface{}) BKTValue {
	return BKTValue{
		Value: value,
	}
}
