
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