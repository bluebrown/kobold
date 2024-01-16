//go:build generate

package build

//go:generate go run github.com/sqlc-dev/sqlc/cmd/sqlc generate
//go:generate go run github.com/swaggo/swag/cmd/swag init --dir ../api/ --output ../api/docs/ --generalInfo handler.go --parseDependency github.com/bluebrown/kobold/store
