package service

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/lbn/memberships/mocks/mock_dynamodbiface"
)

func TestAddMembershipDuplicate(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockDynamo := mock_dynamodbiface.NewMockDynamoDBAPI(ctrl)
	svc := MembershipService{mockDynamo}

	mockDynamo.EXPECT().GetItem(gomock.Any()).Return(&dynamodb.GetItemOutput{
		Item: make(map[string]*dynamodb.AttributeValue),
	}, nil)
	_, err := svc.AddMembership(Membership{
		Name:  "test",
		Level: "L2",
	})
	assert.Equal(t, ErrDuplicate, err)
}
