# libdns-namecheap
Namecheap provider implementation for libdns

**NOTE** This module is being implemented for my personal usage. It should be sufficient to integrate with Caddy2 to use Let's Encrypt DNS challenge on Namecheap.

Namecheap's API is very crude and dangerous. Please don't use this in production for any critical systems.

## Testing

Create an account in https://www.sandbox.namecheap.com/ and enable the API (https://www.namecheap.com/support/api/intro/ under `Enabling API Access`).

Export 2 environment variables: `NAMECHEAP_API_KEY` and `NAMECHEAP_API_USER` with the data from the previous step.

In that account, reserve the following domains:

- `gethosts-0.com`
- `sethosts.com`

Finally, run the `test` make task:

```shell
$ make test
```
