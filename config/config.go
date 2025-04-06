package config

type Shard struct {
	Name string
	Idx  int
}

type Config struct {
	Shards []Shard
}
