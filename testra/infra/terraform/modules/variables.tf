variable "environment" {
  type        = string
  description = "Deployment environment (staging or production)"
}

variable "region" {
  type        = string
  description = "AWS region"
  default     = "ap-southeast-1"
}

variable "service_names" {
  type        = list(string)
  description = "Names of services that require ECR repositories"
  default     = ["api", "worker", "ml", "web", "migrator"]
}
