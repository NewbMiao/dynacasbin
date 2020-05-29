provider "aws" {
  region  = "ap-southeast-1"
  profile = "default"
}
resource "aws_dynamodb_table" "casbin-rules" {
  name           = "casbin-rules"
  billing_mode   = "PROVISIONED"
  read_capacity  = 10
  write_capacity = 5
  hash_key       = "PType"
  range_key      = "V0"


  attribute {
    name = "PType"
    type = "S"
  }

  attribute {
    name = "V0"
    type = "S"
  }



  tags = {
    Name        = "casbin_table"
    Environment = "production"
  }
}
