resource "aws_sqs_queue" "nlb_logs_queue" {
  name = var.sqs_name
  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "sqspolicy",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": "*",
      "Action": "sqs:SendMessage",
      "Resource": "arn:aws:sqs:*:*:${var.sqs_name}",
      "Condition": {
        "ArnEquals": { "aws:SourceArn": "${aws_s3_bucket.nlb_logs.arn}" }
      }
    }
  ]
}
POLICY
}
