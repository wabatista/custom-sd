package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/model"
	"github.com/prometheus/common/version"
	"github.com/prometheus/prometheus/discovery/targetgroup"
	"github.com/prometheus/prometheus/documentation/examples/custom-sd/adapter"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	a             = kingpin.New("sd adapter usage", "Tool to generate file_sd target files from prometheus query.")
	listenAddress = a.Flag("listen.address", "The address for listening.").Default("localhost:9091").String()
	roleMatch     = a.Flag("output.file", "Output file for file_sd compatible file.").Default("jmx_exporter").String()
	targetAddress = a.Flag("target.address", "The address the prometheus HTTP API is listening on for requests.").Default("localhost:9090").String()
	filePath      = a.Flag("output.path", "Path File location for fileSD.").Default("/opt/prometheus/conf/files_sd/").String()
	logger        log.Logger

	// addressLabel is the name for the label containing a target's address.
	addressLabel = model.MetaLabelPrefix + "target_address"
)

// MetricLabels is the result data from prometheus api.
// this struct represents the response from a /api/v1/query
type MetricLabels struct {
	Data struct {
		Result []struct {
			Metric map[string]string
			Value  []interface{}
		}
		ResultType string
	}
	Status string
}

// Note: Config struct for custom SD.
type sdConfig struct {
	Address         string
	RoleLabel       string
	RefreshInterval int
}

// Note: This is the struct with your implementation of the Discoverer interface (see Run function).
// Discovery retrieves target information from prometheus and updates them via watches.
type discovery struct {
	address         string
	refreshInterval int
	roleLabel       string
	logger          log.Logger
	oldSourceList   map[string]bool
}

func (d *discovery) parseServiceNodes(metric map[string]string, name string) (*targetgroup.Group, error) {

	tgroup := targetgroup.Group{
		Source: name,
		Labels: make(model.LabelSet),
	}

	tgroup.Targets = make([]model.LabelSet, 0, len(metric))
	instance := strings.Split(metric["instance"], ":")[0] + ":" + metric["exporter_port"]
	tgroup.Source = instance
	target := model.LabelSet{model.AddressLabel: model.LabelValue(instance)}
	for k, v := range metric {
		if k == "__name__" {
			tgroup.Labels[model.LabelName(k)] = model.LabelValue(v)
			continue
		}
		tgroup.Labels[model.LabelName(model.MetaLabelPrefix+k)] = model.LabelValue(v)
	}
	tgroup.Targets = append(tgroup.Targets, target)

	return &tgroup, nil
}

// Query function on prometheus instant query
// The query is up{role=`Exporter Tag`, exporter_port=~'.+', metrics_path=~'.+', app=~'.+'}
func (d *discovery) Run(ctx context.Context, ch chan<- []*targetgroup.Group) {
	for c := time.Tick(time.Duration(d.refreshInterval) * time.Second); ; {

		endpoint := fmt.Sprintf("http://%v/api/v1/query?", d.address)
		query := fmt.Sprintf("up{role=~'%s', exporter_port=~'.+', metrics_path=~'.+', app=~'.+'}", d.roleLabel)

		q := url.Values{}
		q.Add("query", query)

		req, err := http.NewRequest("GET", endpoint+q.Encode(), nil)

		if err != nil {
			level.Error(d.logger).Log("msg", "Error on connect to prometheus server", "err", err)
			continue
		}
		resp, err := http.Get(req.URL.String())

		if err != nil {
			level.Error(d.logger).Log("msg", "Error getting metrics list", "err", err)
			time.Sleep(time.Duration(d.refreshInterval) * time.Second)
			continue
		}
		var tgs []*targetgroup.Group

		name := d.roleLabel
		newSourceList := make(map[string]bool)

		var metrics *MetricLabels

		dec := json.NewDecoder(resp.Body)
		defer func() {
			io.Copy(ioutil.Discard, resp.Body)
			resp.Body.Close()
		}()
		for dec.More() {
			err := dec.Decode(&metrics)
			if err != nil {
				level.Error(d.logger).Log("msg", "Error reading list", "err", err)
				time.Sleep(time.Duration(d.refreshInterval) * time.Second)
				continue
			}
		}

		for _, v := range metrics.Data.Result {

			tg, err := d.parseServiceNodes(v.Metric, name)
			if err != nil {
				level.Error(d.logger).Log("msg", "Error parsing metrics", "service", name, "err", err)
				break
			}
			tgs = append(tgs, tg)
			newSourceList[tg.Source] = true
		}
		// When targetGroup disappear, send an update with empty targetList.
		for key := range d.oldSourceList {
			if !newSourceList[key] {
				tgs = append(tgs, &targetgroup.Group{
					Source: key,
				})
			}
		}
		d.oldSourceList = newSourceList
		if err == nil {
			// We're returning all exporters as a single targetgroup.
			ch <- tgs
		}
		// Wait for ticker or exit when ctx is closed.
		select {
		case <-c:
			continue
		case <-ctx.Done():
			return
		}
	}
}

func newDiscovery(conf sdConfig) (*discovery, error) {
	cd := &discovery{
		address:         conf.Address,
		refreshInterval: conf.RefreshInterval,
		roleLabel:       conf.RoleLabel,
		logger:          logger,
		oldSourceList:   make(map[string]bool),
	}
	return cd, nil
}

func main() {
	a.HelpFlag.Short('h')

	_, err := a.Parse(os.Args[1:])
	if err != nil {
		fmt.Println("err: ", err)
		return
	}
	metrics := prometheus.NewRegistry()
	metrics.MustRegister(
		version.NewCollector("CustomDiscovery"),
		prometheus.NewGoCollector(),
		prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}),
	)
	// Some packages still use default Register. Replace to have those metrics.
	prometheus.DefaultRegisterer = metrics

	logger = log.NewSyncLogger(log.NewLogfmtLogger(os.Stdout))
	logger = log.With(logger, "ts", log.DefaultTimestampUTC, "caller", log.DefaultCaller)

	ctx := context.Background()
	for _, host := range strings.Split(*targetAddress, ",") {
		for _, role := range strings.Split(*roleMatch, ",") {
			// Instance of new SD config.
			cfg := sdConfig{
				RoleLabel:       role,
				Address:         host,
				RefreshInterval: 30,
			}

			disc, err := newDiscovery(cfg)
			if err != nil {
				fmt.Println("err: ", err)
			}

			sdAdapter := adapter.NewAdapter(ctx, *filePath+role+".metrics.json", role, disc, logger)
			sdAdapter.Run()
		}
	}
	http.Handle("/metrics", promhttp.HandlerFor(
		prometheus.DefaultGatherer,
		promhttp.HandlerOpts{
			// Opt into OpenMetrics to support exemplars.
			EnableOpenMetrics: true,
		},
	))
	http.ListenAndServe(*listenAddress, nil)
	<-ctx.Done()

}
