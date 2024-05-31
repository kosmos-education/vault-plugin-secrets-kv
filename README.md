# Vault Plugin: Key-Value Secrets Backend with secret search feature

This is an evolution of the original [KV backend plugin](https://github.com/hashicorp/vault-plugin-secrets-kv) for use with [Hashicorp Vault](https://www.github.com/hashicorp/vault).
This implementation provides recursive search for all Key-Value secrets backend by adding the /search API endpoint.


**Please note carefully**: this is a highly experimental project and should be only used as a preview feature.
I do not recommend any use in production, as the current behavior might be I/O intensive for the server in certain cases, and could lead to performance degradation. 


## Quick Links
    - KV original Docs: https://github.com/hashicorp/vault-plugin-secrets-kv/blob/main/README.md

## API reference

```http
  GET /${backend}/search/${secret_name}
```

| Parameter     | Type     | Description                           |
|:--------------| :------- |:--------------------------------------|
| `backend`     | `string` | Backend name (KV v2 type)             |
| `secret_name` | `string` | Name of the secret we are looking for |

### Example
```http
  GET /secret/search/demo
```


Response body:
```json
{
"request_id": "49ed8e03-0748-72a3-d645-413f036fa367",
"lease_id": "",
"renewable": false,
"lease_duration": 0,
"data": {
    "keys": [
        "/path1/demo",
        "/path1/path12/demo",
        "/path2/path21/demo"
    ]
},
"wrap_info": null,
"warnings": null,
"auth": null
}
```

## Use with Vault

If you wish to work on this plugin, you'll first need
[Go](https://www.golang.org) installed on your machine
(version 1.10+ is *required*).

For local dev first make sure Go is properly installed, including
setting up a [GOPATH](https://golang.org/doc/code.html#GOPATH).
Next, clone this repository into
`$GOPATH/src/github.com/hashicorp/vault-plugin-secrets-kv`.
You can then download any required build tools by bootstrapping your
environment:

```sh
$ make bootstrap
```

To compile a development version of this plugin, run `make` or `make dev`.
This will put the plugin binary in the `bin` and `$GOPATH/bin` folders. `dev`
mode will only generate the binary for your platform and is faster:

```sh
$ make
$ make dev
```

Once you've done that, there are two approaches to testing your new plugin version
in Vault. You can add a temporary `replace` declaration in your local Vault checkout's
go.mod (above the `require` declarations), such as:

```
replace github.com/hashicorp/vault-plugin-secrets-kv => /path/to/your/project/vault-plugin-secrets-kv
```

Alternatively, you could go through the plugin process. To do this,
put the plugin binary into a location of your choice. This directory
will be specified as the [`plugin_directory`](https://www.vaultproject.io/docs/configuration/index.html#plugin_directory)
in the Vault config used to start the server.

```json
...
plugin_directory = "path/to/plugin/directory"
...
```

Start a Vault server with this config file:
```sh
$ vault server -config=path/to/config.json ...
...
```

Once the server is started, register the plugin in the Vault server's [plugin catalog](https://developer.hashicorp.com/vault/docs/plugins/plugin-architecture#plugin-catalog):

```sh
$ vault plugin register \
        -sha256=<expected SHA256 Hex value of the plugin binary> \
        -command="vault-plugin-secrets-kv" \
        secret \
        kv
```

Note you should generate a new sha256 checksum if you have made changes
to the plugin. Example using openssl:

```sh
openssl dgst -sha256 $GOPATH/vault-plugin-secrets-kv
...
SHA256(.../go/bin/vault-plugin-secrets-kv)= 896c13c0f5305daed381952a128322e02bc28a57d0c862a78cbc2ea66e8c6fa1
```

Enable the auth plugin backend using the secrets enable plugin command:

```sh
$ vault secrets enable -plugin-name='kv' plugin
...

Successfully enabled 'plugin' at 'kv'!
```