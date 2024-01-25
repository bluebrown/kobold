//go:build generate

package build

//go:generate sqlc generate
//go:generate swag init --dir ../http/api/ --output ../http/api/docs/ --generalInfo handler.go --parseDependency github.com/bluebrown/kobold/store/model
