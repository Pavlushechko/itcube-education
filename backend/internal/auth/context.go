package auth

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type ctxKey string

const (
	ctxUserID ctxKey = "user_id"
	ctxRole   ctxKey = "role"
)

// MVP: читаем из заголовков. Потом заменишь на JWT/интроспекцию ядра.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uidStr := r.Header.Get("X-User-Id")
		role := r.Header.Get("X-Role") // user|moderator|admin

		if uidStr == "" {
			// публичные ручки разрешим без user_id (каталог)
			next.ServeHTTP(w, r)
			return
		}
		uid, err := uuid.Parse(uidStr)
		if err != nil {
			http.Error(w, "invalid X-User-Id", http.StatusBadRequest)
			return
		}
		if role == "" {
			role = "user"
		}

		ctx := context.WithValue(r.Context(), ctxUserID, uid)
		ctx = context.WithValue(ctx, ctxRole, role)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func UserID(ctx context.Context) (uuid.UUID, bool) {
	v := ctx.Value(ctxUserID)
	id, ok := v.(uuid.UUID)
	return id, ok
}

func Role(ctx context.Context) string {
	v := ctx.Value(ctxRole)
	if s, ok := v.(string); ok && s != "" {
		return s
	}
	return "anonymous"
}
