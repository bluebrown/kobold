variable "CONTAINER_REGISTRY" { default = "docker.io" }
variable "VERSION" { default = "dev" }
variable "IS_LATEST" { default = "0" }

group "default" { targets = ["gitgo", "gitexec"] }

target "gitgo" {
  labels = {
    "org.opencontainers.image.author"  = "Nico Braun"
    "org.opencontainers.image.source"  = "https://github.com/bluebrown/kobold"
  }
  dockerfile = "build/Dockerfile"
  args       = { CGO_ENABLED = "1", BUILD_TAG = "gitgo", VERSION = VERSION}
  tags       = compact([
    "${CONTAINER_REGISTRY}/bluebrown/kobold:${VERSION}",
    equal(IS_LATEST, "1") ? "${CONTAINER_REGISTRY}/bluebrown/kobold:latest" : null
  ])
}

target "gitexec" {
  inherits = ["gitgo"]
  args     = { BASE_IMAGE = "docker.io/alpine/git:v2.36.3", CGO_ENABLED = "0", BUILD_TAG = "gitexec", VERSION = VERSION}
  tags     = compact([
    "${CONTAINER_REGISTRY}/bluebrown/kobold:${VERSION}-gitexec",
    equal(IS_LATEST, "1") ? "${CONTAINER_REGISTRY}/bluebrown/kobold:latest-gitexec" : null
  ])
}
