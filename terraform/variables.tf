
variable "aws_region" {
  description = "The AWS region to deploy to."
  type        = string
}

variable "bucket_name" {
  description = "The name of the source S3 bucket."
  type        = string
}