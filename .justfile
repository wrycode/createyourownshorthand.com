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
	-Lextern/google-cloud-cpp/lib \
	-Iextern/google-cloud-cpp/include \
	-lgrpc++ -lprotoc -lprotobuf `pkg-config --libs grpc protobuf` -lgrpc \
	-lcrc32c -I /usr/include/absl/


	# -Wsign-conversion \
# -Wall -Wextra -Wconversion \
