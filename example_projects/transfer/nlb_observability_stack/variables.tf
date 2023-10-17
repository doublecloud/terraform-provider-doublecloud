# AWS resource needed variables
variable "tags" {
  type = map(string)
  default = {
    Example = "test"
  }
  description = "Tags for AWS resources"
}

variable "bucket_name" {
  type        = string
  default     = "double-cloud-nlb-logs-bucket"
  description = "Name of the sample bucket where the NLB logs should be stored"
}

variable "bucket_prefix" {
  type        = string
  default     = "metrics"
  description = "Bucket prefix for NLB logs in S3 bucket"
}

variable "nlb_name" {
  type        = string
  default     = "preprod-metric-exp-dt-tls"
  description = "Name of the load balancer to enable acces logs for"
}

variable "sqs_name" {
  type        = string
  default     = "s3-nlb-logs-queue"
  description = "Name of the AWS SQS where the bucktect creation events should be sent to"
}

variable "aws_profile" {
  type        = string
  default     = "your-profile"
  description = "AWS profile to use for API calls"
}

variable "region" {
  type        = string
  default     = "eu-central-1"
  description = "Region where to create S3 bucket"
}

variable "vpc_id" {
  type        = string
  default     = "example"
  description = "ID of the VPC the NLB is tied to"
}

variable "fqdn" {
  type        = string
  default     = "test"
  description = "DNS name associated with the NLB"
}

variable "subnet_ids" {
  type = list(string)
  default = ["test"]
  description = "Subnet ID's for NLB"
}

# DC resource needed variables
variable "project_id" {
  type        = string
  default     = "example"
  description = "ID of the DoubleCloud project in which to create resources"
}

variable "service_account_endpoint" {
  type        = string
  default     = "example.com:123"
  description = "DoubleCloud endpoint for service account key authentication"
}

variable "service_account_token_endpoint" {
  type        = string
  default     = "https://example.com/oauth/token"
  description = "DoubleCloud token verification endpoint"
}

variable "aws_account_id" {
  type        = string
  default     = "1234"
  description = "AWS Account ID to use for S3 bucket creation"
}

# Transfer needed varibales
variable "aws_access_key_id" {
  type        = string
  default     = "access ky id"
  description = "AWS Access account ID for transfer to be able to read from S3 bucket"
}
variable "aws_access_key_secret" {
  type        = string
  default     = "your secret"
  description = "AWS Access account secret key for transfer to be able to read from S3 bucket"
}

variable "endpoint" {
  type        = string
  default     = ""
  description = "Endpoint to connect for fetching S3 objects"
}
