.PHONY: lint

build:
	mkdir -p dist/
package: build
	nfpm pkg --packager deb --target dist/
lint:
	python -m pylint joy2mouse.py
clean:
	rm -f dist/*
