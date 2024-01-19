// Create a VPC named "docdb_vpc"
resource "aws_vpc" "docdb_vpc" {
  // Here we are setting the CIDR block of the VPC
  // to the "vpc_cidr_block" variable
  cidr_block = var.vpc_cidr_block
  // We want DNS hostnames enabled for this VPC
  enable_dns_hostnames = true

  // We are tagging the VPC with the name "docdb_vpc"
  tags = {
    Name = "docdb_vpc"
  }
}

// Create an internet gateway named "docdb_igw"
// and attach it to the "docdb_vpc" VPC
resource "aws_internet_gateway" "docdb_igw" {
  // Here we are attaching the IGW to the
  // docdb_vpc VPC
  vpc_id = aws_vpc.docdb_vpc.id

  // We are tagging the IGW with the name docdb_igw
  tags = {
    Name = "docdb_igw"
  }
}

// Create a group of public subnets based on the variable subnet_count.public
resource "aws_subnet" "docdb_public_subnet" {
  // Put the subnet into the "docdb_vpc" VPC
  vpc_id = aws_vpc.docdb_vpc.id

  cidr_block = var.public_subnet_cidr_block

  // We are grabbing the availability zone from the data object we created earlier
  // Since this is a list, we are grabbing the name of the element based on count,
  // so since count is 1, and our region is us-east-2, this should grab us-east-2a
  availability_zone = data.aws_availability_zones.available.names[0]

  // We are tagging the subnet with a name of "docdb_public_subnet_" and
  // suffixed with the count
  tags = {
    Name = "docdb public subnet"
  }
}

// Create a group of private subnets based on the variable subnet_count.private
resource "aws_subnet" "docdb_private_subnet" {
  count = 2
  // Put the subnet into the "docdb_vpc" VPC
  vpc_id = aws_vpc.docdb_vpc.id

  cidr_block = var.private_subnet_cidr_block[count.index]

  // We are grabbing the availability zone from the data object we created earlier
  // Since this is a list, we are grabbing the name of the element based on count,
  // since count is 2, and our region is us-east-2, the first subnet should
  // grab us-east-2a and the second will grab us-east-2b
  availability_zone = data.aws_availability_zones.available.names[count.index]

  // We are tagging the subnet with a name of "docdb_private_subnet_" and
  // suffixed with the count
  tags = {
    Name = "docdb private subnet"
  }
}

// Create a public route table named "docdb_public_rt"
resource "aws_route_table" "docdb_public_rt" {
  // Put the route table in the "docdb_vpc" VPC
  vpc_id = aws_vpc.docdb_vpc.id
}

// Since this is the public route table, it will need
// access to the internet. So we are adding a route with
// a destination of 0.0.0.0/0 and targeting the Internet
// Gateway "docdb_igw"
resource "aws_route" "docdb_public_rt_igw" {
  provider               = aws
  route_table_id         = aws_route_table.docdb_public_rt.id
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = aws_internet_gateway.docdb_igw.id
}

// Here we are going to add the public subnets to the
// "docdb_public_rt" route table
resource "aws_route_table_association" "public" {
  // Here we are making sure that the route table is
  // "docdb_public_rt" from above
  route_table_id = aws_route_table.docdb_public_rt.id

  // This is the subnet ID. Since the "docdb_public_subnet" is a
  // list of the public subnets, we need to use count to grab the
  // subnet element and then grab the id of that subnet
  subnet_id = aws_subnet.docdb_public_subnet.id
}

// Create a private route table named "docdb_private_rt"
resource "aws_route_table" "docdb_private_rt" {
  // Put the route table in the "docdb_VPC" VPC
  vpc_id = aws_vpc.docdb_vpc.id

  // Since this is going to be a private route table,
  // we will not be adding a route
}

// Here we are going to add the private subnets to the
// route table "docdb_private_rt"
resource "aws_route_table_association" "private" {
  count = 2
  // Here we are making sure that the route table is
  // "docdb_private_rt" from above
  route_table_id = aws_route_table.docdb_private_rt.id

  // This is the subnet ID. Since the "docdb_private_subnet" is a
  // list of private subnets, we need to use count to grab the
  // subnet element and then grab the ID of that subnet
  subnet_id = aws_subnet.docdb_private_subnet[count.index].id
}

