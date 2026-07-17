terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }

  required_version = ">= 1.8"
}

data "aws_caller_identity" "current" {}

data "aws_region" "current" {}
