package raft

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type KeyValue struct {
	Key   *string `json:"key" protobuf:"bytes,1,opt,name=key"`
	Value *string `json:"value" protobuf:"bytes,2,opt,name=value"`
}

func SetValueForCoordinator(key, value string) error {
	dnsName := fmt.Sprintf("%s.%s.svc", os.Getenv("PRIMARY_HOST"), os.Getenv("NAMESPACE"))
	subPath := "/set"
	url := "http://" + dnsName + ":" + strconv.Itoa(PostgresCoordinatorClientPort) + subPath
	keyValue := &KeyValue{
		Key:   &key,
		Value: &value,
	}

	requestByte, err := json.Marshal(keyValue)
	if err != nil {
		return err
	}
	requestBody := bytes.NewReader(requestByte)
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPost, url, requestBody)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Failed on update key %s with value %s", key, value))
	}
	user := os.Getenv("POSTGRES_USER")
	pass := os.Getenv("POSTGRES_PASSWORD")
	req.Header.Add("Content-Type", "application/json; charset=UTF-8")
	req.SetBasicAuth(user, pass)
	client.Timeout = 3 * time.Second
	resp, err := client.Do(req)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Failed on update key %s with value %s", key, value))
	}
	defer resp.Body.Close()
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Failed on update key %s with value %s", key, value))
	}
	return nil
}

func GetValueForCoordinator(key string) (string, error) {
	//key := api.PostgresPgCoordinatorStatus
	dnsName := fmt.Sprintf("%s.%s.svc", os.Getenv("PRIMARY_HOST"), os.Getenv("NAMESPACE"))
	subPath := "/get"
	url := "http://" + dnsName + ":" + strconv.Itoa(PostgresCoordinatorClientPort) + subPath
	keyValue := &KeyValue{
		Key: &key,
	}

	requestByte, err := json.Marshal(keyValue)
	if err != nil {
		return "", err
	}
	requestBody := bytes.NewReader(requestByte)
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodGet, url, requestBody)
	if err != nil {
		return "", errors.Wrap(err, "Failed on getting value from pg-coordinator")
	}
	user := os.Getenv("POSTGRES_USER")
	pass := os.Getenv("POSTGRES_PASSWORD")
	req.Header.Add("Content-Type", "application/json; charset=UTF-8")
	req.SetBasicAuth(user, pass)
	client.Timeout = 3 * time.Second
	resp, err := client.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "Failed on getting value from pg-coordinator")
	}
	defer resp.Body.Close()
	bodyText, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "Failed on getting value from pg-coordinator")
	}
	var responseKeyValue KeyValue
	err = json.Unmarshal(bodyText, &responseKeyValue)
	if err != nil {
		if strings.Contains(string(bodyText), "key not found") {
			return "", nil
		}
		return "", err
	}
	return *responseKeyValue.Value, nil
}
