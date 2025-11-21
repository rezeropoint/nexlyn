package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf

	PostgreSQL struct {
		Host         string
		Port         int
		Database     string
		Username     string
		Password     string
		MaxConns     int `json:",default=10"`
		MaxIdleConns int `json:",default=5"`
		MaxLifetime  int `json:",default=3600"` // 秒
	}

	ClickHouse struct {
		Host         string
		Port         int
		Database     string
		Username     string
		Password     string
		MaxOpenConns int `json:",default=10"`
		MaxIdleConns int `json:",default=5"`
		MaxLifetime  int `json:",default=3600"` // 秒
		ConnTimeout  int `json:",default=10"`   // 秒
		ReadTimeout  int `json:",default=30"`   // 秒
		WriteTimeout int `json:",default=30"`   // 秒
	}
}
