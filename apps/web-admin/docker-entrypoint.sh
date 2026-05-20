#!/bin/sh
if [ "$DEPLOY_MODE" = "docker" ]; then
  cp /etc/nginx/conf.d/default.conf.docker /etc/nginx/conf.d/default.conf
else
  cp /etc/nginx/conf.d/default.conf.k8s /etc/nginx/conf.d/default.conf
fi
exec nginx -g "daemon off;"
