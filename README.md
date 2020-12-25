# libdns-namecheap
Namecheap provider implementation for libdns

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
