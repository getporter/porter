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

// populateGRPCInstallation populates a GRPC Installation (generated from protobuf)
// a native porter DisplayInstallation
func populateGRPCInstallation(ctx context.Context, inst porter.DisplayInstallation, gInst *iGRPCv1alpha1.Installation) error {
	_, log := tracing.StartSpan(ctx)
	defer log.EndSpan()
	bInst, err := json.Marshal(inst)
	if err != nil {
		return log.Errorf("porter.DisplayInstallation marshal error: %e", err)
	}
	pjum := protojson.UnmarshalOptions{}
	err = pjum.Unmarshal(bInst, gInst)
	if err != nil {
		return log.Errorf("installation GRPC Installation unmarshal error: %e", err)
	}
	return nil
}

// populateGRPCPorterValue populates a GRPC PorterValue (generated from protobuf)
// from a native porter DisplayValue
func populateGRPCPorterValue(ctx context.Context, dv porter.DisplayValue, gInstOut *iGRPCv1alpha1.PorterValue) error {
	_, log := tracing.StartSpan(ctx)
	defer log.EndSpan()
	bInstOut, err := json.Marshal(dv)
	if err != nil {
		return log.Errorf("PorterValue marshal error: %e", err)
	}
	pjum := protojson.UnmarshalOptions{}
	err = pjum.Unmarshal(bInstOut, gInstOut)
	if err != nil {
		return log.Errorf("installation GRPC InstallationOutputs unmarshal error: %e", err)
	}
	return nil
}

// ListInstallations takes a GRPC ListInstallationsRequest and returns a filtered list of
// porter installations as a GRPC ListInstallationsResponse
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
		err := populateGRPCInstallation(ctx, pInst, gInst)
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

// ListInstallationLatestOutputs takes a GRPC ListInstallationLatestOutputRequest and returns
// the most recent outputs for the porter installation as a GRPC ListInstallationLatestOutputResponse
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
		err = populateGRPCPorterValue(ctx, dv, gInstOut)
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
