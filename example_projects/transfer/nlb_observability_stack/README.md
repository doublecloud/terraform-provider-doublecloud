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

Before running `terraform init` you should first replace all the necessary variables in the `variables.tf` file with your actual values.

You will also need a `authorized_key.json` file in this folder to be able to create resource in your specified `project_id` space on DoubleCloud.

The `authorized_key.json` is an authentication file required to use a service account. See [DoubleCloud documentation](https://double.cloud/docs/en/public-api/tutorials/transfer-api-quickstart) for the instructions on how to obtain this file.

We assume in this example that the NLB is already existing and that we need to import its state into terraform:

```
terraform import module.nlb_module.aws_lb.nlb  the-arn-of-your-nlb
```

You can then run `terraform apply`

Once this is completes successfully, all you have to do is navigate to the created transfer in the UI and activate the the transfer.
