package bambulabs_api

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/torbenconto/bambulabs_api/internal/protocol"
)

type fakeCommandClient struct{}

func (fakeCommandClient) Send(context.Context, *protocol.Command) error {
	return nil
}

func newTestPrinter(tb testing.TB, model Model, reportFile string) *printer {
	tb.Helper()

	p := &printer{
		cfg: Config{
			Model: model,
		},
		AMS: &AMSSystem{},
	}

	p.decoder = *NewDecoder(model, fakeCommandClient{})

	if reportFile != "" {
		data, err := os.ReadFile(filepath.Join("fixtures", reportFile))
		if err != nil {
			tb.Fatal(err)
		}

		var report protocol.Report
		if err := json.Unmarshal(data, &report); err != nil {
			tb.Fatal(err)
		}

		p.decoder.Apply(p, &report)
	}

	return p
}
