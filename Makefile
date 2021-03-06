BIN_IMAGE = blockbook-build
DEB_IMAGE = blockbook-build-deb
PACKAGER = $(shell id -u):$(shell id -g)
NO_CACHE = false
UPDATE_VENDOR = 1

.PHONY: build build-debug test deb

build: .bin-image
	docker run -t --rm -e PACKAGER=$(PACKAGER) -e UPDATE_VENDOR=$(UPDATE_VENDOR) -v $(CURDIR):/src -v $(CURDIR)/build:/out $(BIN_IMAGE) make build

build-debug: .bin-image
	docker run -t --rm -e PACKAGER=$(PACKAGER) -e UPDATE_VENDOR=$(UPDATE_VENDOR) -v $(CURDIR):/src -v $(CURDIR)/build:/out $(BIN_IMAGE) make build-debug

test: .bin-image
	docker run -t --rm -e PACKAGER=$(PACKAGER) -e UPDATE_VENDOR=$(UPDATE_VENDOR) -v $(CURDIR):/src $(BIN_IMAGE) make test

test-all: .bin-image
	docker run -t --rm -e PACKAGER=$(PACKAGER) -e UPDATE_VENDOR=$(UPDATE_VENDOR) -v $(CURDIR):/src $(BIN_IMAGE) make test-all

deb: .deb-image clean-deb
	docker run -t --rm -e PACKAGER=$(PACKAGER) -e UPDATE_VENDOR=$(UPDATE_VENDOR) -v $(CURDIR):/src -v $(CURDIR)/build:/out $(DEB_IMAGE)

tools:
	docker run -t --rm -e PACKAGER=$(PACKAGER) -e UPDATE_VENDOR=$(UPDATE_VENDOR) -v $(CURDIR):/src -v $(CURDIR)/build:/out $(BIN_IMAGE) make tools

all: build-images deb

build-images:
	rm -f .bin-image .deb-image
	$(MAKE) .bin-image .deb-image

.bin-image:
	docker build --no-cache=$(NO_CACHE) -t $(BIN_IMAGE) build/bin
	@ docker images -q $(BIN_IMAGE) > $@

.deb-image: .bin-image
	docker build --no-cache=$(NO_CACHE) -t $(DEB_IMAGE) build/deb
	@ docker images -q $(DEB_IMAGE) > $@

clean: clean-bin clean-deb

clean-all: clean clean-images

clean-bin:
	find build -maxdepth 1 -type f -executable -delete

clean-deb:
	rm -f build/*.deb

clean-images: clean-bin-image clean-deb-image

clean-bin-image:
	- docker rmi $(BIN_IMAGE)
	@ rm -f .bin-image

clean-deb-image:
	- docker rmi $(DEB_IMAGE)
	@ rm -f .deb-image
