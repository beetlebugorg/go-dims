module "go-dims-lambda" {
  source             = "./go-dims-lambda"
  aws_region         = var.aws_region
  authorization_type = var.authorization_type

  environment = merge(
    {
      DIMS_SIGNING_KEY      = var.signing_key
      DIMS_DEVELOPMENT_MODE = tostring(var.development_mode)
      DIMS_DEBUG_MODE       = tostring(var.debug_mode)
    },
    var.extra_environment,
  )
}
