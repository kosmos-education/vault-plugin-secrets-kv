// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kv

import (
	"context"
	"testing"

	"github.com/go-test/deep"
	"github.com/hashicorp/vault/sdk/helper/testhelpers/schema"
	"github.com/hashicorp/vault/sdk/logical"
)

// TestVersionedKV_Search_NotFound verifies that the search endpoint returns nil keys
// when the path/secret does not exist
func TestVersionedKV_Search_NotFound(t *testing.T) {
	b, storage := getBackend(t)

	data := map[string]interface{}{
		"data": map[string]interface{}{
			"foo": "does-not-matter",
		},
	}

	req := &logical.Request{
		Operation: logical.CreateOperation,
		Path:      "data/a/b/foo",
		Storage:   storage,
		Data:      data,
	}

	resp, err := b.HandleRequest(context.Background(), req)
	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("CreateOperation request failed, err: %v, resp %#v", err, resp)
	}

	req = &logical.Request{
		Operation: logical.ListOperation,
		Path:      "search/fooa",
		Storage:   storage,
	}

	resp, err = b.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected ListOperation response, err: %v, resp %#v", err, resp)
	}

	if diff := deep.Equal(resp.Data["keys"], nil); len(diff) > 0 {
		t.Fatal(diff)
	}

}

// TestVersionedKV_Search_FoundSecret verifies that the search endpoint returns
// the path of the secret that match the keyword
func TestVersionedKV_Search_FoundSecret(t *testing.T) {
	b, storage := getBackend(t)

	data := map[string]interface{}{
		"data": map[string]interface{}{
			"foo": "does-not-matter",
		},
	}

	req := &logical.Request{
		Operation: logical.CreateOperation,
		Path:      "data/a/b/foo",
		Storage:   storage,
		Data:      data,
	}

	resp, err := b.HandleRequest(context.Background(), req)
	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("CreateOperation request failed, err: %v, resp %#v", err, resp)
	}

	req = &logical.Request{
		Operation: logical.ListOperation,
		Path:      "search/foo",
		Storage:   storage,
	}

	resp, err = b.HandleRequest(context.Background(), req)
	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("err:%s resp:%#v\n", err, resp)
	}
	schema.ValidateResponse(
		t,
		schema.GetResponseSchema(t, b.(*versionedKVBackend).Route(req.Path), req.Operation),
		resp,
		true,
	)

	if diff := deep.Equal(resp.Data["keys"], []string{"/a/b/foo"}); len(diff) > 0 {
		t.Fatal(diff)
	}
}

// TestVersionedKV_Search_FoundSecret verifies that the search endpoint returns
// the path that match the keyword
func TestVersionedKV_Search_FoundPath(t *testing.T) {
	b, storage := getBackend(t)

	data := map[string]interface{}{
		"data": map[string]interface{}{
			"foo": "does-not-matter",
		},
	}

	req := &logical.Request{
		Operation: logical.CreateOperation,
		Path:      "data/a/b/foo",
		Storage:   storage,
		Data:      data,
	}

	resp, err := b.HandleRequest(context.Background(), req)
	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("CreateOperation request failed, err: %v, resp %#v", err, resp)
	}

	req = &logical.Request{
		Operation: logical.ListOperation,
		Path:      "search/b",
		Storage:   storage,
	}

	resp, err = b.HandleRequest(context.Background(), req)
	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("err:%s resp:%#v\n", err, resp)
	}
	schema.ValidateResponse(
		t,
		schema.GetResponseSchema(t, b.(*versionedKVBackend).Route(req.Path), req.Operation),
		resp,
		true,
	)

	if diff := deep.Equal(resp.Data["keys"], []string{"/a/b/"}); len(diff) > 0 {
		t.Fatal(diff)
	}
}

// TestVersionedKV_Search_Deleted verifies that the search endpoint returns
// nil keys when the secret is deleted
func TestVersionedKV_Search_Deleted(t *testing.T) {
	b, storage := getBackend(t)

	req := &logical.Request{
		Operation: logical.CreateOperation,
		Path:      "data/a/b/foo",
		Storage:   storage,
		Data: map[string]interface{}{
			"data": map[string]interface{}{
				"foo": "bar",
			},
		},
	}

	resp, err := b.HandleRequest(context.Background(), req)
	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("data CreateOperation request failed, err: %v, resp %#v", err, resp)
	}

	req = &logical.Request{
		Operation: logical.ListOperation,
		Path:      "search/foo",
		Storage:   storage,
	}

	resp, err = b.HandleRequest(context.Background(), req)
	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("search ListOperation request failed, err: %v, resp %#v", err, resp)
	}

	if diff := deep.Equal(resp.Data["keys"], []string{"/a/b/foo"}); len(diff) > 0 {
		t.Fatal(diff)
	}

	req = &logical.Request{
		Operation: logical.DeleteOperation,
		Path:      "metadata/a/b/foo",
		Storage:   storage,
	}

	resp, err = b.HandleRequest(context.Background(), req)
	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("DeleteOperation request failed - err: %v resp:%#v\n", err, resp)
	}

	req = &logical.Request{
		Operation: logical.ListOperation,
		Path:      "search/foo",
		Storage:   storage,
	}

	resp, err = b.HandleRequest(context.Background(), req)

	if diff := deep.Equal(resp.Data["keys"], nil); len(diff) > 0 {
		t.Fatal(diff)
	}
}
