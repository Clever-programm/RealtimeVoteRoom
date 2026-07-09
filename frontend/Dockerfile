FROM ubuntu:22.04 AS builder

RUN apt-get update && apt-get install -y curl git unzip xz-utils zip libglu1-mesa python3
WORKDIR /app

RUN git clone https://github.com/flutter/flutter.git -b stable /flutter
ENV PATH="/flutter/bin:/flutter/bin/cache/dart-sdk/bin:${PATH}"

RUN flutter doctor
RUN flutter config --enable-web

COPY . .
RUN flutter pub get
RUN flutter build web --release

FROM nginx:alpine
COPY --from=builder /app/build/web /usr/share/nginx/html
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]