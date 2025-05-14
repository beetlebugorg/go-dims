
variable "aws_region" {
  description = "The AWS region to deploy to."
  type        = string
}

variable "source_bucket_name" {
  description = "The name of the source S3 bucket."
  type        = string
}

variable "environment" {
  description = "Environment variables for the Lambda function."
  type        = map(string)
}

variable "lambda_zip_file" {
  description = "Path to the Lambda zip file."
  type        = string
}