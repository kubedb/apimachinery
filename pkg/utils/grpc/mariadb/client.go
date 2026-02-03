/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Free Trial License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Free-Trial-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package mariadb

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"time"

	"kubedb.dev/apimachinery/apis/kubedb"
	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1"
	pb "kubedb.dev/apimachinery/pkg/utils/grpc/mariadb/protogen"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	port      = "50051"
	rootCAKey = "ca.crt"
)

func Client(kbClient client.Client, db *dbapi.MariaDB, podName string) (*grpc.ClientConn, error) {
	serverAddr := serverAddress(db, podName)

	interceptor, err := UnaryClientInterceptor(kbClient, db)
	if err != nil {
		return nil, err
	}

	opts := []grpc.DialOption{
		grpc.WithUnaryInterceptor(interceptor),
	}

	if sslEnabledMariaDB(db) {
		clientSecret := corev1.Secret{}
		err = kbClient.Get(context.TODO(), types.NamespacedName{
			Namespace: db.Namespace,
			Name:      db.GetCertSecretName(dbapi.MariaDBClientCert),
		}, &clientSecret)
		if err != nil {
			return nil, err
		}

		serverCa, exists := clientSecret.Data[rootCAKey]
		if !exists {
			return nil, fmt.Errorf("%v in not present in client secret", rootCAKey)
		}

		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(serverCa) {
			return nil, fmt.Errorf("failed to add server CA certificate")
		}

		tlsConfig := &tls.Config{
			RootCAs: caCertPool,
		}

		tlsCreds := credentials.NewTLS(tlsConfig)
		opts = append(opts, grpc.WithTransportCredentials(tlsCreds))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	conn, err := grpc.NewClient(serverAddr, opts...)
	if err != nil {
		return nil, fmt.Errorf("did not connect to %s: %w", serverAddr, err)
	}

	return conn, nil
}

func RunCommand(grpcClient pb.CommandServiceClient, cmd string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()

	req := &pb.CommandRequest{
		Command: cmd,
	}

	resp, err := grpcClient.ExecuteCommand(ctx, req)
	if err != nil {
		klog.Infof("failed to execute command: %v", err)
		return nil, err
	}

	if resp.Status != "success" {
		return nil, fmt.Errorf("failed to execute command: %s, Output: %s, err: %v", cmd, string(resp.Output), resp.Error)
	}

	return resp.Output, nil
}

func RunCommandWithPayload(ctx context.Context, grpcClient pb.CommandServiceClient, cmd string, data []byte) ([]byte, error) {
	req := &pb.CommandRequest{
		Command: cmd,
		Data:    data,
	}

	resp, err := grpcClient.ExecuteCommand(ctx, req)
	if err != nil {
		klog.Infof("failed to execute command: %v", err)
		return nil, err
	}

	if resp.Status != "success" {
		return nil, fmt.Errorf("failed to execute command: %s, Output: %s, err: %v", cmd, string(resp.Output), resp.Error)
	}

	return resp.Output, nil
}

func serverAddress(db *dbapi.MariaDB, podName string) string {
	if db.Spec.Distributed {
		return fmt.Sprintf("%s.%s.%s.svc%s:%v", podName, db.GoverningServiceName(), db.Namespace, kubedb.KubeSliceDomainSuffix, port)
	} else {
		return fmt.Sprintf("%s.%s.%s.svc:%v", podName, db.GoverningServiceName(), db.Namespace, port)
	}
}

func UnaryClientInterceptor(kbClient client.Client, db *dbapi.MariaDB) (grpc.UnaryClientInterceptor, error) {
	username, password, err := getMariaDBBasicAuth(kbClient, db)
	if err != nil {
		return nil, err
	}

	if username == "" || password == "" {
		return nil, fmt.Errorf("MYSQL_ROOT_USERNAME and MYSQL_ROOT_PASSWORD environment variables must be set for the client")
	}

	return func(
		ctx context.Context,
		method string,
		req, reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		return invoker(attachCredentials(ctx, username, password), method, req, reply, cc, opts...)
	}, nil
}

func attachCredentials(ctx context.Context, username, password string) context.Context {
	auth := username + ":" + password
	encodedAuth := base64.StdEncoding.EncodeToString([]byte(auth))
	return metadata.AppendToOutgoingContext(ctx, "authorization", "Basic "+encodedAuth)
}

func sslEnabledMariaDB(md *dbapi.MariaDB) bool {
	return md.Spec.TLS != nil && md.Spec.RequireSSL
}

func getMariaDBBasicAuth(kbClient client.Client, db *dbapi.MariaDB) (string, string, error) {
	var secretName string
	if db.Spec.AuthSecret != nil {
		secretName = db.GetAuthSecretName()
	}

	secret := corev1.Secret{}
	err := kbClient.Get(context.Background(), client.ObjectKey{Namespace: db.Namespace, Name: secretName}, &secret)
	if err != nil {
		return "", "", err
	}

	user, ok := secret.Data[corev1.BasicAuthUsernameKey]
	if !ok {
		return "", "", fmt.Errorf("DB root user is not set")
	}

	pass, ok := secret.Data[corev1.BasicAuthPasswordKey]
	if !ok {
		return "", "", fmt.Errorf("DB root password is not set")
	}

	return string(user), string(pass), nil
}
