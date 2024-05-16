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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	verifier "go.bytebuilders.dev/license-verifier"
	"go.bytebuilders.dev/license-verifier/apis/licenses/v1alpha1"
	"go.bytebuilders.dev/license-verifier/info"
	"go.bytebuilders.dev/license-verifier/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
)

const (
	natsConnectionTimeout       = 350 * time.Millisecond
	natsConnectionRetryInterval = 100 * time.Millisecond
	natsEventPublishTimeout     = 10 * time.Second
	natsRequestTimeout          = 2 * time.Second
)

type NatsConfig struct {
	// LicenseID string     `json:"licenseID"`
	Subject string     `json:"natsSubject"`
	Server  string     `json:"natsServer"`
	Client  *nats.Conn `json:"-"`
}

// NatsCredential represents the api response of the register licensed user api
type NatsCredential struct {
	NatsConfig `json:",inline,omitempty"`
	Credential []byte `json:"credential"`
}

type LicenseIDGetter interface {
	GetLicenseID() string
}

type LicenseUpdater struct {
	le      *kubernetes.LicenseEnforcer
	License v1alpha1.License
	mu      sync.Mutex
}

func (lu *LicenseUpdater) GetLicenseID() string {
	lu.mu.Lock()
	defer lu.mu.Unlock()

	l := lu.License
	if l.Status == v1alpha1.LicenseActive && time.Now().After(l.NotAfter.Time) {
		license, _ := lu.le.LoadLicense()
		lu.License = license
		l = license
	}
	return l.ID
}

func NewNatsConfig(cfg *rest.Config, clusterID string, LicenseFile string) (*NatsConfig, LicenseIDGetter, error) {
	le, err := kubernetes.NewLicenseEnforcer(cfg, LicenseFile)
	if err != nil {
		return nil, nil, err
	}
	license, licenseBytes := le.LoadLicense()
	if license.Status != v1alpha1.LicenseActive {
		return nil, nil, fmt.Errorf("license status is %s", license.Status)
	}

	opts := verifier.Options{
		ClusterUID: clusterID,
		Features:   info.ProductName,
		CACert:     []byte(info.LicenseCA),
		License:    licenseBytes,
	}
	data, err := json.Marshal(opts)
	if err != nil {
		return nil, nil, err
	}

	resp, err := http.Post(info.MustRegistrationAPIEndpoint(), "application/json", bytes.NewReader(data))
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, nil, errors.New(resp.Status + ", " + string(body))
	}

	var natscred NatsCredential
	err = json.Unmarshal(body, &natscred)
	if err != nil {
		return nil, nil, err
	}

	klog.V(5).InfoS("using event receiver", "address", natscred.Server, "subject", natscred.Subject, "licenseID", license.ID)

	natscred.Client, err = NewConnection(license.ID, natscred)
	if err != nil {
		return nil, nil, err
	}

	return &natscred.NatsConfig, &LicenseUpdater{le: le, License: license}, nil
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
