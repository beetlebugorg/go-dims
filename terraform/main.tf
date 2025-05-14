

module "s3-object-lambda" {
  source             = "./s3-object-lambda"
  aws_region         = var.aws_region
  source_bucket_name = var.bucket_name
  environment = {
    DIMS_DEV_MODE    = "true"
    DIMS_DEBUG_MODE  = "true"
    DIMS_SIGNING_KEY = "devmode"
  }
  lambda_zip_file = "../build/lambda.zip"
}