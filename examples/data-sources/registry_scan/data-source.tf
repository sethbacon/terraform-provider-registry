data "registry_scan" "example" {
  id = "00000000-0000-0000-0000-000000000000"
}

output "scan_passed" {
  value = data.registry_scan.example.passed
}
