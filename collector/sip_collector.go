package collector

import (
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/robinmarechal/asterisk_exporter/cmd"
)

// sipCollector collector for all 'sip show ...' commands
type sipCollector struct {
	cmdRunner *cmd.CmdRunner
	logger    log.Logger

	// sip show peers
	totalPeers              *prometheus.Desc
	totalMonitoredOnline    *prometheus.Desc
	totalMonitoredOffline   *prometheus.Desc
	totalUnmonitoredOnline  *prometheus.Desc
	totalUnmonitoredOffline *prometheus.Desc
	totalSipStatusUnknown   *prometheus.Desc
	totalSipStatusQualified *prometheus.Desc
	// for individual peer
	peerStatus *prometheus.Desc

	// sip show channels
	dialogsActive *prometheus.Desc
	// sip show subscriptions
	subscriptionsActive *prometheus.Desc
	// sip show channelstats
	channelsActive *prometheus.Desc

	// sip show users
	users *prometheus.Desc

	// sip show registry
	totalRegistrations        *prometheus.Desc
	onlineRegistrationsCount  *prometheus.Desc
	offlineRegistrationsCount *prometheus.Desc
	registrationStatus        *prometheus.Desc

	collectorError *prometheus.Desc
}

type sipMetrics struct {
	PeersInfo       *cmd.PeersInfo
	SipChannelsInfo *cmd.SipChannelsInfo
	UsersInfo       *cmd.UsersInfo
	RegistriesInfo  *cmd.RegistriesInfo
}

func NewSipCollector(prefix string, cmdRunner *cmd.CmdRunner, logger log.Logger, collectorError *prometheus.Desc) Collector {
	return &sipCollector{
		cmdRunner:      cmdRunner,
		logger:         logger,
		collectorError: collectorError,
		totalPeers: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "sip", "current_peers"),
			"Number of SIP peers",
			nil, nil,
		),
		totalMonitoredOnline: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "sip", "current_monitored_online"),
			"Number of currently monitored online SIP",
			nil, nil,
		),
		totalMonitoredOffline: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "sip", "current_monitored_offline"),
			"Number of currently monitored offline SIP",
			nil, nil,
		),
		totalUnmonitoredOnline: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "sip", "current_unmonitored_online"),
			"Number of currently unmonitored online SIP",
			nil, nil,
		),
		totalUnmonitoredOffline: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "sip", "current_unmonitored_offline"),
			"Number of currently unmonitored offline SIP",
			nil, nil,
		),
		totalSipStatusUnknown: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "sip", "current_unknown"),
			"Current number of unknown SIP",
			nil, nil,
		),
		totalSipStatusQualified: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "sip", "current_qualified"),
			"Current number of qualified SIP",
			nil, nil,
		),
		peerStatus: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "sip", "peer_status"),
			"Status of individual SIP peers",
			[]string{"peer_name", "peer_status"}, nil,
		),
		dialogsActive: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "sip", "active_dialogs"),
			"Number of active SIP dialogs",
			nil, nil,
		),
		subscriptionsActive: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "sip", "active_subscriptions"),
			"Number of active SIP subscriptions",
			nil, nil,
		),
		channelsActive: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "sip", "active_channels"),
			"Number of active SIP channels",
			nil, nil,
		),
		users: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "sip", "users"),
			"Number of users",
			nil, nil,
		),
		totalRegistrations: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "sip", "total_registrations"),
			"Total number of SIP registrations",
			nil, nil,
		),
		onlineRegistrationsCount: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "sip", "registered_count"),
			"Number of successfully registered SIP accounts",
			nil, nil,
		),
		offlineRegistrationsCount: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "sip", "unregistered_count"),
			"Number of unregistered SIP accounts",
			nil, nil,
		),
		registrationStatus: prometheus.NewDesc(
			prometheus.BuildFQName(prefix, "sip", "registration_status"),
			"Status of individual SIP registrations",
			[]string{"username", "state"}, nil,
		),
	}
}

func (c *sipCollector) Name() string {
	return "sip"
}

func (c *sipCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.totalPeers
	ch <- c.totalMonitoredOnline
	ch <- c.totalMonitoredOffline
	ch <- c.totalUnmonitoredOnline
	ch <- c.totalUnmonitoredOffline
	ch <- c.totalSipStatusUnknown
	ch <- c.totalSipStatusQualified
	ch <- c.peerStatus
	ch <- c.dialogsActive
	ch <- c.subscriptionsActive
	ch <- c.channelsActive
	ch <- c.users
	ch <- c.totalRegistrations
	ch <- c.onlineRegistrationsCount
	ch <- c.offlineRegistrationsCount
	ch <- c.registrationStatus
}

