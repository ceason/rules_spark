package main

import (
	"fmt"
	"log"
	"os"
	"io/ioutil"
	"net/http"
	"encoding/json"
	"time"
	"path/filepath"
	"strings"
	"io"
	"crypto/x509"
	"crypto/tls"
)

type k8sClient struct {
	http.Client
	token string
	baseUrl string
}

// Automatically sets auth header
func (c *k8sClient) Do(req *http.Request) (*http.Response, error) {
	if req.Header.Get("Authorization") == "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	}
	return c.Client.Do(req)
}

// Automatically sets correct URL and auth header. Will panic if request is invalid.
func (c *k8sClient) mustNewRequest(method, url string, body io.Reader) *http.Request {
	req, err := http.NewRequest(method, c.baseUrl + url, body)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	return req
}

// Http client which automagically handles authorization
func newInClusterClient() *k8sClient {
	token, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token")
	if err != nil {
		panic(err)
	}
	caCert, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/ca.crt")
	if err != nil {
		log.Fatal(err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	// Setup HTTPS client
	tlsConfig := &tls.Config{
		RootCAs: caCertPool,
	}
	transport := &http.Transport{TLSClientConfig: tlsConfig}
	client := http.Client{Transport: transport}
	return &k8sClient{client, string(token), "https://kubernetes.default.svc"}
}

func withinClusterGetPodInfo() *PodInfo {
	hostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	namespace, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		panic(err)
	}
	client := newInClusterClient()
	req, err := http.NewRequest("GET", fmt.Sprintf("https://kubernetes.default.svc/api/v1/namespaces/%s/pods/%s", namespace, hostname), nil)
	if err != nil {
		panic(err)
	}
	podInfo := &PodInfo{}
PODINFO_REQ:
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(body, podInfo)
	if err != nil {
		panic(err)
	}
	if podInfo.Status.PodIP == "" {
		log.Printf("WARN: PodIP missing, waiting & trying again")
		time.Sleep(time.Second)
		goto PODINFO_REQ
	}
	return podInfo
}

func withinContainerGetMainClass() string {
	return environmentVars(os.Environ()).mustGet("MAIN_CLASS")
}

func withinContainerGetClasspath() string {
	classpath := ""
	classpathFile := ""
	// find the classpath file
	// todo: there has GOT to be a better way to do this :(
	err := filepath.Walk("/app", func(path string, f os.FileInfo, _ error) error {
		if ! f.IsDir() && strings.HasSuffix(path, ".classpath") {
			classpathFile = path
			return io.EOF // returning this error lets us stop the search once we've found our file. we'll nil out the error later
		}
		return nil
	})
	if err == io.EOF {
		err = nil
	}
	if err != nil {
		panic(fmt.Sprintf("Error searching for classpath file: %s", err))
	}
	log.Printf("INFO: Loading classpath from '%s'", classpathFile)
	data, err := ioutil.ReadFile(classpathFile)
	if err != nil {
		panic(err)
	}
	classpath = string(data)
	if classpath == "" {
		panic("Could not find classpath data!!! D:")
	}
	return fmt.Sprintf("/app:%s", classpath)
}
