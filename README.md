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