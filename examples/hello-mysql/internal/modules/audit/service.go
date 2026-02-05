package audit

import (
	"context"
	"fmt"

	"github.com/go-modkit/modkit/examples/hello-mysql/internal/modules/users"
)

type Service interface {
	AuditUserLookup(ctx context.Context, id int64) (string, error)
}

type service struct {
	users users.Service
}

func NewService(usersSvc users.Service) Service {
	return &service{users: usersSvc}
}

func (s *service) AuditUserLookup(ctx context.Context, id int64) (string, error) {
	user, err := s.users.GetUser(ctx, id)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("lookup user id=%d email=%s", user.ID, user.Email), nil
}
