package limiter

import (
	"time"

	"github.com/flaviojohansson/goexpert-rate-limiter/internal/storage"
)

type Limiter struct {
	storage storage.StorageInterface
}

func NewLimiter(storage storage.StorageInterface) *Limiter {
	return &Limiter{
		storage: storage,
	}
}

// Check verifica se a chave (IP/token) pode fazer uma nova requisição
func (l *Limiter) Check(key string, limit int, window, blockDuration time.Duration) (bool, error) {

	// Verifica se está bloqueado
	blocked, err := l.storage.IsBlocked(key)
	if err != nil {
		return false, err
	}
	if blocked {
		return false, nil // Rejeita a requisição
	}

	count, err := l.storage.Increment(key, window)

	if err != nil {
		return false, err
	}

	// Se exceder o limite, bloqueia a chave
	if count > int64(limit) {
		err = l.storage.Block(key, blockDuration)
		return false, err
	}

	return count <= int64(limit), nil
}
