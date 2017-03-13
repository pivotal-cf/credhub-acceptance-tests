package api_integration_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/cloudfoundry-incubator/credhub-acceptance-tests/test_helpers"
	"crypto/tls"
	"log"
	"crypto/x509"
	"io/ioutil"
	"net/http"
	"bytes"
	"encoding/json"
	"testing"
	"os"
	"path"
)
var (
	config Config
	err error
)

var _ = Describe("mutual TLS authentication", func() {

	Describe("with a certificate signed by a trusted CA	", func() {
		Describe("when the certificate has a valid date range", func() {

			BeforeEach(func() {
				config, err = LoadConfig()
				Expect(err).NotTo(HaveOccurred())
			})

			It("allows the client to hit an authenticated endpoint", func() {
				postData := map[string]string{
					"name": "mtlstest",
					"type": "password",
				}
				result, err := mtlsPost(
					config.ApiUrl + "/api/v1/data",
					postData,
					"ca.pem",
					"client.pem",
					"client_key.pem")

				Expect(err).To(BeNil())
				Expect(result).To(MatchRegexp(`"type":"password"`))
			})
		})

		Describe("when the certificate is signed by wrong CA", func() {
			BeforeEach(func() {
				config, err = LoadConfig()
				Expect(err).NotTo(HaveOccurred())
			})

			It("prevents the client from hitting an authenticated endpoint", func() {
				postData := map[string]string{
					"name": "mtlstest",
					"type": "password",
				}
				result, err := mtlsPost(
					config.ApiUrl + "/api/v1/data",
					postData,
					"ca.pem",
					"expired.pem",
					"expired_key.pem")

				Expect(err).ToNot(BeNil())
				Expect(result).To(BeEmpty())
			})
		})
	})
})

func TestMTLS(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "mTLS Test Suite")
}

func handleError(err error) {
	if err != nil {
		log.Fatal("Fatal", err)
	}
}

func mtlsPost(url string, postData map[string]string, serverCaFilename, clientCertFilename, clientKeyPath string) (string, error) {
	err, client := createMtlsClient(serverCaFilename, clientCertFilename, clientKeyPath)

	jsonValue, _ := json.Marshal(postData)

	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func createMtlsClient(serverCaFilename, clientCertFilename, clientKeyFilename string) (error, *http.Client) {
	serverCaPath := path.Join(config.CredentialRoot, serverCaFilename)
	clientCertPath := path.Join(os.Getenv("PWD"), "certs", clientCertFilename)
	clientKeyPath := path.Join(os.Getenv("PWD"), "certs", clientKeyFilename)
	_, err := os.Stat(serverCaPath)
	handleError(err)
	_, err = os.Stat(clientCertPath)
	handleError(err)
	_, err = os.Stat(clientKeyPath)
	handleError(err)

	clientCertificate, err := tls.LoadX509KeyPair(clientCertPath, clientKeyPath)
	handleError(err)

	trustedCAs := x509.NewCertPool()
	serverCA, err := ioutil.ReadFile(serverCaPath)

	ok := trustedCAs.AppendCertsFromPEM([]byte(serverCA))
	if !ok {
		log.Fatal("failed to parse root certificate")
	}

	tlsConf := &tls.Config{
		Certificates: []tls.Certificate{clientCertificate},
		RootCAs:      trustedCAs,
	}

	transport := &http.Transport{TLSClientConfig: tlsConf}
	client := &http.Client{Transport: transport}

	return err, client
}
