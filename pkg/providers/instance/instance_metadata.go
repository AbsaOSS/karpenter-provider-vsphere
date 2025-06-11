package instance

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"html/template"

	v1alpha1 "github.com/absaoss/karpenter-provider-vsphere/pkg/apis/v1alpha1"
	"github.com/vmware/govmomi/vim25/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Config is data used with a VM's guestInfo RPC interface.
type Config []types.BaseOptionValue

const (
	guestInfoIgnitionData      = "guestinfo.ignition.config.data"
	guestInfoIgnitionEncoding  = "guestinfo.ignition.config.data.encoding"
	guestInfoCloudInitData     = "guestinfo.userdata"
	guestInfoMetadata          = "guestinfo.metadata"
	guestInfoMetadataEncoding  = "guestinfo.metadata.encoding"
	guestInfoCloudInitEncoding = "guestinfo.userdata.encoding"
)

func (e *Config) Extract() []types.BaseOptionValue {
	if e == nil {
		return nil
	}
	return *e
}

func (p *DefaultProvider) GetInitData(ctx context.Context, class *v1alpha1.VsphereNodeClass, nodeName string) ([]types.BaseOptionValue, error) {
	metaData := &Config{}
	// Set the metadata with the local hostname
	metaDataRaw := []byte(fmt.Sprintf("local-hostname: \"%s\"", nodeName))
	userDataTplBase64 := class.Spec.UserData.TemplateBase64
	decodedUserData, err := base64.StdEncoding.DecodeString(userDataTplBase64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode user data template: %w", err)
	}
	templateValuesSecret, err := p.kubeClient.CoreV1().Secrets(class.Spec.UserData.Values.Namespace).Get(ctx, class.Spec.UserData.Values.Name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get user data values secret: %w", err)
	}
	values := templateValuesSecret.Data
	strValues := map[string]string{}
	for k, v := range values {
		strValues[k] = string(v)
	}
	tmpl, err := template.New("cloud-data").Parse(string(decodedUserData))
	if err != nil {
		panic(err)
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, strValues)
	if err != nil {
		return nil, fmt.Errorf("failed to execute user data template: %w", err)
	}
	// Ignition userdata
	if class.Spec.UserData.Type == v1alpha1.UserDataTypeCloudInit {
		metaData.SetCloudInitMetadata(buf.Bytes())
	}
	// Cloud-init userdata
	if class.Spec.UserData.Type == v1alpha1.UserDataTypeIgnition {
		metaData.SetIgnitionUserData(buf.Bytes())
	}
	// Set metadata
	metaData.SetMetadata(metaDataRaw)

	return metaData.Extract(), nil

}

func (e *Config) SetCloudInitMetadata(data []byte) {
	*e = append(*e,
		&types.OptionValue{
			Key:   "guestinfo.metadata",
			Value: e.encode(data),
		},
		&types.OptionValue{
			Key:   "guestinfo.metadata.encoding",
			Value: "base64",
		},
	)
}

// setUserData sets the user data at the provided key
// as a base64-encoded string.
func (e *Config) setUserData(userdataKey, encodingKey string, data []byte) {
	*e = append(*e,
		&types.OptionValue{
			Key:   userdataKey,
			Value: e.encode(data),
		},
		&types.OptionValue{
			Key:   encodingKey,
			Value: "base64",
		},
	)
}

func (e *Config) SetMetadata(data []byte) {
	// Set the metadata at the key "guestinfo.metadata" as a base64-encoded string.
	e.setUserData(guestInfoMetadata, guestInfoMetadataEncoding, data)
}

// SetIgnitionUserData sets the ignition user data at the key
// "guestinfo.ignition.config.data" as a base64-encoded string.
func (e *Config) SetIgnitionUserData(data []byte) {
	e.setUserData(guestInfoIgnitionData, guestInfoIgnitionEncoding, data)
}

// encode first attempts to decode the data as many times as necessary
// to ensure it is plain-text before returning the result as a base64
// encoded string.
func (e *Config) encode(data []byte) string {
	if len(data) == 0 {
		return ""
	}
	for {
		decoded, err := base64.StdEncoding.DecodeString(string(data))
		if err != nil {
			break
		}
		data = decoded
	}
	return base64.StdEncoding.EncodeToString(data)
}
