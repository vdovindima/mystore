# This is my simple home file storage project.

Assumes work together with nginx, where nginx will distribute static files, and all other operations will be performed by my application.

## Running nginx
```sh
docker run -d \
    --name nginx \
    -p 80:80  \
    --network host \
    -v ./nginx.conf:/etc/nginx/conf.d/default.conf:ro \
    -v ./socket:/tmp/nginxsoc \ # path to linux socket file
    -v ./data:/var/www/html:rw \ # file storage folder
    nginx:1.23-alpine-slim
```

## Nginx conf

```conf
map $request_method $method_location {
   GET     @get;
   default @all;
}

upstream myapp {
  server unix:/tmp/yoursocketpath;
}

  server {
    listen       80 default_server;
    server_name  _ "";
    root         /var/www/html;

    client_max_body_size 10G;

    location = / {
      return 404;
    }

    location ~ /([A-z0-9]*)/(.*) {
      try_files /dev/null $method_location;
    }

    location @get {
      autoindex on;
      autoindex_exact_size on;
      autoindex_format json;
      autoindex_localtime on;
    }

    location @all {
      proxy_pass http://myapp;
    }
  }
```
