/*
Copyright AppsCode Inc. and Contributors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package lib

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"go.bytebuilders.dev/license-verifier/apis/licenses/v1alpha1"
	"go.bytebuilders.dev/license-verifier/info"
	"go.bytebuilders.dev/license-verifier/kubernetes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	identityapi "kmodules.xyz/resource-metadata/apis/identity/v1alpha1"
	identityclient "kmodules.xyz/resource-metadata/client/clientset/versioned/typed/identity/v1alpha1"
	identitylib "kmodules.xyz/resource-metadata/pkg/identity"
)

const (
	natsConnectionTimeout       = 350 * time.Millisecond
	natsConnectionRetryInterval = 100 * time.Millisecond
	natsEventPublishTimeout     = 10 * time.Second
	natsRequestTimeout          = 2 * time.Second
)

type NatsConfig struct {
	Subject string `json:"natsSubject"`
	Server  string `json:"natsServer"`
}

// NatsCredential represents the api response of the register licensed user api
type NatsCredential struct {
	NatsConfig `json:",inline"`
	Credential []byte `json:"credential"`
}

type NatsClient struct {
	cfg         *rest.Config
	clusterID   string
	LicenseFile string

	le *kubernetes.LicenseEnforcer
	l  *v1alpha1.License

	nc      *nats.Conn
	Subject string
	Server  string
	mu      sync.Mutex
}

func NewNatsClient(cfg *rest.Config, clusterID string, LicenseFile string) *NatsClient {
	return &NatsClient{
		cfg:         cfg,
		clusterID:   clusterID,
		LicenseFile: LicenseFile,
	}
}

func (c *NatsClient) Request(data []byte, timeout time.Duration) (*nats.Msg, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var justConnected bool
	if c.nc == nil {
		if err := c.connect(); err != nil {
			return nil, err
		}
		justConnected = true
	}
	msg, err := c.nc.Request(c.Subject, data, timeout)
	if err != nil && !justConnected && isNatsAuthError(err.Error()) {
		if err := c.connect(); err != nil {
			return nil, err
		}
		msg, err = c.nc.Request(c.Subject, data, timeout)
	}
	return msg, err
}

// src: https://github.com/nats-io/nats.go/blob/main/nats.go#L3693-L3709
func isNatsAuthError(e string) bool {
	if strings.HasPrefix(e, nats.AUTHORIZATION_ERR) {
		return true
	}
	if strings.HasPrefix(e, nats.AUTHENTICATION_EXPIRED_ERR) {
		return true
	}
	if strings.HasPrefix(e, nats.AUTHENTICATION_REVOKED_ERR) {
		return true
	}
	if strings.HasPrefix(e, nats.ACCOUNT_AUTHENTICATION_EXPIRED_ERR) {
		return true
	}
	return false
}

func (c *NatsClient) GetLicenseID() (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.l == nil {
		if err := c.connect(); err != nil {
			return "", err
		}
	}

	if c.l.Status == v1alpha1.LicenseActive && time.Now().After(c.l.NotAfter.Time) {
		license, _ := c.le.LoadLicense()
		c.l = &license
	}
	return c.l.ID, nil
}

func (c *NatsClient) Connect() (*nats.Conn, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.nc == nil {
		if err := c.connect(); err != nil {
			return nil, err
		}
	}
	return c.nc, nil
}

func (c *NatsClient) connect() error {
	le, err := kubernetes.NewLicenseEnforcer(c.cfg, c.LicenseFile)
	if err != nil {
		return err
	}
	license, licenseBytes := le.LoadLicense()
	if license.Status != v1alpha1.LicenseActive {
		return fmt.Errorf("license status is %s", license.Status)
	}

	natscred, err := c.fetchNatsCredential(licenseBytes)
	if err != nil {
		return err
	}

	klog.V(5).InfoS("using event receiver", "address", natscred.Server, "subject", natscred.Subject, "licenseID", license.ID)

	nc, err := NewConnection(license.ID, *natscred)
	if err != nil {
		return err
	}

	c.le = le
	c.l = &license
	c.nc = nc
	c.Subject = natscred.Subject
	c.Server = natscred.Server
	return nil
}

// fetchNatsCredential obtains a NATS credential by first calling the public
// appscode.com Register endpoint. If that call fails because the cluster
// cannot reach appscode.com (DNS failure, connection refused, timeout, etc.),
// it falls back to the in-cluster identity.k8s.appscode.com extended API.
func (c *NatsClient) fetchNatsCredential(licenseBytes []byte) (*NatsCredential, error) {
	natscred, err := registerWithAppsCode(c.clusterID, licenseBytes)
	if err == nil {
		return natscred, nil
	}
	if !isNoConnectivityErr(err) {
		return nil, err
	}
	klog.V(5).InfoS("appscode.com unreachable; falling back to extended API",
		"error", err.Error())
	return registerViaExtendedAPI(c.cfg, licenseBytes)
}

func registerWithAppsCode(clusterID string, licenseBytes []byte) (*NatsCredential, error) {
	resp, err := identitylib.NewDefaultClient().GetAuditTokenForCluster(clusterID, info.ProductName, licenseBytes)
	if err != nil {
		return nil, err
	}
	return &NatsCredential{
		NatsConfig: NatsConfig{
			Subject: resp.Subject,
			Server:  resp.Server,
		},
		Credential: resp.Credential,
	}, nil
}

func registerViaExtendedAPI(cfg *rest.Config, licenseBytes []byte) (*NatsCredential, error) {
	ic, err := identityclient.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	req := &identityapi.AuditTokenRequest{
		Request: &identityapi.AuditTokenRequestRequest{
			Features: info.ProductName,
			License:  licenseBytes,
		},
	}
	result, err := ic.AuditTokenRequests().Create(context.TODO(), req, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}
	if result.Response == nil {
		return nil, errors.New("extended api returned empty AuditTokenRequest response")
	}
	return &NatsCredential{
		NatsConfig: NatsConfig{
			Subject: result.Response.Subject,
			Server:  result.Response.Server,
		},
		Credential: result.Response.Credential,
	}, nil
}

// isNoConnectivityErr reports whether the cluster cannot reach
// https://appscode.com at all (DNS lookup failure, connection refused,
// timeout, etc.). It actively probes the host with a short-timeout HEAD
// request rather than guessing from the registration error, so HTTP error
// responses from appscode.com (auth failures, 4xx, 5xx) don't accidentally
// trigger the fallback. The err argument is accepted for future use but
// currently ignored.
func isNoConnectivityErr(_ error) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, "https://appscode.com", nil)
	if err != nil {
		return false
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return true
	}
	resp.Body.Close() // nolint:errcheck
	return false
}

// NewConnection creates a new NATS connection
func NewConnection(licenseID string, natscred NatsCredential) (nc *nats.Conn, err error) {
	servers := natscred.Server

	opts := []nats.Option{
		nats.Name(fmt.Sprintf("%s.%s", licenseID, info.ProductName)),
		nats.MaxReconnects(-1),
		nats.ErrorHandler(errorHandler),
		nats.ReconnectHandler(reconnectHandler),
		nats.DisconnectErrHandler(disconnectHandler),
		// nats.UseOldRequestStyle(),
	}

	credFile := "/tmp/nats.creds"
	if err = os.WriteFile(credFile, natscred.Credential, 0o600); err != nil {
		return nil, err
	}

	opts = append(opts, nats.UserCredentials(credFile))

	//if os.Getenv("NATS_CERTIFICATE") != "" && os.Getenv("NATS_KEY") != "" {
	//	opts = append(opts, nats.ClientCert(os.Getenv("NATS_CERTIFICATE"), os.Getenv("NATS_KEY")))
	//}
	//
	//if os.Getenv("NATS_CA") != "" {
	//	opts = append(opts, nats.RootCAs(os.Getenv("NATS_CA")))
	//}

	// initial connections can error due to DNS lookups etc, just retry, eventually with backoff
	ctx, cancel := context.WithTimeout(context.Background(), natsConnectionTimeout)
	defer cancel()

	ticker := time.NewTicker(natsConnectionRetryInterval)
	for {
		select {
		case <-ticker.C:
			nc, err := nats.Connect(servers, opts...)
			if err == nil {
				return nc, nil
			}
			klog.V(5).InfoS("failed to connect to event receiver", "error", err)
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

// called during errors subscriptions etc
func errorHandler(nc *nats.Conn, s *nats.Subscription, err error) {
	if s != nil {
		klog.Warningf("error in event receiver connection: %s: subscription: %s: %s", nc.ConnectedUrl(), s.Subject, err)
		return
	}
	klog.Warningf("Error in event receiver connection: %s: %s", nc.ConnectedUrl(), err)
}

// called after reconnection
func reconnectHandler(nc *nats.Conn) {
	klog.Warningf("Reconnected to %s", nc.ConnectedUrl())
}

// called after disconnection
func disconnectHandler(nc *nats.Conn, err error) {
	if err != nil {
		klog.Warningf("Disconnected from event receiver due to error: %v", err)
	} else {
		klog.Warningln("Disconnected from event receiver")
	}
}
