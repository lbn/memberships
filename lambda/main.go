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

type GetResponse AddResponse

type NoRecordError struct{}

func (err NoRecordError) Error() string {
	return "record not found"
}

type ListMembersForLevelRequest struct {
	Level string `json:"level"`
}

type GetRequest struct {
	Name  string `json:"name"`
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

func HandleRequestListMembersForLevel(ctx context.Context, listRequest ListMembersForLevelRequest) (MembershipListResponse, error) {
	svc := memberships.NewMembershipService()
	memberships, err := svc.ListMembersForLevel(listRequest.Level)
	if err != nil {
		return MembershipListResponse{}, fmt.Errorf("Could get memberships for level %s: %v", listRequest.Level, err)
	}
	return MembershipListResponse{Memberships: memberships}, nil
}

func HandleRequestGet(ctx context.Context, getRequest GetRequest) (res GetResponse, err error) {
	svc := memberships.NewMembershipService()
	membership, err := svc.GetMembership(getRequest.Name, getRequest.Level)

	if err != nil {
		if err == memberships.ErrNotFound {
			err = NoRecordError{}
			return
		}
		err = fmt.Errorf("Could get membership %v: %v", getRequest, err)
		return
	}
	return GetResponse{Membership: &membership}, nil
}

func main() {
	functions := map[string]interface{}{
		"add":                    HandleRequestAdd,
		"get":                    HandleRequestGet,
		"list-members-for-level": HandleRequestListMembersForLevel,
	}

	lambda.Start(functions[os.Getenv("function")])
}
