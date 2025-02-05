package main

import (
	"flag"
	"fmt"
	"net/http"
	"time"

	meraki "github.com/meraki/dashboard-api-go/v3/sdk"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

type merakiCollector struct {
	wirelessFailedConn        *prometheus.Desc
	wirelessClientCount       *prometheus.Desc
	deviceUplinkLoss          *prometheus.Desc
	deviceUplinkLatency       *prometheus.Desc
	channelUtilizationWifi    *prometheus.Desc
	channelUtilizationNonWifi *prometheus.Desc
	channelUtilizationTotal   *prometheus.Desc
	vpnPeerStatus             *prometheus.Desc
	thirdPartyVpnPeerStatus   *prometheus.Desc
	vpnPeerStatsAvgJitter     *prometheus.Desc
	vpnPeerStatsMinJitter     *prometheus.Desc
	vpnPeerStatsMaxJitter     *prometheus.Desc
	vpnPeerStatsAvgLatency    *prometheus.Desc
	vpnPeerStatsMinLatency    *prometheus.Desc
	vpnPeerStatsMaxLatency    *prometheus.Desc
	vpnPeerStatsAvgLoss       *prometheus.Desc
	vpnPeerStatsMinLoss       *prometheus.Desc
	vpnPeerStatsMaxLoss       *prometheus.Desc

	merakiClient *meraki.Client
}

func newMultipathCollector(client *meraki.Client) *merakiCollector {
	return &merakiCollector{
		wirelessFailedConn: prometheus.NewDesc("failed_conn",
			"Shows failed connections",
			[]string{"network_id", "device_serial", "failure_step", "failure_type"}, nil,
		),
		wirelessClientCount: prometheus.NewDesc("wireless_client_count",
			"Shows wireless client count",
			[]string{"network_id", "ip", "device_name", "device_serial"}, nil,
		),
		deviceUplinkLoss: prometheus.NewDesc("device_uplink_loss",
			"Shows device uplink loss",
			[]string{"network_id", "ip", "uplink", "device_serial"}, nil,
		),
		deviceUplinkLatency: prometheus.NewDesc("device_uplink_latency",
			"Shows device uplink latency",
			[]string{"network_id", "ip", "uplink", "device_serial"}, nil,
		),
		channelUtilizationWifi: prometheus.NewDesc("channel_utilization_wifi",
			"Shows channel utilization for wifi",
			[]string{"network_id", "device_serial", "band"}, nil,
		),
		channelUtilizationNonWifi: prometheus.NewDesc("channel_utilization_non_wifi",
			"Shows channel utilization for non wifi",
			[]string{"network_id", "device_serial", "band"}, nil,
		),
		channelUtilizationTotal: prometheus.NewDesc("channel_utilization_total",
			"Shows channel utilization total",
			[]string{"network_id", "device_serial", "band"}, nil,
		),
		vpnPeerStatus: prometheus.NewDesc("vpn_peer_status",
			"Shows vpn peer status",
			[]string{"network_id", "network_name", "device_serial", "peer_network_id", "peer_network_name", "status"}, nil,
		),
		thirdPartyVpnPeerStatus: prometheus.NewDesc("third_party_vpn_peer_status",
			"Shows third party vpn peer status",
			[]string{"network_id", "network_name", "device_serial", "peer_name", "peer_public_ip", "status"}, nil,
		),
		vpnPeerStatsAvgJitter: prometheus.NewDesc("vpn_peer_stats_avg_jitter",
			"Shows vpn peer avg jitter",
			[]string{"network_id", "network_name", "peer_network_id", "peer_network_name", "receiver_uplink", "sender_uplink"}, nil,
		),
		vpnPeerStatsMinJitter: prometheus.NewDesc("vpn_peer_stats_min_jitter",
			"Shows vpn peer min jitter",
			[]string{"network_id", "network_name", "peer_network_id", "peer_network_name", "receiver_uplink", "sender_uplink"}, nil,
		),
		vpnPeerStatsMaxJitter: prometheus.NewDesc("vpn_peer_stats_max_jitter",
			"Shows vpn peer max jitter",
			[]string{"network_id", "network_name", "peer_network_id", "peer_network_name", "receiver_uplink", "sender_uplink"}, nil,
		),
		vpnPeerStatsAvgLatency: prometheus.NewDesc("vpn_peer_stats_avg_latency",
			"Shows vpn peer avg latency",
			[]string{"network_id", "network_name", "peer_network_id", "peer_network_name", "receiver_uplink", "sender_uplink"}, nil,
		),
		vpnPeerStatsMinLatency: prometheus.NewDesc("vpn_peer_stats_min_latency",
			"Shows vpn peer min latency",
			[]string{"network_id", "network_name", "peer_network_id", "peer_network_name", "receiver_uplink", "sender_uplink"}, nil,
		),
		vpnPeerStatsMaxLatency: prometheus.NewDesc("vpn_peer_stats_max_latency",
			"Shows vpn peer max latency",
			[]string{"network_id", "network_name", "peer_network_id", "peer_network_name", "receiver_uplink", "sender_uplink"}, nil,
		),
		vpnPeerStatsAvgLoss: prometheus.NewDesc("vpn_peer_stats_avg_loss",
			"Shows vpn peer avg loss",
			[]string{"network_id", "network_name", "peer_network_id", "peer_network_name", "receiver_uplink", "sender_uplink"}, nil,
		),
		vpnPeerStatsMinLoss: prometheus.NewDesc("vpn_peer_stats_min_loss",
			"Shows vpn peer min loss",
			[]string{"network_id", "network_name", "peer_network_id", "peer_network_name", "receiver_uplink", "sender_uplink"}, nil,
		),
		vpnPeerStatsMaxLoss: prometheus.NewDesc("vpn_peer_stats_max_loss",
			"Shows vpn peer max loss",
			[]string{"network_id", "network_name", "peer_network_id", "peer_network_name", "receiver_uplink", "sender_uplink"}, nil,
		),
		merakiClient: client,
	}
}

func (collector *merakiCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.wirelessFailedConn
	ch <- collector.wirelessClientCount
	ch <- collector.deviceUplinkLoss
	ch <- collector.deviceUplinkLatency
	ch <- collector.channelUtilizationWifi
	ch <- collector.channelUtilizationNonWifi
	ch <- collector.channelUtilizationTotal
	ch <- collector.vpnPeerStatus
	ch <- collector.thirdPartyVpnPeerStatus
	ch <- collector.vpnPeerStatsAvgJitter
	ch <- collector.vpnPeerStatsMinJitter
	ch <- collector.vpnPeerStatsMaxJitter
}

func (collector *merakiCollector) Collect(ch chan<- prometheus.Metric) {
	orgs, _, err := collector.merakiClient.Organizations.GetOrganizations(nil)
	if err != nil {
		log.Error(err)
		return
	}
	if orgs == nil {
		log.Error("no data in orgs")
		return
	}

	collector.devicesUplinksLossAndLatency(ch, orgs)
	collector.wirelessDevicesChannelUtilizationByDevice(ch, orgs)
	collector.organizationApplianceVpnStatuses(ch, orgs)
	collector.getOrganizationApplianceVpnStats(ch, orgs)

	for _, org := range *orgs {
		devices, _, err := collector.merakiClient.Organizations.GetOrganizationDevices(org.ID, &meraki.GetOrganizationDevicesQueryParams{
			ProductTypes: []string{"wireless"},
		})
		if err != nil {
			log.Errorf("failed to get devices for %s: %v", org.Name, err)
			return
		}
		log.Infof("found %d wireless devices for %s", len(*devices), org.Name)

		for _, device := range *devices {
			collector.wirelessClientCountHistory(ch, &org, &device)
			collector.wirelessFailedConnections(ch, &org, &device)
			time.Sleep(time.Millisecond * 100)
		}
	}
}

var merakiApiKey = flag.String("meraki-api-key", "", "Access key for API")
var listenAddress = flag.String("web.listen-address", "9101", "Listen address")

func main() {
	flag.Parse()

	client, err := meraki.NewClientWithOptions("https://api.meraki.com/", *merakiApiKey, "false", "meraki prometheus-exporter")
	if err != nil {
		log.Fatal(err)
	}

	c := newMultipathCollector(client)
	prometheus.MustRegister(c)

	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", *listenAddress), nil))
}