func (c *sipCollector) Collect(ch chan<- prometheus.Metric) {
	level.Debug(c.logger).Log("msg", "collecting sip metrics")
	metrics, err := collectSipMetrics(c.cmdRunner)

	if err != nil {
		ch <- prometheus.MustNewConstMetric(c.collectorError, prometheus.GaugeValue, 1, c.Name())
		level.Error(c.logger).Log("err", err)
		return
	}

	level.Debug(c.logger).Log("msg", "sip metrics collected")

	ch <- prometheus.MustNewConstMetric(c.collectorError, prometheus.GaugeValue, 0, c.Name())

	c.updateMetrics(metrics, ch)
}

func collectSipMetrics(c *cmd.CmdRunner) (*sipMetrics, error) {
	metrics := &sipMetrics{
		PeersInfo:       c.PeersInfo(),
		SipChannelsInfo: c.SipChannelsInfo(),
		UsersInfo:       c.UsersInfo(),
		RegistriesInfo:  c.RegistriesInfo(),
	}

	return metrics, nil
}

func (c *sipCollector) updateMetrics(values *sipMetrics, ch chan<- prometheus.Metric) {
	ch <- prometheus.MustNewConstMetric(c.totalMonitoredOnline, prometheus.GaugeValue, float64(values.PeersInfo.MonitoredOnline))
	ch <- prometheus.MustNewConstMetric(c.totalMonitoredOffline, prometheus.GaugeValue, float64(values.PeersInfo.MonitoredOffline))
	ch <- prometheus.MustNewConstMetric(c.totalUnmonitoredOnline, prometheus.GaugeValue, float64(values.PeersInfo.UnmonitoredOnline))
	ch <- prometheus.MustNewConstMetric(c.totalUnmonitoredOffline, prometheus.GaugeValue, float64(values.PeersInfo.UnmonitoredOffline))
	ch <- prometheus.MustNewConstMetric(c.totalSipStatusUnknown, prometheus.GaugeValue, float64(values.PeersInfo.PeersStatusUnknown))
	ch <- prometheus.MustNewConstMetric(c.totalSipStatusQualified, prometheus.GaugeValue, float64(values.PeersInfo.PeersStatusQualified))

	for _, peer := range values.PeersInfo.IndividualPeers {
		ch <- prometheus.MustNewConstMetric(
			c.peerStatus,
			prometheus.GaugeValue,
			1, // 1 указывает на наличие пира, 0 на его отсутствие.
			peer.Name, peer.Status,
		)
	}

	ch <- prometheus.MustNewConstMetric(c.dialogsActive, prometheus.GaugeValue, float64(values.SipChannelsInfo.ActiveSipDialogs))
	ch <- prometheus.MustNewConstMetric(c.subscriptionsActive, prometheus.GaugeValue, float64(values.SipChannelsInfo.ActiveSipSubscriptions))
	ch <- prometheus.MustNewConstMetric(c.channelsActive, prometheus.GaugeValue, float64(values.SipChannelsInfo.ActiveSipChannels))

	ch <- prometheus.MustNewConstMetric(c.users, prometheus.GaugeValue, float64(values.UsersInfo.Users))

	ch <- prometheus.MustNewConstMetric(c.totalRegistrations, prometheus.GaugeValue, float64(values.RegistriesInfo.TotalRegistrations))
	ch <- prometheus.MustNewConstMetric(c.onlineRegistrationsCount, prometheus.GaugeValue, float64(values.RegistriesInfo.OnlineRegistrations))
	ch <- prometheus.MustNewConstMetric(c.offlineRegistrationsCount, prometheus.GaugeValue, float64(values.RegistriesInfo.OfflineRegistrations))

	for _, registry := range values.RegistriesInfo.IndividualRegistrations {
		ch <- prometheus.MustNewConstMetric(
			c.registrationStatus,
			prometheus.GaugeValue,
			1, // 1 указывает на наличие регистрации.
			registry.Username, registry.State,
		)
	}

	level.Debug(c.logger).Log("msg", "sip metrics built")
}
