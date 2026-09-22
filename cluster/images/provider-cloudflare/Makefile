# ====================================================================================
# Setup Project

PLATFORMS := linux_amd64 linux_arm64
include ../../../build/makelib/common.mk

# ====================================================================================
# Options
IMAGE ?= $(BUILD_REGISTRY)/provider-cloudflare:$(VERSION)
TARGETOS := $(word 1, $(subst _, ,$(PLATFORM)))
TARGETARCH := $(word 2, $(subst _, ,$(PLATFORM)))
include ../../../build/makelib/imagelight.mk

# ====================================================================================
# Image Build
img.build:
	@$(INFO) docker build $(IMAGE)
	@$(MAKE) -C ../../.. go.build PLATFORM=$(PLATFORM)
	@$(MAKE) img.build.shared
	@$(OK) docker build $(IMAGE)

img.build.shared:
	@cp Dockerfile $(IMAGE_TEMP_DIR)
	@cp -r $(OUTPUT_DIR)/bin/ $(IMAGE_TEMP_DIR)/
	@cp -r $(ROOT_DIR)/package/crds $(IMAGE_TEMP_DIR)/
	@cp $(ROOT_DIR)/package/crossplane.yaml $(IMAGE_TEMP_DIR)/
	@docker build $(BUILD_ARGS) --build-arg TARGETOS=$(TARGETOS) --build-arg TARGETARCH=$(TARGETARCH) -t $(IMAGE) $(IMAGE_TEMP_DIR)

# ====================================================================================
# Image Publish
# The vendored imagelight.mk invokes img.publish once with IMAGE_PLATFORMS=linux/amd64,linux/arm64
# but PLATFORM is only set for one arch at a time. We override img.publish to:
#   1. Build the amd64 image and push it as $(IMAGE)-amd64
#   2. Build the arm64 image and push it as $(IMAGE)-arm64
#   3. Create a multi-arch manifest index at $(IMAGE) combining the two
# This produces a real OCI image index (mediaType application/vnd.docker.distribution.manifest.list.v2+json)
# rather than a single-arch image masquerading as multi-arch.
img.publish:
	@$(INFO) docker push $(IMAGE) [multi-arch index]
	@$(MAKE) PLATFORM=linux_amd64 img.build IMAGE=$(IMAGE)-amd64
	@docker push $(IMAGE)-amd64
	@$(MAKE) PLATFORM=linux_arm64 img.build IMAGE=$(IMAGE)-arm64
	@docker push $(IMAGE)-arm64
	@docker buildx imagetools create -t $(IMAGE) $(IMAGE)-amd64 $(IMAGE)-arm64
	@$(OK) docker push $(IMAGE) [multi-arch index]

# ====================================================================================
# Image Promote
img.promote:
	@$(INFO) Promoting image from $(FROM_IMAGE) to $(TO_IMAGE)
	@docker buildx imagetools create -t $(TO_IMAGE) $(FROM_IMAGE)
	@$(OK) Promoted image from $(FROM_IMAGE) to $(TO_IMAGE)