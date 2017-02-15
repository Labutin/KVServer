package api

type Errors uint8

// Errors (codes)
const (
	KeyNotFound = Errors(iota)
)

// Errors in string format
var errors = map[Errors]string{
	KeyNotFound: "Key not found",
}

func (t Errors) String() string {
	return errors[t]
}
