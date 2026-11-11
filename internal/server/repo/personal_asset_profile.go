package repo

import "context"

type personalAssetProfileKey struct{}

type personalAssetProfile struct {
	user, profile int
}

func (r *repo) PersonalAssetProfileContext(ctx context.Context, user int) (context.Context, error) {
	profile, err := r.getDefaultProfileId(ctx, user)
	if err != nil {
		return nil, err
	}
	return context.WithValue(ctx, personalAssetProfileKey{}, personalAssetProfile{user, profile}), nil
}

func (r *repo) personalAssetProfile(ctx context.Context, user int) (int, error) {
	if p, ok := ctx.Value(personalAssetProfileKey{}).(personalAssetProfile); ok && p.user == user {
		return p.profile, nil
	}
	return r.getDefaultProfileId(ctx, user)
}
