package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/Romasmi/anti-bruteforce/pkg/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const defaultAddr = "localhost:50051"

func runClear(args []string) error {
	fs := flag.NewFlagSet("clear", flag.ExitOnError)
	addr := fs.String("addr", defaultAddr, "anti-bruteforce gRPC address (host:port)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	rest := fs.Args()
	if len(rest) < 2 {
		return fmt.Errorf("usage: clear [--addr host:port] <login|ip> <value>")
	}
	kind, value := rest[0], rest[1]

	req := &api.ClearRateRequest{}
	switch kind {
	case "login":
		req.Login = value
	case "ip":
		req.Ip = value
	default:
		return fmt.Errorf("unknown kind %q: must be login or ip", kind)
	}

	c, err := dial(*addr)
	if err != nil {
		return err
	}
	defer c.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := api.NewAntiBruteforceClient(c).ClearRate(ctx, req); err != nil {
		return fmt.Errorf("clear rate: %w", err)
	}

	fmt.Printf("cleared %s bucket for %q\n", kind, value)
	return nil
}

func runBlacklist(args []string) error {
	return runIPList("blacklist", args)
}

func runWhitelist(args []string) error {
	return runIPList("whitelist", args)
}

func runIPList(list string, args []string) error {
	fs := flag.NewFlagSet(list, flag.ExitOnError)
	addr := fs.String("addr", defaultAddr, "anti-bruteforce gRPC address (host:port)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	rest := fs.Args()
	if len(rest) < 2 {
		return fmt.Errorf("usage: %s [--addr host:port] <add|remove> <subnet>", list)
	}
	op, subnet := rest[0], rest[1]

	c, err := dial(*addr)
	if err != nil {
		return err
	}
	defer c.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

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
		return fmt.Errorf("unknown operation %q: must be add or remove", op)
	}

	if err != nil {
		return fmt.Errorf("%s %s: %w", op, list, err)
	}

	fmt.Printf("%sed %s %q\n", op, list, subnet)
	return nil
}

func dial(addr string) (*grpc.ClientConn, error) {
	c, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("dial %s: %w", addr, err)
	}
	return c, nil
}
