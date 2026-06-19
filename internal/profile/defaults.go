package profile

func ApplyDefaults(p *Profile) {
	if p.APIVersion == "" {
		p.APIVersion = "configurator.pucora.io/v1"
	}
	if p.Kind == "" {
		p.Kind = "GatewayProfile"
	}
	if p.Metadata.Name == "" {
		p.Metadata.Name = "velonetics-gateway"
	}
	if p.Gateway.Port == 0 {
		p.Gateway.Port = 8080
	}
	if p.Gateway.Timeout == "" {
		p.Gateway.Timeout = "3s"
	}
	if p.Gateway.CacheTTL == "" {
		p.Gateway.CacheTTL = "3600s"
	}

	if p.CORS != nil && p.CORS.Enabled {
		if len(p.CORS.AllowMethods) == 0 {
			p.CORS.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
		}
		if len(p.CORS.AllowHeaders) == 0 {
			p.CORS.AllowHeaders = []string{"Origin", "Authorization", "Content-Type", "Accept"}
		}
		if p.CORS.MaxAge == "" {
			p.CORS.MaxAge = "12h"
		}
	}

	if p.Telemetry == nil {
		p.Telemetry = &Telemetry{}
	}
	if p.Telemetry.Logging == nil {
		p.Telemetry.Logging = &Logging{Level: "INFO", Stdout: true}
	}
	if p.Telemetry.Logging.Level == "" {
		p.Telemetry.Logging.Level = "INFO"
	}
	if p.Telemetry.Metrics == nil {
		p.Telemetry.Metrics = &Metrics{Enabled: true, ListenAddress: ":8090"}
	}
	if p.Telemetry.Metrics.ListenAddress == "" {
		p.Telemetry.Metrics.ListenAddress = ":8090"
	}
	if p.Telemetry.Usage == nil {
		p.Telemetry.Usage = &Usage{Enabled: false}
	}

	for i := range p.Routes {
		r := &p.Routes[i]
		if r.Method == "" {
			r.Method = "GET"
		}
		if r.OutputEncoding == "" {
			r.OutputEncoding = "json"
		}
		if r.Backend.Type == "" {
			r.Backend.Type = "http"
		}
		if r.Backend.Path == "" {
			r.Backend.Path = "/"
		}
		if r.Headers != nil && len(r.Headers.Forward) == 0 {
			r.Headers = nil
		}
		if r.QueryStrings != nil && len(r.QueryStrings.Forward) == 0 {
			r.QueryStrings = nil
		}
	}
}
