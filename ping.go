package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"regexp"
	"strings"
	"time"

	mp "github.com/mackerelio/go-mackerel-plugin"
	fping "github.com/tatsushid/go-fastping"
)

type PingPlugin struct {
	Prefix      string
	Hosts       []string
	Labels      []string
	Groups      map[string]string
	Tempfile    string
	Count       int
	WaitTime    int
	AcceptCount int
	SourceIP    string
	Stacked	    bool
}

func (pp PingPlugin) FetchMetrics() (map[string]float64, error) {
	stat := make(map[string]float64)
	total := make(map[string]float64)
	count := make(map[string]int)

	pinger := fping.NewPinger()
	pinger.OnRecv = func(addr *net.IPAddr, rtt time.Duration) {
		rttMilliSec := float64(rtt.Nanoseconds()) / 1000.0 / 1000.0
		total[escapeHostName(addr.String())] += rttMilliSec
		count[escapeHostName(addr.String())] += 1
	}

	for _, host := range pp.Hosts {
		total[escapeHostName(host)] = 0
		count[escapeHostName(host)] = 0
		pinger.AddIP(host)
	}

	pinger.MaxRTT = time.Millisecond * time.Duration(pp.WaitTime)
	if pp.SourceIP != "" {
		pinger.Source(pp.SourceIP)
	}

	for i := 0; i < pp.Count; i++ {
		err := pinger.Run()
		if err != nil {
			return nil, err
		}
	}
	for k, v := range total {
		if count[k] >= (pp.Count - pp.AcceptCount) {
			stat[k] = v / float64(count[k])
		} else {
			stat[k] = -1.0
		}
	}

	return stat, nil
}

func (pp PingPlugin) GraphDefinition() map[string]mp.Graphs {
	metrics := map[string][]mp.Metrics{}
	for i := 0; i < len(pp.Hosts); i++ {
		name := escapeHostName(pp.Hosts[i])
		key := "rtt"
		if pp.Groups[name] != ":def" && pp.Groups[name] != "" {
			key = fmt.Sprintf("%s.%s", key, pp.Groups[name])
		}
		metrics[key] = append(metrics[key], mp.Metrics{
			Name:    name,
			Label:   pp.Labels[i],
			Diff:    false,
			Stacked: pp.Stacked,
		})
		if os.Getenv("MACKEREL_AGENT_PLUGIN_META") != "" && key != "rtt" {
			metrics["rtt.#"] = append(metrics["rtt.#"], mp.Metrics{
				Name:    name,
				Label:   pp.Labels[i],
				Diff:    false,
				Stacked: pp.Stacked,
			})
		}
	}

	graphs := map[string]mp.Graphs{}
	for k, v := range metrics {
		label := "Ping Round Trip Times"
		if k != "rtt.#" && k != "rtt" {
			label = fmt.Sprintf("%s (Group: %s)", label, strings.SplitN(k, ".", 2)[1])
		} else if k == "rtt" {
			label = fmt.Sprintf("%s (Group: default)", label)
		}
		graphs[k] = mp.Graphs{
			Label:    label,
			Unit:     "float",
			Metrics:  v,
		}
	}

	return graphs
}

func (pp PingPlugin) MetricKeyPrefix() string {
	if pp.Prefix == "" {
		pp.Prefix = "ping"
	}
	return pp.Prefix
}

func escapeHostName(host string) string {
	return strings.Replace(strings.Replace(host, ".", "_", -1), ":", "_", -1)
}

func parseHostsString(optHost string, ipv6 bool, strict ...string) ([]string, map[string]string, []string, error) {
	hosts := strings.Split(optHost, ",")
	ips, groups, labels := make([]string, len(hosts)), make(map[string]string), make([]string, len(hosts))

	for i := 0; i < len(hosts); i++ {
		v := strings.SplitN(hosts[i], ":", 3)
		version := "ip4"
		if ipv6 {
			version = "ip6"
			if strings.Count(hosts[i], "[") == 1 {
				r := regexp.MustCompile(`\[(.*)\](:.*)(:.*)`)
				v1 := strings.Replace(r.ReplaceAllString(hosts[i], "$2"), ":", "", 1)
				v2 := strings.Replace(r.ReplaceAllString(hosts[i], "$3"), ":", "", 1)
				v = []string{r.ReplaceAllString(hosts[i], "$1"), v1, v2}
			}
		}
		ip, err := net.ResolveIPAddr(version, v[0])
		if err != nil {
			if strict[0] != "" {
				return nil, nil, nil, err
			}
			continue
		}
		ips[i] = ip.String()
		labels[i] = v[0]
		groups[escapeHostName(ips[i])] = ":def"
		if len(v) == 2 && v[1] != "" {
			labels[i] = v[1]
		} else if len(v) == 3 {
			if v[1] != "" {
				groups[escapeHostName(ips[i])] = v[1]
			}
			if v[2] != "" {
				labels[i] = v[2]
			}
		}
	}

	return ips, groups, labels, nil
}

func main() {
	optHost := flag.String("host", "127.0.0.1:localhost", "IPv4 Address[[:Group]:Metric Label],[IPv6 Address]:[[:Group]:Label],FQDN[[:Group]:Label],...")
	optTempfile := flag.String("tempfile", "", "Temp file name")
	optCount := flag.Int("count", 1, "Sending (and receiving) count ping packets.")
	optWaitTime := flag.Int("waittime", 1000, "Wait time, Max RTT(ms)")
	optAcceptCount := flag.Int("acceptmiss", 0, "Accept out of wait time count ping packets.")
	optIPv6 := flag.Bool("6", false, "Enable IPv6.")
	optSourceIP := flag.String("source", "", "Source IP Address. If the IP Address is invalid, it will be ignored.")
	optPrefix := flag.String("prefix", "", "Prefix of graph metrics.")
	optStacked := flag.Bool("stacked", false, "Use Stacked graph.")
	flag.Parse()

	hosts, groups, labels, err := parseHostsString(*optHost, *optIPv6, os.Getenv("MACKEREL_AGENT_PLUGIN_META"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err.Error())
		os.Exit(1)
	}

	var pp PingPlugin
	pp.Prefix = *optPrefix
	pp.Hosts = hosts
	pp.Groups = groups
	pp.Labels = labels
	pp.Count = *optCount
	pp.WaitTime = *optWaitTime
	pp.AcceptCount = *optAcceptCount
	pp.SourceIP = *optSourceIP
	pp.Stacked = *optStacked

	helper := mp.NewMackerelPlugin(pp)

	if *optTempfile != "" {
		helper.Tempfile = *optTempfile
	} else {
		helper.Tempfile = fmt.Sprintf("/tmp/mackerel-plugin-ping-%s", escapeHostName(strings.Join(hosts[:], "-")))
	}

	if os.Getenv("MACKEREL_AGENT_PLUGIN_META") != "" {
		helper.OutputDefinitions()
	} else {
		helper.OutputValues()
	}
}
