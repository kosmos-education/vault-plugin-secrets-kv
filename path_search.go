// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kv

import (
	"context"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
	"net/http"
	"strings"
)

// pathSearch returns the path configuration for the secret/path search endpoint
func pathSearch(b *versionedKVBackend) *framework.Path {
	return &framework.Path{
		Pattern: "search/" + framework.MatchAllRegex("keyword"),
		Fields: map[string]*framework.FieldSchema{
			"keyword": {
				Type:        framework.TypeString,
				Description: "Path or secret name",
			},
		},
		Operations: map[logical.Operation]framework.OperationHandler{
			logical.ListOperation: &framework.PathOperation{
				Callback: b.upgradeCheck(b.pathSearchListRecursive()),
				DisplayAttrs: &framework.DisplayAttributes{
					OperationVerb: "list",
				},
				Responses: map[int][]framework.Response{
					http.StatusOK: {{
						Description: http.StatusText(http.StatusOK),
						Fields: map[string]*framework.FieldSchema{
							"keys": {
								Type:     framework.TypeSlice,
								Required: true,
							},
						},
					}},
				},
			},
		},
		HelpSynopsis:    searchHelpSyn,
		HelpDescription: searchHelpDesc,
	}
}

// pathSearchListRecursive calls the recursive lookup function and returns a logical response
func (b *versionedKVBackend) pathSearchListRecursive() framework.OperationFunc {
	return func(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
		keyword := strings.Trim(data.Get("keyword").(string), "/")

		// Get an encrypted key storage object
		wrapper, err := b.getKeyEncryptor(ctx, req.Storage)
		if err != nil {
			return nil, err
		}

		es := wrapper.Wrap(req.Storage)

		// Use encrypted key storage to list the keys
		keys, err := es.List(ctx, "/")
		if err != nil {
			return logical.ListResponse(keys), err
		}

		keys, err = recursiveLookup(ctx, "/", keyword, keys, es)

		if err != nil {
			return logical.ListResponse(keys), err
		}
		return logical.ListResponse(keys), nil
	}
}

// recursiveLookup is a recursive function that looks up all subfolders from the given paths.
// It returns a list of all paths and keys that match the specified keyword.
func recursiveLookup(ctx context.Context, path string, secretName string, keys []string, es logical.Storage) ([]string, error) {
	var lookedUpKeys []string

	for _, key := range keys {
		fullPath := path + key

		if strings.Contains(strings.Trim(key, "/"), secretName) {
			lookedUpKeys = append(lookedUpKeys, fullPath)
		}
		// Check if the current key is a subfolder
		if strings.HasSuffix(key, "/") {
			// Load keys from subfolder recursively
			subFolderKeys, err := es.List(ctx, fullPath)
			if err != nil {
				return lookedUpKeys, err
			}

			// Check if returned keys contains a subfolder and search keys recursively
			recursiveLookupKeys, err := recursiveLookup(ctx, fullPath, secretName, subFolderKeys, es)
			if err != nil {
				return lookedUpKeys, err
			}
			lookedUpKeys = append(lookedUpKeys, recursiveLookupKeys...)
		}
	}
	return lookedUpKeys, nil
}

const searchHelpSyn = `List all paths and keys that match the given keyword`
const searchHelpDesc = `
This endpoint allows for searching paths and keys in a specific key-value store backend
store
`
