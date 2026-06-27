package bench

import (
	"os"
	"testing"

	caarlos0env "github.com/caarlos0/env/v11"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/muzzii255/goenv"
)

func TestMain(m *testing.M) {
	setBenchEnv()
	os.Exit(m.Run())
}

func BenchmarkSmallStdlib(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = loadSmallStdlib()
	}
}

func BenchmarkSmallCarl(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var cfg CarlSmallConfig
		_ = caarlos0env.Parse(&cfg)
	}
}

func BenchmarkSmallCleanenv(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var cfg CleanSmallConfig
		_ = cleanenv.ReadEnv(&cfg)
	}
}

func BenchmarkSmallGoenv(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		cfg := &GoenvSmallConfig{}
		_ = goenv.LoadEnv(cfg, false)
	}
}

func BenchmarkSmallEnvconfig(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var cfg EnvconfigSmallConfig
		_ = envconfig.Process("", &cfg)
	}
}

func BenchmarkSmallViper(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = loadSmallViper()
	}
}

func BenchmarkSmallEnvgen(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = LoadSmallConfig()
	}
}

func BenchmarkMediumStdlib(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = loadMediumStdlib()
	}
}

func BenchmarkMediumCarl(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var cfg MediumConfig
		_ = caarlos0env.Parse(&cfg)
	}
}

func BenchmarkMediumCleanenv(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var cfg CleanMediumConfig
		_ = cleanenv.ReadEnv(&cfg)
	}
}

func BenchmarkMediumGoenv(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		cfg := &GoenvMediumConfig{}
		_ = goenv.LoadEnv(cfg, false)
	}
}

func BenchmarkMediumEnvconfig(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var cfg EnvconfigMediumConfig
		_ = envconfig.Process("", &cfg)
	}
}

func BenchmarkMediumViper(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = loadMediumViper()
	}
}

func BenchmarkMediumEnvgen(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = LoadMediumConfig()
	}
}

func BenchmarkLargeStdlib(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = loadLargeStdlib()
	}
}

func BenchmarkLargeCarl(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var cfg LargeConfig
		_ = caarlos0env.Parse(&cfg)
	}
}

func BenchmarkLargeCleanenv(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var cfg CleanLargeConfig
		_ = cleanenv.ReadEnv(&cfg)
	}
}

func BenchmarkLargeGoenv(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		cfg := &GoenvLargeConfig{}
		_ = goenv.LoadEnv(cfg, false)
	}
}

func BenchmarkLargeEnvconfig(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var cfg EnvconfigLargeConfig
		_ = envconfig.Process("", &cfg)
	}
}

func BenchmarkLargeViper(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = loadLargeViper()
	}
}

func BenchmarkLargeEnvgen(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = LoadLargeConfig()
	}
}
