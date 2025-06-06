package instance

import (
	"encoding/base64"

	"github.com/vmware/govmomi/vim25/types"
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
