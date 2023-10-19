# Full end to end terraform example for setting up AWS NLB access logs observability stack

This Example will create the necessary components for a fully fledged observability stack analyzing AWS NLB access logs.

The created components of this transfer are:

- S3 bucket where the logs are stored by AWS as our Transfer source
- SQS queue for notifications on new logs file creation (in order to replicate data from the source to our target)
- VPC Network for the ClickHouse target
- ClickHouse as target database for analyzing the access logs
- S3 source for the transfer
- ClickHouse destination for the transfer
- S3 to ClickHouse transfer

Before running `terraform init` you should first set the necessary variables defined in the `variables.tf`according to your setup.

You will also need a `authorized_key.json` file in this folder to be able to create resource in your specified `project_id` space on DoubleCloud.

The `authorized_key.json` is an authentication file required to use a service account. See [DoubleCloud documentation](https://double.cloud/docs/en/public-api/tutorials/transfer-api-quickstart) for the instructions on how to obtain this file.

You can then run `terraform apply`

This will then create the S3 bucket, SQS queue, the ClickHouse Cluster and the necessary resources for a transfer from S3 to CH.
Whats left to do is to configure your NLB to push access logs into the newly created bucket. You can do this in your onw terraform module
by configuring the `aws_lb` resource with the necesary `access_logs` block:

```
resource "aws_lb" "nlb" {
  name = var.name
  ...
  access_logs {
    bucket = "bucket-name"
    prefix = "your-bucket-prefix"
    enabled = true
  }
}
```

Now all thats left to do is navigate to the created transfer in the UI and activate the the transfer.
