.PHONY: docs
docs:
	hugo --source docs/
docs-preview:
	hugo serve --source docs/
