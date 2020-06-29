//go:generate go run github.com/golang/mock/mockgen -destination models.go -package mock github.com/rucciva/giteaty/pkg/gitea Models
package mock

import "github.com/rucciva/giteaty/pkg/gitea"

var (
	_ gitea.Models = (*MockModels)(nil)
)
