resource "aws_lambda_function" "default" {
  function_name = "aws-iam-keys-rotation"
  s3_bucket     = var.storage_bucket
  s3_key        = "${var.build_definition_name}-${var.build_number}.zip"
  role          = aws_iam_role.default.arn
  handler       = "main"
  runtime       = "go1.x"
  description   = "Ado Iam Keys Rotation"

  environment {
    variables = {
      GO_ENVIRONMENT            = var.environment
      ADO_ORG                   = var.ado.org
      ADO_PROJECT               = var.ado.project
      ADO_SERVICE_ENDPOINT_NAME = var.ado.endpoint_name
      ADO_USER_SSM              = var.ado.user
      ADO_TOKEN_SSM             = var.ado.token
      IAM_USER_NAME             = var.iam_user_name
      DISCORD_WEBHOOK_URL_SSM   = var.discord_webhook_url
    }
  }

  memory_size = 128
  timeout     = 300
}

resource "aws_iam_role" "default" {
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
  managed_policy_arns = [
    aws_iam_policy.role.arn,
    "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
  ]
}

resource "aws_iam_policy" "role" {
  policy = data.aws_iam_policy_document.role.json
}

resource "aws_lambda_permission" "default" {
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.default.arn
  principal     = "events.amazonaws.com"
  source_arn    = aws_cloudwatch_event_rule.default.arn
}

resource "aws_cloudwatch_event_rule" "default" {
  description         = "Event To Trigger Lambda Function"
  schedule_expression = "cron(0 0 ? * * *)"
  is_enabled          = true
}
