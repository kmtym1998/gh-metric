spec-workflow-dashboard:
	npx -y @pimzino/spec-workflow-mcp@latest $(PWD) --dashboard --port 5001

.PHONY: build
build:
	go build -o gh-metric .
	chmod +x gh-metric

.PHONY: install
install: build
	gh extension install .

.PHONY: reinstall
reinstall: build
	gh extension remove metric || true
	gh extension install .

