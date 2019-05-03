package config

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dollarshaveclub/pvc"
	"github.com/pkg/errors"
)

type ServerConfig struct {
	HTTPSPort                  uint
	HTTPSAddr                  string
	DisableTLS                 bool
	TLSCert                    tls.Certificate
	WordnetPath                string
	FuranAddrs                 []string
	APIKeys                    []string
	ReaperIntervalSecs         uint
	EventRateLimitPerSecond    uint
	GlobalEnvironmentLimit     uint
	HostnameTemplate           string
	DatadogServiceName         string
	DebugEndpoints             bool
	DebugEndpointsIPWhitelists []string
	NitroFeatureFlag           bool
	NotificationsDefaultsJSON  string
}

type PGConfig struct {
	PostgresURI            string
	PostgresMigrationsPath string
	DatadogServiceName     string
	EnableTracing          bool
}

// K8sClientConfig models the configuration required for a kubernetes client
// to communicate with the API server
type K8sClientConfig struct {
	JWTPath string
}

// K8sSecret models a kubernetes secret
type K8sSecret struct {
	Data map[string][]byte `json:"data"`
	Type string            `json:"type"`
}

type K8sConfig struct {
	// GroupBindings is a map of k8s group name to cluster role
	GroupBindings map[string]string
	// PrivilegedRepoWhitelist is a list of GitHub repositories whose environment service accounts will be given cluster-admin privileges
	PrivilegedRepoWhitelist []string
	// SecretInjections is a map of secret name to value that will be injected into each environment namespace
	SecretInjections map[string]K8sSecret
	// Labels is a map used to store key/value pairs that will be attached to k8s objects that are created by Acyl itself.
	// Labels should not be empty and should contain a unique combination of labels. These labels can be used by Acyl
	// to remove orphaned resources.
	Labels map[string]string
}

// ProcessPrivilegedRepos takes a comma-separated list of repositories and populates the PrivilegedRepoWhitelist field
func (kc *K8sConfig) ProcessPrivilegedRepos(repostr string) error {
	kc.PrivilegedRepoWhitelist = strings.Split(repostr, ",")
	for i, pr := range kc.PrivilegedRepoWhitelist {
		if rsl := strings.Split(pr, "/"); len(rsl) != 2 {
			return fmt.Errorf("malformed repo at offset %v: %v", i, pr)
		}
	}
	return nil
}

// ProcessGroupBindings takes a comma-separated list of group bindings and populates the GroupBindings field
func (kc *K8sConfig) ProcessGroupBindings(gbstr string) error {
	kc.GroupBindings = make(map[string]string)
	for i, gb := range strings.Split(gbstr, ",") {
		if gb == "" {
			continue
		}
		gbsl := strings.Split(gb, "=")
		if len(gbsl) != 2 {
			return fmt.Errorf("malformed group binding at offset %v: %v", i, gb)
		}
		if len(gbsl[0]) == 0 || len(gbsl[1]) == 0 {
			return fmt.Errorf("empty binding at offset %v: %v", i, gb)
		}
		kc.GroupBindings[gbsl[0]] = gbsl[1]
	}
	return nil
}

// ProcessLabels takes a comma-separated list of labels and popultes the Labels field.
// We want to ensure that at least one label is provided, otherwise, all resources
// will be managed by Acyl and could be deleted during cleanup.
func (kc *K8sConfig) ProcessLabels(labelsStr string) error {
	kc.Labels = make(map[string]string)
	labels := strings.Split(labelsStr, ",")
	if len(labels) == 0 {
		return fmt.Errorf("at least one label should be provided")
	}
	for _, labelStr := range labels {
		keyValPair := strings.Split(labelStr, "=")
		if len(keyValPair) != 2 {
			return fmt.Errorf("malformed label %s in %s", labelStr, labelsStr)
		}
		key, value := keyValPair[0], keyValPair[1]
		kc.Labels[key] = value
	}
	return nil
}

// SecretFetcher describes an object that fetches secrets
type SecretFetcher interface {
	Get(id string) ([]byte, error)
}

// ProcessSecretInjections takes a comma-separated list of injections and uses sf to populate the SecretInjections field
func (kc *K8sConfig) ProcessSecretInjections(sf SecretFetcher, injstr string) error {
	kc.SecretInjections = make(map[string]K8sSecret)
	for i, sstr := range strings.Split(injstr, ",") {
		if sstr == "" {
			continue
		}
		ssl := strings.Split(sstr, "=")
		if len(ssl) != 2 {
			return fmt.Errorf("malformed secret injection at offset %v: %v", i, sstr)
		}
		if len(ssl[0]) == 0 || len(ssl[1]) == 0 {
			return fmt.Errorf("empty secret injection at offset %v: %v", i, sstr)
		}
		val, err := sf.Get(ssl[1])
		if err != nil {
			return errors.Wrapf(err, "error fetching secret for id: %v", ssl[1])
		}
		secret := K8sSecret{}
		if err := json.Unmarshal(val, &secret); err != nil {
			return errors.Wrapf(err, "error unmarshaling secret for id: %v", ssl[1])
		}
		kc.SecretInjections[ssl[0]] = secret
	}
	return nil
}

type AminoConfig struct {
	HelmChartToRepoRaw       string
	HelmChartToRepo          map[string]string
	AminoDeploymentToRepoRaw string
	AminoDeploymentToRepo    map[string]string
	AminoJobToRepoRaw        string
	AminoJobToRepo           map[string]string
}

func (a *AminoConfig) Parse() error {
	if err := json.Unmarshal([]byte(a.HelmChartToRepoRaw), &a.HelmChartToRepo); err != nil {
		return fmt.Errorf("error unmarshaling HelmChartToRepo: %v", err)
	}
	if err := json.Unmarshal([]byte(a.AminoDeploymentToRepoRaw), &a.AminoDeploymentToRepo); err != nil {
		return fmt.Errorf("error unmarshaling AminoDeploymentToRepo: %v", err)
	}
	if err := json.Unmarshal([]byte(a.AminoJobToRepoRaw), &a.AminoJobToRepo); err != nil {
		return fmt.Errorf("error unmarshaling AminoJobToRepo: %v", err)
	}
	return nil
}

type ConsulConfig struct {
	Addr       string
	LockPrefix string
}

type SlackConfig struct {
	Username                    string
	IconURL                     string
	Token                       string
	Channel                     string
	MapperRepo                  string
	MapperRepoRef               string
	MapperMapPath               string
	MapperUpdateIntervalSeconds uint
}

type AWSCreds struct {
	AccessKeyID     string
	SecretAccessKey string
}

type AWSConfig struct {
	Region     string
	MaxRetries uint
}

type S3Config struct {
	Region, Bucket, KeyPrefix string
}

type VaultConfig struct {
	Addr        string
	Token       string
	TokenAuth   bool
	K8sAuth     bool
	K8sJWTPath  string
	K8sAuthPath string
	K8sRole     string
	AppID       string
	UserIDPath  string
}

type GithubConfig struct {
	HookSecret string
	Token      string
	TypePath   string // relative path within repo to look for the QAType definition
}

type BackendConfig struct {
	AminoAddr string
}

type MigrateConfig struct {
	CheckPending  bool
	MetaDataTable string
}

// SecretsConfig contains configuration values for retrieving secrets
type SecretsConfig struct {
	Backend pvc.SecretsClientOption
	Mapping string
}
