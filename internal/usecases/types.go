package usecases

type Type int

const (
	Healthcheck Type = iota
	Hello
	CheckAuth
	ClearRate
	AddToBlacklist
	RemoveFromBlacklist
	AddToWhitelist
	RemoveFromWhitelist
)

const (
	StrategyLogin    = "login"
	StrategyPassword = "password"
	StrategyIP       = "ip"
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

type IPListInput struct {
	Subnet string
}
