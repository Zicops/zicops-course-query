package global

import (
	"context"
	"math/rand"
	"sync"

	"github.com/scylladb/gocqlx/v2"
	cry "github.com/zicops/zicops-course-query/lib/crypto"
)

// some global variables commonly used
var (
	CTX             context.Context
	CassSession     *gocqlx.Session
	CryptSession    *cry.Cryptography
	Cancel          context.CancelFunc
	WaitGroupServer sync.WaitGroup
	Rand            *rand.Rand
)

// initializes global package to read environment variables as needed
func init() {
}
