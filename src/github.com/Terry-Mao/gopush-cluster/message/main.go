// Copyright © 2014 Terry Mao, LiuDing All rights reserved.
// This file is part of gopush-cluster.

// gopush-cluster is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// gopush-cluster is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with gopush-cluster.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	log "github.com/alecthomas/log4go"
	"flag"
	"github.com/Terry-Mao/gopush-cluster/perf"
	"github.com/Terry-Mao/gopush-cluster/process"
	"github.com/Terry-Mao/gopush-cluster/ver"
	"runtime"
)


/**
  Message系统的结构:
  main.go 整体集成
  config.go 配置文件定义
  rpc.go 对外接口的定义
  signal.go 似乎每个模块都有重复定义
  zk.go 也是重复定义，只不过引用了本package内部的模块
  storage.go 做一个storage的工厂，定义了接口和FactoryMethod
   redis.go
   mysql.go 实现了具体的storage
 */

func main() {
	flag.Parse()
	log.Info("message ver: \"%s\" start", ver.Version)
	if err := InitConfig(); err != nil {
		panic(err)
	}
	// Set max routine
	runtime.GOMAXPROCS(Conf.MaxProc)
	// init log
	log.LoadConfiguration(Conf.Log)
	defer log.Close()

	// start pprof http
	// 性能pref
	perf.Init(Conf.PprofBind)


	// Initialize redis
	// 本地额存储
	if err := InitStorage(); err != nil {
		panic(err)
	}

	// init rpc service
	// 初始化RPC服务器，对外的API
	if err := InitRPC(); err != nil {
		panic(err)
	}

	// init zookeeper
	zk, err := InitZK()
	if err != nil {
		if zk != nil {
			zk.Close()
		}
		panic(err)
	}

	// process init
	if err = process.Init(Conf.User, Conf.Dir, Conf.PidFile); err != nil {
		panic(err)
	}

	// 运维相关的管理
	// init signals, block wait signals
	sig := InitSignal()
	HandleSignal(sig)
	// exit
	log.Info("message stop")
}
