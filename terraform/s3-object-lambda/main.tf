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

resource "aws_iam_role" "dims_lambda_role" {
  name = "dims_lambda_exec_role"

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

resource "aws_iam_policy" "s3_object_lambda_write_response" {
  name   = "allow-s3-object-lambda-write-response"
  policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Effect   = "Allow",
        Action   = [
          "s3-object-lambda:WriteGetObjectResponse",
          "s3:GetObject",
          "s3:ListBucket"
        ]
        Resource = "*"
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "lambda_write_response" {
  role       = aws_iam_role.dims_lambda_role.name
  policy_arn = aws_iam_policy.s3_object_lambda_write_response.arn
}

resource "aws_iam_role_policy_attachment" "lambda_basic" {
  role       = aws_iam_role.dims_lambda_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_lambda_function" "dims_lambda" {
  function_name = "dims-s3-object-lambda"
  role          = aws_iam_role.dims_lambda_role.arn

  timeout       = 30
  memory_size   = 512
  package_type  = "Zip"
  runtime       = "provided.al2023"
  handler       = "bootstrap"
  architectures = ["arm64"]

  filename         = var.lambda_zip_file
  source_code_hash = filebase64sha256(var.lambda_zip_file)

  environment {
    variables = var.environment
  }
}

resource "aws_s3_access_point" "dims_access_point" {
  name   = "dims-s3-object-lambda"
  bucket = var.source_bucket_name
}

resource "aws_s3control_object_lambda_access_point" "dims_object_lambda" {
  name = "dims-s3-object-lambda"

  configuration {
    supporting_access_point = aws_s3_access_point.dims_access_point.arn

    transformation_configuration {
      actions = ["GetObject"]

      content_transformation {
        aws_lambda {
          function_arn = aws_lambda_function.dims_lambda.arn
        }
      }
    }
  }

  depends_on = [aws_lambda_function.dims_lambda]
}

output "lambda_function_name" {
  value = aws_lambda_function.dims_lambda.function_name
}

output "object_lambda_arn" {
  value = aws_s3control_object_lambda_access_point.dims_object_lambda.arn
}
