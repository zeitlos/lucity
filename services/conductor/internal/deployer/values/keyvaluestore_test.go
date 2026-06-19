package values

import (
	"testing"

	"k8s.io/apimachinery/pkg/api/resource"
)

func TestCreateKeyValueStore(t *testing.T) {
	env := New()

	spec := KeyValueStoreSpec{
		Version:  "8",
		Size:     resource.MustParse("1Gi"),
		Password: "s3cret",
	}

	if err := CreateKeyValueStore(env, "cache", spec); err != nil {
		t.Fatalf("CreateKeyValueStore: %v", err)
	}

	vk, ok := env.Databases.Valkey["cache"]

	if !ok {
		t.Fatal("key-value store entry not created")
	}

	if vk.Version != "8" {
		t.Errorf("version = %q, want 8", vk.Version)
	}

	if vk.Size != "1Gi" {
		t.Errorf("size = %q, want 1Gi", vk.Size)
	}

	if vk.Password != "s3cret" {
		t.Errorf("password = %q, want s3cret", vk.Password)
	}

	if vk.Labels[labelKeyValueStore] != "cache" {
		t.Errorf("discovery label = %q, want cache", vk.Labels[labelKeyValueStore])
	}
}

func TestCreateKeyValueStoreIdempotent(t *testing.T) {
	env := New()

	if err := CreateKeyValueStore(env, "cache", KeyValueStoreSpec{Password: "first"}); err != nil {
		t.Fatalf("CreateKeyValueStore: %v", err)
	}

	// A second create (e.g. a reconcile re-apply) must not rotate the password.
	if err := CreateKeyValueStore(env, "cache", KeyValueStoreSpec{Password: "second"}); err != nil {
		t.Fatalf("CreateKeyValueStore (idempotent): %v", err)
	}

	if got := env.Databases.Valkey["cache"].Password; got != "first" {
		t.Errorf("password rotated to %q on re-create; want stable %q", got, "first")
	}
}

func TestCreateKeyValueStoreInvalidName(t *testing.T) {
	env := New()

	if err := CreateKeyValueStore(env, "Invalid_Name", KeyValueStoreSpec{}); err == nil {
		t.Fatal("expected error for invalid key-value store name")
	}
}

func TestKeyValueStoreRoundTripPreservesPassword(t *testing.T) {
	env := New()

	if err := CreateKeyValueStore(env, "cache", KeyValueStoreSpec{
		Version:  "8",
		Size:     resource.MustParse("2Gi"),
		Password: "keepme",
	}); err != nil {
		t.Fatalf("CreateKeyValueStore: %v", err)
	}

	data, err := Marshal(env)

	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	// Reconcile loads the current release values, mutates, and re-applies.
	// The generated password must survive that round-trip.
	reloaded, err := Parse(data)

	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	vk, ok := reloaded.Databases.Valkey["cache"]

	if !ok {
		t.Fatal("key-value store entry lost on round-trip")
	}

	if vk.Password != "keepme" {
		t.Errorf("password = %q after round-trip, want keepme", vk.Password)
	}

	if vk.Size != "2Gi" {
		t.Errorf("size = %q after round-trip, want 2Gi", vk.Size)
	}
}

func TestDeleteKeyValueStore(t *testing.T) {
	env := New()

	if err := CreateKeyValueStore(env, "cache", KeyValueStoreSpec{Password: "x"}); err != nil {
		t.Fatalf("CreateKeyValueStore: %v", err)
	}

	if err := DeleteKeyValueStore(env, "cache"); err != nil {
		t.Fatalf("DeleteKeyValueStore: %v", err)
	}

	if _, ok := env.Databases.Valkey["cache"]; ok {
		t.Fatal("key-value store entry not deleted")
	}

	if err := DeleteKeyValueStore(env, "cache"); err == nil {
		t.Fatal("expected error deleting absent key-value store")
	}
}
