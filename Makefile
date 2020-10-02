package:
	env GOOS=linux GOARCH=amd64 go build -o build/memberships-lambda ./lambda
	cd build && zip -r ../infra/memberships_lambda.zip memberships-lambda

deploy: package
	cd infra && terraform apply
