/*-
 * Copyright © 2016-2017, Jörg Pernfuß <code.jpe@gmail.com>
 * All rights reserved.
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package erebos // import "github.com/mjolnir42/erebos"

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/url"
	"path/filepath"

	"github.com/client9/reopen"
	ucl "github.com/nahanni/go-ucl"
	"github.com/ulule/deepcopier"
)

// Config holds the runtime configuration which is expected to be
// read from a UCL formatted file
type Config struct {
	// Log is the namespace for logging options
	Log struct {
		// Name of the logfile
		File string `json:"file"`
		// Path in wich to open the logfile
		Path string `json:"path"`
		// Reopen the logfile if SIGUSR2 is received
		Rotate bool `json:"rotate.on.usr2,string"`
		// Handle to the logfile
		FH *reopen.FileWriter `json:"-"`
		// Switch to enable debug logging
		Debug bool `json:"debug,string"`
	} `json:"log"`
	// Zookeeper is the namespace with options for Apache Zookeeper
	Zookeeper struct {
		// How often to publish offset updates to Zookeeper
		CommitInterval int `json:"commit.ms,string"`
		// Conncect string for the Zookeeper-Ensemble to use
		Connect string `json:"connect.string"`
		// If true, the Zookeeper stored offset will be ignored and
		// the newest message consumed
		ResetOffset bool `json:"reset.offset.on.startup,string"`
	} `json:"zookeeper"`
	// Kafka is the namespace with options for Apache Kafka
	Kafka struct {
		// Name of the consumergroup to join
		ConsumerGroup string `json:"consumer.group.name"`
		// Which topics to consume from
		ConsumerTopics string `json:"consumer.topics"`
		// Where to start consuming: Oldest, Newest
		ConsumerOffsetStrategy string `json:"consumer.offset.strategy"`
		// Which topic to produce to
		ProducerTopic string `json:"producer.topic"`
		// Producer-Response behaviour: NoResponse, WaitForLocal or WaitForAll
		ProducerResponseStrategy string `json:"producer.response.strategy"`
		// Producer retry attempts
		ProducerRetry int `json:"producer.retry.attempts,string"`
		// Keepalive interval in milliseconds
		Keepalive int `json:"keepalive.ms,string"`
	} `json:"kafka"`
	// Twister is the namespace with configuration options relating to
	// the splitting of metric batches
	Twister struct {
		HandlerQueueLength int      `json:"handler.queue.length,string"`
		QueryMetrics       []string `json:"query.metric.profiles"`
	} `json:"twister"`
	// Mistral is the namespace with configuration options relating to
	// accepting incoming messages via HTTP API
	Mistral struct {
		HandlerQueueLength int    `json:"handler.queue.length,string"`
		ListenAddress      string `json:"listen.address"`
		ListenPort         string `json:"listen.port"`
		ListenScheme       string `json:"listen.scheme"`
		EndpointPath       string `json:"api.endpoint.path"`
		ShutdownGrace      int    `json:"graceful.shutdown.delay.seconds,string"`
		Authentication     string `json:"authentication.style"`
	} `json:"mistral"`
	// DustDevil is the namespace with configuration options relating to
	// forwarding Kafka read messages to an HTTP API
	DustDevil struct {
		HandlerQueueLength int    `json:"handler.queue.length,string"`
		Endpoint           string `json:"api.endpoint"`
		RetryCount         int    `json:"post.request.retry.count,string"`
		RetryMinWaitTime   int    `json:"retry.min.wait.time.ms,string"`
		RetryMaxWaitTime   int    `json:"retry.max.wait.time.ms,string"`
		RequestTimeout     int    `json:"request.timeout.ms,string"`
		StripStringMetrics bool   `json:"strip.string.metrics,string"`
		ConcurrencyLimit   uint32 `json:"post.request.concurrency.limit,string"`
		ForwardElastic     bool   `json:"api.endpoint.is.elasticsearch,string"`
		InputFormat        string `json:"input.format"`
	} `json:"dustdevil"`
	// Cyclone is the namespace with configuration options relating
	// to threshold evaluation of metrics
	Cyclone struct {
		MetricsMaxAge      int      `json:"metrics.max.age.minutes,string"`
		DestinationURI     string   `json:"alarming.destination"`
		TestMode           bool     `json:"testmode,string"`
		APIVersion         string   `json:"api.version"`
		HandlerQueueLength int      `json:"handler.queue.length,string"`
		ConcurrencyLimit   uint32   `json:"post.request.concurrency.limit,string"`
		RetryCount         int      `json:"post.request.retry.count,string"`
		RetryMinWaitTime   int      `json:"retry.min.wait.time.ms,string"`
		RetryMaxWaitTime   int      `json:"retry.max.wait.time.ms,string"`
		RequestTimeout     int      `json:"request.timeout.ms,string"`
		DiscardMetrics     []string `json:"discard.metrics"`
	} `json:"cyclone"`
	// Redis is the namespace with configuration options relating
	// to Redis
	Redis struct {
		Connect      string `json:"connect"`
		Password     string `json:"password"`
		DB           int    `json:"db.number,string"`
		CacheTimeout uint64 `json:"cache.timeout.seconds,string"`
	} `json:"redis"`
	// Legacy is the namespace with configuration options relating to
	// legacy data formats
	Legacy struct {
		// Path for the legacy.MetricSocket
		SocketPath string `json:"socket.path"`
		// Print out metrics on STDERR. Requires the application to set a
		// debug formatting function
		MetricsDebug bool `json:"metrics.debug.stderr,string"`
		// Frequency to print out debug metrics
		MetricsFrequency int `json:"metrics.debug.frequency.seconds,string"`
	} `json:"legacy"`
	// ElasticSearch is the namespace with configuration options relating
	// to ElasticSearch
	ElasticSearch struct {
		Endpoint       string `json:"endpoint"`
		ClusterMetrics bool   `json:"collect.cluster.metrics,string"`
		LocalMetrics   bool   `json:"collect.local.metrics,string"`
	} `json:"elasticsearch"`
	// Misc is the namespace with miscellaneous settings
	Misc struct {
		// Whether to produce metrics or not
		ProduceMetrics bool `json:"produce.metrics,string"`
		// Name of the application instance
		InstanceName string `json:"instance.name"`
		// ReadOnly sets the application to readonly if supported
		ReadOnly bool `json:"readonly,string"`
	} `json:"misc"`
	// Hurricane is the namespace with configuration options relating
	// to the calculation of derived metrics
	Hurricane struct {
		HandlerQueueLength int  `json:"handler.queue.length,string"`
		DeriveCTX          bool `json:"derive.ctx.metrics,string"`
		DeriveCPU          bool `json:"derive.cpu.metrics,string"`
		DeriveMEM          bool `json:"derive.mem.metrics,string"`
		DeriveDISK         bool `json:"derive.disk.metrics,string"`
		DeriveNETIF        bool `json:"derive.netif.metrics,string"`
	} `json:"hurricane"`
	// Eyewall is the namespace for configuration options relating to
	// the eye lookup client library
	Eyewall struct {
		Host string `json:"host"`
		Port string `json:"port"`
		Path string `json:"path"`
		// number of allowed concurrent lookups to Eye per Eyewall
		// instance, 0 for unlimited
		ConcurrencyLimit uint32 `json:"eye.request.concurrency.limit,string"`
		// ApplicationName overrides the application name with which the
		// lookup client registers with Eye
		ApplicationName string `json:"eye.registration.application.name"`
		// Enable this switch for applications that do not cache Eye
		// information
		NoLocalRedis bool `json:"no.local.redis,string"`
	} `json:"eyewall"`
	// Eye is the namespace for configuration options relating to
	// the eye configuration profile server
	Eye struct {
		QueueLen          int    `json:"handler.queue.length,string"`
		SomaURL           string `json:"soma.address"`
		SomaPrefix        string `json:"soma.default.feedback.path.prefix"`
		ConcurrencyLimit  uint32 `json:"post.request.concurrency.limit,string"`
		RetryCount        int    `json:"post.request.retry.count,string"`
		RetryMinWaitTime  int    `json:"retry.min.wait.time.ms,string"`
		RetryMaxWaitTime  int    `json:"retry.max.wait.time.ms,string"`
		RequestTimeout    int    `json:"request.timeout.ms,string"`
		AlarmEndpoint     string `json:"alarm.notification.endpoint"`
		AlarmContentType  string `json:"alarm.notification.content.type"`
		AlarmTemplateFile string `json:"alarm.notification.template.file"`
		Daemon            struct {
			URL    *url.URL `json:"-"`
			Listen string   `json:"listen"`
			Port   string   `json:"port"`
			TLS    bool     `json:"tls,string"`
			Cert   string   `json:"cert-file"`
			Key    string   `json:"key-file"`
		} `json:"daemon"`
	} `json:"eye"`
	// PostgreSQL is the namespace for configuration options relating to
	// connections to a pgSQL database
	PostgreSQL struct {
		Host    string `json:"host" valid:"dns"`
		User    string `json:"user" valid:"alphanum"`
		Name    string `json:"name" valid:"alphanum"`
		Port    string `json:"port" valid:"port"`
		Pass    string `json:"password" valid:"-"`
		Timeout string `json:"timeout" valid:"numeric"`
		TLSMode string `json:"tlsmode" valid:"alpha"`
	} `json:"postgresql"`
	// Geocontrol is the namespace for configuration options relating
	// to Zookeeper-based inter-cluster synchronization
	Geocontrol struct {
		// Connect string for the Zookeeper-Ensemble to use
		Connect string `json:"connect.string"`
		// Name of the tool this is an instance of
		ToolName string `json:"tool.name"`
		// Name of the managed tool instance
		InstanceName string `json:"instance.name"`
		// PublicKey for the tool instance
		InstancePubKey string `json:"instance.public.key"`
		// PrivateKey for the tool instance
		InstancePrvKey string `json:"instance.private.key"`
		// Datacenter in which this runs
		Datacenter string `json:"datacenter"`
	} `json:"geocontrol"`
	// Stormchaser ...
	Stormchaser struct {
		HandlerQueueLength     int    `json:"handler.queue.length,string"`
		ConcurrencyLimit       uint32 `json:"request.concurrency.limit,string"`
		AlarmAPIDestinationURI string `json:"alarm.api.destination.uri"`
		Profiles               []struct {
			EyeHost                      string `json:"eye.host"`
			EyePort                      string `json:"eye.port"`
			TargetApplication            string `json:"target.application"`
			ExpectedRegistrations        int64  `json:"expected.application.registration.count,string"`
			ExpectedMetricProgress       int64  `json:"expected.metric.progress.per.check,string"`
			RegistrationAlertMissing     bool   `json:"alert.missing.registrations,string"`
			RegistrationAlertUnavailable bool   `json:"alert.unavailable.registrations,string"`
			HeartbeatAlertStale          bool   `json:"alert.stale.heartbeats,string"`
			MetricsAlertProgress         bool   `json:"alert.insufficient.metric.progress,string"`
		} `json:"profiles"`
	} `json:"stormchaser"`
	// BasicAuth contains static basic auth configuration
	BasicAuth struct {
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"basicauth"`
	// TLS contains the TLS configuration and settings
	TLS struct {
		CertificateChains []struct {
			ChainFile string `json:"certificate.chain.file"`
			KeyFile   string `json:"certificate.key.file"`
		} `json:"certificate.chains"`
		RootCAs    []string `json:"root.certificates"`
		MinVersion string   `json:"min.version"`
		MaxVersion string   `json:"max.version"`
		ServerName string   `json:"server.name"`
		Ciphers    string   `json:"cipher.style"`
	} `json:"tls"`
}

// FromFile sets Config c based on the file contents
func (c *Config) FromFile(fname string) error {
	var (
		file, uclJSON []byte
		err           error
		fileBytes     *bytes.Buffer
		parser        *ucl.Parser
		uclData       map[string]interface{}
	)
	if fname, err = filepath.Abs(fname); err != nil {
		return err
	}
	if fname, err = filepath.EvalSymlinks(fname); err != nil {
		return err
	}
	if file, err = ioutil.ReadFile(fname); err != nil {
		return err
	}

	fileBytes = bytes.NewBuffer(file)
	parser = ucl.NewParser(fileBytes)
	if uclData, err = parser.Ucl(); err != nil {
		return err
	}

	if uclJSON, err = json.Marshal(uclData); err != nil {
		return err
	}
	return json.Unmarshal(uclJSON, &c)
}

// Clone returns a copy of c
func (c *Config) Clone() *Config {
	clone := Config{}
	deepcopier.Copy(*c).To(clone)
	return &clone
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
