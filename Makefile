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
	git tag v0.0.1
	git push --tags
	@echo "\n tags 发布中..."
