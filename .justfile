build:
	bear -- g++ -std=c++23 -Wall -Wextra -Wconversion \
	-Wsign-conversion -ggdb write.cpp \
	-o write -Iextern/plog/include
run:
	./write

stream:
	mkdir -p build && cmake . -B build && make -C build
