package pluginstore

import (
	"context"

	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/storage/plugins"
	"get.porter.sh/porter/pkg/storage/plugins/proto"
	"go.mongodb.org/mongo-driver/bson"
	"google.golang.org/protobuf/types/known/structpb"
)

var _ plugins.StorageProtocol = &GClient{}

// GClient is a gRPC implementation of the storage client.
type GClient struct {
	client proto.StorageProtocolClient
}

func NewClient(client proto.StorageProtocolClient) *GClient {
	return &GClient{client}
}

func (m *GClient) EnsureIndex(ctx context.Context, opts plugins.EnsureIndexOptions) error {
	req := &proto.EnsureIndexRequest{
		Indices: make([]*proto.Index, len(opts.Indices)),
	}

	for i, index := range opts.Indices {
		req.Indices[i] = &proto.Index{
			Collection: index.Collection,
			Keys:       FromOrderedMap(index.Keys),
			Unique:     index.Unique,
		}
	}

	_, err := m.client.EnsureIndex(ctx, req)
	return err
}

func (m *GClient) Aggregate(ctx context.Context, opts plugins.AggregateOptions) ([]bson.Raw, error) {
	req := &proto.AggregateRequest{
		Collection: opts.Collection,
		Pipeline:   NewPipeline(opts.Pipeline),
	}
	resp, err := m.client.Aggregate(ctx, req)
	if err != nil {
		return nil, err
	}

	results := convertToBsonRawList(resp.Results)
	return results, nil
}

func (m *GClient) Count(ctx context.Context, opts plugins.CountOptions) (int64, error) {
	resp, err := m.client.Count(ctx, &proto.CountRequest{
		Collection: opts.Collection,
		Filter:     FromMap(opts.Filter),
	})
	if err != nil {
		return 0, err
	}
	return resp.Count, nil
}

func (m *GClient) Find(ctx context.Context, opts plugins.FindOptions) ([]bson.Raw, error) {
	req := &proto.FindRequest{
		Collection: opts.Collection,
		Sort:       FromOrderedMap(opts.Sort),
		Skip:       opts.Skip,
		Limit:      opts.Limit,
		Select:     FromOrderedMap(opts.Select),
		Filter:     FromMap(opts.Filter),
		Group:      FromOrderedMap(opts.Group),
	}
	resp, err := m.client.Find(ctx, req)
	if err != nil {
		return nil, err
	}

	results := convertToBsonRawList(resp.Results)
	return results, nil
}

func (m *GClient) Insert(ctx context.Context, opts plugins.InsertOptions) error {
	req := &proto.InsertRequest{
		Collection: opts.Collection,
		Documents:  FromMapList(opts.Documents),
	}
	_, err := m.client.Insert(ctx, req)
	return err
}

func (m *GClient) Patch(ctx context.Context, opts plugins.PatchOptions) error {
	req := &proto.PatchRequest{
		Collection:     opts.Collection,
		QueryDocument:  FromMap(opts.QueryDocument),
		Transformation: FromOrderedMap(opts.Transformation),
	}
	_, err := m.client.Patch(ctx, req)
	return err
}

func (m *GClient) Remove(ctx context.Context, opts plugins.RemoveOptions) error {
	req := &proto.RemoveRequest{
		Collection: opts.Collection,
		Filter:     FromMap(opts.Filter),
		All:        opts.All,
	}
	_, err := m.client.Remove(ctx, req)
	return err
}

func (m *GClient) Update(ctx context.Context, opts plugins.UpdateOptions) error {
	req := &proto.UpdateRequest{
		Collection: opts.Collection,
		Filter:     FromMap(opts.Filter),
		Upsert:     opts.Upsert,
		Document:   FromMap(opts.Document),
	}
	_, err := m.client.Update(ctx, req)
	return err
}

// GServer is a gRPC wrapper around a StorageProtocol plugin
type GServer struct {
	impl plugins.StorageProtocol
	proto.UnsafeStorageProtocolServer
	c *portercontext.Context
}

func NewServer(c *portercontext.Context, impl plugins.StorageProtocol) *GServer {
	return &GServer{c: c, impl: impl}
}

func (m *GServer) EnsureIndex(ctx context.Context, request *proto.EnsureIndexRequest) (*proto.EnsureIndexResponse, error) {
	opts := plugins.EnsureIndexOptions{
		Indices: make([]plugins.Index, len(request.Indices)),
	}

	for i, index := range request.Indices {
		opts.Indices[i] = plugins.Index{
			Collection: index.Collection,
			Keys:       AsOrderedMap(index.Keys),
			Unique:     index.Unique,
		}
	}

	err := m.impl.EnsureIndex(ctx, opts)
	return &proto.EnsureIndexResponse{}, err
}

