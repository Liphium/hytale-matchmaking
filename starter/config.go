package starter

import (
	"fmt"

	"github.com/Liphium/magic/v2"
	"github.com/Liphium/magic/v2/mconfig"
)

func BuildMagicConfig() magic.Config {
	return magic.Config{
		AppName: "hytale-matchmaking",
		PlanDeployment: func(ctx *mconfig.Context) {

			port := ctx.ValuePort(3000)
			listen := mconfig.ValueWithBase([]mconfig.EnvironmentValue{port}, func(s []string) string {
				return fmt.Sprintf("127.0.0.1:%s", s[0])
			})

			tokenFileLoc := mconfig.ValueStatic(".")
			if ctx.Profile() == "test" {
				tokenFileLoc = mconfig.ValueStatic("test-runner")
			}

			ctx.WithEnvironment(mconfig.Environment{
				"LISTEN":              listen,
				"TOKEN_FILE_LOCATION": tokenFileLoc,
				"CREDENTIAL":          mconfig.ValueStatic("test"),
			})

			ctx.LoadSecretsToEnvironment(".env")
		},
		StartFunction: Start,
	}
}
