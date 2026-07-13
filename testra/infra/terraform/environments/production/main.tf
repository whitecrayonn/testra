terraform {
  backend "s3" {}
}

module "testra" {
  source = "../../modules"

  environment = "production"
  region      = "ap-southeast-1"
}
