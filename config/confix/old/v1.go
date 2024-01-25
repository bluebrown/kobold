package old

// NOTE: The v1 user-config has the same layout as the normalized config itself
// the below is done to establish a pattern that can be used for other versions
// and formats.

type UserConfigV1 NormalizedConfig

func (c UserConfigV1) Normalize() *NormalizedConfig {
	n := NormalizedConfig(c)
	return &n
}
