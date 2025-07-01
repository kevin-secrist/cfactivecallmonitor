terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.0"
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

resource "aws_sns_topic" "ops_critical" {
  name = "ops-critical"
}

resource "aws_sns_topic_policy" "cloudwatch_policy" {
  arn    = aws_sns_topic.ops_critical.arn
  policy = data.aws_iam_policy_document.cloudwatch_policy.json
}

data "aws_iam_policy_document" "cloudwatch_policy" {
  statement {
    sid     = "OpsSNSAllowCloudWatch"
    effect  = "Allow"
    actions = ["sns:Publish"]

    principals {
      type        = "Service"
      identifiers = ["cloudwatch.amazonaws.com"]
    }

    resources = [
      aws_sns_topic.ops_critical.arn,
    ]
  }
}

resource "aws_sns_topic_subscription" "email_alerts" {
  topic_arn = aws_sns_topic.ops_critical.arn
  protocol  = "email"
  endpoint  = var.OPS_EMAIL
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
  source_file      = "../build/bin/harvestcalls/bootstrap"
  output_file_mode = "0666"
  output_path      = "../build/bin/harvestcalls.zip"
}

resource "aws_lambda_function" "harvestcalls" {
  function_name    = "HarvestCalls"
  description      = "Pulls active Police/Fire calls from chesterfield.gov and stores them"
  filename         = data.archive_file.harvestcalls.output_path
  memory_size      = 128
  runtime          = "provided.al2023"
  handler          = "bootstrap"
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

# this should only go into an alarm state if the lambda
# is totally broken in some way
resource "aws_cloudwatch_metric_alarm" "harvest_lambda_errors" {
  alarm_name          = "harvest-lambda-errors"
  comparison_operator = "GreaterThanOrEqualToThreshold"
  evaluation_periods  = 6
  metric_name         = "Errors"
  namespace           = "AWS/Lambda"
  period              = 3600
  statistic           = "Sum"
  threshold           = 15
  treat_missing_data  = "notBreaching"
  alarm_description   = "Monitors for errors in the harvest lambda"
  alarm_actions = [
    aws_sns_topic.ops_critical.arn
  ]

  dimensions = {
    FunctionName = aws_lambda_function.harvestcalls.function_name
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
}

resource "aws_iam_role_policy_attachments_exclusive" "harvester" {
  role_name = aws_iam_role.harvester.name
  policy_arns = [
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
  source_file      = "../build/bin/active_call_notifier/bootstrap"
  output_file_mode = "0666"
  output_path      = "../build/bin/active_call_notifier.zip"
}

resource "aws_lambda_function" "active_call_notifier" {
  function_name    = "ActiveCallNotifier"
  description      = "Sends SMS Notifications from Call Events"
  filename         = data.archive_file.active_call_notifier.output_path
  memory_size      = 128
  runtime          = "provided.al2023"
  handler          = "bootstrap"
  role             = aws_iam_role.active_call_notifier.arn
  source_code_hash = data.archive_file.active_call_notifier.output_base64sha256
  timeout          = 60

  environment {
    variables = {
      SMS_FROM           = var.SMS_FROM
      SMS_TO             = var.SMS_TO
      TWILIO_ACCOUNT_SID = var.TWILIO_ACCOUNT_SID
      TWILIO_API_KEY     = var.TWILIO_API_KEY
      TWILIO_API_SECRET  = var.TWILIO_API_SECRET
    }
  }
}

# the notifier should never fail, this is always worth looking at
resource "aws_cloudwatch_metric_alarm" "notifier_lambda_errors" {
  alarm_name          = "notifier-lambda-errors"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 1
  metric_name         = "Errors"
  namespace           = "AWS/Lambda"
  period              = 1800
  statistic           = "Sum"
  treat_missing_data  = "notBreaching"
  threshold           = 3
  alarm_description   = "Monitors for errors in the notifier lambda"
  alarm_actions = [
    aws_sns_topic.ops_critical.arn
  ]

  dimensions = {
    FunctionName = aws_lambda_function.active_call_notifier.function_name
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
}

resource "aws_iam_role_policy_attachments_exclusive" "active_call_notifier" {
  role_name = aws_iam_role.active_call_notifier.name
  policy_arns = [
    local.lambda_default_role_arn,
    aws_iam_policy.active_call_notifier.arn
  ]
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
        "dynamodb" : {
          "NewImage" : {
            "streetName" : {
              "S" : var.STREET_NAMES
            }
          }
        }
      })
    }
  }
}
