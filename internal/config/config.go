package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Port string
	Env  string

	DBURL string

	FirebaseCredentials string

	KakaoRestAPIKey string

	MapboxAccessToken string

	R2AccountID       string
	R2AccessKeyID     string
	R2SecretAccessKey string
	R2BucketName      string
	R2PublicURL       string

	CORSOrigins string
}

func Load() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	// Defaults
	viper.SetDefault("PORT", "8080")
	viper.SetDefault("ENV", "development")

	_ = viper.ReadInConfig() // .env 없어도 env vars로 동작

	cfg := &Config{
		Port: viper.GetString("PORT"),
		Env:  viper.GetString("ENV"),

		DBURL: viper.GetString("DB_URL"),

		FirebaseCredentials: viper.GetString("FIREBASE_CREDENTIALS"),

		KakaoRestAPIKey: viper.GetString("KAKAO_REST_API_KEY"),

		MapboxAccessToken: viper.GetString("MAPBOX_ACCESS_TOKEN"),

		R2AccountID:       viper.GetString("R2_ACCOUNT_ID"),
		R2AccessKeyID:     viper.GetString("R2_ACCESS_KEY_ID"),
		R2SecretAccessKey: viper.GetString("R2_SECRET_ACCESS_KEY"),
		R2BucketName:      viper.GetString("R2_BUCKET_NAME"),
		R2PublicURL:       viper.GetString("R2_PUBLIC_URL"),

		CORSOrigins: viper.GetString("CORS_ORIGINS"),
	}

	return cfg, nil
}
