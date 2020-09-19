package main

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	memberships "github.com/lbn/memberships/service"
)

type AddResponse struct {
	Membership *memberships.Membership `json:"membership"`
}

type LevelRequest struct {
	Level string `json:"level"`
}

type MembershipListResponse struct {
	Memberships []memberships.Membership `json:"memberships"`
}

func HandleRequestAdd(ctx context.Context, membership memberships.Membership) (AddResponse, error) {
	svc := memberships.NewMembershipService()
	membershipOut, err := svc.AddMembership(membership)
	if err != nil {
		return AddResponse{}, fmt.Errorf("Could not add membership: %v", err)
	}
	return AddResponse{Membership: membershipOut}, nil
}

func HandleRequestListMembersForLevel(ctx context.Context, levelRequest LevelRequest) (MembershipListResponse, error) {
	svc := memberships.NewMembershipService()
	memberships, err := svc.ListMembersForLevel(levelRequest.Level)
	if err != nil {
		return MembershipListResponse{}, fmt.Errorf("Could get memberships for level %s: %v", levelRequest.Level, err)
	}
	return MembershipListResponse{Memberships: memberships}, nil
}

func main() {
	functions := map[string]interface{}{
		"add":                    HandleRequestAdd,
		"list-members-for-level": HandleRequestListMembersForLevel,
	}

	lambda.Start(functions[os.Getenv("function")])
}
