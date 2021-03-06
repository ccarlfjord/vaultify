package run

import (
	"context"

	"github.com/hashicorp/go-hclog"

	"github.com/ahilsend/vaultify/pkg/prometheus"
	"github.com/ahilsend/vaultify/pkg/secrets"
	"github.com/ahilsend/vaultify/pkg/template"
	"github.com/ahilsend/vaultify/pkg/vault"
	"github.com/ahilsend/vaultify/pkg/http"
)

func Run(logger hclog.Logger, options *Options) error {
	config := options.VaultApiConfig()
	vaultClient, err := vault.NewClient(logger, options.Role, config)
	if err != nil {
		return err
	}

	ctx := context.Background()

	prometheus.RegisterHandler(options.MetricsPath)
	go http.Serve(options.MetricsAddress)
	go vaultClient.StartAuthRenewal(ctx)

	secretReader := secrets.NewVaultReader(vaultClient)
	vaultTemplate := template.New(logger, secretReader)
	resultSecrets, err := vaultTemplate.RenderToPath(options.CommonTemplateOptions)
	if err != nil {
		return err
	}

	go vaultClient.RenewLeases(ctx, resultSecrets.Secrets)

	return vaultClient.Wait(ctx)
}
