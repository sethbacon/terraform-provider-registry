resource "registry_storage_config" "s3" {
  backend  = "s3"
  activate = true
  config = {
    bucket     = "my-registry-bucket"
    region     = "us-east-1"
    access_key = var.aws_access_key
    secret_key = var.aws_secret_key
  }
}

variable "aws_access_key" {
  type      = string
  sensitive = true
}

variable "aws_secret_key" {
  type      = string
  sensitive = true
}
