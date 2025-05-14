package model

type SourceIDType int32

const (
	SourceIDGoServer SourceIDType = 5
)

func (s SourceIDType) Int32() int32 {
	return int32(s)
}
