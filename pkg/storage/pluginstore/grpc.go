package pluginstore

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"

	"get.porter.sh/porter/pkg/storage/plugins"
	"get.porter.sh/porter/pkg/storage/plugins/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

// GClient is a gRPC implementation of the storage client.
type GClient struct {
	client proto.StorageProtocolClient
}

func (m *GClient) EnsureIndex(ctx context.Context, opts plugins.EnsureIndexOptions) error {
	req := &proto.EnsureIndexRequest{
		Indices: make([]*proto.Index, len(opts.Indices)),
	}

	for i := range opts.Indices {
		keys, err := convertFromBsonD(opts.Indices[i].Keys)
		if err != nil {
			return fmt.Errorf("error converting Indicies.Keys to protobuf: %w", err)
		}

		req.Indices[i] = &proto.Index{
			Collection: opts.Indices[i].Collection,
			Keys:       keys,
			Unique:     false,
		}
	}

	_, err := m.client.EnsureIndex(ctx, req)
	return err
}

func (m *GClient) Aggregate(ctx context.Context, opts plugins.AggregateOptions) ([]bson.Raw, error) {
	pipeline, err := convertFromBsonMList(opts.Pipeline)
	if err != nil {
		return nil, fmt.Errorf("error converting AggregateOptions.Pipeline to protobuf: %w", err)
	}

	req := &proto.AggregateRequest{
		Collection: opts.Collection,
		Pipeline:   pipeline,
	}
	resp, err := m.client.Aggregate(ctx, req)
	if err != nil {
		return nil, err
	}

	results := convertToBsonRawList(resp.Results)
	return results, nil
}

func (m *GClient) Count(ctx context.Context, opts plugins.CountOptions) (int64, error) {
	filter, err := convertFromBsonM(opts.Filter)
	if err != nil {
		return 0, fmt.Errorf("error converting CountOptions.Filter to a protobuf: %w", err)
	}

	resp, err := m.client.Count(ctx, &proto.CountRequest{
		Collection: opts.Collection,
		Filter:     filter,
	})
	if err != nil {
		return 0, err
	}
	return resp.Count, nil
}

func (m *GClient) Find(ctx context.Context, opts plugins.FindOptions) ([]bson.Raw, error) {
	sort, err := convertFromBsonD(opts.Sort)
	if err != nil {
		return nil, fmt.Errorf("error converting FindOptions.Sort to protobuf: %w", err)
	}

	sel, err := convertFromBsonD(opts.Select)
	if err != nil {
		return nil, fmt.Errorf("error converting FindOptions.Select to protobuf: %w", err)
	}

	filter, err := convertFromBsonM(opts.Filter)
	if err != nil {
		return nil, fmt.Errorf("error converting FindOptions.Filter to protobuf: %w", err)
	}

	group, err := convertFromBsonD(opts.Group)
	if err != nil {
		return nil, fmt.Errorf("error converting FindOptions.Group to protobuf: %w", err)
	}

	req := &proto.FindRequest{
		Collection: opts.Collection,
		Sort:       sort,
		Skip:       opts.Skip,
		Limit:      opts.Limit,
		Select:     sel,
		Filter:     filter,
		Group:      group,
	}
	resp, err := m.client.Find(ctx, req)
	if err != nil {
		return nil, err
	}

	results := convertToBsonRawList(resp.Results)
	return results, nil
}

func (m *GClient) Insert(ctx context.Context, opts plugins.InsertOptions) error {
	docs, err := convertFromBsonMList(opts.Documents)
	if err != nil {
		return fmt.Errorf("error converting InsertOptions.Documents to protobuf: %w", err)
	}

	req := &proto.InsertRequest{
		Collection: opts.Collection,
		Documents:  docs,
	}
	_, err = m.client.Insert(ctx, req)
	return err
}

func (m *GClient) Patch(ctx context.Context, opts plugins.PatchOptions) error {
	query, err := convertFromBsonM(opts.QueryDocument)
	if err != nil {
		return fmt.Errorf("error converting PatchOptions.Query to protobuf: %w", err)
	}

	transformation, err := convertFromBsonD(opts.Transformation)
	if err != nil {
		return fmt.Errorf("error converting PatchOptions.Transformation to protobuf: %w", err)
	}

	req := &proto.PatchRequest{
		Collection:     opts.Collection,
		QueryDocument:  query,
		Transformation: transformation,
	}
	_, err = m.client.Patch(ctx, req)
	return err
}

func (m *GClient) Remove(ctx context.Context, opts plugins.RemoveOptions) error {
	filter, err := convertFromBsonM(opts.Filter)
	if err != nil {
		return fmt.Errorf("error converting RemoveOptions.Filter to protobuf: %w", err)
	}

	req := &proto.RemoveRequest{
		Collection: opts.Collection,
		Filter:     filter,
		All:        opts.All,
	}
	_, err = m.client.Remove(ctx, req)
	return err
}

func (m *GClient) Update(ctx context.Context, opts plugins.UpdateOptions) error {
	filter, err := convertFromBsonM(opts.Filter)
	if err != nil {
		return fmt.Errorf("error converting UpdateOptions.Filter to protobuf: %w", err)
	}

	doc, err := convertFromBsonM(opts.Document)
	if err != nil {
		return fmt.Errorf("error converting UpdateOptions.Document to protobuf: %w", err)
	}

	req := &proto.UpdateRequest{
		Collection: opts.Collection,
		Filter:     filter,
		Upsert:     opts.Upsert,
		Document:   doc,
	}
	_, err = m.client.Update(ctx, req)
	return err
}