func (m *GServer) Aggregate(ctx context.Context, request *proto.AggregateRequest) (*proto.AggregateResponse, error) {
	opts := plugins.AggregateOptions{
		Collection: request.Collection,
		Pipeline:   AsOrderedMapList(request.Pipeline),
	}

	results, err := m.impl.Aggregate(ctx, opts)
	resp := &proto.AggregateResponse{Results: make([][]byte, len(results))}
	for i := range results {
		resp.Results[i] = results[i]
	}
	return resp, err
}

func (m *GServer) Count(ctx context.Context, req *proto.CountRequest) (*proto.CountResponse, error) {
	opts := plugins.CountOptions{
		Collection: req.Collection,
		Filter:     AsMap(req.Filter),
	}
	count, err := m.impl.Count(ctx, opts)
	return &proto.CountResponse{Count: count}, err
}

func (m *GServer) Find(ctx context.Context, request *proto.FindRequest) (*proto.FindResponse, error) {
	opts := plugins.FindOptions{
		Collection: request.Collection,
		Sort:       AsOrderedMap(request.Sort),
		Skip:       request.Skip,
		Limit:      request.Limit,
		Select:     AsOrderedMap(request.Select),
		Filter:     AsMap(request.Filter),
		Group:      AsOrderedMap(request.Group),
	}

	results, err := m.impl.Find(ctx, opts)
	resp := &proto.FindResponse{Results: make([][]byte, len(results))}
	for i := range results {
		resp.Results[i] = results[i]
	}

	return resp, err
}

func (m *GServer) Insert(ctx context.Context, request *proto.InsertRequest) (*proto.InsertResponse, error) {
	opts := plugins.InsertOptions{
		Collection: request.Collection,
		Documents:  AsMapList(request.Documents),
	}

	err := m.impl.Insert(ctx, opts)
	return &proto.InsertResponse{}, err
}

func (m *GServer) Patch(ctx context.Context, request *proto.PatchRequest) (*proto.PatchResponse, error) {
	opts := plugins.PatchOptions{
		Collection:     request.Collection,
		QueryDocument:  AsMap(request.QueryDocument),
		Transformation: AsOrderedMap(request.Transformation),
	}

	err := m.impl.Patch(ctx, opts)
	return &proto.PatchResponse{}, err
}

func (m *GServer) Remove(ctx context.Context, request *proto.RemoveRequest) (*proto.RemoveResponse, error) {
	opts := plugins.RemoveOptions{
		Collection: request.Collection,
		Filter:     AsMap(request.Filter),
		All:        request.All,
	}

	err := m.impl.Remove(ctx, opts)
	return &proto.RemoveResponse{}, err
}

func (m *GServer) Update(ctx context.Context, request *proto.UpdateRequest) (*proto.UpdateResponse, error) {
	opts := plugins.UpdateOptions{
		Collection: request.Collection,
		Filter:     AsMap(request.Filter),
		Upsert:     request.Upsert,
		Document:   AsMap(request.Document),
	}

	err := m.impl.Update(ctx, opts)
	return &proto.UpdateResponse{}, err
}

func NewPipeline(src []bson.D) []*proto.Stage {
	pipeline := make([]*proto.Stage, len(src))
	for i, srcStage := range src {
		stage := &proto.Stage{
			Steps: FromOrderedMap(srcStage),
		}
		pipeline[i] = stage
	}

	return pipeline
}

// ConvertBsonToPrimitives converts from bson primitives to pure Go primitives
func ConvertBsonToPrimitives(src interface{}) interface{} {
	switch t := src.(type) {
	case bson.E:
		rawValue := ConvertBsonToPrimitives(t.Value)
		return map[string]interface{}{t.Key: rawValue}
	case []bson.D:
		raw := make([]interface{}, len(t))
		for i, item := range t {
			raw[i] = ConvertBsonToPrimitives(item)
		}
		return raw
	case bson.D:
		raw := make([]interface{}, len(t))
		for i, item := range t {
			raw[i] = ConvertBsonToPrimitives(item)
		}
		return raw
	case []bson.M:
		raw := make([]interface{}, len(t))
		for i, item := range t {
			raw[i] = ConvertBsonToPrimitives(item)
		}
		return raw
	case bson.M:
		raw := make(map[string]interface{}, len(t))
		for k, v := range t {
			raw[k] = ConvertBsonToPrimitives(v)
		}
		return raw
	default:
		return src
	}
}