// Create a security for the EC2 instances called "docdb_sg"
resource "aws_security_group" "docdb_sg" {
  // Basic details like the name and description of the SG
  name        = "docdb_sg"
  description = "Security group for DocDB"
  // We want the SG to be in the "docdb_vpc" VPC
  vpc_id = aws_vpc.docdb_vpc.id

  // The first requirement we need to meet is "EC2 instances should
  // be accessible anywhere on the internet via HTTP." So we will
  // create an inbound rule that allows all traffic through
  // TCP port 80.
  ingress {
    description = "Allow all traffic through HTTP"
    from_port   = "80"
    to_port     = "80"
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  // The second requirement we need to meet is "Only you should be
  // "able to access the EC2 instances via SSH." So we will create an
  // inbound rule that allows SSH traffic ONLY from your IP address
  ingress {
    description = "Allow SSH from my computer"
    from_port   = "22"
    to_port     = "22"
    protocol    = "tcp"
    // This is using the variable "my_ip"
    cidr_blocks = ["${var.my_ip}/32"]
  }

  // This outbound rule is allowing all outbound traffic
  // with the EC2 instances
  egress {
    description = "Allow all outbound traffic"
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  // Here we are tagging the SG with the name "docdb_sg"
  tags = {
    Name = "docdb_sg"
  }
}

// Create a security group for the RDS instances called "docdb_db_sg"
resource "aws_security_group" "docdb_db_sg" {
  // Basic details like the name and description of the SG
  name        = "docdb_db_sg"
  description = "Security group for tutorial databases"
  // We want the SG to be in the "docdb_vpc" VPC
  vpc_id = aws_vpc.docdb_vpc.id

  // The third requirement was "RDS should be on a private subnet and
  // inaccessible via the internet." To accomplish that, we will
  // not add any inbound or outbound rules for outside traffic.

  // The fourth and finally requirement was "Only the EC2 instances
  // should be able to communicate with RDS." So we will create an
  // inbound rule that allows traffic from the EC2 security group
  // through TCP port 5432, which is the port that DocDB
  // communicates through
  ingress {
    description     = "Allow DocDB traffic from only the web sg"
    from_port       = "27017"
    to_port         = "27017"
    protocol        = "tcp"
    security_groups = [aws_security_group.docdb_sg.id]
  }

  // So we have VPC in DC and we need to allow it to Postgres
  ingress {
    description = "Allow DocDB traffic from Data Network"
    from_port   = "27017"
    to_port     = "27017"
    protocol    = "tcp"
    cidr_blocks = [var.dwh_ipv4_cidr]
  }

  // Here we are tagging the SG with the name "docdb_db_sg"
  tags = {
    Name = "docdb_db_sg"
  }
}

// Create a db subnet group named "docdb_db_subnet_group"
resource "aws_docdb_subnet_group" "docdb_db_subnet_group" {
  // The name and description of the db subnet group
  name        = "docdb_docdb_subnet_group"
  description = "DB subnet group for tutorial"

  // Since the db subnet group requires 2 or more subnets, we are going to
  // loop through our private subnets in "docdb_private_subnet" and
  // add them to this db subnet group
  subnet_ids = [for subnet in aws_subnet.docdb_private_subnet : subnet.id]
}


resource "aws_docdb_cluster" "service" {
  skip_final_snapshot    = true
  db_subnet_group_name   = aws_docdb_subnet_group.docdb_db_subnet_group.name
  cluster_identifier     = "tf-demo-docdb"
  engine                 = "docdb"
  master_username        = var.db_username
  master_password        = var.db_password
  vpc_security_group_ids = [aws_security_group.docdb_db_sg.id]
}

resource "aws_docdb_cluster_instance" "service" {
  count              = 1
  identifier         = "${aws_docdb_cluster.service.cluster_identifier}-${count.index}"
  cluster_identifier = aws_docdb_cluster.service.id
  instance_class     = var.settings.database.instance_class
}


// Create a key pair named "docdb_kp"
resource "aws_key_pair" "docdb_kp" {
  // Give the key pair a name
  key_name = "docdb_kp_doc_db"

  // This is going to be the public key of our
  // ssh key. The file directive grabs the file
  // from a specific path. Since the public key
  // was created in the same directory as main.tf
  // we can just put the name
  public_key = file("~/.ssh/id_rsa.pub")
}

// Create an EC2 instance named "jump_host"
// this instance jump-host to docdb
resource "aws_instance" "jump_host" {
  // Here we need to select the ami for the EC2. We are going to use the
  // ami data object we created called ubuntu, which is grabbing the latest
  // Ubuntu 20.04 ami
  ami = data.aws_ami.ubuntu.id

  // This is the instance type of the EC2 instance. The variable
  // settings.web_app.instance_type is set to "t2.micro"
  instance_type = var.settings.jump_host.instance_type

  // The subnet ID for the EC2 instance. Since "docdb_public_subnet" is a list
  // of public subnets, we want to grab the element based on the count variable.
  // Since count is 1, we will be grabbing the first subnet in
  // "docdb_public_subnet" and putting the EC2 instance in there
  subnet_id = aws_subnet.docdb_public_subnet.id

  // The key pair to connect to the EC2 instance. We are using the "docdb_kp" key
  // pair that we created
  key_name = aws_key_pair.docdb_kp.key_name

  // The security groups of the EC2 instance. This takes a list, however we only
  // have 1 security group for the EC2 instances.
  vpc_security_group_ids = [aws_security_group.docdb_sg.id]

  // We are tagging the EC2 instance with the name "docdb_db_" followed by
  // the count index
  tags = {
    Name = "jump_host_docdb"
  }
}

// Create an Elastic IP named "jump_host" for each
// EC2 jump_host instance
resource "aws_eip" "jump_host" {
  instance = aws_instance.jump_host.id

  // We want the Elastic IP to be in the VPC
  vpc = true

  // Here we are tagging the Elastic IP with the name
  // "docdb_jump_host_eip"
  tags = {
    Name = "docdb_jump_host_eip"
  }
}
