build:
	bear -- g++ -std=c++23 -Wall -Wextra -Wconversion \
	-Wsign-conversion -ggdb write.cpp \
	-o write -Iextern/plog/include
run:
	./write

stream:
	bear -- g++ -std=c++23 \
	-ggdb stream.cpp \
	-o stream -Iextern/plog/include  \
	-Iextern/google-cloud-cpp/include \
	-Lextern/google-cloud-cpp/lib \
	-lgrpc++ -lprotoc -lprotobuf `pkg-config --libs grpc protobuf` -lgrpc \
	-lcrc32c 


	# -Wsign-conversion \
# -Wall -Wextra -Wconversion \
