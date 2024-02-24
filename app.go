package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/geoffjay/crm/config"
	"github.com/geoffjay/crm/handlers"
	"github.com/geoffjay/crm/repository"
	"github.com/geoffjay/crm/util"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/session"
	log "github.com/sirupsen/logrus"
)

type app struct{}

func (s *app) init() {
	log.WithFields(log.Fields{
		"service": "crm",
		"context": "crm.init",
	}).Debug("initializing")

	// TODO: remove this once there's a database.
	repository.Initialize()
}

func (s *app) run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	log.WithFields(log.Fields{
		"service": "crm",
		"context": "crm.run",
	}).Debug("starting")

	wg.Add(1)
	go s.runApp(ctx, wg)

	<-ctx.Done()

	log.WithFields(log.Fields{
		"service": "crm",
		"context": "crm.run",
	}).Debug("exiting")
}

func (s *app) runApp(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	conf := config.GetConfig()

	fields := log.Fields{"service": "crm", "context": "crm.run-app"}
	bindAddress := util.Getenv("CRM_BIND_ADDRESS", "127.0.0.1")
	bindPort, err := strconv.Atoi(util.Getenv("CRM_BIND_PORT", "8443"))
	if err != nil {
		log.WithFields(fields).Fatal(err)
	}

	log.WithFields(fields).Debug("starting server")

	go func() {
		app := fiber.New()

		handlers.SessionStore = session.New(conf.Session.ToSessionConfig())

		app.Use(helmet.New())
		app.Use(cors.New(conf.Cors.ToCorsConfig()))
		app.Use(logger.New())
		app.Use(recover.New())
		app.Use(etag.New())
		app.Use(limiter.New(limiter.Config{
			Expiration: 30 * time.Second,
			Max:        50,
		}))

		initRouter(app)

		cert := initializeCert()
		tlsConfig := &tls.Config{Certificates: []tls.Certificate{cert}}
		address := fmt.Sprintf("%s:%d", bindAddress, bindPort)

		ln, err := tls.Listen("tcp", address, tlsConfig)
		if err != nil {
			panic(err)
		}

		log.WithFields(fields).Fatal(app.Listener(ln))
	}()

	<-ctx.Done()

	log.WithFields(fields).Debug("exiting server")
}

func initializeCert() tls.Certificate {
	conf := config.GetConfig()
	fields := log.Fields{"service": "crm", "context": "crm.init-cert"}

	certFile := util.Getenv("CRM_TLS_CERT", "cert/app-cert.pem")
	keyFile := util.Getenv("CRM_TLS_KEY", "cert/app-key.pem")

	if conf.Env == "development" || conf.Env == "test" {
		if _, err := os.Stat(certFile); os.IsNotExist(err) {
			log.WithFields(fields).Info("Self-signed certificate not found, generating...")
			if err := generateSelfSignedCert(certFile, keyFile); err != nil {
				log.WithFields(fields).Fatal(err)
			}
			log.WithFields(fields).Info("Self-signed certificate generated successfully")
			log.WithFields(fields).Info("You will need to accept the self-signed certificate in your browser")
		}
	}

	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		log.WithFields(fields).Fatal(err)
	}

	return cert
}

// generateSelfSignedCert generates a self-signed certificate and key
// and saves them to the specified files
//
// This is only for testing purposes and should not be used in production.
func generateSelfSignedCert(certFile string, keyFile string) error {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"CRM Org"},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(time.Hour * 24 * 180),

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return err
	}

	certOut, err := os.Create(certFile)
	if err != nil {
		return err
	}
	defer certOut.Close()

	_ = pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})

	keyOut, err := os.Create(keyFile)
	if err != nil {
		return err
	}
	defer keyOut.Close()

	_ = pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})

	return nil
}
