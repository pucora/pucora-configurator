package profile

var validBackendTypes = map[string]bool{
	"http": true, "grpc": true, "websocket": true,
	"kafka": true, "nats": true, "rabbit": true,
	"gcp": true, "aws_sns": true, "aws_sqs": true, "azure": true,
	"graphql": true, "soap": true, "lambda": true,
}

var validMethods = map[string]bool{
	"GET": true, "POST": true, "PUT": true, "PATCH": true,
	"DELETE": true, "HEAD": true, "OPTIONS": true,
}

func Validate(p *Profile) error {
	if errs := ValidateStructured(p); errs.HasErrors() {
		return errs
	}
	return nil
}
