// temp

package wavefront

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/ezeev/go-metrics-wavefront/tags"
	"github.com/rcrowley/go-metrics"
)

// A simple wrapper around go-metrics metrics.Register function
func RegisterMetric(key string, metric interface{}, tags ...tags.Tag) {
	key = tags.EncodeTags(key, tags)
	fmt.Println(key)
	metrics.Register(key, metric)
}

// WavefrontConfig provides a container with configuration parameters for
// the Wavefront exporter
type WavefrontConfig struct {
	Addr          *net.TCPAddr     // Network address to connect to
	Registry      metrics.Registry // Registry to be exported
	FlushInterval time.Duration    // Flush interval
	DurationUnit  time.Duration    // Time conversion unit for durations
	Prefix        string           // Prefix to be prepended to metric names
	Percentiles   []float64        // Percentiles to export from timers and histograms
}

// Wavefront is a blocking exporter function which reports metrics in r
// to a wavefront server located at addr, flushing them every d duration
// and prepending metric names with prefix.
func Wavefront(r metrics.Registry, d time.Duration, prefix string, addr *net.TCPAddr) {
	WavefrontWithConfig(WavefrontConfig{
		Addr:          addr,
		Registry:      r,
		FlushInterval: d,
		DurationUnit:  time.Nanosecond,
		Prefix:        prefix,
		Percentiles:   []float64{0.5, 0.75, 0.95, 0.99, 0.999},
	})
}

// WavefrontWithConfig is a blocking exporter function just like Wavefront,
// but it takes a WavefrontConfig instead.
func WavefrontWithConfig(c WavefrontConfig) {
	for _ = range time.Tick(c.FlushInterval) {
		if err := wavefront(&c); nil != err {
			log.Println(err)
		}
	}
}

// WavefrontOnce performs a single submission to Wavefront, returning a
// non-nil error on failed connections. This can be used in a loop
// similar to WavefrontWithConfig for custom error handling.
func WavefrontOnce(c WavefrontConfig) error {
	return wavefront(&c)
}

func wavefront(c *WavefrontConfig) error {
	now := time.Now().Unix()
	du := float64(c.DurationUnit)
	conn, err := net.DialTCP("tcp", nil, c.Addr)
	if nil != err {
		return err
	}
	defer conn.Close()
	w := bufio.NewWriter(conn)
	c.Registry.Each(func(name string, i interface{}) {
		switch metric := i.(type) {
		case metrics.Counter:
			fmt.Fprintf(w, "%s.%s.count %d %d\n", c.Prefix, name, metric.Count(), now)
		case metrics.Gauge:
			fmt.Fprintf(w, "%s.%s.value %d %d\n", c.Prefix, name, metric.Value(), now)
		case metrics.GaugeFloat64:
			fmt.Fprintf(w, "%s.%s.value %f %d\n", c.Prefix, name, metric.Value(), now)
		case metrics.Histogram:
			h := metric.Snapshot()
			ps := h.Percentiles(c.Percentiles)
			fmt.Fprintf(w, "%s.%s.count %d %d\n", c.Prefix, name, h.Count(), now)
			fmt.Fprintf(w, "%s.%s.min %d %d\n", c.Prefix, name, h.Min(), now)
			fmt.Fprintf(w, "%s.%s.max %d %d\n", c.Prefix, name, h.Max(), now)
			fmt.Fprintf(w, "%s.%s.mean %.2f %d\n", c.Prefix, name, h.Mean(), now)
			fmt.Fprintf(w, "%s.%s.std-dev %.2f %d\n", c.Prefix, name, h.StdDev(), now)
			for psIdx, psKey := range c.Percentiles {
				key := strings.Replace(strconv.FormatFloat(psKey*100.0, 'f', -1, 64), ".", "", 1)
				fmt.Fprintf(w, "%s.%s.%s-percentile %.2f %d\n", c.Prefix, name, key, ps[psIdx], now)
			}
		case metrics.Meter:
			m := metric.Snapshot()
			fmt.Fprintf(w, "%s.%s.count %d %d\n", c.Prefix, name, m.Count(), now)
			fmt.Fprintf(w, "%s.%s.one-minute %.2f %d\n", c.Prefix, name, m.Rate1(), now)
			fmt.Fprintf(w, "%s.%s.five-minute %.2f %d\n", c.Prefix, name, m.Rate5(), now)
			fmt.Fprintf(w, "%s.%s.fifteen-minute %.2f %d\n", c.Prefix, name, m.Rate15(), now)
			fmt.Fprintf(w, "%s.%s.mean %.2f %d\n", c.Prefix, name, m.RateMean(), now)
		case metrics.Timer:
			t := metric.Snapshot()
			ps := t.Percentiles(c.Percentiles)
			fmt.Fprintf(w, "%s.%s.count %d %d\n", c.Prefix, name, t.Count(), now)
			fmt.Fprintf(w, "%s.%s.min %d %d\n", c.Prefix, name, t.Min()/int64(du), now)
			fmt.Fprintf(w, "%s.%s.max %d %d\n", c.Prefix, name, t.Max()/int64(du), now)
			fmt.Fprintf(w, "%s.%s.mean %.2f %d\n", c.Prefix, name, t.Mean()/du, now)
			fmt.Fprintf(w, "%s.%s.std-dev %.2f %d\n", c.Prefix, name, t.StdDev()/du, now)
			for psIdx, psKey := range c.Percentiles {
				key := strings.Replace(strconv.FormatFloat(psKey*100.0, 'f', -1, 64), ".", "", 1)
				fmt.Fprintf(w, "%s.%s.%s-percentile %.2f %d\n", c.Prefix, name, key, ps[psIdx]/du, now)
			}
			fmt.Fprintf(w, "%s.%s.one-minute %.2f %d\n", c.Prefix, name, t.Rate1(), now)
			fmt.Fprintf(w, "%s.%s.five-minute %.2f %d\n", c.Prefix, name, t.Rate5(), now)
			fmt.Fprintf(w, "%s.%s.fifteen-minute %.2f %d\n", c.Prefix, name, t.Rate15(), now)
			fmt.Fprintf(w, "%s.%s.mean-rate %.2f %d\n", c.Prefix, name, t.RateMean(), now)
		}
		w.Flush()
	})
	return nil
}
