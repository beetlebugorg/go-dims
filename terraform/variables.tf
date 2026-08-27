
variable "aws_region" {
  description = "The AWS region to deploy to."
  type        = string
  default     = "us-east-1"
}

variable "signing_key" {
  description = "Value for DIMS_SIGNING_KEY. Use at least 32 characters of random data."
  type        = string
  sensitive   = true

  validation {
    condition     = length(var.signing_key) >= 16
    error_message = "The signing key must be at least 16 characters."
  }
}

variable "development_mode" {
  description = "Set DIMS_DEVELOPMENT_MODE. Development mode accepts every request without a signature. Keep this false outside a local test."
  type        = bool
  default     = false
}

variable "debug_mode" {
  description = "Set DIMS_DEBUG_MODE. Debug mode writes request details to the log."
  type        = bool
  default     = false
}

variable "authorization_type" {
  description = "Authorization for the Lambda function URL. AWS_IAM requires a signed caller. NONE serves the endpoint to the public internet."
  type        = string
  default     = "AWS_IAM"

  validation {
    condition     = contains(["AWS_IAM", "NONE"], var.authorization_type)
    error_message = "The authorization type must be AWS_IAM or NONE."
  }
}

variable "platform" {
  description = "Architecture to deploy. The build writes build/lambda-linux-<platform>.zip."
  type        = string
  default     = "arm64"

  validation {
    condition     = contains(["arm64", "amd64"], var.platform)
    error_message = "The platform must be arm64 or amd64."
  }
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

variable "extra_environment" {
  description = "Additional environment variables for the Lambda function."
  type        = map(string)
  default     = {}
}