// ConvertSliceToBsonD converts go slices to bson primitive.
// it also works around a weirdness in how numbers are represented
// by structpb.Value, where integer values are stored in float64. When we
// deserialize from protobuf, this walks the specified value, finds ints
// that were encoded as floats, and converts them back to ints.
func ConvertSliceToBsonD(src interface{}) interface{} {
	dest := ConvertFloatToInt(src)
	switch tv := dest.(type) {
	case []interface{}:
		toBson := make(bson.D, 0, len(tv))
		for i, item := range tv {
			converted := ConvertSliceToBsonD(item)
			if m, ok := converted.(map[string]interface{}); ok {
				for k, v := range m {
					toBson = append(toBson, bson.E{Key: k, Value: v})
				}
				continue
			}
			tv[i] = converted
		}
		if len(toBson) > 0 {
			return toBson
		}
		return tv
	case map[string]interface{}:
		for k, v := range tv {
			tv[k] = ConvertSliceToBsonD(v)
		}
		return tv
	default:
		return tv
	}
}

// ConvertFloatToInt works around a weirdness in how numbers are represented
// by structpb.Value, where integer values are stored in float64. When we
// deserialize from protobuf, this walks the specified value, finds ints
// that were encoded as floats, and converts them back to ints.
func ConvertFloatToInt(src interface{}) interface{} {
	switch tv := src.(type) {
	case float64:
		intVal := int64(tv)
		if tv == float64(intVal) {
			return intVal
		}
		return tv
	case []interface{}:
		for i, item := range tv {
			tv[i] = ConvertFloatToInt(item)
		}
		return tv
	case map[string]interface{}:
		for k, v := range tv {
			tv[k] = ConvertFloatToInt(v)
		}
		return tv
	default:
		return tv
	}
}

// FromMap represents bson.M in a data structure that protobuf understands
// (which is a plain struct).
func FromMap(src bson.M) *structpb.Struct {
	rawSrc := make(map[string]interface{}, len(src))
	for k, v := range src {
		rawSrc[k] = ConvertBsonToPrimitives(v)
	}

	dest, err := structpb.NewStruct(rawSrc)
	if err != nil {
		// panic because if we hit this, there's no recovering or handling possible
		panic(err)
	}

	return dest
}

// FromMapList represents []bson.M in a data structure that protobuf understands
// (an array of structs).
func FromMapList(src []bson.M) []*structpb.Struct {
	dest := make([]*structpb.Struct, len(src))
	for i, item := range src {
		dest[i] = FromMap(item)
	}
	return dest
}

// FromOrderedMap represents bson.D, an ordered map, in a data structure that
// protobuf understands (an array of structs).
func FromOrderedMap(src bson.D) []*structpb.Struct {
	dest := make([]*structpb.Struct, len(src))
	for i, item := range src {
		rawValue := ConvertBsonToPrimitives(item.Value)
		dest[i] = NewStruct(map[string]interface{}{item.Key: rawValue})
	}
	return dest
}

func NewStruct(src map[string]interface{}) *structpb.Struct {
	dest, err := structpb.NewStruct(src)
	if err != nil {
		panic(err)
	}
	return dest
}

// AsMap converts a protobuf struct into its original representation, bson.M.
func AsMap(src *structpb.Struct, c ...converter) bson.M {
	dest := src.AsMap()
	converts := []converter{ConvertFloatToInt}
	if c != nil {
		converts = append(converts, c...)
	}
	for k, v := range dest {
		for _, convert := range converts {
			v = convert(v)
		}
		dest[k] = v
	}
	return dest
}

func AsMapList(src []*structpb.Struct) []bson.M {
	dest := make([]bson.M, len(src))
	for i, item := range src {
		dest[i] = AsMap(item)
	}
	return dest
}

type converter func(src interface{}) interface{}

// AsOrderedMap converts an array of protobuf structs into its original
// representation, bson.D.
func AsOrderedMap(src []*structpb.Struct, c ...converter) bson.D {
	dest := make(bson.D, 0, len(src))
	if c == nil {
		c = []converter{ConvertFloatToInt}
	}
	for _, item := range src {
		for k, v := range AsMap(item, c...) {
			dest = append(dest, bson.E{Key: k, Value: v})
		}
	}
	return dest
}

// AsOrderedMapList converts a protobuf Pipeline into its original
// representation, []bson.D
func AsOrderedMapList(src []*proto.Stage) []bson.D {
	dest := make([]bson.D, len(src))
	for i, item := range src {
		dest[i] = AsOrderedMap(item.Steps, ConvertSliceToBsonD)
	}
	return dest
}

func convertToBsonRawList(src [][]byte) []bson.Raw {
	dest := make([]bson.Raw, len(src))
	for i := range src {
		dest[i] = src[i]
	}
	return dest
}
