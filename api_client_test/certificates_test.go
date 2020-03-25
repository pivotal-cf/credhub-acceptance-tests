package acceptance_test

import (
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"code.cloudfoundry.org/credhub-cli/credhub"
	"code.cloudfoundry.org/credhub-cli/credhub/credentials"
	"code.cloudfoundry.org/credhub-cli/credhub/credentials/generate"
	"code.cloudfoundry.org/credhub-cli/credhub/credentials/values"
	"github.com/cloudfoundry-incubator/credhub-acceptance-tests/test_helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Certificates", func() {
	Describe("getting certificate metadata", func() {
		It("gets certificate metadata", func() {
			name := testCredentialPath(time.Now().UnixNano(), "some-certificate")

			generateCert := generate.Certificate{
				CommonName: "example.com",
				SelfSign:   true,
			}
			certificate, err := credhubClient.GenerateCertificate(name, generateCert, credhub.Overwrite)
			Expect(err).ToNot(HaveOccurred())
			Expect(certificate.Value.Certificate).ToNot(BeEmpty())
			Expect(certificate.Value.PrivateKey).ToNot(BeEmpty())

			queryParams := url.Values{}
			queryParams.Add("name", certificate.Name)
			data, err := credhubClient.Request(http.MethodGet, "/api/v1/certificates/", queryParams, nil, true)
			Expect(err).ToNot(HaveOccurred())

			dec := json.NewDecoder(data.Body)
			response := make(map[string][]credentials.CertificateMetadata)

			err = dec.Decode(&response)
			Expect(err).ToNot(HaveOccurred())

			metadataArray, _ := response["certificates"]
			expected := metadataArray[0]

			actual, err := credhubClient.GetAllCertificatesMetadata()
			Expect(err).ToNot(HaveOccurred())

			Expect(actual).To(ContainElement(expected))

		})
		It("properly returns self_signed and is_ca", func() {
			name := testCredentialPath(time.Now().UnixNano(), "some-intermediate-ca")

			setCertificate := values.Certificate{
				Certificate: test_helpers.INTERMEDIATE_CA,
				PrivateKey:  test_helpers.INTERMEDIATE_CA_PRIVATE_KEY,
			}
			_, err := credhubClient.SetCertificate(name, setCertificate)
			Expect(err).ToNot(HaveOccurred())

			queryParams := url.Values{}
			queryParams.Add("name", name)
			data, err := credhubClient.Request(http.MethodGet, "/api/v1/certificates/", queryParams, nil, true)
			Expect(err).ToNot(HaveOccurred())

			dec := json.NewDecoder(data.Body)
			response := make(map[string][]credentials.CertificateMetadata)

			err = dec.Decode(&response)
			Expect(err).ToNot(HaveOccurred())

			metadataArray, _ := response["certificates"]
			actual := metadataArray[0]

			Expect(actual.Versions[0].SelfSigned).To(BeFalse())
			Expect(actual.Versions[0].CertificateAuthority).To(BeTrue())
		})
	})
})
