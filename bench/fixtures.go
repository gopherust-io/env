package bench

//go:generate go run generate_competitors.go .
//go:generate go run ../cmd/envgen -type SmallConfig -output small_env_gen.go
//go:generate go run ../cmd/envgen -type MediumConfig -output medium_env_gen.go
//go:generate go run ../cmd/envgen -type LargeConfig -output large_env_gen.go
