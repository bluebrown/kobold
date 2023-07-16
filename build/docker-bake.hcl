variable "CONTAINER_REGISTRY" { default = "docker.io" }
variable "VERSION" { default = "dev" }

group "default" { targets = ["gitgo", "gitexec"] }

target "gitgo" {
  labels = {
    "org.opencontainers.image.author"  = "Nico Braun"
    "org.opencontainers.image.source"  = "https://github.com/bluebrown/kobold"
    "org.opencontainers.image.created" = timestamp()
  }
  dockerfile = "build/Dockerfile"
  args       = { CGO_ENABLED = "1", BUILD_TAG = "gitgo", VERSION = VERSION}
  tags       = ["${CONTAINER_REGISTRY}/bluebrown/kobold:${VERSION}"]
  # attest = ["type=provenance,mode=min", "type=sbom"]
}

target "gitexec" {
  inherits = ["gitgo"]
  args     = { BASE_IMAGE = "docker.io/alpine/git:v2.36.3", CGO_ENABLED = "0", BUILD_TAG = "gitexec", VERSION = VERSION}
  tags     = ["${CONTAINER_REGISTRY}/bluebrown/kobold:${VERSION}-gitexec"]
  # output   = ["type=tar,dest=.dist/container-image-kobold-gitexec.tar"]
}
