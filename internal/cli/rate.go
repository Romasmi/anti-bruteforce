package cli

import (
	"context"
	"fmt"

	"github.com/Romasmi/anti-bruteforce/pkg/api"
)

func ClearRate(ctx context.Context, addr, kind, value string) error {
	c, err := dial(addr)
	if err != nil {
		return err
	}
	defer c.Close()

	req := &api.ClearRateRequest{}
	switch kind {
	case "login":
		req.Login = value
	case "ip":
		req.Ip = value
	default:
		return fmt.Errorf("unknown kind %q: must be login or ip", kind)
	}

	if _, err := api.NewAntiBruteforceClient(c).ClearRate(ctx, req); err != nil {
		return fmt.Errorf("clear rate: %w", err)
	}

	fmt.Printf("cleared %s bucket for %q\n", kind, value)
	return nil
}
