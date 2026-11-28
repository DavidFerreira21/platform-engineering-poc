variable "name" { type = string }
variable "azs" {
  type = list(string)
  validation {
    condition     = length(var.azs) == 3 && length(distinct(var.azs)) == 3
    error_message = "Informe três AZs distintas em ordem estável."
  }
}
variable "config" {
  type = object({
    vpc       = object({ enabled = bool, cidr = string, dnsHostnames = bool, dnsSupport = bool })
    nat       = object({ enabled = bool, mode = optional(string, "zonal") })
    layout    = optional(string, "spoke")
    endpoints = object({ s3 = object({ enabled = bool }), ecr = object({ enabled = bool }) })
  })
  validation {
    condition = (
      contains(["hub", "spoke"], var.config.layout) &&
      contains(["regional", "zonal"], var.config.nat.mode) &&
      (var.config.layout != "hub" || !var.config.nat.enabled || var.config.nat.mode == "regional")
    )
    error_message = "Use o layout hub ou spoke. O layout hub requer NAT regional quando habilitado."
  }
}
