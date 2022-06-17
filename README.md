# Slow Proxy

Toy project to test redirecting to a slow proxy. 

# Setup 

- Run the golang server, this is the slow proxy. 

```shell 
go run main.go <port>
```

- Start ngrok. This allows you to proxy to a local app. 

```shell
ngrok http <port>
```

- Change your netlify.toml to point to the ngrok url.

- Hit the endpoint:

hcurl cdn-glo-aws-sfo-11 https://cbosss-slow-proxy.netlify.app/proxy/slow/1m -X PATCH