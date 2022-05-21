all:
	rm -rf build
	mkdir build
	./build.sh
	@echo "Build finished"

clean:
	rm -rf build
