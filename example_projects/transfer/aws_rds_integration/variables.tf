// This variable is to set the
// AWS region that everything will be
// created in
variable "aws_region" {
  default = "eu-west-2" // london
}

// This variable is to set the
// CIDR block for the VPC
variable "vpc_cidr_block" {
  description = "CIDR block for VPC"
  type        = string
  default     = "10.0.0.0/16"
}

// This variable holds the
// number of public and private subnets
variable "subnet_count" {
  description = "Number of subnets"
  type        = map(number)
  default = {
    public  = 1,
    private = 2
  }
}

// This variable contains the configuration
// settings for the EC2 and RDS instances
variable "settings" {
  description = "Configuration settings"
  type        = map(any)
  default = {
    "database" = {
      allocated_storage   = 10            // storage in gigabytes
      engine              = "postgres"    // engine type
      engine_version      = "15.4"        // engine version
      instance_class      = "db.t3.micro" // rds instance type
      db_name             = "chinook"     // database name
      identifier           = "chinook"     // database identifier
      skip_final_snapshot = true
    },
    "web_app" = {
      count         = 1          // the number of EC2 instances
      instance_type = "t3.micro" // the EC2 instance
    }
  }
}

// This variable contains the CIDR blocks for
// the public subnet. I have only included 4
// for this tutorial, but if you need more you
// would add them here
variable "public_subnet_cidr_blocks" {
  description = "Available CIDR blocks for public subnets"
  type        = list(string)
  default = [
    "10.0.1.0/24",
    "10.0.2.0/24",
    "10.0.3.0/24",
    "10.0.4.0/24"
  ]
}

// This variable contains the CIDR blocks for
// the public subnet. I have only included 4
// for this tutorial, but if you need more you
// would add them here
variable "private_subnet_cidr_blocks" {
  description = "Available CIDR blocks for private subnets"
  type        = list(string)
  default = [
    "10.0.101.0/24",
    "10.0.102.0/24",
    "10.0.103.0/24",
    "10.0.104.0/24",
  ]
}

// This variable contains your IP address. This
// is used when setting up the SSH rule on the
// web security group
variable "my_ip" {
  description = "Your IP address"
  type        = string
  sensitive   = true
}
// This variable contains your IP address. This
// is used when setting up the SSH rule on the
// web security group
variable "my_ipv6" {
  description = "Your IPv6 address"
  type        = string
  sensitive   = true
}

// This variable contains the database master user
// We will be storing this in a secrets file
variable "db_username" {
  description = "Database master user"
  type        = string
  sensitive   = true
}

// This variable contains the database master password
// We will be storing this in a secrets file
variable "db_password" {
  description = "Database master user password"
  type        = string
  sensitive   = true
}

// Stage 2 Variables

variable "dwh_ipv4_cidr" {
  type        = string
  description = "CIDR of a used vpc"
  default     = "172.16.0.0/16"
}
variable "dc_project_id" {
  type        = string
  description = "ID of the DoubleCloud project in which to create resources"
}
