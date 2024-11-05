.PHONY: up tag

up:
	git pull origin main
	git add .
	git commit -am "update"
	git push origin main
	@echo "\n 发布中..."

tag:
	#git pull origin main
	git add .
	git commit -am "init"
	git push origin main
	git tag v0.0.7
	git push --tags
	@echo "\n tags 发布中..."


proto:
	@echo ===================================== Compiling Proto Files ============================================
	@find packet -name "*.proto" -type f | while read f; do \
		echo "Compiling: $$f"; \
		protoc --proto_path=. \
			   --go_out=. \
   			   --go_opt=paths=source_relative \
			   "$$f" || exit 1; \
		echo "Compiled: $$f"; \
	done
	@echo "All proto files compiled successfully!"
