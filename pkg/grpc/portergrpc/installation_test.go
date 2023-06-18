package portergrpc

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	iGRPC "get.porter.sh/porter/gen/proto/go/porterapis/installation/v1alpha1"
	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/porter"
	"get.porter.sh/porter/pkg/storage"
	"get.porter.sh/porter/tests"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/encoding/protojson"
	"k8s.io/utils/pointer"
)

type instInfo struct {
	namespace string
	name      string
}

func TestListInstallationReturnsListOfCorrectPorterInstallations(t *testing.T) {
	testNamespace := "test"
	testInstallationName := "foo"
	filterOpts := []struct {
		name string
		opts *iGRPC.ListInstallationsRequest
	}{
		{
			name: "FilterByAllNamespaces",
			opts: &iGRPC.ListInstallationsRequest{AllNamespaces: pointer.Bool(true)},
		},
		{
			name: "FilterByNamespace",
			opts: &iGRPC.ListInstallationsRequest{Namespace: &testNamespace},
		},
		{
			name: "FilterByInstallationNameAndNamespace",
			opts: &iGRPC.ListInstallationsRequest{Namespace: &testNamespace, Name: testInstallationName},
		},
	}
	tests := []struct {
		name     string
		instInfo []instInfo
		// Number of expected installations when filtering by all namespaces, all installations in a single namespace, and a single installation in a namespace
		numExpInsts []int
	}{
		{
			name:        "NoInstallations",
			instInfo:    []instInfo{},
			numExpInsts: []int{0, 0, 0},
		},
		{
			name: "SingleInstallationDefaultNamespace",
			instInfo: []instInfo{
				{namespace: "", name: "test"},
			},
			numExpInsts: []int{1, 0, 0},
		},
		{
			name: "SingleInstallationInMultipleNamespaces",
			instInfo: []instInfo{
				{namespace: testNamespace, name: testInstallationName},
				{namespace: "bar", name: "test"},
			},
			numExpInsts: []int{2, 1, 1},
		},
		{
			name: "MultipleInstallationSInMultipleNamespaces",
			instInfo: []instInfo{
				{namespace: "foo", name: "test1"},
				{namespace: "foo", name: "test2"},
				{namespace: testNamespace, name: testInstallationName},
				{namespace: testNamespace, name: "test4"},
			},
			numExpInsts: []int{4, 2, 1},
		},
	}
	for _, test := range tests {
		ctx, insts := setupTestPorterWithInstallations(t, test.instInfo)
		for i, opts := range filterOpts {
			t.Run(fmt.Sprintf("%s%s", test.name, opts.name), func(t *testing.T) {
				instSvc := PorterServer{}
				resp, err := instSvc.ListInstallations(ctx, opts.opts)
				installations := resp.GetInstallation()
				assert.Nil(t, err)
				assert.Len(t, installations, test.numExpInsts[i])
				verifyInstallations(t, installations, insts)
			})
		}
	}
}

func TestListInstallationsReturnsErrorIfUnableToGetPorterConnectionFromRequestContext(t *testing.T) {
	instSvc := PorterServer{}
	req := &iGRPC.ListInstallationsRequest{}
	ctx := context.TODO()
	resp, err := instSvc.ListInstallations(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)

}

func setupTestPorterWithInstallations(t *testing.T, installations []instInfo) (context.Context, map[string]porter.DisplayInstallation) {
	p := porter.NewTestPorter(t)
	insts := map[string]porter.DisplayInstallation{}
	for _, inst := range installations {
		installation := storage.NewInstallation(inst.namespace, inst.name)
		storeInst := p.TestInstallations.CreateInstallation(installation, p.TestInstallations.SetMutableInstallationValues, func(i *storage.Installation) {
			// Overwrite the default ID set by SetMutableInstallationValues because it is always the same
			i.ID = uuid.NewString()
			i.Status.BundleVersion = "v0.1.0"
			i.Status.ResultStatus = cnab.StatusSucceeded
			i.Bundle.Repository = "test-bundle"
			i.Bundle.Version = "v0.1.0"
		})
		insts[storeInst.ID] = porter.NewDisplayInstallation(storeInst)
	}
	ctx := AddPorterConnectionToContext(p.Porter, context.TODO())
	return ctx, insts
}

func verifyInstallations(t *testing.T, installations []*iGRPC.Installation, allInsts map[string]porter.DisplayInstallation) {
	for _, inst := range installations {
		i, ok := allInsts[inst.Id]
		assert.True(t, ok)
		bExpInst, err := json.Marshal(i)
		assert.NoError(t, err)
		grpcExpInst, err := tests.GRPCDisplayInstallationExpectedJSON(bExpInst)
		assert.NoError(t, err)
		pjm := protojson.MarshalOptions{EmitUnpopulated: true}
		bActInst, err := pjm.Marshal(inst)
		assert.NoError(t, err)
		assert.JSONEq(t, string(grpcExpInst), string(bActInst))
	}
}
