module lotus_data_sync

go 1.17

replace github.com/filecoin-project/filecoin-ffi => ./extern/lotus/extern/filecoin-ffi

require (
	github.com/arthurkiller/rollingwriter v1.1.2
	github.com/astaxie/beego v1.12.3
	github.com/fatih/color v1.9.0
	github.com/filecoin-project/go-address v0.0.5
	github.com/filecoin-project/go-state-types v0.1.1-0.20210915140513-d354ccf10379
	github.com/filecoin-project/lotus v1.12.0
	github.com/filecoin-project/specs-actors v0.9.14
	github.com/globalsign/mgo v0.0.0-20181015135952-eeefdecb41b8
	github.com/go-redis/redis v6.15.9+incompatible
	github.com/hako/durafmt v0.0.0-20210608085754-5c1018a4e16b
	github.com/ipfs/go-block-format v0.0.3
	github.com/ipfs/go-cid v0.1.0
	github.com/ipfs/go-ipfs-blockstore v1.0.4
	github.com/ipfs/go-log v1.0.5
	github.com/libp2p/go-libp2p-core v0.8.6
	github.com/natefinch/lumberjack v2.0.0+incompatible
	github.com/oschwald/geoip2-golang v1.5.0
	github.com/shopspring/decimal v1.2.0
	github.com/whyrusleeping/cbor-gen v0.0.0-20210713220151-be142a5ae1a8
	go.uber.org/zap v1.16.0
	gopkg.in/mgo.v2 v2.0.0-20190816093944-a6b53ec6cb22
	xorm.io/xorm v1.2.5
)

