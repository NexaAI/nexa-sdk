### Option 1: use ngrok reverse proxy

```sh
make build && ./build/nexa serve --ngrok
```

The server will start and display the ngrok URL, e.g.:
API documentation available at url=https://xxxxx.ngrok-free.app/docs/ui

### Option 2: set up a subdomain on nexasdk.com

1. start from project root
2. Set CF_Token and CF_Zone_ID in shell variable
3. `ssl/httpsserve.sh <subdomain>`, e.g. `ssl/httpsserve.sh leo`
