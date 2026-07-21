package proxy

// ProxyService is responsible for routing traffic to containers
type ProxyService interface {
	AssignPort() error
	ReverseProxy() error
	CustomDomain() error
	HTTPS() error
	SSL() error
}
