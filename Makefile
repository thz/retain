IMAGE=thzpub/retain
VERSION=1.1.0

ARCHS:=amd64 i386 arm32v7 arm64v8


build: $(ARCHS:%=dockerbuild-%) build-solaris_amd64
push: $(ARCHS:%=push-%)

build-solaris_amd64:
	CGO_ENABLED=0 GOOS=solaris GOARCH=amd64 go build -o retain-$(@:build-%=%)

# manifest creation needs all referenced images pushed to
# the registry first.
define manifest_create
	docker manifest create $(1) \
		$(IMAGE):i386-$(VERSION) \
		$(IMAGE):amd64-$(VERSION) \
		$(IMAGE):arm32v7-$(VERSION) \
		$(IMAGE):arm64v8-$(VERSION)
	docker manifest annotate $(1) $(IMAGE):i386-$(VERSION)    --os=linux --arch=386
	docker manifest annotate $(1) $(IMAGE):amd64-$(VERSION)   --os=linux --arch=amd64
	docker manifest annotate $(1) $(IMAGE):arm32v7-$(VERSION) --os=linux --arch=arm
	docker manifest annotate $(1) $(IMAGE):arm64v8-$(VERSION) --os=linux --arch=arm64
endef
manifest: push-i386 push-amd64 push-arm32v7 push-arm64v8
	$(call manifest_create,$(IMAGE):$(VERSION))
	$(call manifest_create,$(IMAGE):latest)
	@echo "manifest can be purged with: make manifest-purge"
	@echo "manifest can be pushed with: make manifest-push"
manifest-push:
	docker manifest push $(IMAGE):$(VERSION)
	docker manifest push $(IMAGE):latest
manifest-purge:
	docker manifest push --purge $(IMAGE):$(VERSION)
	docker manifest push --purge $(IMAGE):latest


define go_arch_build
	docker build --build-arg GOARCH=$(1) -t $(IMAGE):$(2)-$(VERSION) .
endef

dockerbuild-i386: qemu
	$(call go_arch_build,386,i386)

dockerbuild-amd64: qemu
	$(call go_arch_build,amd64,amd64)

dockerbuild-arm32v7: qemu
	$(call go_arch_build,arm,arm32v7)

dockerbuild-arm64v8: qemu
	$(call go_arch_build,arm64,arm64v8)

push-%: dockerbuild-%
	docker push $(IMAGE):$(@:push-%=%)-$(VERSION)

release: build-amd64 manifest-purge manifest manifest-push

# binfmt support for cross-building docker images
qemu:
	mkdir -p qemu-statics
	cp /usr/bin/qemu-arm-static qemu-statics/
	cp /usr/bin/qemu-aarch64-static qemu-statics/
