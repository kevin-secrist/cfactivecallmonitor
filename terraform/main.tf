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

locals {
  lambda_default_role_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_dynamodb_table" "savedcalls" {
  name           = "SavedCalls"
  billing_mode   = "PROVISIONED"
  read_capacity  = 1
  write_capacity = 1
  hash_key       = "streetName"
  range_key      = "sortKey"

  stream_enabled   = true
  stream_view_type = "NEW_AND_OLD_IMAGES"

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
          "dynamodb:GetItem",
          "dynamodb:PutItem",
          "dynamodb:Query",
          "dynamodb:UpdateItem"
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
    local.lambda_default_role_arn,
    aws_iam_policy.harvester_data_access_policy.arn
  ]
}

resource "aws_cloudwatch_event_rule" "every_five_minutes" {
  name                = "every-five-minutes"
  description         = "Fires every five minutes"
  schedule_expression = "rate(5 minutes)"
}

resource "aws_cloudwatch_event_target" "trigger_harvestcalls" {
  rule      = aws_cloudwatch_event_rule.every_five_minutes.name
  target_id = "harvestcalls"
  arn       = aws_lambda_function.harvestcalls.arn
}

resource "aws_lambda_permission" "trigger_harvestcalls_permission" {
  statement_id  = "AllowExecutionFromCloudWatch"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.harvestcalls.arn
  principal     = "events.amazonaws.com"
  source_arn    = aws_cloudwatch_event_rule.every_five_minutes.arn
}

data "archive_file" "active_call_notifier" {
  type             = "zip"
  source_file      = "../build/bin/active_call_notifier"
  output_file_mode = "0666"
  output_path      = "../build/bin/active_call_notifier.zip"
}

variable "SMS_TO" {
  type = string
}
variable "TWILIO_ACCOUNT_SID" {
  type = string
}
variable "TWILIO_API_KEY" {
  type = string
}
variable "TWILIO_API_SECRET" {
  type = string
}

resource "aws_lambda_function" "active_call_notifier" {
  function_name    = "ActiveCallNotifier"
  description      = "Sends SMS Notifications from Call Events"
  filename         = data.archive_file.active_call_notifier.output_path
  memory_size      = 128
  runtime          = "go1.x"
  handler          = "active_call_notifier"
  role             = aws_iam_role.active_call_notifier.arn
  source_code_hash = data.archive_file.active_call_notifier.output_base64sha256
  timeout          = 60

  environment {
    variables = {
      SMS_TO             = var.SMS_TO
      TWILIO_ACCOUNT_SID = var.TWILIO_ACCOUNT_SID
      TWILIO_API_KEY     = var.TWILIO_API_KEY
      TWILIO_API_SECRET  = var.TWILIO_API_SECRET
    }
  }
}

resource "aws_cloudwatch_log_group" "active_call_notifier" {
  name              = "/aws/lambda/${aws_lambda_function.active_call_notifier.function_name}"
  retention_in_days = 7
}

resource "aws_iam_policy" "active_call_notifier" {
  name = "ActiveCallNotifier"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "dynamodb:DescribeStream",
          "dynamodb:GetRecords",
          "dynamodb:GetShardIterator",
          "dynamodb:ListStreams"
        ],
        Effect = "Allow",
        Resource = [
          "${aws_dynamodb_table.savedcalls.arn}/stream/*"
        ]
      }
    ]
  })
}

resource "aws_iam_role" "active_call_notifier" {
  name = "ActiveCallNotifier"

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
    local.lambda_default_role_arn,
    aws_iam_policy.active_call_notifier.arn
  ]
}

variable "STREET_NAMES" {
  type = list(string)
}

resource "aws_lambda_event_source_mapping" "active_call_notifier_trigger" {
  event_source_arn  = aws_dynamodb_table.savedcalls.stream_arn
  function_name     = aws_lambda_function.active_call_notifier.arn
  starting_position = "TRIM_HORIZON"

  maximum_batching_window_in_seconds = 10
  maximum_record_age_in_seconds      = 3600
  maximum_retry_attempts             = 5

  filter_criteria {
    filter {
      pattern = jsonencode({
        "data" : {
          "streetName" : var.STREET_NAMES
        }
      })
    }
  }
}
