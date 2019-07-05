package cache

import "time"

type Builder struct {
	expireAfterAccessDuration time.Duration
	expireAfterWriteDuration  time.Duration
}

func NewBuilder() *Builder {
	return &Builder{
		expireAfterAccessDuration: 0,
		expireAfterWriteDuration:  0,
	}
}

func (builder *Builder) ExpireAfterAccess(duration time.Duration) *Builder {
	builder.expireAfterAccessDuration = duration
	return builder
}

func (builder *Builder) ExpireAfterWrite(duration time.Duration) *Builder {
	builder.expireAfterWriteDuration = duration
	return builder
}

func (builder *Builder) Build(loader Loader) *localCache {
	return &localCache{
		expireAfterAccessDuration: builder.expireAfterAccessDuration,
		expireAfterWriteDuration:  builder.expireAfterWriteDuration,
		loader:                    loader,
	}
}
