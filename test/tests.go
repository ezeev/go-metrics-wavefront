package main

import (
	"log"
	"net"
	"os"
	"time"

	"github.com/ezeev/go-metrics-wavefront"
	"github.com/ezeev/go-metrics-wavefront/tags"
	"github.com/rcrowley/go-metrics"
)

func main() {

	/*tags.EncodeTags(
	"mykey",
	tags.Tag{"tag1", "val1"},
	tags.Tag{"tag2", "val2"})
	*/
	c := metrics.NewCounter()
	//metrics.Register("foo", c)
	wavefront.RegisterMetric(
		"foo", c,
		tags.Tag{"key1", "val1"},
		tags.Tag{"key2", "val2"})

	c.Inc(47)

	g := metrics.NewGauge()
	metrics.Register("bar", g)
	g.Update(47)

	s := metrics.NewExpDecaySample(1028, 0.015) // or metrics.NewUniformSample(1028)
	h := metrics.NewHistogram(s)
	metrics.Register("baz", h)
	h.Update(47)

	m := metrics.NewMeter()
	metrics.Register("quux", m)
	m.Mark(47)

	t := metrics.NewTimer()
	metrics.Register("bang", t)
	t.Time(func() {})
	t.Update(47)

	addr, _ := net.ResolveTCPAddr("tcp", "192.168.99.100:2878")
	go wavefront.Wavefront(metrics.DefaultRegistry, 1*time.Second, "some.prefix", addr)
	go metrics.Log(metrics.DefaultRegistry, 5*time.Second, log.New(os.Stderr, "metrics: ", log.Lmicroseconds))

	for {

	}

}
