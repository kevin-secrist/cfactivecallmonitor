terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.8"
    }
  }

  backend "s3" {
    bucket = "cfactivecallmonitor-terraform"
    key    = "production.tfstate"
    region = "us-east-1"
  }

  required_version = ">= 1.1.7"
}

provider "aws" {
  profile = "default"
  region  = "us-east-1"
}

resource "aws_dynamodb_table" "savedcalls" {
  name           = "SavedCalls"
  billing_mode   = "PROVISIONED"
  read_capacity  = 1
  write_capacity = 1
  hash_key       = "streetName"
  range_key      = "sortKey"

  attribute {
    name = "streetName"
    type = "S"
  }

  attribute {
    name = "sortKey"
    type = "S"
  }

  attribute {
    name = "isActive"
    type = "S"
  }

  attribute {
    name = "callReceived"
    type = "S"
  }

  global_secondary_index {
    name            = "ActiveIndex"
    hash_key        = "isActive"
    range_key       = "callReceived"
    write_capacity  = 1
    read_capacity   = 1
    projection_type = "ALL"
  }

  lifecycle {
    prevent_destroy = true
  }
}
