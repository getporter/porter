package portergrpc

import (
	"context"
	"encoding/json"
	"fmt"

	iGRPCv1alpha1 "get.porter.sh/porter/gen/proto/go/porterapis/installation/v1alpha1"
	"get.porter.sh/porter/pkg/porter"
	"get.porter.sh/porter/pkg/tracing"
	"google.golang.org/protobuf/encoding/protojson"
)

func newInstallationOptsFromLabels(labels map[string]string) []string {
	var retLabels []string
	for k, v := range labels {
		retLabels = append(retLabels, fmt.Sprintf("%s=%s", k, v))
	}
	return retLabels
}

// newGRPCPorterValue creates a GRPC PorterValue (generated from protobufs)
// from native porter DisplayValue
func newGRPCPorterValue(value porter.DisplayValue) (*iGRPCv1alpha1.PorterValue, error) {
	b, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	pv := &iGRPCv1alpha1.PorterValue{}
	err = protojson.Unmarshal(b, pv)
	if err != nil {
		return nil, err
	}
	return pv, nil
}

// newGRPCPorterValues creates a slice of GRPC PorterValue from native porter DisplayValues
// to accommodate differences in generated structs from protobufs
func newGRPCPorterValues(values porter.DisplayValues) []*iGRPCv1alpha1.PorterValue {
	var retPVs []*iGRPCv1alpha1.PorterValue
	for _, dv := range values {
		//TODO: handle error
		pv, _ := newGRPCPorterValue(dv)
		retPVs = append(retPVs, pv)
	}
	return retPVs
}

// populateGRPCInstallation populates a GRPC Installation (generated from protobuf)
// a native porter DisplayInstallation
func populateGRPCInstallation(inst porter.DisplayInstallation, gInst *iGRPCv1alpha1.Installation) error {
	bInst, err := json.Marshal(inst)
	if err != nil {
		return fmt.Errorf("porter.DisplayInstallation marshal error: %e", err)
	}
	pjum := protojson.UnmarshalOptions{}
	err = pjum.Unmarshal(bInst, gInst)
	if err != nil {
		return fmt.Errorf("installation GRPC Installation unmarshal error: %e", err)
	}
	return nil
}

// populateGRPCPorterValue populates a GRPC PorterValue (generated from protobuf)
// from a native porter DisplayValue
func populateGRPCPorterValue(dv porter.DisplayValue, gInstOut *iGRPCv1alpha1.PorterValue) error {
	bInstOut, err := json.Marshal(dv)
	if err != nil {
		return fmt.Errorf("PorterValue marshal error: %e", err)
	}
	pjum := protojson.UnmarshalOptions{}
	err = pjum.Unmarshal(bInstOut, gInstOut)
	if err != nil {
		return fmt.Errorf("installation GRPC InstallationOutputs unmarshal error: %e", err)
	}
	return nil
}

func (s *PorterServer) ListInstallations(ctx context.Context, req *iGRPCv1alpha1.ListInstallationsRequest) (*iGRPCv1alpha1.ListInstallationsResponse, error) {
	ctx, log := tracing.StartSpan(ctx)
	defer log.EndSpan()
	p, err := GetPorterConnectionFromContext(ctx)
	// Maybe try to setup a new porter connection instead of erring?
	if err != nil {
		return nil, err
	}
	opts := porter.ListOptions{
		Name:          req.GetName(),
		Namespace:     req.GetNamespace(),
		Labels:        newInstallationOptsFromLabels(req.GetLabels()),
		AllNamespaces: req.GetAllNamespaces(),
		Skip:          req.GetSkip(),
		Limit:         req.GetLimit(),
	}
	installations, err := p.ListInstallations(ctx, opts)
	if err != nil {
		return nil, err
	}
	insts := []*iGRPCv1alpha1.Installation{}
	for _, pInst := range installations {
		gInst := &iGRPCv1alpha1.Installation{}
		err := populateGRPCInstallation(pInst, gInst)
		if err != nil {
			return nil, err
		}
		insts = append(insts, gInst)
	}
	res := iGRPCv1alpha1.ListInstallationsResponse{
		Installation: insts,
	}
	return &res, nil
}

func (s *PorterServer) ListInstallationLatestOutputs(ctx context.Context, req *iGRPCv1alpha1.ListInstallationLatestOutputRequest) (*iGRPCv1alpha1.ListInstallationLatestOutputResponse, error) {
	ctx, log := tracing.StartSpan(ctx)
	defer log.EndSpan()
	p, err := GetPorterConnectionFromContext(ctx)
	// Maybe try to setup a new porter connection instead of erring?
	if err != nil {
		return nil, err
	}

	opts := porter.OutputListOptions{}
	opts.Name = req.GetName()
	opts.Namespace = req.GetNamespace()
	opts.Format = "json"
	pdv, err := p.ListBundleOutputs(ctx, &opts)
	if err != nil {
		return nil, err
	}
	gInstOuts := []*iGRPCv1alpha1.PorterValue{}
	for _, dv := range pdv {
		gInstOut := &iGRPCv1alpha1.PorterValue{}
		err = populateGRPCPorterValue(dv, gInstOut)
		if err != nil {
			return nil, err
		}
		gInstOuts = append(gInstOuts, gInstOut)
	}
	res := &iGRPCv1alpha1.ListInstallationLatestOutputResponse{
		Outputs: gInstOuts,
	}
	return res, nil
}
