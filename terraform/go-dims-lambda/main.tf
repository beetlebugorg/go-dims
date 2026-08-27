// Copyright 2025 Jeremy Collins. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

provider "aws" {
  region = var.aws_region
}

resource "aws_iam_role" "go-dims" {
  name = "go-dims-lambda"

  assume_role_policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Action = "sts:AssumeRole",
        Effect = "Allow",
        Principal = {
          Service = "lambda.amazonaws.com"
        }
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "go-dims" {
  role       = aws_iam_role.go-dims.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_lambda_function" "go-dims" {
  function_name = "go-dims"
  role          = aws_iam_role.go-dims.arn

  timeout       = 30
  memory_size   = 512
  package_type  = "Zip"
  runtime       = "provided.al2023"
  handler       = "bootstrap"
  architectures = [var.platform]

  filename         = "../build/lambda-${var.platform}.zip"
  source_code_hash = filebase64sha256("../build/lambda-${var.platform}.zip")

  environment {
    variables = var.environment
  }
}

resource "aws_lambda_function_url" "go-dims" {
  function_name      = aws_lambda_function.go-dims.function_name
  authorization_type = var.authorization_type
  invoke_mode        = "RESPONSE_STREAM"

  cors {
    allow_origins = ["*"]
    allow_methods = ["GET"]
    allow_headers = ["*"]
  }

  lifecycle {
    precondition {
      condition     = var.authorization_type != "NONE" || lookup(var.environment, "DIMS_DEVELOPMENT_MODE", "false") != "true"
      error_message = "A public function URL requires signature validation. Set authorization_type to AWS_IAM, or set development_mode to false."
    }
  }
}

output "function_name" {
  value = aws_lambda_function.go-dims.function_name
}

output "function_url" {
  value = aws_lambda_function_url.go-dims.function_url
}