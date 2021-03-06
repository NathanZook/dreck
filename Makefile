GITHUB:=miekg
NAME:=dreck

build:
	go build

test: deps
	go test -race -cover

.PHONY: deps
deps:
	@(cd $(GOPATH)/src/github.com/mholt/caddy 2>/dev/null  && git checkout -q master 2>/dev/null || true)
	@go get -u github.com/mholt/caddy
	@(cd $(GOPATH)/src/github.com/mholt/caddy              && git checkout -q v0.10.11)

.PHONY: release
release:
	mkdir -p release
	cp version.go release

.PHONY: upload
upload:
	@echo Releasing: $(VERSION)
	$(eval RELEASE:=$(shell curl -s -d '{"tag_name": "v$(VERSION)", "name": "v$(VERSION)"}' "https://api.github.com/repos/$(GITHUB)/$(NAME)/releases?access_token=${GITHUB_ACCESS_TOKEN}" | grep -m 1 '"id"' | tr -cd '[[:digit:]]'))
	@echo ReleaseID: $(RELEASE)
	@for asset in `ls -A release`; do \
	    curl -o /dev/null -X POST \
	      -H "Content-Type: application/gzip" \
	      --data-binary "@release/$$asset" \
	      "https://uploads.github.com/repos/$(GITHUB)/$(NAME)/releases/$(RELEASE)/assets?name=$${asset}&access_token=${GITHUB_ACCESS_TOKEN}" ; \
	done
