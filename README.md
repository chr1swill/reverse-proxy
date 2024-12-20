# Reverse Proxy

## Motivation:

Maybe I am just bad at a little stupid, but for the life of me I could not find out how to use NGINX to make a HTTPS reverse proxy that sits in front of multiple HTTP application under different domains with a single Let Encypt certificate. SSL should be set up for the reverse proxy only for convenience. So as any good man would, I am making my own. And so far it turning out to be pretty straight forward.

Update: I figured out how to build what is was looking for all along, but I like this much better.

## Benefits 

- Only depends on the go standard libary
- Dead simple, no configuration files, just a single command cli

## Build

```bash
./build.sh
```

## Ussage

```bash
./reverse-proxy
```
