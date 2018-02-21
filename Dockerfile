FROM golang:1.9.4-alpine3.7 AS runtime

ENV APP_NAME=keep-client \
	BIN_PATH=/usr/local/bin \
	LIB_DIR=/usr/local/lib/ \
	INCLUDE_DIR=/usr/local/include/

RUN apk add --update --no-cache \
	gmp \
	gmp-dev \
	libgmpxx  \
	libstdc++ \
	openssl \
	openssl-dev && \
	rm -rf /var/cache/apk && mkdir /var/cache/apk && \
	rm -rf /usr/share/man

FROM runtime AS cbuild

ENV BN_VERSION=d1a44d2f242692601b3e150b59044ab82f265b65

RUN apk add --update --no-cache \
	clang \
	g++ \
	git \
	llvm \
	make && \
	rm -rf /var/cache/apk && mkdir /var/cache/apk && \
	rm -rf /usr/share/man

RUN git clone https://github.com/dfinity/bn /bn && \
	cd /bn && \
	git reset --hard $BN_VERSION && \
	make install && make && \
	rm -rf /bn

FROM runtime AS gobuild

ENV GOBIN=/go/bin \
	APP_DIR=/go/src/github.com/keep-network/keep-client

RUN apk add --update --no-cache \
	g++ \
	git && \
	rm -rf /var/cache/apk/ && mkdir /var/cache/apk/ && \
	rm -rf /usr/share/man

RUN mkdir -p $APP_DIR

COPY . $APP_DIR
WORKDIR $APP_DIR

RUN go get -u github.com/golang/dep/cmd/dep
COPY ./go/Gopkg.toml ./go/Gopkg.lock ./
RUN dep ensure --vendor-only

COPY --from=cbuild $LIB_DIR $LIB_DIR
COPY --from=cbuild $INCLUDE_DIR $INCLUDE_DIR

RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o $APP_NAME ./go && \
	mv $APP_NAME $BIN_PATH && \
	rm -rf $APP_DIR

FROM runtime

COPY --from=gobuild $BIN_PATH/$APP_NAME $BIN_PATH
COPY --from=cbuild $LIB_DIR $LIB_DIR
COPY --from=cbuild $INCLUDE_DIR $INCLUDE_DIR

# ENTRYPOINT cant handle ENV variables.
ENTRYPOINT ["keep-client", "-config",  "/keepclient/config.toml"]

# docker caches more when using CMD [] resulting in a faster build.
CMD []
