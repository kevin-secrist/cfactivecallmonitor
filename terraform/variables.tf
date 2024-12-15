variable "CPD_API_KEY" {
  type      = string
  sensitive = true
}

variable "CFD_API_KEY" {
  type      = string
  sensitive = true
}

variable "SMS_FROM" {
  type      = string
  sensitive = true
}

variable "SMS_TO" {
  type      = string
  sensitive = true
}

variable "TWILIO_ACCOUNT_SID" {
  type      = string
  sensitive = true
}

variable "TWILIO_API_KEY" {
  type      = string
  sensitive = true
}

variable "TWILIO_API_SECRET" {
  type      = string
  sensitive = true
}

variable "STREET_NAMES" {
  type      = list(string)
  sensitive = true
}

variable "OPS_EMAIL" {
  type      = string
  sensitive = true
}
