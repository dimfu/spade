package templates

import "errors"

type TemplateType = int

const (
	TOP_8  = 8
	TOP_16 = 16
	TOP_32 = 32
)

type Match struct {
	Seats    [2]int
	WinnerTo int
}

type Matches = []Match

var Templates = map[int]Matches{
	TOP_8:  Top8,
	TOP_16: Top16,
	TOP_32: Top32,
}

func WithTemplate(t TemplateType) (Matches, error) {
	if t, exists := Templates[t]; exists {
		return t, nil
	}
	return nil, errors.New("template not found")
}
