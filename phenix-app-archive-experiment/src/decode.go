package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"phenix/store"
	"phenix/types"
	ifaces "phenix/types/interfaces"
	"phenix/types/version"

	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v3"
)

func DecodeExperiment(body []byte) (*types.Experiment, error) {
	var mapper map[string]interface{}

	if err := json.Unmarshal(body, &mapper); err != nil {
		return nil, fmt.Errorf("unable to parse JSON: %w", err)
	}

	iface, err := version.GetVersionedSpecForKind("Experiment", "v1")
	if err != nil {
		return nil, fmt.Errorf("getting versioned spec for experiment: %w", err)
	}

	if err := mapstructure.Decode(mapper["spec"], &iface); err != nil {
		return nil, fmt.Errorf("decoding versioned spec: %w", err)
	}

	spec, ok := iface.(ifaces.ExperimentSpec)
	if !ok {
		return nil, fmt.Errorf("invalid experiment spec")
	}

	iface, err = version.GetVersionedStatusForKind("Experiment", "v1")
	if err != nil {
		return nil, fmt.Errorf("getting versioned status for experiment: %w", err)
	}

	if err := mapstructure.Decode(mapper["status"], &iface); err != nil {
		return nil, fmt.Errorf("decoding versioned status: %w", err)
	}

	status, ok := iface.(ifaces.ExperimentStatus)
	if !ok {
		return nil, fmt.Errorf("invalid experiment status")
	}

	var metadata store.ConfigMetadata

	if err := mapstructure.Decode(mapper["metadata"], &metadata); err != nil {
		return nil, fmt.Errorf("decoding experiment metadata: %w", err)
	}

	return &types.Experiment{Spec: spec, Status: status, Metadata: metadata}, nil
}

func DecodeExperimentFromFile(filePath string) (*types.Experiment, error) {

	var mapper map[string]interface{}

	body, err := ioutil.ReadFile(filePath)
	if err != nil {
		logger.Fatalf("unable to read YAML from %s", filePath)
	}

	if err := yaml.Unmarshal(body, &mapper); err != nil {
		return nil, fmt.Errorf("unable to parse YAML: %w", err)
	}

	iface, err := version.GetVersionedSpecForKind("Experiment", "v1")
	if err != nil {
		return nil, fmt.Errorf("getting versioned spec for experiment: %w", err)
	}

	if err := mapstructure.Decode(mapper["spec"], &iface); err != nil {
		return nil, fmt.Errorf("decoding versioned spec: %w", err)
	}

	spec, ok := iface.(ifaces.ExperimentSpec)
	if !ok {
		return nil, fmt.Errorf("invalid experiment spec")
	}

	iface, err = version.GetVersionedStatusForKind("Experiment", "v1")
	if err != nil {
		return nil, fmt.Errorf("getting versioned status for experiment: %w", err)
	}

	if err := mapstructure.Decode(mapper["status"], &iface); err != nil {
		return nil, fmt.Errorf("decoding versioned status: %w", err)
	}

	status, ok := iface.(ifaces.ExperimentStatus)
	if !ok {
		return nil, fmt.Errorf("invalid experiment status")
	}

	var metadata store.ConfigMetadata

	if err := mapstructure.Decode(mapper["metadata"], &metadata); err != nil {
		return nil, fmt.Errorf("decoding experiment metadata: %w", err)
	}

	return &types.Experiment{Spec: spec, Status: status, Metadata: metadata}, nil

}
