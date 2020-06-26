package gitea

import "code.gitea.io/gitea/models"

type Models interface {
	UserSignIn(username, password string) (*models.User, error)

	SearchUsers(opts *models.SearchUserOptions) (users []*models.User, count int64, err error)
	GetOrgsByUserID(userID int64, showAll bool) ([]*models.User, error)
	GetUserTeams(userID int64, listOptions models.ListOptions) ([]*models.Team, error)
}
