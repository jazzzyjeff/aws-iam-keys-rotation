variable "region" {
  type = string
}

variable "ado" {
  type = object({
    org           = string
    project       = string
    endpoint_name = string
    user          = string
    token         = string
  })
}

variable "environment" {
  type = string
}

variable "iam_user_name" {
  type = string
}

variable "discord_webhook_url" {
  type = string
}

variable "build_definition_name" {
  type = string
}

variable "build_number" {
  type = string
}

variable "storage_bucket" {
  type = string
}
