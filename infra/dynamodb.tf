resource "aws_dynamodb_table" "memberships" {
  name           = "Memberships"
  read_capacity  = 10
  write_capacity = 10
  hash_key       = "Level"
  range_key      = "Name"

  attribute {
    name = "Level"
    type = "S"
  }
  attribute {
    name = "Name"
    type = "S"
  }
}