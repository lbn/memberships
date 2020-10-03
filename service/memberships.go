package service

import (
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

const (
	tableName                     string = "Memberships"
	defaultMembershipDurationDays int    = 7 * 4
)

type MembershipService struct {
	client *dynamodb.DynamoDB
}

type Membership struct {
	Name      string     `dynamodbav:"Name" json:"name"`
	Level     string     `dynamodbav:"Level" json:"level"`
	StartDate *time.Time `json:"startDate"`
	EndDate   *time.Time `json:"endDate"`
}

func (svc *MembershipService) AddMembership(membership Membership) (*Membership, error) {
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

func (svc *MembershipService) listMembers() ([]Membership, error) {
	res, err := svc.client.Scan(&dynamodb.ScanInput{
		TableName: aws.String(tableName),
	})
	if err != nil {
		return nil, err
	}
	memberships := []Membership{}
	err = dynamodbattribute.UnmarshalListOfMaps(res.Items, &memberships)
	return memberships, err
}

func NewMembershipService() (svc MembershipService) {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	endpointURL := os.Getenv("AWS_ENDPOINT_URL")
	svc.client = dynamodb.New(sess, &aws.Config{Endpoint: &endpointURL})
	return
}
