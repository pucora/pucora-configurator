package compose

import (
	"os"
	"path/filepath"
)

const servicesReadme = `# Local development services

Copy mock service folders from Pucora CE examples into this directory:

| Service | Copy from |
|---------|-----------|
| mock-backend | pucora-ce/examples/websocket/mock-backend |
| mock-webhook | pucora-ce/examples/pubsub/async-kafka/mock-webhook |

Example:

` + "```bash" + `
cp -r ../pucora-ce/examples/websocket/mock-backend ./services/
cp -r ../pucora-ce/examples/pubsub/async-kafka/mock-webhook ./services/
` + "```" + `

Then run:

` + "```bash" + `
docker compose up --build
` + "```" + `
`

func WriteScaffold(outputDir string, req Requirements) error {
	if !req.MockBackend && !req.MockWebhook {
		return nil
	}
	servicesDir := filepath.Join(outputDir, "services")
	if err := os.MkdirAll(servicesDir, 0o755); err != nil {
		return err
	}
	readmePath := filepath.Join(servicesDir, "README.md")
	if _, err := os.Stat(readmePath); os.IsNotExist(err) {
		return os.WriteFile(readmePath, []byte(servicesReadme), 0o644)
	}
	return nil
}
