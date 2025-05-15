

module "s3-object-lambda" {
  source             = "./s3-object-lambda"
  aws_region         = var.aws_region
  source_bucket_name = var.bucket_name
  environment = {
    DIMS_DEVELOPMENT_MODE = "true"
    DIMS_DEBUG_MODE       = "true"
    DIMS_SIGNING_KEY      = "devmode"
    DIMS_S3_BUCKET        = var.bucket_name
  }
  lambda_zip_file = "../build/lambda.zip"
}