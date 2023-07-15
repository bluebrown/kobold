variable "CONTAINER_REGISTRY" { default = "docker.io" }
variable "VERSION" { default = "dev" }

group "default" { targets = ["gitgo", "gitexec"] }

target "default" {
  dockerfile  = "build/Dockerfile"
  args = { version = VERSION }
  attest = [
    "type=provenance,mode=min",
    "type=sbom"
  ]
  labels = {
    "org.opencontainers.image.author"  = "Nico Braun"
    "org.opencontainers.image.source"  = "https://github.com/bluebrown/kobold"
    "org.opencontainers.image.created" = timestamp()
  }
}

target "gitgo" {
  inherits  = ["default"]
  target    = "gitgo"
  tags      = ["${CONTAINER_REGISTRY}/bluebrown/kobold:${VERSION}"]
  output    = ["type=tar,dest=.dist/container-image-kobold.tar"]

}

target "gitexec" {
  inherits = ["default"]
  target   = "gitexec"
  tags     = ["${CONTAINER_REGISTRY}/bluebrown/kobold:${VERSION}-gitexec"]
  output    = ["type=tar,dest=.dist/container-image-kobold-gitexec.tar"]
}
