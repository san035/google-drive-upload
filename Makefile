# Create new version tag in version.txt and push to repo
.PHONY: create-new-version-tag
create-new-version-tag:
	@set -x && cp version.txt version.old && perl -pe 's/(\d+)(?=[^.\d]*$$)/$$1+1/e' version.old > version.txt
	git add version.txt && git commit -m "New version"
	@set -x && VERSION=$$(cat version.txt | tr -d '\n\r') && && git tag -a "$$VERSION" -m "Release version $$VERSION" && git push origin main $$VERSION
	@rm version.old
	@echo "Tag $$(cat version.txt | tr -d '\n\r') created and pushed successfully!"

