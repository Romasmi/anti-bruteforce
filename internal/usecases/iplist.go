package usecases

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type IPListRepository interface {
	AddToWhitelist(cidr string) error
	RemoveFromWhitelist(cidr string) error
	AddToBlacklist(cidr string) error
	RemoveFromBlacklist(cidr string) error
}

type AddToBlacklistUsecase struct{ Repo IPListRepository }

func (u *AddToBlacklistUsecase) Do(_ context.Context, req any) (any, error) {
	r, ok := req.(*IPListInput)
	if !ok {
		return nil, status.Error(codes.InvalidArgument, "invalid request type")
	}
	if err := validateSubnet(r.Subnet); err != nil {
		return nil, err
	}
	if err := u.Repo.AddToBlacklist(r.Subnet); err != nil {
		return nil, status.Errorf(codes.Internal, "add to blacklist: %v", err)
	}
	return nil, nil
}

type RemoveFromBlacklistUsecase struct{ Repo IPListRepository }

func (u *RemoveFromBlacklistUsecase) Do(_ context.Context, req any) (any, error) {
	r, ok := req.(*IPListInput)
	if !ok {
		return nil, status.Error(codes.InvalidArgument, "invalid request type")
	}
	if err := validateSubnet(r.Subnet); err != nil {
		return nil, err
	}
	if err := u.Repo.RemoveFromBlacklist(r.Subnet); err != nil {
		return nil, status.Errorf(codes.Internal, "remove from blacklist: %v", err)
	}
	return nil, nil
}

type AddToWhitelistUsecase struct{ Repo IPListRepository }

func (u *AddToWhitelistUsecase) Do(_ context.Context, req any) (any, error) {
	r, ok := req.(*IPListInput)
	if !ok {
		return nil, status.Error(codes.InvalidArgument, "invalid request type")
	}
	if err := validateSubnet(r.Subnet); err != nil {
		return nil, err
	}
	if err := u.Repo.AddToWhitelist(r.Subnet); err != nil {
		return nil, status.Errorf(codes.Internal, "add to whitelist: %v", err)
	}
	return nil, nil
}

type RemoveFromWhitelistUsecase struct{ Repo IPListRepository }

func (u *RemoveFromWhitelistUsecase) Do(_ context.Context, req any) (any, error) {
	r, ok := req.(*IPListInput)
	if !ok {
		return nil, status.Error(codes.InvalidArgument, "invalid request type")
	}
	if err := validateSubnet(r.Subnet); err != nil {
		return nil, err
	}
	if err := u.Repo.RemoveFromWhitelist(r.Subnet); err != nil {
		return nil, status.Errorf(codes.Internal, "remove from whitelist: %v", err)
	}
	return nil, nil
}
