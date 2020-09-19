locals {
  lambda_zip = "memberships_lambda.zip"
}
provider "aws" {
  version = "~> 3.0"
  region  = "eu-central-1"
}

resource "aws_s3_bucket" "lambda_bucket" {
  bucket = var.lambda_deploy_bucket
  acl    = "private"
}

resource "aws_s3_bucket_object" "lambda_zip" {
  bucket = aws_s3_bucket.lambda_bucket.bucket
  key    = local.lambda_zip
  source = "../build/${local.lambda_zip}"

  etag = filemd5("../build/${local.lambda_zip}")
}

resource "aws_iam_role" "iam_for_lambda" {
  name = "iam_for_lambda"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "iam_for_lambda" {
  role       = aws_iam_role.iam_for_lambda.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonDynamoDBFullAccess"
}

resource "aws_lambda_function" "lambda_add" {
  s3_bucket     = aws_s3_bucket.lambda_bucket.bucket
  s3_key        = local.lambda_zip
  function_name = "memberships_add"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "memberships-lambda"

  source_code_hash = filebase64sha256("../build/${local.lambda_zip}")

  runtime = "go1.x"

  depends_on = [
    aws_s3_bucket_object.lambda_zip
  ]

  environment {
    variables = {
      function = "add"
    }
  }
}

resource "aws_lambda_function" "lambda_list_members_for_level" {
  s3_bucket     = aws_s3_bucket.lambda_bucket.bucket
  s3_key        = local.lambda_zip
  function_name = "memberships_list_members_for_level"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "memberships-lambda"

  source_code_hash = filebase64sha256("../build/${local.lambda_zip}")

  runtime = "go1.x"

  depends_on = [
    aws_s3_bucket_object.lambda_zip
  ]

  environment {
    variables = {
      function = "list-members-for-level"
    }
  }
}
