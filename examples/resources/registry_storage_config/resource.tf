resource "registry_storage_config" "s3" {
  backend  = "s3"
  activate = true
  config = {
    s3_bucket          = "my-registry-bucket"
    s3_region          = "us-east-1"
    s3_access_key_id   = var.aws_access_key
    s3_secret_access_key = var.aws_secret_key
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
