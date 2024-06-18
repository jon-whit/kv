# kv
A simple gPRC persistent Key/Value store.

## Get Started
1. Start the KV service.
```
./kv serve
```
This will start the KV service by default on `:50052`.

2. Write a key/value pair.
```console
grpcurl --plaintext -use-reflection -d '{"key": "foo", "value": {"string_value": "bar"}}' localhost:50052 kv.v1alpha1.KVService/Set
```
> ℹ️ The `value` field of Set is defined as a [google.protobuf.Value](https://protobuf.dev/reference/protobuf/google.protobuf/#value) type. This allows developers to set dynamic values according to well-known protobuf types.

1. Get a key/value pair.
```console
grpcurl --plaintext -use-reflection -d '{"key": "foo"}' localhost:50052 kv.v1alpha1.KVService/Get
```

1. Delete a key.
```console
grpcurl --plaintext -use-reflection -d '{"key": "foo"}' localhost:50052 kv.v1alpha1.KVService/Delete
```

## FAQ
1. Can I run the server on a different port than the default `:50052`?
Yes, to run the server on a different port you can provide the `--addr` flag. For example, `./kv --addr ":8080"`

2. Is there a per-request size limit for the keys and/or values being stored?

Yes, To protect the server, the maximum request message size is limited to `1.5MiB`. If the total size of a request is greater than this value then you can expect an error to be returned from the server.

## Developers
### Build
```
make build
```

### Lint
> ℹ️ You need to install `golangci-lint` for this target to work.
> See https://golangci-lint.run/welcome/install/
```
make lint
```

### Test
```
make test
```