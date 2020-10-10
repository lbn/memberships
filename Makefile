mockgen:
	mkdir -p mocks/mock_dynamodbiface/
	mockgen -destination mocks/mock_dynamodbiface/mock_dynamodbiface.go \
		github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface DynamoDBAPI

test:
	go test ./service ./lambda

package:
	env GOOS=linux GOARCH=amd64 go build -o build/memberships-lambda ./lambda
	cd build && zip -r ../infra/memberships_lambda.zip memberships-lambda

deploy: package
	cd infra && terraform apply
