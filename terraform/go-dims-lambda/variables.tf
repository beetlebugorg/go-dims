
variable "aws_region" {
  description = "The AWS region to deploy to."
  type        = string
}

variable "environment" {
  description = "Environment variables for the Lambda function."
  type        = map(string)
}

variable "platform" {
  description = "Platform architecture to deploy. Options: arm64 or amd64."
  type        = string
  default     = "arm64"
}
variable "authorization_type" {
  description = "Authorization for the Lambda function URL. AWS_IAM requires a signed caller. NONE serves the endpoint to the public internet."
  type        = string
  default     = "AWS_IAM"
}

variable "allow_origins" {
  description = "Origins the Lambda function URL answers a cross-origin request from. Name each site to scope it."
  type        = list(string)
  default     = ["*"]
}

variable "allow_headers" {
  description = "Request headers a cross-origin caller may send."
  type        = list(string)
  default     = ["*"]
}
