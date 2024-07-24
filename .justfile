build:
	bear -- g++ -std=c++23 -Wall -Wextra -Wconversion \
	-Wsign-conversion -ggdb write.cpp \
	-o write -Iextern/plog/include
run:
	./write

stream:
	bear -- g++ -std=c++23 -Wall -Wextra -Wconversion \
	-Wsign-conversion -ggdb stream.cpp \
	-o stream -Iextern/plog/include  \
	-lgrpc++ -lprotoc -lprotobuf `pkg-config --libs grpc` -lgrpc \
	-Iextern/include