require (
	contrib.go.opencensus.io/exporter/prometheus v0.1.0 // indirect
	github.com/BurntSushi/toml v0.3.1 // indirect
	github.com/GeertJohan/go.incremental v1.0.0 // indirect
	github.com/GeertJohan/go.rice v1.0.0 // indirect
	github.com/StackExchange/wmi v0.0.0-20190523213315-cbe66965904d // indirect
	github.com/akavel/rsrc v0.8.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/btcsuite/btcd v0.21.0-beta // indirect
	github.com/cespare/xxhash/v2 v2.1.1 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.0 // indirect
	github.com/crackcomm/go-gitignore v0.0.0-20170627025303-887ab5e44cc3 // indirect
	github.com/daaku/go.zipexe v1.0.0 // indirect
	github.com/detailyang/go-fallocate v0.0.0-20180908115635-432fa640bd2e // indirect
	github.com/elastic/go-sysinfo v1.3.0 // indirect
	github.com/elastic/go-windows v1.0.0 // indirect
	github.com/filecoin-project/filecoin-ffi v0.30.4-0.20200910194244-f640612a1a1f // indirect
	github.com/filecoin-project/go-amt-ipld/v2 v2.1.0 // indirect
	github.com/filecoin-project/go-amt-ipld/v3 v3.1.0 // indirect
	github.com/filecoin-project/go-bitfield v0.2.4 // indirect
	github.com/filecoin-project/go-cbor-util v0.0.0-20191219014500-08c40a1e63a2 // indirect
	github.com/filecoin-project/go-commp-utils v0.1.1-0.20210427191551-70bf140d31c7 // indirect
	github.com/filecoin-project/go-data-transfer v1.10.1 // indirect
	github.com/filecoin-project/go-fil-commcid v0.1.0 // indirect
	github.com/filecoin-project/go-fil-markets v1.12.0 // indirect
	github.com/filecoin-project/go-hamt-ipld v0.1.5 // indirect
	github.com/filecoin-project/go-hamt-ipld/v2 v2.0.0 // indirect
	github.com/filecoin-project/go-hamt-ipld/v3 v3.1.0 // indirect
	github.com/filecoin-project/go-jsonrpc v0.1.4-0.20210217175800-45ea43ac2bec // indirect
	github.com/filecoin-project/go-padreader v0.0.0-20210723183308-812a16dc01b1 // indirect
	github.com/filecoin-project/go-statemachine v1.0.1 // indirect
	github.com/filecoin-project/go-statestore v0.1.1 // indirect
	github.com/filecoin-project/specs-actors/v2 v2.3.5 // indirect
	github.com/filecoin-project/specs-actors/v3 v3.1.1 // indirect
	github.com/filecoin-project/specs-actors/v4 v4.0.1 // indirect
	github.com/filecoin-project/specs-actors/v5 v5.0.4 // indirect
	github.com/filecoin-project/specs-actors/v6 v6.0.0 // indirect
	github.com/filecoin-project/specs-storage v0.1.1-0.20201105051918-5188d9774506 // indirect
	github.com/gbrlsnchs/jwt/v3 v3.0.0-beta.1 // indirect
	github.com/go-ole/go-ole v1.2.4 // indirect
	github.com/go-stack/stack v1.8.0 // indirect
	github.com/goccy/go-json v0.7.4 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e // indirect
	github.com/golang/mock v1.6.0 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/golang/snappy v0.0.2-0.20190904063534-ff6b7dc882cf // indirect
	github.com/google/uuid v1.2.0 // indirect
	github.com/gorilla/mux v1.7.4 // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/hannahhoward/cbor-gen-for v0.0.0-20200817222906-ea96cece81f1 // indirect
	github.com/hannahhoward/go-pubsub v0.0.0-20200423002714-8d62886cc36e // indirect
	github.com/hashicorp/errwrap v1.0.0 // indirect
	github.com/hashicorp/go-multierror v1.1.0 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/ipfs/bbloom v0.0.4 // indirect
	github.com/ipfs/go-blockservice v0.1.5 // indirect
	github.com/ipfs/go-datastore v0.4.5 // indirect
	github.com/ipfs/go-graphsync v0.9.3 // indirect
	github.com/ipfs/go-ipfs-cmds v0.1.0 // indirect
	github.com/ipfs/go-ipfs-ds-help v1.0.0 // indirect
	github.com/ipfs/go-ipfs-exchange-interface v0.0.1 // indirect
	github.com/ipfs/go-ipfs-files v0.0.8 // indirect
	github.com/ipfs/go-ipfs-http-client v0.0.5 // indirect
	github.com/ipfs/go-ipfs-posinfo v0.0.1 // indirect
	github.com/ipfs/go-ipfs-util v0.0.2 // indirect
	github.com/ipfs/go-ipld-cbor v0.0.5 // indirect
	github.com/ipfs/go-ipld-format v0.2.0 // indirect
	github.com/ipfs/go-log/v2 v2.3.0 // indirect
	github.com/ipfs/go-merkledag v0.3.2 // indirect
	github.com/ipfs/go-metrics-interface v0.0.1 // indirect
	github.com/ipfs/go-path v0.0.7 // indirect
	github.com/ipfs/go-unixfs v0.2.6 // indirect
	github.com/ipfs/go-verifcid v0.0.1 // indirect
	github.com/ipfs/interface-go-ipfs-core v0.2.3 // indirect
	github.com/ipld/go-car v0.3.1-0.20210601190600-f512dac51e8e // indirect
	github.com/ipld/go-codec-dagpb v1.3.0 // indirect
	github.com/ipld/go-ipld-prime v0.12.0 // indirect
	github.com/jbenet/goprocess v0.1.4 // indirect
	github.com/jessevdk/go-flags v1.4.0 // indirect
	github.com/joeshaw/multierror v0.0.0-20140124173710-69b34d4ec901 // indirect
	github.com/jpillora/backoff v1.0.0 // indirect
	github.com/json-iterator/go v1.1.11 // indirect
	github.com/kelseyhightower/envconfig v1.4.0 // indirect
	github.com/klauspost/compress v1.13.6 // indirect
	github.com/klauspost/cpuid/v2 v2.0.8 // indirect
	github.com/libp2p/go-buffer-pool v0.0.2 // indirect
	github.com/libp2p/go-flow-metrics v0.0.3 // indirect
	github.com/libp2p/go-libp2p-discovery v0.5.1 // indirect
	github.com/libp2p/go-libp2p-peerstore v0.2.8 // indirect
	github.com/libp2p/go-libp2p-pubsub v0.5.4 // indirect
	github.com/libp2p/go-msgio v0.0.6 // indirect
	github.com/libp2p/go-openssl v0.0.7 // indirect
	github.com/mattn/go-colorable v0.1.8 // indirect
	github.com/mattn/go-isatty v0.0.13 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.1 // indirect
	github.com/miekg/dns v1.1.41 // indirect
	github.com/minio/blake2b-simd v0.0.0-20160723061019-3f5f724cb5b1 // indirect
	github.com/minio/sha256-simd v1.0.0 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/mr-tron/base58 v1.2.0 // indirect
	github.com/multiformats/go-base32 v0.0.3 // indirect
	github.com/multiformats/go-base36 v0.1.0 // indirect
	github.com/multiformats/go-multiaddr v0.3.3 // indirect
	github.com/multiformats/go-multiaddr-dns v0.3.1 // indirect
	github.com/multiformats/go-multiaddr-fmt v0.1.0 // indirect
	github.com/multiformats/go-multiaddr-net v0.2.0 // indirect
	github.com/multiformats/go-multibase v0.0.3 // indirect
	github.com/multiformats/go-multihash v0.0.15 // indirect
	github.com/multiformats/go-varint v0.0.6 // indirect
	github.com/nkovacs/streamquote v0.0.0-20170412213628-49af9bddb229 // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	github.com/oschwald/maxminddb-golang v1.8.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/polydawn/refmt v0.0.0-20201211092308-30ac6d18308e // indirect
	github.com/prometheus/client_golang v1.10.0 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.18.0 // indirect
	github.com/prometheus/procfs v0.6.0 // indirect
	github.com/raulk/clock v1.1.0 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20200410134404-eec4a21b6bb0 // indirect
	github.com/robfig/cron v1.1.0 // indirect
	github.com/rs/cors v1.6.0 // indirect
	github.com/russross/blackfriday/v2 v2.0.1 // indirect
	github.com/shiena/ansicolor v0.0.0-20151119151921-a422bbe96644 // indirect
	github.com/shirou/gopsutil v2.18.12+incompatible // indirect
	github.com/shurcooL/sanitized_anchor_name v1.0.0 // indirect
	github.com/spacemonkeygo/spacelog v0.0.0-20180420211403-2296661a0572 // indirect
	github.com/spaolacci/murmur3 v1.1.0 // indirect
	github.com/syndtr/goleveldb v1.0.0 // indirect
	github.com/tj/go-spin v1.1.0 // indirect
	github.com/urfave/cli/v2 v2.2.0 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v1.0.1 // indirect
	github.com/whyrusleeping/bencher v0.0.0-20190829221104-bb6607aa8bba // indirect
	github.com/whyrusleeping/pubsub v0.0.0-20190708150250-92bcb0691325 // indirect
	github.com/whyrusleeping/timecache v0.0.0-20160911033111-cfcb2f1abfee // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.0.2 // indirect
	github.com/xdg-go/stringprep v1.0.2 // indirect
	github.com/xlab/c-for-go v0.0.0-20201112171043-ea6dce5809cb // indirect
	github.com/xlab/pkgconfig v0.0.0-20170226114623-cea12a0fd245 // indirect
	github.com/youmark/pkcs8 v0.0.0-20181117223130-1be2e3e5546d // indirect
	go.mongodb.org/mongo-driver v1.7.3 // indirect
	go.opencensus.io v0.23.0 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.7.0 // indirect
	golang.org/x/crypto v0.0.0-20210711020723-a769d52b0f97 // indirect
	golang.org/x/mod v0.4.2 // indirect
	golang.org/x/net v0.0.0-20210428140749-89ef3d95e781 // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c // indirect
	golang.org/x/sys v0.0.0-20210630005230-0f9fa26af87c // indirect
	golang.org/x/text v0.3.6 // indirect
	golang.org/x/tools v0.1.5 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	google.golang.org/protobuf v1.27.1 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.0.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	howett.net/plist v0.0.0-20181124034731-591f970eefbb // indirect
	modernc.org/cc v1.0.0 // indirect
	modernc.org/golex v1.0.1 // indirect
	modernc.org/mathutil v1.4.0 // indirect
	modernc.org/strutil v1.1.1 // indirect
	modernc.org/xc v1.0.0 // indirect
	xorm.io/builder v0.3.9 // indirect
)
