build:
	bear -- g++ -std=c++23 -Wall -Wextra -Wconversion \
	-Wsign-conversion -ggdb write.cpp \
	-o write -Iextern/plog/include
run:
	./write

stream:
	bear -- \
	g++ -std=c++23 -ggdb \
	-Lextern/google-cloud-cpp/lib \
	-Iextern/google-cloud-cpp/include \
	-I /usr/include/absl/ \
	stream.cpp \
	`pkg-config --libs grpc protobuf` \
	-l:libgoogle_cloud_cpp_speech.a -l:libgoogle_cloud_cpp_speech_protos.a \
	-lgrpc++ -lprotoc -lprotobuf -lgrpc -lcrc32c \
	-o stream

	# -lgoogle_cloud_cpp_speech_protos -lgoogle_cloud_cpp_speech \
# g++ -std=c++23 \
# 	-ggdb stream.cpp -o stream \
# 	-Lextern/google-cloud-cpp/lib \
# 	-Iextern/google-cloud-cpp/include \
# 	-lgrpc++ -lprotoc -lprotobuf `pkg-config --libs grpc protobuf` -lgrpc \
# 	-lcrc32c -I /usr/include/absl/


# -Wsign-conversion \
# -Wall -Wextra -Wconversion \
