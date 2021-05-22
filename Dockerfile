FROM golang:1.16.4-alpine3.13 as dev

RUN apk add --no-cache gcc libc-dev

WORKDIR /workspace

FROM dev as build

ADD . /workspace/
RUN CGO_ENABLED=0 go build -o api.cgi -ldflags "-X main.version=${VERSION:-dev} -X main.commit=${COMMIT:-none}" ./cmd/cgi

FROM httpd:2.4.46-alpine as server

COPY --from=build /workspace/random.cgi /usr/local/apache2/htdocs/random.cgi
RUN sh -c 'echo "<Directory \"/usr/local/apache2/htdocs\">" >>/usr/local/apache2/conf/httpd.conf' \
  && sh -c 'echo "Options +ExecCGI" >>/usr/local/apache2/conf/httpd.conf' \
  && sh -c 'echo "</Directory>" >>/usr/local/apache2/conf/httpd.conf' \
  && sh -c 'echo "LoadModule cgid_module modules/mod_cgid.so" >>/usr/local/apache2/conf/httpd.conf' \
  && sh -c 'echo "AddHandler cgi-script .cgi" >>/usr/local/apache2/conf/httpd.conf'
