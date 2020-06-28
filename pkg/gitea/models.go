package gitea

import "code.gitea.io/gitea/models"

type Models interface {
	UserSignIn(username, password string) (*models.User, error)

	SearchUsers(opts *models.SearchUserOptions) (users []*models.User, count int64, err error)
	GetUserTeams(userID int64, listOptions models.ListOptions) ([]*models.Team, error)
}
