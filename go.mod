module github.com/hooto/hpress

go 1.26.0

// replace github.com/hooto/httpsrv v0.12.6 => /opt/workspace/src/github.com/hooto/httpsrv

replace github.com/lynkdb/lynksearch v0.1.0 => /opt/workspace/src/github.com/lynkdb/lynksearch

replace github.com/lynkdb/lynkapi v0.0.9 => /opt/workspace/src/github.com/lynkdb/lynkapi

replace github.com/lynkdb/lynkui v0.0.1 => /opt/workspace/src/github.com/lynkdb/lynkui

replace github.com/hooto/iam/v2 v2.0.0 => /opt/workspace/src/github.com/hooto/iam

replace github.com/sysinner/incore/v2 v2.0.0-alpha.2 => /opt/workspace/src/github.com/sysinner/incore

require (
	github.com/bamiaux/rez v0.0.0-20170731184118-29f4463c688b
	github.com/hooto/hcaptcha v0.1.6
	github.com/hooto/hchart v0.1.2
	github.com/hooto/hflag4g v0.10.1
	github.com/hooto/hini4g v0.1.2
	github.com/hooto/hlang4g v0.1.1
	github.com/hooto/hlog4g v0.9.5
	github.com/hooto/htoml4g v0.9.5
	github.com/hooto/httpsrv v0.13.0
	github.com/hooto/iam/v2 v2.0.0
	github.com/lessos/lessgo v1.0.1
	github.com/lynkdb/iomix v0.0.0-20210408130459-cc48edfc442f
	github.com/lynkdb/kvgo/v2 v2.0.15
	github.com/lynkdb/lynkapi v0.0.11
	github.com/lynkdb/lynksearch v0.1.0
	github.com/lynkdb/lynkui v0.0.1
	github.com/lynkdb/mysqlgo v0.0.0-20210408130716-96edd6491cba
	github.com/lynkdb/pgsqlgo v0.0.0-20210408130625-1c1f97eedf2c
	github.com/microcosm-cc/bluemonday v1.0.27
	github.com/nfnt/resize v0.0.0-20180221191011-83c6a9932646
	github.com/oliamb/cutter v0.2.2
	github.com/shurcooL/sanitized_anchor_name v1.0.0
	github.com/sysinner/incore/v2 v2.0.0-alpha.2
	github.com/ulikunitz/xz v0.5.15
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/DataDog/zstd v1.5.7 // indirect
	github.com/ServiceWeaver/weaver v0.24.6 // indirect
	github.com/andybalholm/brotli v1.2.1 // indirect
	github.com/aymerick/douceur v0.2.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/cockroachdb/errors v1.12.0 // indirect
	github.com/cockroachdb/fifo v0.0.0-20240816210425-c5d0cb0b6fc0 // indirect
	github.com/cockroachdb/logtags v0.0.0-20241215232642-bb51bb14a506 // indirect
	github.com/cockroachdb/pebble v1.1.5 // indirect
	github.com/cockroachdb/redact v1.1.8 // indirect
	github.com/cockroachdb/tokenbucket v0.0.0-20250429170803-42689b6311bb // indirect
	github.com/ebitengine/purego v0.10.0 // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/gabriel-vasile/mimetype v1.4.13 // indirect
	github.com/getsentry/sentry-go v0.45.1 // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.30.2 // indirect
	github.com/go-sql-driver/mysql v1.9.3 // indirect
	github.com/goccy/go-json v0.10.6 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0 // indirect
	github.com/golang/snappy v1.0.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/css v1.0.1 // indirect
	github.com/hooto/hauth/go v0.1.5 // indirect
	github.com/hooto/hauth/go/v2 v2.0.0-20260125120444-4cbf92d8d081 // indirect
	github.com/hooto/hmetrics v0.0.2 // indirect
	github.com/klauspost/compress v1.18.5 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/lufia/plan9stats v0.0.0-20260330125221-c963978e514e // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/power-devops/perfstat v0.0.0-20240221224432-82ca36839d55 // indirect
	github.com/prometheus/client_golang v1.23.2 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.67.5 // indirect
	github.com/prometheus/procfs v0.20.1 // indirect
	github.com/rakyll/statik v0.1.8 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	github.com/shirou/gopsutil/v4 v4.26.3 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/tklauser/go-sysconf v0.3.16 // indirect
	github.com/tklauser/numcpus v0.11.0 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	go.yaml.in/yaml/v2 v2.4.4 // indirect
	golang.org/x/crypto v0.50.0 // indirect
	golang.org/x/exp v0.0.0-20260410095643-746e56fc9e2f // indirect
	golang.org/x/image v0.35.0 // indirect
	golang.org/x/mod v0.35.0 // indirect
	golang.org/x/net v0.53.0 // indirect
	golang.org/x/sys v0.44.0 // indirect
	golang.org/x/text v0.36.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260511170946-3700d4141b60 // indirect
	google.golang.org/grpc v1.81.0 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)
