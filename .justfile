build:
	bear -- g++ -std=c++23 -Wall -Wextra -Wconversion \
	-Wsign-conversion -ggdb write.cpp \
	-o write -Iextern/plog/include
run:
	./write

	# bear -- g++ -std=c++23 -Wall -Weffc++ -Werror -Wextra -Wconversion \
	# -Wsign-conversion -pedantic-errors -ggdb write.cpp \
	# -o write -Iextern/plog/include
