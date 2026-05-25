package cli

import (
	"context"
	"fmt"

	"github.com/Romasmi/anti-bruteforce/pkg/api"
)

func UpdateIPList(ctx context.Context, addr, list, op, subnet string) error {
	c, err := dial(addr)
	if err != nil {
		return err
	}
	defer c.Close()

	cl := api.NewAntiBruteforceClient(c)
	req := &api.IPListRequest{Subnet: subnet}

	switch list + "/" + op {
	case "blacklist/add":
		_, err = cl.AddToBlacklist(ctx, req)
	case "blacklist/remove":
		_, err = cl.RemoveFromBlacklist(ctx, req)
	case "whitelist/add":
		_, err = cl.AddToWhitelist(ctx, req)
	case "whitelist/remove":
		_, err = cl.RemoveFromWhitelist(ctx, req)
	default:
		return fmt.Errorf("unknown operation %q for %s: must be add or remove", op, list)
	}

	if err != nil {
		return fmt.Errorf("%s %s: %w", op, list, err)
	}

	fmt.Printf("%sed %s %q\n", op, list, subnet)
	return nil
}
