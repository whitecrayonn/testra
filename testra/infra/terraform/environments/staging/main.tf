terraform {
  backend "s3" {}
}

module "testra" {
  source = "../../modules"

  environment = "staging"
  region      = "ap-southeast-1"
}
