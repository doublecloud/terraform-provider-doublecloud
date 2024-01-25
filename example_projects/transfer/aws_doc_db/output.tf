// This will output the public IP of the web server
output "public_ip" {
  description = "The public IP address of the web server"
  // We are grabbing it from the Elastic IP
  value       = aws_eip.jump_host.public_ip

  // This output waits for the Elastic IPs to be created and distributed
  depends_on = [aws_eip.jump_host]
}

// This will output the the public DNS address of the web server
output "public_dns" {
  description = "The public DNS address of the web server"
  // We are grabbing it from the Elastic IP
  value       = aws_eip.jump_host.public_dns

  depends_on = [aws_eip.jump_host]
}

// This will output the database endpoint
output "doc_db_connection" {
  description = "The endpoint of the database"
  value       = aws_docdb_cluster.service.endpoint
}
