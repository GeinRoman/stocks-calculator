package app

import "context"

func (s *app) PersonalAssetProfileContext(ctx context.Context, user int) (context.Context, error) {
	return s.repo.PersonalAssetProfileContext(ctx, user)
}
