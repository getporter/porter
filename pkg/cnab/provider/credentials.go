package cnabprovider

import (
	"context"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/secrets"
	"get.porter.sh/porter/pkg/storage"
	"get.porter.sh/porter/pkg/tracing"
	"go.mongodb.org/mongo-driver/bson"
)

func (r *Runtime) loadCredentials(ctx context.Context, b cnab.ExtendedBundle, args ActionArguments) (secrets.Set, error) {
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	if len(args.Installation.CredentialSets) == 0 {
		return nil, storage.Validate(nil, b.Credentials, args.Action)
	}

	// The strategy here is "last one wins". We loop through each credential file and
	// calculate its credentials. Then we insert them into the creds map in the order
	// in which they were supplied on the CLI.
	resolvedCredentials := secrets.Set{}
	for _, name := range args.Installation.CredentialSets {
		var cset storage.CredentialSet
		// Try to get the creds in the local namespace first, fallback to the global creds
		query := storage.FindOptions{
			Sort: []string{"-namespace"},
			Filter: bson.M{
				"name": name,
				"$or": []bson.M{
					{"namespace": ""},
					{"namespace": args.Installation.Namespace},
				},
			},
		}
		store := r.credentials.GetDataStore()
		err := store.FindOne(ctx, storage.CollectionCredentials, query, &cset)
		if err != nil {
			return nil, err
		}

		rc, err := r.credentials.ResolveAll(ctx, cset)
		if err != nil {
			return nil, err
		}

		for k, v := range rc {
			resolvedCredentials[k] = v
		}
	}

	return resolvedCredentials, storage.Validate(resolvedCredentials, b.Credentials, args.Action)
}
