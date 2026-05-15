package usecases

type Type int

const (
	Healthcheck Type = iota
	Hello
	CheckAuth
	ClearRate
)

type IdentifierType int32

const (
	IdentifierTypeUnspecified IdentifierType = 0
	IdentifierTypeLogin       IdentifierType = 1
	IdentifierTypePassword    IdentifierType = 2
	IdentifierTypeIP          IdentifierType = 3
)

type CheckAuthInput struct {
	Type  IdentifierType
	Value string
}

func (r *CheckAuthInput) validate() error {
	return validateIdentifier(r.Type, r.Value)
}

type ClearRateInput struct {
	Type  IdentifierType
	Value string
}

func (r *ClearRateInput) validate() error {
	return validateIdentifier(r.Type, r.Value)
}
