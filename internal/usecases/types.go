package usecases

type Type int

const (
	Healthcheck Type = iota
	Hello
	CheckAuth
	ClearRate
)

type CheckAuthInput struct {
	Login    string
	Password string
	IP       string
}

func (r *CheckAuthInput) validate() error {
	return validateCheckAuth(r.Login, r.Password, r.IP)
}

type ClearRateInput struct {
	Login string
	IP    string
}

func (r *ClearRateInput) validate() error {
	return validateClearRate(r.Login, r.IP)
}
