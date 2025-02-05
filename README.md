# Meraki Exporter

The Meraki Exporter is a Prometheus exporter for Cisco Meraki networks. It exposes a variety of metrics related to wireless connectivity, device uplinks, and VPN peer status. These metrics are designed to help you monitor the health and performance of your Meraki devices and network.

# Metrics

The following metrics are exposed by the Meraki Exporter:

Wireless Metrics

    wirelessFailedConn:
        Description: The number of failed wireless connections.
        Type: Gauge
        Units: Count

    wirelessClientCount:
        Description: The number of clients currently connected to the wireless network.
        Type: Gauge
        Units: Count

Device Uplink Metrics

    deviceUplinkLoss:
        Description: The packet loss percentage on the device's uplink.
        Type: Gauge
        Units: Percentage (%)

    deviceUplinkLatency:
        Description: The latency (in milliseconds) on the device's uplink.
        Type: Gauge
        Units: Milliseconds (ms)

Channel Utilization Metrics

    channelUtilizationWifi:
        Description: The channel utilization for Wi-Fi channels.
        Type: Gauge
        Units: Percentage (%)

    channelUtilizationNonWifi:
        Description: The channel utilization for non-Wi-Fi channels.
        Type: Gauge
        Units: Percentage (%)

    channelUtilizationTotal:
        Description: The total channel utilization (Wi-Fi + non-Wi-Fi).
        Type: Gauge
        Units: Percentage (%)

VPN Metrics

    vpnPeerStatus:
        Description: The status of a VPN peer connection (up/down).
        Type: Gauge
        Units: 0 = down, 1 = up

    thirdPartyVpnPeerStatus:
        Description: The status of a third-party VPN peer connection (up/down).
        Type: Gauge
        Units: 0 = down, 1 = up

VPN Peer Stats (Jitter, Latency, Loss)

    vpnPeerStatsAvgJitter:
        Description: The average jitter (in milliseconds) for a VPN peer connection.
        Type: Gauge
        Units: Milliseconds (ms)

    vpnPeerStatsMinJitter:
        Description: The minimum jitter (in milliseconds) for a VPN peer connection.
        Type: Gauge
        Units: Milliseconds (ms)

    vpnPeerStatsMaxJitter:
        Description: The maximum jitter (in milliseconds) for a VPN peer connection.
        Type: Gauge
        Units: Milliseconds (ms)

    vpnPeerStatsAvgLatency:
        Description: The average latency (in milliseconds) for a VPN peer connection.
        Type: Gauge
        Units: Milliseconds (ms)

    vpnPeerStatsMinLatency:
        Description: The minimum latency (in milliseconds) for a VPN peer connection.
        Type: Gauge
        Units: Milliseconds (ms)

    vpnPeerStatsMaxLatency:
        Description: The maximum latency (in milliseconds) for a VPN peer connection.
        Type: Gauge
        Units: Milliseconds (ms)

    vpnPeerStatsAvgLoss:
        Description: The average packet loss percentage for a VPN peer connection.
        Type: Gauge
        Units: Percentage (%)

    vpnPeerStatsMinLoss:
        Description: The minimum packet loss percentage for a VPN peer connection.
        Type: Gauge
        Units: Percentage (%)

    vpnPeerStatsMaxLoss:
        Description: The maximum packet loss percentage for a VPN peer connection.
        Type: Gauge
        Units: Percentage (%)


# Installation

Clone the repository:


```shell
git clone https://github.com/yourusername/meraki-exporter.git
cd meraki-exporter
```

Build the exporter:

```shell
go build
```

Run the exporter:

```shell
./meraki-exporter --meraki-api-key <your-api-key> --metrics-port 9100
```

Replace <your-api-key> with your actual Meraki API key.

# Configuration

The exporter exposes metrics at the /metrics endpoint. You can configure the exporter to collect Meraki data using the API key you provide. Make sure that the API key has appropriate access to the Meraki organization and network.

Flags

    --meraki-api-key: Your Meraki API key (required)
    --web.listen-address: Address to bind the exporter to (default: :9101)

# Usage with Prometheus

To scrape metrics from the exporter, add the following job definition to your Prometheus configuration:

    scrape_configs:
      - job_name: 'meraki-exporter'
        static_configs:
          - targets: ['<MERAKI_EXPORTER_HOST>:9101']

Replace <MERAKI_EXPORTER_HOST> with the actual host where the exporter is running.

# License

MIT License. See LICENSE for more details.
