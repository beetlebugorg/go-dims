

module "go-dims-lambda" {
  source     = "./go-dims-lambda"
  aws_region = var.aws_region
  environment = {
    DIMS_DEVELOPMENT_MODE = "true"
    DIMS_DEBUG_MODE       = "true"
    DIMS_SIGNING_KEY      = "devmode"
  }
}