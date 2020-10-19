package main

import (
        "bytes"
        cryptorand "crypto/rand"
        rsa "crypto/rsa"
        "crypto/x509"
        "encoding/pem"
        "io/ioutil"
        v1 "k8s.io/api/core/v1"
        meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
        scheme "k8s.io/client-go/kubernetes/scheme"
        rest "k8s.io/client-go/rest"
        "log"
        "os"
        pkcs12 "software.sslmate.com/src/go-pkcs12"
)

// secrets implements SecretInterface
type secrets struct {
        client rest.Interface
        ns     string
}

// newSecrets returns a Secrets
/*
func newSecrets(c *CoreV1Client, namespace string) *secrets {
        return &secrets{
                client: c.RESTClient(),
                ns:     namespace,
        }
}
*/
func (c *secrets) Get(name string, options meta_v1.GetOptions) (result *v1.Secret, err error) {
        result = &v1.Secret{}
        err = c.client.Get().
                Namespace(c.ns).
                Resource("secrets").
                Name(name).
                VersionedParams(&options, scheme.ParameterCodec).
                Do().
                Into(result)
        return
}

// Create takes the representation of a secret and creates it.  Returns the server's representation of the secret, and an error, if there is any.
func (c *secrets) Create(secret *v1.Secret) (result *v1.Secret, err error) {
        result = &v1.Secret{}
        err = c.client.Post().
                Namespace(c.ns).
                Resource("secrets").
                Body(secret).
                Do().
                Into(result)
        return
}
// ssh-keygetn -t rsa  # save it to priv.pem
// openssl req -new -x509 -key priv.pem -out cert.pem -days 1725
//   createPfx("/tmp/cert.pem", "/tmp/priv.pem")
// verify with openssl pkcs12 -in  /tmp/keycert.p12 -nokeys
func createPfx(certPath string, keyPath string) (pfxData []byte, err error) {
        // certPEM, err := Get("certificate")
        certPEM, err := ioutil.ReadFile(certPath)
        if err != nil {
                log.Printf("certificate not found")
                return []byte{}, err
        }
        // keyPEM, err := Get("key")
        keyPEM, err := ioutil.ReadFile(keyPath)
        if err != nil {
                log.Printf("key not found")
                return []byte{}, err
        }
        block, _ := pem.Decode([]byte(keyPEM))
        var key *rsa.PrivateKey
        if key, err = x509.ParsePKCS1PrivateKey(block.Bytes); err != nil {
                log.Printf("Failed to parse private key: " + err.Error())
                return []byte{}, err
        }
        block, _ = pem.Decode([]byte(certPEM))
        if block == nil {
                log.Printf("failed to parse certificate PEM")
                return []byte{}, err
        }
        certificate, err := x509.ParseCertificate(block.Bytes)
        if err != nil {
                log.Printf("failed to parse certificate: " + err.Error())
                return []byte{}, err
        }
        pfxPEM, err := pkcs12.Encode(cryptorand.Reader, key, certificate, []*x509.Certificate{}, "changeit")
        if err != nil {
                log.Printf("No pfx:" + err.Error())
                return []byte{}, err
        }
        log.Printf("pfx generated.")
        pfxPEM = bytes.Trim(pfxPEM, " \r\t\n\x00")
        // result, err := Create(pfxPEM)
        return pfxPEM, nil
}

func writePfx(pfxData []byte, filePath string) {
        const folderPath string = "/tmp"
        os.MkdirAll(folderPath, 0755)
        err := ioutil.WriteFile(filePath, pfxData, 0644)
        if err != nil {
                log.Printf("Store for internal ssl connectivity could not be written")
        }
}

func main() {
        pk, err := createPfx("/tmp/cert.pem", "/tmp/priv.pem")
        if err != nil {
                log.Printf("error:" + err.Error())
        }
        log.Printf("pk=%s", pk)
        if pk != nil {
                log.Printf("success")
        }
        writePfx(pk, "/tmp/keycert.p12")
}
