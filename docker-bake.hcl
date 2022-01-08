group "default" {
  targets = ["image-local"]
}

// Special target: https://github.com/docker/metadata-action#bake-definition
target "docker-metadata-action" {
  tags = ["faq:local"]
}

group "validate" {
  targets = ["lint"]
}

target "lint" {
  target = "lint"
  output = ["type=cacheonly"]
}

target "test" {
  target = "test-coverage"
  output = ["."]
}

target "artifact" {
  target = "artifacts"
  secret = ["id=GITHUB_TOKEN,env=GITHUB_TOKEN"]
  output = ["./dist"]
}

target "artifact-all" {
  inherits = ["artifact"]
  platforms = [
    "darwin/amd64",
    "darwin/arm64",
    "linux/amd64",
    "linux/arm/v7",
    "linux/arm64",
    "linux/ppc64le",
    "linux/riscv64",
    "linux/s390x"
  ]
}

target "image" {
  inherits = ["docker-metadata-action"]
  secret = ["id=GITHUB_TOKEN,env=GITHUB_TOKEN"]
}

target "image-local" {
  inherits = ["image"]
  output = ["type=docker"]
}

target "image-all" {
  inherits = ["image"]
  platforms = [
    "linux/amd64",
    "linux/arm/v7",
    "linux/arm64",
    "linux/ppc64le",
    "linux/riscv64",
    "linux/s390x"
  ]
}
