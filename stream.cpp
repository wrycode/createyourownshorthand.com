#include <iostream>
#include <memory>
#include <string>

#include <grpcpp/grpcpp.h>
#include "google/cloud/speech/v1/cloud_speech.grpc.pb.h"

#include <cstdlib> // for getenv
#include <string>


using google::cloud::speech::v1::Speech;
using google::cloud::speech::v1::RecognizeRequest;
using google::cloud::speech::v1::RecognizeResponse;
using grpc::Status;

std::string getGoogleAppCredentialsJson() {
  const char* googleAppCredentialJson = std::getenv("YOUR_GOOGLE_APPLICATION_CREDENTIALS_JSON");
  if (googleAppCredentialJson) {
    return std::string(googleAppCredentialJson);
  } else {
    return std::string("~/.config/gcloud/application_default_credentials.json");
  }
}

class SpeechClient {
 public:
  SpeechClient(std::shared_ptr<grpc::Channel> channel)
      : stub_(Speech::NewStub(channel)) {}

  // Assembles the client's payload and sends it to the server.
  void Recognize() {
    RecognizeRequest request;

    // Fill the request object with the relevant data, e.g.:
    // * Config parameters (encoding, sample rate, language, etc.)
    // * Audio data

    RecognizeResponse response;
    grpc::ClientContext context;

    Status status = stub_->Recognize(&context, request, &response);

    if (status.ok()) {
      // Handle the RecognizeResponse object.
      std::cout << "Recognized speech successfully.\n";
    } else {
      std::cout << status.error_code() << ": " << status.error_message() << std::endl;
    }
  }

 private:
  std::unique_ptr<Speech::Stub> stub_;
};

int main() {
  auto json_key = grpc::ServiceAccountJWTAccessCredentials(
      getGoogleAppCredentialsJson(), std::chrono::hours(1).count() * 3600);

  std::shared_ptr<grpc::ChannelCredentials> creds = grpc::SslCredentials(grpc::SslCredentialsOptions());
  creds = grpc::CompositeChannelCredentials(creds, json_key);

  auto channel = grpc::CreateChannel("speech.googleapis.com", creds);

  SpeechClient client(channel);
  client.Recognize();

  return 0;
}
