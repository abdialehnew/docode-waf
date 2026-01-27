package services

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"log"
	"sync"

	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"
)

// MyUser represents a user for Let's Encrypt
type MyUser struct {
	Email        string
	Registration *registration.Resource
	key          crypto.PrivateKey
}

func (u *MyUser) GetEmail() string {
	return u.Email
}
func (u *MyUser) GetRegistration() *registration.Resource {
	return u.Registration
}
func (u *MyUser) GetPrivateKey() crypto.PrivateKey {
	return u.key
}

// InMemoryHTTPProvider implements http01.Provider
type InMemoryHTTPProvider struct {
	tokens map[string]string
	mu     sync.RWMutex
}

func NewInMemoryHTTPProvider() *InMemoryHTTPProvider {
	return &InMemoryHTTPProvider{
		tokens: make(map[string]string),
	}
}

func (p *InMemoryHTTPProvider) Present(domain, token, keyAuth string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.tokens[token] = keyAuth
	log.Printf("[ACME] Presenting challenge for domain %s with token %s", domain, token)
	return nil
}

func (p *InMemoryHTTPProvider) CleanUp(domain, token, keyAuth string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.tokens, token)
	log.Printf("[ACME] Cleaning up challenge for domain %s", domain)
	return nil
}

// GetKeyAuth returns the key auth for a given token
func (p *InMemoryHTTPProvider) GetKeyAuth(token string) (string, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	val, ok := p.tokens[token]
	return val, ok
}

type ACMEService struct {
	httpProvider *InMemoryHTTPProvider
}

func NewACMEService() *ACMEService {
	return &ACMEService{
		httpProvider: NewInMemoryHTTPProvider(),
	}
}

func (s *ACMEService) GetHTTPProvider() *InMemoryHTTPProvider {
	return s.httpProvider
}

func (s *ACMEService) ObtainCertificate(domain, email string) (*certificate.Resource, error) {
	// Create a user. In a real application, you'd want to persist the user/private key.
	// For now, we generate a new one for each request (ephemeral user).
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	user := &MyUser{
		Email: email,
		key:   privateKey,
	}

	config := lego.NewConfig(user)

	// Use Staging CA for testing/dev, Production for prod.
	// TODO: Make this configurable
	config.CADirURL = lego.LEDirectoryProduction
	config.Certificate.KeyType = certcrypto.EC256

	client, err := lego.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create lego client: %w", err)
	}

	// Register
	reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		return nil, fmt.Errorf("failed to register user: %w", err)
	}
	user.Registration = reg

	// Set HTTP provider
	err = client.Challenge.SetHTTP01Provider(s.httpProvider)
	if err != nil {
		return nil, fmt.Errorf("failed to set HTTP-01 provider: %w", err)
	}

	request := certificate.ObtainRequest{
		Domains: []string{domain},
		Bundle:  true,
	}

	log.Printf("[ACME] Requesting certificate for %s", domain)
	certificates, err := client.Certificate.Obtain(request)
	if err != nil {
		return nil, fmt.Errorf("failed to obtain certificate: %w", err)
	}
	log.Printf("[ACME] Successfully obtained certificate for %s", domain)

	return certificates, nil
}
