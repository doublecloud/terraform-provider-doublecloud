# AWS resource needed variables
variable "tags" {
  type        = map(string)
  description = "Tags for AWS resources"
}

variable "bucket_name" {
  type        = string
  description = "Name of the sample bucket where the NLB logs should be stored"
}

variable "bucket_prefix" {
  type        = string
  description = "Bucket prefix for NLB logs in S3 bucket"
}

variable "bucket_encrypted" {
  type        = bool
  default     = true
  description = "Set if bucket should be server side encrpyted"
}

variable "sqs_name" {
  type        = string
  default     = "s3-nlb-logs-queue"
  description = "Name of the AWS SQS where the bucktect creation events should be sent to"
}

variable "aws_profile" {
  type        = string
  description = "AWS profile to use for API calls"
}

variable "region" {
  type        = string
  default     = "eu-central-1"
  description = "Region where to create S3 bucket"
}

# DC resource needed variables
variable "project_id" {
  type        = string
  description = "ID of the DoubleCloud project in which to create resources"
}

variable "cloud_type" {
  type        = string
  default     = "aws"
  description = "Specifies cloud provider"
}

variable "ipv4_cidr" {
  type        = string
  default     = "10.0.0.0/16"
  description = "CIDR of used vpc"
}

variable "aws_account_id" {
  type        = string
  description = "AWS Account ID to use for S3 bucket creation"
}

variable "network_name" {
  type        = string
  default     = "nlb-example-network"
  description = "Name for the network created in DoubleCloud"
}

variable "clickhouse_cluster_name" {
  type        = string
  default     = "nlb-logs-clickhouse-cluster"
  description = "Name for the managed ClickHouse cluster created in DoubleCloud"
}

variable "clickhouse_cluster_resource_preset" {
  type        = string
  default     = "s2-c2-m4"
  description = "Specs for the managed ClickHouse cluster created in DoubleCloud"
}

# Transfer needed varibales
variable "aws_access_key_id" {
  type        = string
  sensitive   = true
  description = "AWS Access account ID for transfer to be able to read from S3 bucket"
}
variable "aws_access_key_secret" {
  type        = string
  sensitive   = true
  description = "AWS Access account secret key for transfer to be able to read from S3 bucket"
}

variable "endpoint" {
  type        = string
  default     = ""
  description = "Endpoint to connect for fetching S3 objects, leave empty for AWS"
}

variable "transfer_source_name" {
  type        = string
  default     = "nlb-s3-s32ch-source"
  description = "Name of the source endpoint for the DoubleCloud transfer"
}

variable "transfer_source_table_name" {
  type        = string
  default     = "nlb_access_logs"
  description = "Name for the resulting table in the destination endpoint"
}

variable "transfer_source_table_namespace" {
  type        = string
  default     = "aws"
  description = "Name for the resulting db schema in the destination endpoint"
}

variable "transfer_target_name" {
  type        = string
  default     = "nlb-ch-s32ch-target"
  description = "Name of the target endpoint for the Doublecloud transfer"
}

variable "transfer_name" {
  type        = string
  default     = "nlb-logs-s32ch"
  description = "DoubleCloud transfer name"
}
