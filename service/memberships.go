package service

import (
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
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

func (svc *MembershipService) createTableIfNotExists() error {
	input := &dynamodb.CreateTableInput{
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String("Level"),
				AttributeType: aws.String("S"),
			},
			{
				AttributeName: aws.String("Name"),
				AttributeType: aws.String("S"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("Level"),
				KeyType:       aws.String("HASH"),
			},
			{
				AttributeName: aws.String("Name"),
				KeyType:       aws.String("RANGE"),
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(10),
			WriteCapacityUnits: aws.Int64(10),
		},
		TableName: aws.String(tableName),
	}

	_, err := svc.client.CreateTable(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == dynamodb.ErrCodeResourceInUseException {
			return nil
		}
		return err
	}
	return svc.client.WaitUntilTableExists(&dynamodb.DescribeTableInput{
		TableName: aws.String(tableName),
	})
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
	// err := svc.createTableIfNotExists()
	// if err != nil {
	// 	log.Fatalf("Could not prepare table: %v", err)
	// }
	return
}
