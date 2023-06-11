package grpc

import (
	"bytes"
	"context"
	"encoding/json"
	"strconv"
	"testing"

	iGRPC "get.porter.sh/porter/gen/proto/go/porterapis/installation/v1alpha1"
	pGRPC "get.porter.sh/porter/gen/proto/go/porterapis/porter/v1alpha1"
	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/porter"
	"get.porter.sh/porter/pkg/storage"
	"get.porter.sh/porter/tests"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/bundle/definition"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"
)

type testOutputOpt struct {
	name         string
	value        string
	bundleOutput bundle.Output
}
type testInstallationOpts struct {
	bundleDefs *map[string]*definition.Schema
	outputs    *[]testOutputOpt
}

// TODO: add opts structure for different installation options
func newTestInstallation(t *testing.T, namespace, name string, grpcSvr *TestPorterGRPCServer, opts testInstallationOpts) storage.Installation {
	//Bundle Definition with required porter-state
	bd := definition.Definitions{
		"porter-state": &definition.Schema{
			Type:    "string",
			Comment: "porter-internal",
		},
	}
	for name, schema := range *opts.bundleDefs {
		bd[name] = schema
	}
	//Bundle Output with required porter-state
	bo := map[string]bundle.Output{
		"porter-state": {
			Definition: "porter-state",
			Path:       "/cnab/app/outputs/porter-state.tgz",
		},
	}
	for _, out := range *opts.outputs {
		bo[out.name] = out.bundleOutput
	}
	b := bundle.Bundle{
		Definitions: bd,
		Outputs:     bo,
	}
	extB := cnab.NewBundle(b)
	storeInst := grpcSvr.TestPorter.TestInstallations.CreateInstallation(storage.NewInstallation(namespace, name), grpcSvr.TestPorter.TestInstallations.SetMutableInstallationValues, func(i *storage.Installation) {
		i.Status.BundleVersion = "v0.1.0"
		i.Status.ResultStatus = cnab.StatusSucceeded
		i.Bundle.Repository = "test-bundle"
		i.Bundle.Version = "v0.1.0"
	})
	c := grpcSvr.TestPorter.TestInstallations.CreateRun(storeInst.NewRun(cnab.ActionInstall, cnab.ExtendedBundle{}), func(sRun *storage.Run) {
		sRun.Bundle = b
		sRun.ParameterOverrides.Parameters = grpcSvr.TestPorter.SanitizeParameters(sRun.ParameterOverrides.Parameters, sRun.ID, extB)
	})
	sRes := grpcSvr.TestPorter.TestInstallations.CreateResult(c.NewResult(cnab.StatusSucceeded))
	for _, out := range *opts.outputs {
		grpcSvr.TestPorter.CreateOutput(sRes.NewOutput(out.name, []byte(out.value)), extB)
	}
	return storeInst
}

func TestInstall_installationMessage(t *testing.T) {
	writeOnly := true
	basicInstOpts := testInstallationOpts{
		bundleDefs: &map[string]*definition.Schema{
			"foo": {Type: "string", WriteOnly: &writeOnly},
			"bar": {Type: "string", WriteOnly: &writeOnly},
		},
		outputs: &[]testOutputOpt{
			{
				name:         "foo",
				value:        "foo-data",
				bundleOutput: bundle.Output{Definition: "foo", Path: "/path/to/foo"},
			},
			{
				name:         "bar",
				value:        "bar-data",
				bundleOutput: bundle.Output{Definition: "bar", Path: "/path/to/bar"},
			},
		},
	}
	tests := []struct {
		testName      string
		instName      string
		instNamespace string ""
		instOpts      testInstallationOpts
	}{
		{
			testName: "basic installation",
			instName: "test",
			instOpts: basicInstOpts,
		},
		{
			testName: "another installation",
			instName: "another-test",
			instOpts: basicInstOpts,
		},
	}
	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			//Server setup
			grpcSvr, err := NewTestGRPCServer(t)
			require.NoError(t, err)
			server := grpcSvr.ListenAndServe()
			defer server.Stop()

			//Client setup
			ctx := context.TODO()
			client, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
			require.NoError(t, err)
			defer client.Close()
			instClient := pGRPC.NewPorterClient(client)

			inst := newTestInstallation(t, test.instNamespace, test.instName, grpcSvr, test.instOpts)

			//Call ListInstallations
			resp, err := instClient.ListInstallations(ctx, &iGRPC.ListInstallationsRequest{})
			require.NoError(t, err)
			assert.Len(t, resp.Installation, 1)
			// Validation
			validateInstallations(t, inst, resp.GetInstallation()[0])

			//Call ListInstallationLatestOutputRequest
			req := &iGRPC.ListInstallationLatestOutputRequest{Name: test.instName, Namespace: &test.instNamespace}
			oresp, err := instClient.ListInstallationLatestOutputs(ctx, req)
			require.NoError(t, err)
			assert.Len(t, oresp.GetOutputs(), len(*test.instOpts.outputs))

			oOpts := &porter.OutputListOptions{}
			oOpts.Name = test.instName
			oOpts.Namespace = test.instNamespace
			oOpts.Format = "json"
			dvs, err := grpcSvr.TestPorter.ListBundleOutputs(ctx, oOpts)
			require.NoError(t, err)

			//Validation
			validateOutputs(t, dvs, oresp)
		})
	}
}

func validateInstallations(t *testing.T, expected storage.Installation, actual *iGRPC.Installation) {
	assert.Equal(t, actual.Name, expected.Name)
	bExpInst, err := json.Marshal(porter.NewDisplayInstallation(expected))
	require.NoError(t, err)
	bExpInst, err = tests.GRPCDisplayInstallationExpectedJSON(bExpInst)
	require.NoError(t, err)
	pjm := protojson.MarshalOptions{EmitUnpopulated: true}
	bActInst, err := pjm.Marshal(actual)
	require.NoError(t, err)
	var pJson bytes.Buffer
	json.Indent(&pJson, bActInst, "", "  ")
	assert.JSONEq(t, string(bExpInst), string(bActInst))
}

func validateOutputs(t *testing.T, dvs porter.DisplayValues, actual *iGRPC.ListInstallationLatestOutputResponse) {
	//Get expected json
	bExpOuts, err := json.MarshalIndent(dvs, "", "  ")
	require.NoError(t, err)
	pjm := protojson.MarshalOptions{EmitUnpopulated: true, Multiline: true, Indent: "  "}
	//Get actual json response
	for i, gPV := range actual.GetOutputs() {
		bActOut, err := pjm.Marshal(gPV)
		require.NoError(t, err)
		//TODO: make this not dependant on order
		bExpOut := gjson.GetBytes(bExpOuts, strconv.Itoa(i)).String()
		assert.JSONEq(t, bExpOut, string(bActOut))
	}
}
