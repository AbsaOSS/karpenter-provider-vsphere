package userdata

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"text/template"

	"github.com/coreos/butane/config/common"
	fcos "github.com/coreos/butane/config/fcos/v1_5"
	ignition "github.com/coreos/ignition/v2/config/v3_4"
	ignitionTypes "github.com/coreos/ignition/v2/config/v3_4/types"
)

type ButaneRenderer struct{}

func (r *ButaneRenderer) Render(
	data *DistroConfig,
	additional string,
) ([]byte, error) {

	butaneBytes, err := renderButane(data)
	if err != nil {
		return nil, err
	}
	joinData, err := butaneToIgnition(butaneBytes)

	if additional != "" {
		addCfg, err := butaneToIgnition([]byte(additional))
		if err != nil {
			return nil, fmt.Errorf("converting additional config to Ignition: %w", err)
		}

		joinData = ignition.Merge(joinData, addCfg)
	}

	return json.Marshal(joinData)
}

func renderButane(input *DistroConfig) ([]byte, error) {
	t := template.Must(template.New("template").Funcs(defaultTemplateFuncMap()).Parse(butaneTemplate))

	var out bytes.Buffer
	if err := t.Execute(&out, input); err != nil {
		return nil, fmt.Errorf("failed to render template: %w", err)
	}

	return out.Bytes(), nil
}

func butaneToIgnition(data []byte) (ignitionTypes.Config, error) {
	ignBytes, reports, err := fcos.ToIgn3_4Bytes(data, common.TranslateBytesOptions{})
	if err != nil {
		return ignitionTypes.Config{}, fmt.Errorf("error converting to Ignition: %w", err)
	}

	cfg, parseReport, err := ignition.Parse(ignBytes)
	if err != nil {
		return ignitionTypes.Config{}, fmt.Errorf("error parsing resulting Ignition: %w", err)
	}

	reports.Merge(parseReport)

	return cfg, nil
}

func defaultTemplateFuncMap() template.FuncMap {
	return template.FuncMap{
		"Indent":     templateYAMLIndent,
		"Split":      strings.Split,
		"Join":       strings.Join,
		"ParseOwner": parseOwner,
	}
}
func templateYAMLIndent(i int, input string) string {
	split := strings.Split(input, "\n")
	ident := "\n" + strings.Repeat(" ", i)

	return strings.Join(split, ident)
}

type owner struct {
	User  *string
	Group *string
}

func parseOwner(ownerRaw string) owner {
	if ownerRaw == "" {
		return owner{}
	}

	ownerSlice := strings.Split(ownerRaw, ":")

	parseEntity := func(entity string) *string {
		if entity == "" {
			return nil
		}

		entityTrimmed := strings.TrimSpace(entity)

		return &entityTrimmed
	}

	if len(ownerSlice) == 1 {
		return owner{
			User: parseEntity(ownerSlice[0]),
		}
	}

	return owner{
		User:  parseEntity(ownerSlice[0]),
		Group: parseEntity(ownerSlice[1]),
	}
}
