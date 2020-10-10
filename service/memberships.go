package service

import (
	"errors"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

const (
	tableName                     string = "Memberships"
	defaultMembershipDurationDays int    = 7 * 4
)

var (
	ErrNotFound  = errors.New("record not found")
	ErrDuplicate = errors.New("duplicate record")
)

type MembershipService struct {
	client dynamodbiface.DynamoDBAPI
}

type Membership struct {
	Name      string     `dynamodbav:"Name" json:"name"`
	Level     string     `dynamodbav:"Level" json:"level"`
	StartDate *time.Time `json:"startDate"`
	EndDate   *time.Time `json:"endDate"`
}

func (svc *MembershipService) AddMembership(membership Membership) (*Membership, error) {
	res, _ := svc.getItemByNameAndLevel(membership.Name, membership.Level)
	if res.Item != nil {
		return nil, ErrDuplicate
	}
	if membership.StartDate == nil {
		membership.StartDate = aws.Time(time.Now())
	}
	if membership.EndDate == nil {
		endDate := membership.StartDate.AddDate(0, 0, defaultMembershipDurationDays)
		membership.EndDate = aws.Time(time.Date(endDate.Year(), endDate.Month(),
			endDate.Day(), 23, 59, 59, endDate.Nanosecond(), endDate.Location()))
	}
	av, err := dynamodbattribute.MarshalMap(membership)
	if err != nil {
		return nil, err
	}
	_, err = svc.client.PutItem(&dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(tableName),
	})
	return &membership, err
}

func (svc *MembershipService) ListMembersForLevel(level string) ([]Membership, error) {
	res, err := svc.client.Query(&dynamodb.QueryInput{
		TableName: aws.String(tableName),
		KeyConditions: map[string]*dynamodb.Condition{
			"Level": {
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						S: aws.String(level),
					},
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}
	memberships := []Membership{}
	err = dynamodbattribute.UnmarshalListOfMaps(res.Items, &memberships)
	return memberships, err
}

func (svc *MembershipService) getItemByNameAndLevel(name, level string) (*dynamodb.GetItemOutput, error) {
	return svc.client.GetItem(&dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"Name": {
				S: aws.String(name),
			},
			"Level": {
				S: aws.String(level),
			},
		},
		TableName: aws.String(tableName),
	})
}

func (svc *MembershipService) GetMembership(name, level string) (membership Membership, err error) {
	res, err := svc.getItemByNameAndLevel(name, level)

	if err != nil {
		return
	}
	err = dynamodbattribute.UnmarshalMap(res.Item, &membership)
	if err != nil {
		return
	}
	if res.Item == nil {
		err = ErrNotFound
		return
	}
	return
}

func NewMembershipService() (svc MembershipService) {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	endpointURL := os.Getenv("AWS_ENDPOINT_URL")
	svc.client = dynamodb.New(sess, &aws.Config{Endpoint: &endpointURL})
	return
}
