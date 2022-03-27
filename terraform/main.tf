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
