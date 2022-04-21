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
  region = "us-east-1"
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

data "archive_file" "harvestcalls" {
  type             = "zip"
  source_file      = "../build/bin/harvestcalls"
  output_file_mode = "0666"
  output_path      = "../build/bin/harvestcalls.zip"
}

variable "CPD_API_KEY" {
  type = string
}

variable "CFD_API_KEY" {
  type = string
}

resource "aws_lambda_function" "harvestcalls" {
  function_name    = "HarvestCalls"
  description      = "Pulls active Police/Fire calls from chesterfield.gov and stores them"
  filename         = data.archive_file.harvestcalls.output_path
  memory_size      = 128
  runtime          = "go1.x"
  handler          = "harvestcalls"
  role             = aws_iam_role.harvester.arn
  source_code_hash = data.archive_file.harvestcalls.output_base64sha256
  timeout          = 15

  environment {
    variables = {
      CPD_API_KEY = var.CPD_API_KEY
      CFD_API_KEY = var.CFD_API_KEY
    }
  }
}

resource "aws_cloudwatch_log_group" "harvestcalls" {
  name              = "/aws/lambda/${aws_lambda_function.harvestcalls.function_name}"
  retention_in_days = 7
}

resource "aws_iam_policy" "harvester_data_access_policy" {
  name = "HarvesterDataAccess"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "dynamodb:*"
        ],
        Effect = "Allow",
        Resource = [
          aws_dynamodb_table.savedcalls.arn,
          "${aws_dynamodb_table.savedcalls.arn}/*"
        ]
      }
    ]
  })
}

resource "aws_iam_role" "harvester" {
  name = "Harvester"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "lambda.amazonaws.com"
      }
    }]
  })

  managed_policy_arns = [
    "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole",
    aws_iam_policy.harvester_data_access_policy.arn
  ]
}
