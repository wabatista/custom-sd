module pdi

require (
	github.com/alecthomas/units v0.0.0-20201120081800-1786d5ef83d4 // indirect
	github.com/go-kit/kit v0.10.0
	github.com/google/go-cmp v0.5.4 // indirect
	github.com/prometheus/client_golang v1.8.0
	github.com/prometheus/common v0.15.0
	github.com/prometheus/prometheus v1.8.2-0.20201119142752-3ad25a6dc3d9
	golang.org/x/mod v0.4.0 // indirect
	golang.org/x/net v0.0.0-20201207224615-747e23833adb // indirect
	golang.org/x/sys v0.0.0-20201207223542-d4d67f95c62d // indirect
	golang.org/x/tools v0.0.0-20201208002638-66f931576d67 // indirect
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

replace (
	// Using a 3rd-party branch for custom dialer - see https://github.com/bradfitz/gomemcache/pull/86.
	// Required by Cortex https://github.com/cortexproject/cortex/pull/3051.
	github.com/bradfitz/gomemcache => github.com/themihai/gomemcache v0.0.0-20180902122335-24332e2d58ab
	// Update to v1.1.1 to make sure windows CI pass.
	github.com/elastic/go-sysinfo => github.com/elastic/go-sysinfo v1.1.1
	// Make sure Prometheus version is pinned as Prometheus semver does not include Go APIs.
	github.com/prometheus/prometheus => github.com/prometheus/prometheus v1.8.2-0.20201119142752-3ad25a6dc3d9
	github.com/sercand/kuberesolver => github.com/sercand/kuberesolver v2.4.0+incompatible
	google.golang.org/grpc => google.golang.org/grpc v1.29.1

	// From Prometheus.
	k8s.io/klog => github.com/simonpasquier/klog-gokit v0.3.0
	k8s.io/klog/v2 => github.com/simonpasquier/klog-gokit/v2 v2.0.1
)

go 1.15
