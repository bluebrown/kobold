//go:build generate

package build

//go:generate go run github.com/sqlc-dev/sqlc/cmd/sqlc generate
//go:generate go run github.com/swaggo/swag/cmd/swag init --dir ../http/api/ --output ../http/api/docs/ --generalInfo handler.go --parseDependency github.com/bluebrown/kobold/store/model
