package config_test

import (
	"go-kvdb/config"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

func createConfig(t *testing.T, contents string) config.Config {
	t.Helper()

	f, err := ioutil.TempFile(os.TempDir(), "config.toml")
	if err != nil {
		t.Fatalf("Couldn't create a temp file: %v", err)
	}
	defer f.Close()

	name := f.Name()
	defer os.Remove(name)

	_, err = f.WriteString(contents)
	if err != nil {
		t.Fatalf("Could not write the config contents: %v", err)
	}

	c, err := config.ParseFile(name)
	if err != nil {
		t.Fatalf("Could not parse config: %v", err)
	}

	return c
}

func TestParseConfig(t *testing.T) {
	got := createConfig(t, `[[shards]]
		name = "shard1"
		idx = 0
		address = "localhost:8080"`)

	want := config.Config{
		Shards: []config.Shard{
			{
				Name:    "shard1",
				Idx:     0,
				Address: "localhost:8080",
			},
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("The config does match: got: %#v, want: %#v", got, want)
	}
}

func TestParseShards(t *testing.T) {
	c := createConfig(t, `
	[[shards]]
		name = "shard1"
		idx = 0
		address = "localhost:8080"
	[[shards]]
		name = "shard2"
		idx = 1
		address = "localhost:8081"`)

	got, err := config.ParseShards(c.Shards, "shard2")
	if err != nil {
		t.Fatalf("Could not parse shards %#v: %v", c.Shards, err)
	}

	want := &config.Shards{
		Count:  2,
		CurIdx: 1,
		Addrs: map[int]string{
			0: "localhost:8080",
			1: "localhost:8081",
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("The shards config does match: got: %#v, want: %#v", got, want)
	}
}