type GServer struct {
	Impl plugins.StorageProtocol
	proto.UnsafeStorageProtocolServer
}

func (m *GServer) EnsureIndex(ctx context.Context, request *proto.EnsureIndexRequest) (*proto.EnsureIndexResponse, error) {
	opts := plugins.EnsureIndexOptions{
		Indices: make([]plugins.Index, len(request.Indices)),
	}

	for i := range request.Indices {
		opts.Indices[i] = plugins.Index{
			Collection: request.Indices[i].Collection,
			Keys:       convertToBsonD(request.Indices[i].Keys),
			Unique:     false,
		}
	}

	err := m.Impl.EnsureIndex(opts)
	return &proto.EnsureIndexResponse{}, err
}

func (m *GServer) Aggregate(ctx context.Context, request *proto.AggregateRequest) (*proto.AggregateResponse, error) {
	opts := plugins.AggregateOptions{
		Collection: request.Collection,
		Pipeline:   convertToBsonMList(request.Pipeline),
	}

	results, err := m.Impl.Aggregate(opts)
	resp := &proto.AggregateResponse{Results: make([][]byte, len(results))}
	for i := range results {
		resp.Results[i] = results[i]
	}
	return resp, err
}

func (m *GServer) Count(ctx context.Context, req *proto.CountRequest) (*proto.CountResponse, error) {
	opts := plugins.CountOptions{
		Collection: req.Collection,
		Filter:     convertToBsonM(req.Filter),
	}
	count, err := m.Impl.Count(opts)
	return &proto.CountResponse{Count: count}, err
}

func (m *GServer) Find(ctx context.Context, request *proto.FindRequest) (*proto.FindResponse, error) {
	opts := plugins.FindOptions{
		Collection: request.Collection,
		Sort:       convertToBsonD(request.Sort),
		Skip:       request.Skip,
		Limit:      request.Limit,
		Select:     convertToBsonD(request.Select),
		Filter:     convertToBsonM(request.Filter),
		Group:      convertToBsonD(request.Group),
	}

	results, err := m.Impl.Find(opts)
	resp := &proto.FindResponse{Results: make([][]byte, len(results))}
	for i := range results {
		resp.Results[i] = results[i]
	}
	return resp, err
}

func (m *GServer) Insert(ctx context.Context, request *proto.InsertRequest) (*proto.InsertResponse, error) {
	opts := plugins.InsertOptions{
		Collection: request.Collection,
		Documents:  convertToBsonMList(request.Documents),
	}

	err := m.Impl.Insert(opts)
	return &proto.InsertResponse{}, err
}

func (m *GServer) Patch(ctx context.Context, request *proto.PatchRequest) (*proto.PatchResponse, error) {
	opts := plugins.PatchOptions{
		Collection:     request.Collection,
		QueryDocument:  convertToBsonM(request.QueryDocument),
		Transformation: convertToBsonD(request.Transformation),
	}

	err := m.Impl.Patch(opts)
	return &proto.PatchResponse{}, err
}

func (m *GServer) Remove(ctx context.Context, request *proto.RemoveRequest) (*proto.RemoveResponse, error) {
	opts := plugins.RemoveOptions{
		Collection: request.Collection,
		Filter:     convertToBsonM(request.Filter),
		All:        request.All,
	}

	err := m.Impl.Remove(opts)
	return &proto.RemoveResponse{}, err
}

func (m *GServer) Update(ctx context.Context, request *proto.UpdateRequest) (*proto.UpdateResponse, error) {
	opts := plugins.UpdateOptions{
		Collection: request.Collection,
		Filter:     convertToBsonM(request.Filter),
		Upsert:     request.Upsert,
		Document:   convertToBsonM(request.Document),
	}

	err := m.Impl.Update(opts)
	return &proto.UpdateResponse{}, err
}

func convertToBsonD(src []*proto.KeyValue) bson.D {
	var dest bson.D
	for i := range src {
		dest[i] = bson.E{
			Key:   src[i].Key,
			Value: src[i].Value.AsInterface(),
		}
	}
	return dest
}

func convertToBsonM(src *structpb.Struct) bson.M {
	return src.AsMap()
}

func convertToBsonMList(src []*structpb.Struct) []bson.M {
	dest := make([]bson.M, len(src))
	for i := range src {
		dest[i] = convertToBsonM(src[i])
	}
	return dest
}

func convertFromBsonD(src bson.D) ([]*proto.KeyValue, error) {
	dest := make([]*proto.KeyValue, len(src))
	for i := range src {
		value, err := structpb.NewValue(src[i].Value)
		if err != nil {
			return nil, fmt.Errorf("error converting value for key %s to a protobuf Struct: %w", src[i].Key, err)
		}
		dest[i] = &proto.KeyValue{
			Key:   src[i].Key,
			Value: value,
		}
	}
	return dest, nil
}

func convertFromBsonM(src bson.M) (*structpb.Struct, error) {
	dest, err := structpb.NewStruct(src)
	if err != nil {
		return nil, fmt.Errorf("error converting from bson.M to protobuf: %w", err)
	}
	return dest, nil
}

func convertFromBsonMList(src []bson.M) ([]*structpb.Struct, error) {
	dest := make([]*structpb.Struct, len(src))
	for i := range src {
		item, err := convertFromBsonM(src[i])
		if err != nil {
			return nil, fmt.Errorf("error converting bson.D[%d] to protobuf: %w", i, err)
		}
		dest[i] = item
	}

	return dest, nil
}

func convertToBsonRawList(src [][]byte) []bson.Raw {
	dest := make([]bson.Raw, len(src))
	for i := range src {
		dest[i] = src[i]
	}

	return dest
}
