package main

import (
	"strings"

	meraki "github.com/meraki/dashboard-api-go/v3/sdk"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

func (collector *merakiCollector) wirelessFailedConnections(ch chan<- prometheus.Metric, org *meraki.ResponseItemOrganizationsGetOrganizations, device *meraki.ResponseItemOrganizationsGetOrganizationDevices) {
	failedConnData, _, err := collector.merakiClient.Wireless.GetNetworkWirelessFailedConnections(device.NetworkID, &meraki.GetNetworkWirelessFailedConnectionsQueryParams{
		Timespan: 300,
	})
	if err != nil || failedConnData == nil {
		log.Errorf("failed to get failed conn count for %s network %s device %s: %v", org.Name, device.NetworkID, device.Serial, err)
		return
	}
	failureStep := make(map[string]map[string]int)
	for _, item := range *failedConnData {
		if failureStep[item.FailureStep] == nil {
			failureStep[item.FailureStep] = make(map[string]int)
		}

		failureStep[item.FailureStep][item.Type] += 1
	}

	for fstep, data := range failureStep {
		for ftype, value := range data {
			failedConn := prometheus.MustNewConstMetric(collector.wirelessFailedConn, prometheus.GaugeValue, float64(value),
				device.NetworkID, device.Serial, fstep, ftype)
			ch <- failedConn
		}
	}
}

func (collector *merakiCollector) wirelessClientCountHistory(ch chan<- prometheus.Metric, org *meraki.ResponseItemOrganizationsGetOrganizations, device *meraki.ResponseItemOrganizationsGetOrganizationDevices) {
	clientCountData, _, err := collector.merakiClient.Wireless.GetNetworkWirelessClientCountHistory(device.NetworkID, &meraki.GetNetworkWirelessClientCountHistoryQueryParams{
		Timespan:     300,
		Resolution:   300,
		DeviceSerial: device.Serial,
	})
	if err != nil || clientCountData == nil {
		log.Errorf("failed to get client count for %s network %s device %s: %v", org.Name, device.NetworkID, device.Serial, err)
		return
	}
	for _, item := range *clientCountData {
		if item.ClientCount != nil {
			clientCount := prometheus.MustNewConstMetric(collector.wirelessClientCount, prometheus.GaugeValue, float64(*item.ClientCount),
				device.NetworkID, device.LanIP, device.Name, device.Serial)
			ch <- clientCount
		}
	}
}

func (collector *merakiCollector) wirelessDevicesChannelUtilizationByDevice(ch chan<- prometheus.Metric, orgs *meraki.ResponseOrganizationsGetOrganizations) {
	for _, org := range *orgs {
		channelUtilizationData, _, err := collector.merakiClient.Wireless.GetOrganizationWirelessDevicesChannelUtilizationByDevice(org.ID, &meraki.GetOrganizationWirelessDevicesChannelUtilizationByDeviceQueryParams{
			Timespan: 300,
			Interval: 300,
		})
		if err != nil || channelUtilizationData == nil {
			log.Errorf("failed to get channel utilization for %s: %v", org.Name, err)
			return
		}

		for _, device := range *channelUtilizationData {
			for _, band := range *device.ByBand {
				channelUtilizationWifi := prometheus.MustNewConstMetric(collector.channelUtilizationWifi, prometheus.GaugeValue, float64(*band.Wifi.Percentage),
					device.Network.ID, device.Serial, band.Band)
				ch <- channelUtilizationWifi

				channelUtilizationNonWifi := prometheus.MustNewConstMetric(collector.channelUtilizationNonWifi, prometheus.GaugeValue, float64(*band.NonWifi.Percentage),
					device.Network.ID, device.Serial, band.Band)
				ch <- channelUtilizationNonWifi

				channelUtilizationTotal := prometheus.MustNewConstMetric(collector.channelUtilizationTotal, prometheus.GaugeValue, float64(*band.Total.Percentage),
					device.Network.ID, device.Serial, band.Band)
				ch <- channelUtilizationTotal
			}
		}
	}
}

func (collector *merakiCollector) devicesUplinksLossAndLatency(ch chan<- prometheus.Metric, orgs *meraki.ResponseOrganizationsGetOrganizations) {
	for _, org := range *orgs {
		res, _, err := collector.merakiClient.Organizations.GetOrganizationDevicesUplinksLossAndLatency(org.ID, &meraki.GetOrganizationDevicesUplinksLossAndLatencyQueryParams{
			Timespan: 300,
		})
		if err != nil {
			log.Errorf("failed to get devices uplinks loss and latency for %s: %v", org.Name, err)
			return
		}

		for _, device := range *res {
			var loss, latency float64
			for _, item := range *device.TimeSeries {
				if item.LossPercent != nil {
					if *item.LossPercent > loss {
						loss = *item.LossPercent
					}
				}
				if item.LatencyMs != nil {
					if *item.LatencyMs > latency {
						latency = *item.LatencyMs
					}
				}
			}

			uplinkLoss := prometheus.MustNewConstMetric(collector.deviceUplinkLoss, prometheus.GaugeValue, loss,
				device.NetworkID, device.IP, device.Uplink, device.Serial)

			uplinkLatency := prometheus.MustNewConstMetric(collector.deviceUplinkLatency, prometheus.GaugeValue, latency,
				device.NetworkID, device.IP, device.Uplink, device.Serial)

			ch <- uplinkLoss
			ch <- uplinkLatency
		}
	}
}

func (collector *merakiCollector) organizationApplianceVpnStatuses(ch chan<- prometheus.Metric, orgs *meraki.ResponseOrganizationsGetOrganizations) {
	for _, org := range *orgs {
		res, _, err := collector.merakiClient.Appliance.GetOrganizationApplianceVpnStatuses(org.ID, nil)
		if err != nil {
			log.Errorf("failed to get appliance vpn statuses for %s: %v", org.Name, err)
			return
		}

		for _, network := range *res {
			for _, peer := range *network.MerakiVpnpeers {
				status := strings.ToLower(peer.Reachability)
				if status == "reachable" {
					reachable := prometheus.MustNewConstMetric(collector.vpnPeerStatus, prometheus.GaugeValue, 1,
						network.NetworkID, network.NetworkName, network.DeviceSerial, peer.NetworkID, peer.NetworkName, "reachable")
					ch <- reachable
				} else {
					unreachable := prometheus.MustNewConstMetric(collector.vpnPeerStatus, prometheus.GaugeValue, 1,
						network.NetworkID, network.NetworkName, network.DeviceSerial, peer.NetworkID, peer.NetworkName, status)
					ch <- unreachable
				}
			}

			for _, peer := range *network.ThirdPartyVpnpeers {
				status := strings.ToLower(peer.Reachability)
				if status == "reachable" {
					reachable := prometheus.MustNewConstMetric(collector.thirdPartyVpnPeerStatus, prometheus.GaugeValue, 1,
						network.NetworkID, network.NetworkName, network.DeviceSerial, peer.Name, peer.PublicIP, "reachable")
					ch <- reachable
				} else {
					unreachable := prometheus.MustNewConstMetric(collector.thirdPartyVpnPeerStatus, prometheus.GaugeValue, 1,
						network.NetworkID, network.NetworkName, network.DeviceSerial, peer.Name, peer.PublicIP, status)
					ch <- unreachable
				}
			}
		}
	}
}

func (collector *merakiCollector) getOrganizationApplianceVpnStats(ch chan<- prometheus.Metric, orgs *meraki.ResponseOrganizationsGetOrganizations) {
	for _, org := range *orgs {
		res, _, err := collector.merakiClient.Appliance.GetOrganizationApplianceVpnStats(org.ID, nil)
		if err != nil {
			log.Errorf("failed to get appliance vpn statuses for %s: %v", org.Name, err)
			return
		}

		for _, network := range *res {
			for _, peer := range *network.MerakiVpnpeers {
				for _, jitter := range *peer.JitterSummaries {
					avgJitter := prometheus.MustNewConstMetric(collector.vpnPeerStatsAvgJitter, prometheus.GaugeValue, *jitter.AvgJitter,
						network.NetworkID, network.NetworkName, peer.NetworkID, peer.NetworkName, jitter.ReceiverUplink, jitter.SenderUplink)
					ch <- avgJitter

					minJitter := prometheus.MustNewConstMetric(collector.vpnPeerStatsMinJitter, prometheus.GaugeValue, *jitter.MinJitter,
						network.NetworkID, network.NetworkName, peer.NetworkID, peer.NetworkName, jitter.ReceiverUplink, jitter.SenderUplink)
					ch <- minJitter

					maxJitter := prometheus.MustNewConstMetric(collector.vpnPeerStatsMaxJitter, prometheus.GaugeValue, *jitter.MaxJitter,
						network.NetworkID, network.NetworkName, peer.NetworkID, peer.NetworkName, jitter.ReceiverUplink, jitter.SenderUplink)
					ch <- maxJitter
				}

				for _, latency := range *peer.LatencySummaries {
					avgLatency := prometheus.MustNewConstMetric(collector.vpnPeerStatsAvgLatency, prometheus.GaugeValue, *latency.AvgLatencyMs,
						network.NetworkID, network.NetworkName, peer.NetworkID, peer.NetworkName, latency.ReceiverUplink, latency.SenderUplink)
					ch <- avgLatency

					minLatency := prometheus.MustNewConstMetric(collector.vpnPeerStatsMinLatency, prometheus.GaugeValue, *latency.MinLatencyMs,
						network.NetworkID, network.NetworkName, peer.NetworkID, peer.NetworkName, latency.ReceiverUplink, latency.SenderUplink)
					ch <- minLatency

					maxLatency := prometheus.MustNewConstMetric(collector.vpnPeerStatsMaxLatency, prometheus.GaugeValue, *latency.MaxLatencyMs,
						network.NetworkID, network.NetworkName, peer.NetworkID, peer.NetworkName, latency.ReceiverUplink, latency.SenderUplink)
					ch <- maxLatency
				}

				for _, loss := range *peer.LossPercentageSummaries {
					avgLoss := prometheus.MustNewConstMetric(collector.vpnPeerStatsAvgLoss, prometheus.GaugeValue, *loss.AvgLossPercentage,
						network.NetworkID, network.NetworkName, peer.NetworkID, peer.NetworkName, loss.ReceiverUplink, loss.SenderUplink)
					ch <- avgLoss

					minLoss := prometheus.MustNewConstMetric(collector.vpnPeerStatsMinLoss, prometheus.GaugeValue, *loss.MinLossPercentage,
						network.NetworkID, network.NetworkName, peer.NetworkID, peer.NetworkName, loss.ReceiverUplink, loss.SenderUplink)
					ch <- minLoss

					maxLoss := prometheus.MustNewConstMetric(collector.vpnPeerStatsMaxLoss, prometheus.GaugeValue, *loss.MaxLossPercentage,
						network.NetworkID, network.NetworkName, peer.NetworkID, peer.NetworkName, loss.ReceiverUplink, loss.SenderUplink)
					ch <- maxLoss
				}
			}
		}
	}
}
