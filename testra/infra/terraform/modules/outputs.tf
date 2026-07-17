output "ecr_repository_urls" {
  description = "URLs of the created ECR repositories"
  value       = { for name, repo in aws_ecr_repository.testra : name => repo.repository_url }
}
