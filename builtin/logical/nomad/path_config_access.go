package nomad

import (
	"context"

	"github.com/hashicorp/errwrap"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

const configAccessKey = "config/access"

func pathConfigAccess(b *backend) *framework.Path {
	return &framework.Path{
		Pattern: "config/access",
		Fields: map[string]*framework.FieldSchema{
			"address": {
				Type:        framework.TypeString,
				Description: "Nomad server address",
			},

			"token": {
				Type:        framework.TypeString,
				Description: "Token for API calls",
			},

			"max_token_name_length": {
				Type:        framework.TypeInt,
				Description: "Max length for name of generated Nomad tokens",
			},
			"ca_cert": {
				Type: framework.TypeString,
				Description: `CA certificate to use when verifying Nomad server certificate,
must be x509 PEM encoded.`,
			},
			"client_cert": {
				Type: framework.TypeString,
				Description: `Client certificate used for Nomad's TLS communication,
must be x509 PEM encoded and if this is set you need to also set client_key.`,
			},
			"client_key": {
				Type: framework.TypeString,
				Description: `Client key used for Nomad's TLS communication,
must be x509 PEM encoded and if this is set you need to also set client_cert.`,
			},
		},

		Callbacks: map[logical.Operation]framework.OperationFunc{
			logical.ReadOperation:   b.pathConfigAccessRead,
			logical.CreateOperation: b.pathConfigAccessWrite,
			logical.UpdateOperation: b.pathConfigAccessWrite,
			logical.DeleteOperation: b.pathConfigAccessDelete,
		},

		ExistenceCheck: b.configExistenceCheck,
	}
}

func (b *backend) configExistenceCheck(ctx context.Context, req *logical.Request, data *framework.FieldData) (bool, error) {
	entry, err := b.readConfigAccess(ctx, req.Storage)
	if err != nil {
		return false, err
	}

	return entry != nil, nil
}

func (b *backend) readConfigAccess(ctx context.Context, storage logical.Storage) (*accessConfig, error) {
	entry, err := storage.Get(ctx, configAccessKey)
	if err != nil {
		return nil, err
	}
	if entry == nil {
		return nil, nil
	}

	conf := &accessConfig{}
	if err := entry.DecodeJSON(conf); err != nil {
		return nil, errwrap.Wrapf("error reading nomad access configuration: {{err}}", err)
	}

	return conf, nil
}

func (b *backend) pathConfigAccessRead(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	conf, err := b.readConfigAccess(ctx, req.Storage)
	if err != nil {
		return nil, err
	}
	if conf == nil {
		return nil, nil
	}

	return &logical.Response{
		Data: map[string]interface{}{
			"address":               conf.Address,
			"max_token_name_length": conf.MaxTokenNameLength,
		},
	}, nil
}

func (b *backend) pathConfigAccessWrite(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	conf, err := b.readConfigAccess(ctx, req.Storage)
	if err != nil {
		return nil, err
	}
	if conf == nil {
		conf = &accessConfig{}
	}

	address, ok := data.GetOk("address")
	if ok {
		conf.Address = address.(string)
	}
	token, ok := data.GetOk("token")
	if ok {
		conf.Token = token.(string)
	}
	caCert, ok := data.GetOk("ca_cert")
	if ok {
		conf.CACert = caCert.(string)
	}
	clientCert, ok := data.GetOk("client_cert")
	if ok {
		conf.ClientCert = clientCert.(string)
	}
	clientKey, ok := data.GetOk("client_key")
	if ok {
		conf.ClientKey = clientKey.(string)
	}

	conf.MaxTokenNameLength = data.Get("max_token_name_length").(int)

	entry, err := logical.StorageEntryJSON("config/access", conf)
	if err != nil {
		return nil, err
	}
	if err := req.Storage.Put(ctx, entry); err != nil {
		return nil, err
	}

	return nil, nil
}

func (b *backend) pathConfigAccessDelete(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	if err := req.Storage.Delete(ctx, configAccessKey); err != nil {
		return nil, err
	}
	return nil, nil
}

type accessConfig struct {
	Address            string `json:"address"`
	Token              string `json:"token"`
	MaxTokenNameLength int    `json:"max_token_name_length"`
	CACert             string `json:"ca_cert"`
	ClientCert         string `json:"client_cert"`
	ClientKey          string `json:"client_key"`
}
